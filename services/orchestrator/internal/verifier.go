package internal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	verifyPollInterval = 3 * time.Second
	verifyTimeout      = 120 * time.Second
)

// WaitUntilHealthy polls Kubernetes until the expected post-remediation state is
// confirmed or the timeout expires. It returns (true, "") on success or
// (false, reason) when the timeout is reached.
//
// Runbook mapping:
//   - scale-deployment  → waits for AvailableReplicas ≥ desired replicas param
//   - rollback-deployment → waits for AvailableReplicas ≥ 1
//   - restart-pod       → waits for all matching pods to be Ready
//   - retry-batch-job   → no K8s state to check; resolves immediately
func WaitUntilHealthy(kubeClient kubernetes.Interface, runbookName string, params map[string]string) (bool, string) {
	if kubeClient == nil {
		// Dry-run / kube unavailable: skip verification.
		return true, ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), verifyTimeout)
	defer cancel()

	switch runbookName {
	case "scale-deployment":
		want := desiredReplicas(params)
		return waitForDeploymentReplicas(ctx, kubeClient, params, want)
	case "rollback-deployment":
		return waitForDeploymentReplicas(ctx, kubeClient, params, 1)
	case "restart-pod":
		return waitForPodsReady(ctx, kubeClient, params)
	default:
		// retry-batch-job and unknown runbooks: nothing to verify.
		return true, ""
	}
}

// desiredReplicas parses the "replicas" param; defaults to 1 if absent or invalid.
func desiredReplicas(params map[string]string) int32 {
	v, err := strconv.ParseInt(params["replicas"], 10, 32)
	if err != nil || v < 1 {
		return 1
	}
	return int32(v)
}

// waitForDeploymentReplicas blocks until the deployment has at least `want`
// available replicas or the context deadline is exceeded.
func waitForDeploymentReplicas(ctx context.Context, kubeClient kubernetes.Interface, params map[string]string, want int32) (bool, string) {
	ns := params["namespace"]
	dep := params["deployment"]
	if ns == "" || dep == "" {
		// Cannot verify without params; assume success.
		return true, ""
	}

	ticker := time.NewTicker(verifyPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Sprintf("deployment %s/%s did not reach %d available replica(s) within %s",
				ns, dep, want, verifyTimeout)
		case <-ticker.C:
			d, err := kubeClient.AppsV1().Deployments(ns).Get(ctx, dep, metav1.GetOptions{})
			if err != nil {
				continue // transient API error; keep polling
			}
			if d.Status.AvailableReplicas >= want {
				return true, ""
			}
		}
	}
}

// waitForPodsReady blocks until all pods matching "app=<name>" in <namespace>
// are in the Ready condition, or the context deadline is exceeded.
func waitForPodsReady(ctx context.Context, kubeClient kubernetes.Interface, params map[string]string) (bool, string) {
	ns := params["namespace"]
	name := params["name"]
	if ns == "" || name == "" {
		return true, ""
	}

	ticker := time.NewTicker(verifyPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Sprintf("pods for %s/%s did not become ready within %s", ns, name, verifyTimeout)
		case <-ticker.C:
			pods, err := kubeClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", name),
			})
			if err != nil || len(pods.Items) == 0 {
				continue
			}
			if allPodsReady(pods.Items) {
				return true, ""
			}
		}
	}
}

func allPodsReady(pods []corev1.Pod) bool {
	for _, pod := range pods {
		ready := false
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
		if !ready {
			return false
		}
	}
	return len(pods) > 0
}
