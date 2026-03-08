// Package detector implements autonomous fault detection for sample-app.
//
// It monitors three layers following the SRE golden signals model:
//
//  1. Application signals (via Prometheus): error rate, latency p99, traffic.
//  2. Availability (direct HTTP health check): service reachable / returning 2xx.
//  3. Infrastructure (Kubernetes API): pod container states, restart loops,
//     unschedulable pods, zero-replica deployments.
//
// When a threshold is breached, the detector creates an IncidentEvent in the
// event-store. The existing processUnanalyzed watcher picks it up and triggers
// AI diagnosis automatically — closing the full detect → diagnose → remediate loop
// without any manual fault injection.
package detector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ravi-poc/contracts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ---- Tuning constants ----

const (
	watchedNamespace  = "workloads"
	watchedDeployment = "sample-app"
	pollInterval      = 5 * time.Second
	cooldownPeriod    = 5 * time.Second // suppress re-fire for the same scenario within this window

	// Golden signal thresholds.
	errorRatioThreshold = 0.10  // 10 % of requests returning errors triggers an incident
	latencyP99Threshold = 0.500 // 500 ms p99 latency triggers an incident
	podRestartThreshold = int32(3)
)

// Detector monitors sample-app using SRE golden signals and Kubernetes
// infrastructure signals and fires IncidentEvents into the event-store.
type Detector struct {
	prometheusURL string
	eventStoreURL string
	sampleAppURL  string // direct HTTP base URL, e.g. http://localhost:8080
	k8s           *kubernetes.Clientset
	httpClient    *http.Client

	mu        sync.Mutex
	cooldowns map[string]time.Time // scenario → last fired timestamp
}

// New constructs a Detector. Kubernetes connectivity failures are non-fatal;
// the detector continues with Prometheus + direct-HTTP signals only.
func New(prometheusURL, eventStoreURL, sampleAppURL string) *Detector {
	d := &Detector{
		prometheusURL: prometheusURL,
		eventStoreURL: eventStoreURL,
		sampleAppURL:  sampleAppURL,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		cooldowns:     make(map[string]time.Time),
	}
	k8sClient, err := buildK8sClient()
	if err != nil {
		log.Printf("detector: k8s client unavailable (%v) — infrastructure signals disabled", err)
	} else {
		d.k8s = k8sClient
		log.Println("detector: k8s client connected — infrastructure signals enabled")
	}
	return d
}

func buildK8sClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("kubeconfig: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}

// Start begins the polling detection loop. It blocks until ctx is cancelled.
func (d *Detector) Start(ctx context.Context) {
	log.Printf("detector: SRE golden signal watcher started (interval=%s cooldown=%s)", pollInterval, cooldownPeriod)
	d.runChecks(ctx) // immediate check on startup catches pre-existing issues
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.runChecks(ctx)
		case <-ctx.Done():
			log.Println("detector: shutting down")
			return
		}
	}
}

// runChecks runs all signal checks in sequence.
func (d *Detector) runChecks(ctx context.Context) {
	// ---- Golden Signal #3: Errors (Prometheus) ----
	d.checkErrorRate(ctx)
	// ---- Golden Signal #1: Latency (Prometheus) ----
	d.checkLatency(ctx)
	// ---- Golden Signal #2: Traffic / Availability (direct HTTP) ----
	d.checkServiceAvailability(ctx)
	// ---- Golden Signal #4: Saturation + Infrastructure (Kubernetes API) ----
	if d.k8s != nil {
		d.checkK8sPodHealth(ctx)
	}
}

// ---- Signal checks ----

// checkErrorRate fires an "error-rate-spike" incident when the ratio of errors
// to total requests over the last 2 minutes exceeds errorRatioThreshold.
func (d *Detector) checkErrorRate(ctx context.Context) {
	ratio, err := d.queryPrometheusFloat(ctx,
		`sum(rate(sample_app_errors_total[2m])) / sum(rate(sample_app_requests_total[2m]))`)
	if err != nil || ratio <= 0 {
		return
	}
	if ratio < errorRatioThreshold {
		return
	}
	absRate, _ := d.queryPrometheusFloat(ctx, `sum(rate(sample_app_errors_total[2m]))`)
	reqRate, _ := d.queryPrometheusFloat(ctx, `sum(rate(sample_app_requests_total[2m]))`)
	d.maybeFireIncident(ctx, contracts.IncidentEvent{
		Scenario:  "error-rate-spike",
		Service:   "sample-app",
		Namespace: watchedNamespace,
		Severity:  severityFromErrorRatio(ratio),
		Evidence: []contracts.EvidenceItem{
			{Source: "metrics", Signal: "error_ratio_2m", Value: fmt.Sprintf("%.4f (%.1f%%)", ratio, ratio*100), Timestamp: time.Now().UTC()},
			{Source: "metrics", Signal: "error_rate_rps", Value: fmt.Sprintf("%.4f req/s", absRate), Timestamp: time.Now().UTC()},
			{Source: "metrics", Signal: "request_rate_rps", Value: fmt.Sprintf("%.4f req/s", reqRate), Timestamp: time.Now().UTC()},
			{Source: "detector", Signal: "threshold", Value: fmt.Sprintf("%.0f%% error ratio", errorRatioThreshold*100), Timestamp: time.Now().UTC()},
		},
	})
}

// checkLatency fires a "high-latency" incident when p99 request duration
// exceeds latencyP99Threshold over the last 2 minutes.
func (d *Detector) checkLatency(ctx context.Context) {
	p99, err := d.queryPrometheusFloat(ctx,
		`histogram_quantile(0.99, rate(sample_app_request_duration_seconds_bucket[2m]))`)
	if err != nil || p99 <= 0 {
		return
	}
	if p99 < latencyP99Threshold {
		return
	}
	p50, _ := d.queryPrometheusFloat(ctx,
		`histogram_quantile(0.50, rate(sample_app_request_duration_seconds_bucket[2m]))`)
	reqRate, _ := d.queryPrometheusFloat(ctx, `sum(rate(sample_app_requests_total[2m]))`)
	d.maybeFireIncident(ctx, contracts.IncidentEvent{
		Scenario:  "high-latency",
		Service:   "sample-app",
		Namespace: watchedNamespace,
		Severity:  severityFromLatency(p99),
		Evidence: []contracts.EvidenceItem{
			{Source: "metrics", Signal: "latency_p99_seconds", Value: fmt.Sprintf("%.4fs", p99), Timestamp: time.Now().UTC()},
			{Source: "metrics", Signal: "latency_p50_seconds", Value: fmt.Sprintf("%.4fs", p50), Timestamp: time.Now().UTC()},
			{Source: "metrics", Signal: "request_rate_rps", Value: fmt.Sprintf("%.4f req/s", reqRate), Timestamp: time.Now().UTC()},
			{Source: "detector", Signal: "threshold_seconds", Value: fmt.Sprintf("%.3fs p99", latencyP99Threshold), Timestamp: time.Now().UTC()},
		},
	})
}

// checkServiceAvailability makes a direct HTTP call to sample-app's /health
// endpoint. This is the most reliable availability signal in this setup because
// Prometheus runs inside Docker and cannot always scrape host-based services.
func (d *Detector) checkServiceAvailability(ctx context.Context) {
	healthURL := d.sampleAppURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		// Connection refused / timeout means the service process is not running.
		d.maybeFireIncident(ctx, contracts.IncidentEvent{
			Scenario:  "service-unavailable",
			Service:   "sample-app",
			Namespace: watchedNamespace,
			Severity:  contracts.SeverityCritical,
			Evidence: []contracts.EvidenceItem{
				{Source: "detector", Signal: "health_check_url", Value: healthURL, Timestamp: time.Now().UTC()},
				{Source: "detector", Signal: "health_check_error", Value: err.Error(), Timestamp: time.Now().UTC()},
				{Source: "detector", Signal: "description", Value: "sample-app /health endpoint unreachable — process may be crashed or container not running", Timestamp: time.Now().UTC()},
			},
		})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		d.maybeFireIncident(ctx, contracts.IncidentEvent{
			Scenario:  "service-degraded",
			Service:   "sample-app",
			Namespace: watchedNamespace,
			Severity:  contracts.SeverityHigh,
			Evidence: []contracts.EvidenceItem{
				{Source: "detector", Signal: "health_check_url", Value: healthURL, Timestamp: time.Now().UTC()},
				{Source: "detector", Signal: "health_check_status", Value: strconv.Itoa(resp.StatusCode), Timestamp: time.Now().UTC()},
				{Source: "detector", Signal: "health_check_body", Value: strings.TrimSpace(string(body)), Timestamp: time.Now().UTC()},
			},
		})
	}
}

// checkK8sPodHealth inspects Kubernetes pod and deployment state for sample-app
// and fires incidents for: CrashLoopBackOff, CreateContainerConfigError,
// OOMKilled, ImagePullBackOff, pod restart loops, unschedulable pods,
// and zero-available-replica deployments.
func (d *Detector) checkK8sPodHealth(ctx context.Context) {
	pods, err := d.k8s.CoreV1().Pods(watchedNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=" + watchedDeployment,
	})
	if err != nil {
		log.Printf("detector: k8s pods list: %v", err)
		return
	}

	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {

			// Container is stuck waiting — CrashLoopBackOff, config error, OOM, pull failure.
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				switch reason {
				case "CrashLoopBackOff", "CreateContainerConfigError",
					"OOMKilled", "ImagePullBackOff", "ErrImagePull":
					scenario := "bad-rollout"
					severity := contracts.SeverityCritical
					if reason == "OOMKilled" {
						scenario = "resource-saturation"
						severity = contracts.SeverityHigh
					}
					d.maybeFireIncident(ctx, contracts.IncidentEvent{
						Scenario:  scenario,
						Service:   "sample-app",
						Namespace: watchedNamespace,
						Severity:  severity,
						Evidence: []contracts.EvidenceItem{
							{Source: "k8s", Signal: "pod_name", Value: pod.Name, Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "container_name", Value: cs.Name, Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "container_state", Value: "Waiting", Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "wait_reason", Value: reason, Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "restart_count", Value: strconv.Itoa(int(cs.RestartCount)), Timestamp: time.Now().UTC()},
						},
					})
					return // one incident per check cycle to avoid flooding
				}
			}

			// Container has restarted too many times — something is flapping.
			if cs.RestartCount >= podRestartThreshold {
				lastReason := ""
				if cs.LastTerminationState.Terminated != nil {
					lastReason = cs.LastTerminationState.Terminated.Reason
				}
				d.maybeFireIncident(ctx, contracts.IncidentEvent{
					Scenario:  "pod-restart-loop",
					Service:   "sample-app",
					Namespace: watchedNamespace,
					Severity:  contracts.SeverityHigh,
					Evidence: []contracts.EvidenceItem{
						{Source: "k8s", Signal: "pod_name", Value: pod.Name, Timestamp: time.Now().UTC()},
						{Source: "k8s", Signal: "container_name", Value: cs.Name, Timestamp: time.Now().UTC()},
						{Source: "k8s", Signal: "restart_count", Value: strconv.Itoa(int(cs.RestartCount)), Timestamp: time.Now().UTC()},
						{Source: "k8s", Signal: "last_exit_reason", Value: lastReason, Timestamp: time.Now().UTC()},
					},
				})
				return
			}
		}

		// Pod is stuck Pending — node doesn't have capacity (resource saturation).
		if pod.Status.Phase == corev1.PodPending {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
					d.maybeFireIncident(ctx, contracts.IncidentEvent{
						Scenario:  "resource-saturation",
						Service:   "sample-app",
						Namespace: watchedNamespace,
						Severity:  contracts.SeverityHigh,
						Evidence: []contracts.EvidenceItem{
							{Source: "k8s", Signal: "pod_name", Value: pod.Name, Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "pod_phase", Value: "Pending", Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "schedule_reason", Value: cond.Reason, Timestamp: time.Now().UTC()},
							{Source: "k8s", Signal: "schedule_message", Value: cond.Message, Timestamp: time.Now().UTC()},
						},
					})
					return
				}
			}
		}
	}

	// Deployment-level: all replicas unavailable.
	dep, err := d.k8s.AppsV1().Deployments(watchedNamespace).Get(ctx, watchedDeployment, metav1.GetOptions{})
	if err != nil {
		log.Printf("detector: k8s deployment get: %v", err)
		return
	}
	desired := int32(0)
	if dep.Spec.Replicas != nil {
		desired = *dep.Spec.Replicas
	}
	if desired == 0 {
		// Deployment spec.replicas was explicitly set to 0 — intentional scale-to-zero.
		// rollback-deployment would NOT fix this because kubectl rollout undo only reverts
		// template changes, not the replica count. The correct fix is scale-deployment.
		d.maybeFireIncident(ctx, contracts.IncidentEvent{
			Scenario:  "scale-to-zero",
			Service:   "sample-app",
			Namespace: watchedNamespace,
			Severity:  contracts.SeverityCritical,
			Evidence: []contracts.EvidenceItem{
				{Source: "k8s", Signal: "desired_replicas", Value: "0", Timestamp: time.Now().UTC()},
				{Source: "k8s", Signal: "available_replicas", Value: strconv.Itoa(int(dep.Status.AvailableReplicas)), Timestamp: time.Now().UTC()},
				{Source: "k8s", Signal: "ready_replicas", Value: strconv.Itoa(int(dep.Status.ReadyReplicas)), Timestamp: time.Now().UTC()},
				{Source: "k8s", Signal: "description", Value: "deployment spec.replicas=0 — service intentionally scaled to zero, no pods will be scheduled; must use scale-deployment to restore", Timestamp: time.Now().UTC()},
			},
		})
	} else if dep.Status.AvailableReplicas == 0 {
		// Desired replicas > 0 but none are available — deployment is broken (crash, bad image, etc.).
		d.maybeFireIncident(ctx, contracts.IncidentEvent{
			Scenario:  "service-unavailable",
			Service:   "sample-app",
			Namespace: watchedNamespace,
			Severity:  contracts.SeverityCritical,
			Evidence: []contracts.EvidenceItem{
				{Source: "k8s", Signal: "available_replicas", Value: "0", Timestamp: time.Now().UTC()},
				{Source: "k8s", Signal: "desired_replicas", Value: strconv.Itoa(int(desired)), Timestamp: time.Now().UTC()},
				{Source: "k8s", Signal: "ready_replicas", Value: strconv.Itoa(int(dep.Status.ReadyReplicas)), Timestamp: time.Now().UTC()},
				{Source: "k8s", Signal: "description", Value: "deployment has 0 available replicas despite desired>0 — pods are crashing or failing to start", Timestamp: time.Now().UTC()},
			},
		})
	}
}

// ---- Incident deduplication and firing ----

// maybeFireIncident applies two layers of deduplication before posting to the
// event-store:
//  1. Durable check: queries the event-store first — if a non-resolved incident
//     already exists for the same scenario+service pair, suppress immediately
//     without touching the cooldown clock.
//  2. In-memory cooldown: once the active incident is gone, prevents re-firing
//     the same scenario faster than cooldownPeriod.
func (d *Detector) maybeFireIncident(ctx context.Context, inc contracts.IncidentEvent) {
	// Check durable state first so the cooldown clock is not consumed while an
	// incident is still open (awaiting_approval, analyzing, remediating, etc.).
	if d.hasActiveIncident(ctx, inc.Scenario, inc.Service) {
		return
	}

	d.mu.Lock()
	if last, ok := d.cooldowns[inc.Scenario]; ok && time.Since(last) < cooldownPeriod {
		d.mu.Unlock()
		return
	}
	d.cooldowns[inc.Scenario] = time.Now()
	d.mu.Unlock()

	now := time.Now().UTC()
	inc.ID = fmt.Sprintf("%s-%d", inc.Scenario, now.UnixNano())
	inc.DetectedAt = now

	data, err := json.Marshal(inc)
	if err != nil {
		log.Printf("detector: marshal: %v", err)
		return
	}
	resp, err := d.httpClient.Post(d.eventStoreURL+"/incidents", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("detector: post to event-store: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("detector: incident detected — id=%s scenario=%s service=%s severity=%s",
		inc.ID, inc.Scenario, inc.Service, inc.Severity)
}

// hasActiveIncident returns true when any incident with the same scenario+service
// is in a non-terminal pipeline state, OR when any incident for the same service
// is actively remediating or verifying. The second condition prevents the detector
// from firing a new incident (e.g. "service-unavailable") while a different but
// related incident (e.g. "scale-to-zero") is mid-remediation and pods are still
// starting up.
func (d *Detector) hasActiveIncident(ctx context.Context, scenario, service string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.eventStoreURL+"/incidents", nil)
	if err != nil {
		return false
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var incidents []contracts.IncidentEvent
	if err := json.NewDecoder(resp.Body).Decode(&incidents); err != nil {
		return false
	}
	for _, inc := range incidents {
		if inc.Service != service {
			continue
		}
		st := d.getIncidentStatus(ctx, inc.ID)
		// If any incident for this service is being actively remediated or
		// verified, suppress all new incidents regardless of scenario.
		if st == contracts.StatusRemediating || st == contracts.StatusVerifying {
			return true
		}
		// For matching scenarios, suppress in any non-terminal state.
		if inc.Scenario == scenario && st != contracts.StatusResolved && st != contracts.StatusFailed {
			return true
		}
	}
	return false
}

func (d *Detector) getIncidentStatus(ctx context.Context, id string) contracts.IncidentStatus {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.eventStoreURL+"/status/"+id, nil)
	if err != nil {
		return contracts.StatusDetected
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return contracts.StatusDetected
	}
	defer resp.Body.Close()
	var su contracts.IncidentStatusUpdate
	json.NewDecoder(resp.Body).Decode(&su)
	return su.Status
}

// ---- Prometheus query helper ----

// queryPrometheusFloat runs an instant PromQL query and returns the first scalar
// result as float64. Returns 0,nil when the result set is empty.
func (d *Detector) queryPrometheusFloat(ctx context.Context, query string) (float64, error) {
	params := url.Values{}
	params.Set("query", query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(d.prometheusURL, "/")+"/api/v1/query",
		strings.NewReader(params.Encode()))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("prometheus parse: %w", err)
	}
	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return 0, nil
	}
	s, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected prometheus value type")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, nil
	}
	return f, nil
}

// ---- Severity helpers ----

func severityFromErrorRatio(ratio float64) contracts.SeverityLevel {
	switch {
	case ratio >= 0.50:
		return contracts.SeverityCritical
	case ratio >= 0.20:
		return contracts.SeverityHigh
	default:
		return contracts.SeverityMedium
	}
}

func severityFromLatency(p99 float64) contracts.SeverityLevel {
	switch {
	case p99 >= 2.0:
		return contracts.SeverityCritical
	case p99 >= 1.0:
		return contracts.SeverityHigh
	default:
		return contracts.SeverityMedium
	}
}
