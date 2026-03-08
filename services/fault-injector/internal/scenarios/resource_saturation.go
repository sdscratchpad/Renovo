package scenarios

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// ResourceSaturation patches sample-app resource limits to near-zero
// (1m CPU, 16Mi RAM), causing CPU throttling and OOMKill. Idempotent.
func ResourceSaturation(ctx context.Context, client *kubernetes.Clientset) error {
	patch := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []map[string]any{{
						"name": SampleAppDeployment,
						"resources": map[string]any{
							"limits": map[string]any{
								"cpu":    "1m",
								"memory": "16Mi",
							},
							"requests": map[string]any{
								"cpu":    "1m",
								"memory": "16Mi",
							},
						},
					}},
				},
			},
		},
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("resource-saturation: marshal: %w", err)
	}
	_, err = client.AppsV1().Deployments(Namespace).Patch(
		ctx, SampleAppDeployment, types.StrategicMergePatchType, b, metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("resource-saturation: patch deployment: %w", err)
	}
	return nil
}
