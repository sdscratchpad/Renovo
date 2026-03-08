// Package internal contains policy enforcement and execution logic for the orchestrator.
package internal

import (
	"errors"
	"os"

	"github.com/ravi-poc/contracts"
)

// allowedNamespaces is the static allow-list of Kubernetes namespaces the orchestrator may act on.
var allowedNamespaces = map[string]bool{
	"workloads": true,
	"demo":      true,
}

// Sentinel errors returned by CheckPolicy.
var (
	ErrNamespaceNotAllowed = errors.New("namespace not in allow-list")
	ErrApprovalRequired    = errors.New("high-risk action requires approval=approved")
)

// AutoModeEnabled returns true when the AUTO_MODE env var is set to "true".
// Low and medium risk actions can execute without explicit approval in auto mode.
func AutoModeEnabled() bool {
	return os.Getenv("AUTO_MODE") == "true"
}

// CheckPolicy validates that the given action is permitted to execute.
//   - The action's "namespace" param must be in allowedNamespaces.
//   - High-risk actions require approval == ApprovalApproved.
//   - Low/medium risk actions pass if AUTO_MODE=true OR approval == ApprovalApproved.
func CheckPolicy(action contracts.RemediationAction, approval contracts.ApprovalStatus) error {
	ns := action.Params["namespace"]
	if !allowedNamespaces[ns] {
		return ErrNamespaceNotAllowed
	}

	switch action.Risk {
	case contracts.RiskHigh:
		if approval != contracts.ApprovalApproved {
			return ErrApprovalRequired
		}
	case contracts.RiskMedium, contracts.RiskLow:
		if approval != contracts.ApprovalApproved && !AutoModeEnabled() {
			return ErrApprovalRequired
		}
	}

	return nil
}
