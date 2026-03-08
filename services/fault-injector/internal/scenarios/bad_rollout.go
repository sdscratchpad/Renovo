package scenarios

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	Namespace           = "workloads"
	SampleAppDeployment = "sample-app"
)

// BadRollout patches sample-app with a missing-secret env var to cause
// CreateContainerConfigError / CrashLoopBackOff. Idempotent.
func BadRollout(ctx context.Context, client *kubernetes.Clientset) error {
	patch := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []map[string]any{{
						"name": SampleAppDeployment,
						"env": []map[string]any{{
							"name": "BREAK_APP",
							"valueFrom": map[string]any{
								"secretKeyRef": map[string]any{
									"name":     "deliberately-missing-secret",
									"key":      "data",
									"optional": false,
								},
							},
						}},
					}},
				},
			},
		},
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("bad-rollout: marshal: %w", err)
	}
	_, err = client.AppsV1().Deployments(Namespace).Patch(
		ctx, SampleAppDeployment, types.StrategicMergePatchType, b, metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("bad-rollout: patch deployment: %w", err)
	}
	return nil
}
