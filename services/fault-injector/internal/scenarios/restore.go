package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// RestoreBadRollout undoes the bad-rollout fault by rolling back the deployment.
func RestoreBadRollout(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "kubectl", "rollout", "undo",
		fmt.Sprintf("deployment/%s", SampleAppDeployment),
		"-n", Namespace,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restore-bad-rollout: %w: %s", err, string(out))
	}
	return nil
}

// RestoreResourceSaturation patches sample-app back to baseline resource limits.
func RestoreResourceSaturation(ctx context.Context, client *kubernetes.Clientset) error {
	patch := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []map[string]any{{
						"name": SampleAppDeployment,
						"resources": map[string]any{
							"limits":   map[string]any{"cpu": "500m", "memory": "128Mi"},
							"requests": map[string]any{"cpu": "50m", "memory": "64Mi"},
						},
					}},
				},
			},
		},
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("restore-resource-saturation: marshal: %w", err)
	}
	_, err = client.AppsV1().Deployments(Namespace).Patch(
		ctx, SampleAppDeployment, types.StrategicMergePatchType, b, metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("restore-resource-saturation: patch: %w", err)
	}
	return nil
}

// RestoreBatchTimeout removes the FAIL_MODE env variable from the batch-worker CronJob.
func RestoreBatchTimeout(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "kubectl", "set", "env",
		fmt.Sprintf("cronjob/%s", BatchWorkerCronJob),
		"-n", Namespace,
		"FAIL_MODE-",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restore-batch-timeout: %w: %s", err, string(out))
	}
	return nil
}
