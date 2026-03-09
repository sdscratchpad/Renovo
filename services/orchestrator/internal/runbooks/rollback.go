// Package runbooks contains individual remediation runbook implementations.
package runbooks

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Rollback implements the "rollback-deployment" runbook.
// It removes the injected bad env var from the deployment using a JSON patch,
// then waits for the rollout to complete. This approach is idempotent — multiple
// concurrent executions all converge on the same clean state, unlike
// `kubectl rollout undo` which toggles between the last two revisions and can
// oscillate when multiple incidents are queued.
type Rollback struct{}

// Name returns the canonical runbook identifier.
func (r *Rollback) Name() string { return "rollback-deployment" }

// Execute removes the env block from the first container of the named deployment.
// Required params: "namespace", "deployment".
func (r *Rollback) Execute(_ context.Context, params map[string]string) error {
	ns := params["namespace"]
	deployment := params["deployment"]
	if ns == "" || deployment == "" {
		return fmt.Errorf("rollback-deployment: params 'namespace' and 'deployment' are required")
	}

	// Remove the env block idempotently. If no env exists the patch is a no-op.
	patch := `[{"op":"remove","path":"/spec/template/spec/containers/0/env"}]`
	cmd := exec.Command("kubectl", "patch", "deployment", deployment,
		"-n", ns,
		"--type=json",
		"-p", patch,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// "remove" on a non-existent path returns an error — treat as already clean.
		if strings.Contains(string(out), "remove operation does not apply") ||
			strings.Contains(string(out), "no change") {
			return nil
		}
		return fmt.Errorf("rollback-deployment: kubectl patch: %w: %s", err, string(out))
	}

	// Wait for the new pods to become ready before returning success.
	wait := exec.Command("kubectl", "rollout", "status",
		fmt.Sprintf("deployment/%s", deployment),
		"-n", ns,
		"--timeout=120s",
	)
	if out, err := wait.CombinedOutput(); err != nil {
		return fmt.Errorf("rollback-deployment: rollout status: %w: %s", err, string(out))
	}
	return nil
}
