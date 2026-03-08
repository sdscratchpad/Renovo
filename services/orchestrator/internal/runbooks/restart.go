package runbooks

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Restart implements the "restart-pod" runbook.
// It deletes all pods matching the label selector so Kubernetes recreates them.
type Restart struct {
	KubeClient kubernetes.Interface
}

// Name returns the canonical runbook identifier.
func (r *Restart) Name() string { return "restart-pod" }

// Execute deletes pods matching label "app=<name>" in the given namespace.
// Required params: "namespace", "name" (label value for app=<name>).
func (r *Restart) Execute(ctx context.Context, params map[string]string) error {
	ns := params["namespace"]
	name := params["name"]
	if ns == "" || name == "" {
		return fmt.Errorf("restart-pod: params 'namespace' and 'name' are required")
	}

	labelSelector := fmt.Sprintf("app=%s", name)
	err := r.KubeClient.CoreV1().Pods(ns).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return fmt.Errorf("restart-pod: delete pods: %w", err)
	}
	return nil
}
