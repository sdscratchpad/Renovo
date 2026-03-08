package runbooks

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// RetryBatch implements the "retry-batch-job" runbook.
// It clears the FAIL_MODE env var from the batch-worker CronJob so the
// batch worker can resume processing after a simulated batch-timeout fault.
type RetryBatch struct {
	KubeClient kubernetes.Interface
}

// Name returns the canonical runbook identifier.
func (rb *RetryBatch) Name() string { return "retry-batch-job" }

// Execute patches the batch-worker CronJob to remove the FAIL_MODE env var,
// allowing the worker to resume normal operation.
// Required params: "namespace". Optional: "deployment" (name of the CronJob, defaults to "batch-worker").
func (rb *RetryBatch) Execute(ctx context.Context, params map[string]string) error {
	ns := params["namespace"]
	if ns == "" {
		return fmt.Errorf("retry-batch-job: param 'namespace' is required")
	}
	cronJobName := params["deployment"]
	if cronJobName == "" {
		cronJobName = "batch-worker"
	}

	// Fetch current CronJob to find its containers.
	cj, err := rb.KubeClient.BatchV1().CronJobs(ns).Get(ctx, cronJobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("retry-batch-job: get cronjob: %w", err)
	}

	// Build a filtered env list removing FAIL_MODE and BLOCK_DEPENDENCY.
	containers := cj.Spec.JobTemplate.Spec.Template.Spec.Containers
	for i, c := range containers {
		filtered := c.Env[:0]
		for _, e := range c.Env {
			if e.Name != "FAIL_MODE" && e.Name != "BLOCK_DEPENDENCY" {
				filtered = append(filtered, e)
			}
		}
		containers[i].Env = filtered
	}

	// Strategic merge patch to update just the container envs.
	type envEntry struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	type containerPatch struct {
		Name string     `json:"name"`
		Env  []envEntry `json:"env"`
	}
	type specPatch struct {
		Containers []containerPatch `json:"containers"`
	}
	type templatePatch struct {
		Spec specPatch `json:"spec"`
	}
	type jobTemplateSpec struct {
		Template templatePatch `json:"template"`
	}
	type jobTemplate struct {
		Spec jobTemplateSpec `json:"spec"`
	}
	type patchBody struct {
		Spec struct {
			JobTemplate jobTemplate `json:"jobTemplate"`
		} `json:"spec"`
	}

	var patch patchBody
	for _, c := range containers {
		cp := containerPatch{Name: c.Name}
		for _, e := range c.Env {
			cp.Env = append(cp.Env, envEntry{Name: e.Name, Value: e.Value})
		}
		patch.Spec.JobTemplate.Spec.Template.Spec.Containers = append(
			patch.Spec.JobTemplate.Spec.Template.Spec.Containers, cp)
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("retry-batch-job: marshal patch: %w", err)
	}

	_, err = rb.KubeClient.BatchV1().CronJobs(ns).Patch(
		ctx,
		cronJobName,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("retry-batch-job: patch cronjob: %w", err)
	}
	return nil
}
