package scenarios

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const BatchWorkerCronJob = "batch-worker"

// BatchTimeout patches the batch-worker CronJob to set FAIL_MODE=timeout,
// causing future job runs to exceed activeDeadlineSeconds. Idempotent.
func BatchTimeout(ctx context.Context, client *kubernetes.Clientset) error {
	patch := map[string]any{
		"spec": map[string]any{
			"jobTemplate": map[string]any{
				"spec": map[string]any{
					"template": map[string]any{
						"spec": map[string]any{
							"containers": []map[string]any{{
								"name": BatchWorkerCronJob,
								"env": []map[string]any{
									{"name": "FAIL_MODE", "value": "timeout"},
									{"name": "PORT",      "value": "8081"},
								},
							}},
						},
					},
				},
			},
		},
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("batch-timeout: marshal: %w", err)
	}
	_, err = client.BatchV1().CronJobs(Namespace).Patch(
		ctx, BatchWorkerCronJob, types.StrategicMergePatchType, b, metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("batch-timeout: patch cronjob: %w", err)
	}
	return nil
}
