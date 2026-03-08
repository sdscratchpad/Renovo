// Package runbooks contains individual remediation runbook implementations.
package runbooks

import (
	"context"
	"fmt"
	"os/exec"
)

// Rollback implements the "rollback-deployment" runbook.
// It runs `kubectl rollout undo deployment/<name> -n <namespace>` because
// the rollback operation is not directly available in client-go.
type Rollback struct{}

// Name returns the canonical runbook identifier.
func (r *Rollback) Name() string { return "rollback-deployment" }

// Execute performs a rollout undo on the named deployment.
// Required params: "namespace", "deployment".
func (r *Rollback) Execute(_ context.Context, params map[string]string) error {
	ns := params["namespace"]
	deployment := params["deployment"]
	if ns == "" || deployment == "" {
		return fmt.Errorf("rollback-deployment: params 'namespace' and 'deployment' are required")
	}

	cmd := exec.Command("kubectl", "rollout", "undo",
		fmt.Sprintf("deployment/%s", deployment),
		"-n", ns,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rollback-deployment: kubectl: %w: %s", err, string(out))
	}
	return nil
}
