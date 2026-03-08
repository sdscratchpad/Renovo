package injector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ravi-poc/contracts"
	"github.com/ravi-poc/fault-injector/internal/scenarios"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type InjectionResult struct {
	Scenario   string    `json:"scenario"`
	InjectedAt time.Time `json:"injected_at"`
}

type ScenarioInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Injector struct {
	k8s           *kubernetes.Clientset
	eventStoreURL string
}

func New(eventStoreURL string) (*Injector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("injector: build kubeconfig: %w", err)
		}
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("injector: build k8s client: %w", err)
	}
	return &Injector{k8s: client, eventStoreURL: eventStoreURL}, nil
}

func (inj *Injector) Scenarios() []ScenarioInfo {
	return []ScenarioInfo{
		{Name: "bad-rollout", Description: "Patch sample-app with missing-secret env var, causes CrashLoopBackOff"},
		{Name: "resource-saturation", Description: "Patch sample-app limits to near-zero, causes CPU throttling"},
		{Name: "batch-timeout", Description: "Set FAIL_MODE=timeout on batch-worker CronJob"},
	}
}

func (inj *Injector) Inject(ctx context.Context, scenario string) (*InjectionResult, error) {
	var err error
	switch scenario {
	case "bad-rollout":
		err = scenarios.BadRollout(ctx, inj.k8s)
	case "resource-saturation":
		err = scenarios.ResourceSaturation(ctx, inj.k8s)
	case "batch-timeout":
		err = scenarios.BatchTimeout(ctx, inj.k8s)
	default:
		return nil, fmt.Errorf("unknown scenario %q", scenario)
	}
	if err != nil {
		return nil, err
	}
	result := &InjectionResult{Scenario: scenario, InjectedAt: time.Now().UTC()}
	inj.notifyEventStore(scenario)
	return result, nil
}

// Restore undoes the named scenario's fault injection.
func (inj *Injector) Restore(ctx context.Context, scenario string) error {
	switch scenario {
	case "bad-rollout":
		return scenarios.RestoreBadRollout(ctx)
	case "resource-saturation":
		return scenarios.RestoreResourceSaturation(ctx, inj.k8s)
	case "batch-timeout":
		return scenarios.RestoreBatchTimeout(ctx)
	default:
		return fmt.Errorf("unknown scenario %q", scenario)
	}
}

// RestoreAll restores all scenarios; logs but does not stop on individual errors.
func (inj *Injector) RestoreAll(ctx context.Context) {
	for _, s := range []string{"bad-rollout", "resource-saturation", "batch-timeout"} {
		if err := inj.Restore(ctx, s); err != nil {
			log.Printf("injector: restore %s: %v", s, err)
		}
	}
}

func (inj *Injector) notifyEventStore(scenario string) {
	// Map each scenario to the actual service it affects (not the fault-injector itself).
	serviceMap := map[string]string{
		"bad-rollout":         "sample-app",
		"resource-saturation": "sample-app",
		"batch-timeout":       "batch-worker",
	}
	severityMap := map[string]contracts.SeverityLevel{
		"bad-rollout":         contracts.SeverityCritical,
		"resource-saturation": contracts.SeverityHigh,
		"batch-timeout":       contracts.SeverityMedium,
	}
	affectedService, ok := serviceMap[scenario]
	if !ok {
		affectedService = scenario
	}
	severity, ok := severityMap[scenario]
	if !ok {
		severity = contracts.SeverityHigh
	}

	event := contracts.IncidentEvent{
		ID:         fmt.Sprintf("%s-%d", scenario, time.Now().UnixNano()),
		Scenario:   scenario,
		Service:    affectedService,
		Namespace:  "workloads",
		Severity:   severity,
		DetectedAt: time.Now().UTC(),
		Evidence: []contracts.EvidenceItem{{
			Source:    "fault-injector",
			Signal:    "scenario-triggered",
			Value:     scenario,
			Timestamp: time.Now().UTC(),
		}},
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("injector: marshal: %v", err)
		return
	}
	resp, err := http.Post(inj.eventStoreURL+"/incidents", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("injector: post: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("injector: event-store %d for %s", resp.StatusCode, scenario)
	}
}
