package runbooks

import (
	"context"
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Scale implements the "scale-deployment" runbook.
// It sets the replica count on a Deployment using the Scale sub-resource.
type Scale struct {
	KubeClient kubernetes.Interface
}

// Name returns the canonical runbook identifier.
func (s *Scale) Name() string { return "scale-deployment" }

// Execute scales the named deployment to the requested replica count.
// Required params: "namespace", "deployment", "replicas".
func (s *Scale) Execute(ctx context.Context, params map[string]string) error {
	ns := params["namespace"]
	deployment := params["deployment"]
	replicasStr := params["replicas"]
	if ns == "" || deployment == "" || replicasStr == "" {
		return fmt.Errorf("scale-deployment: params 'namespace', 'deployment', and 'replicas' are required")
	}

	replicas64, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		return fmt.Errorf("scale-deployment: invalid replicas %q: %w", replicasStr, err)
	}
	replicas := int32(replicas64)

	scale, err := s.KubeClient.AppsV1().Deployments(ns).GetScale(ctx, deployment, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("scale-deployment: get scale: %w", err)
	}

	scale.Spec.Replicas = replicas
	_, err = s.KubeClient.AppsV1().Deployments(ns).UpdateScale(ctx, deployment, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("scale-deployment: update scale: %w", err)
	}
	return nil
}
