package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ravi-poc/contracts"
	"github.com/ravi-poc/orchestrator/internal/runbooks"
	"k8s.io/client-go/kubernetes"
)

// Runbook is the interface every remediation runbook must satisfy.
type Runbook interface {
	Name() string
	Execute(ctx context.Context, params map[string]string) error
}

// Executor dispatches RemediationActions to the matching runbook and persists
// the outcome to the event-store.
type Executor struct {
	EventStoreURL string
	runbooks      map[string]Runbook
}

// NewExecutor constructs an Executor wired to all known runbooks.
func NewExecutor(eventStoreURL string, kubeClient kubernetes.Interface) *Executor {
	rb := []Runbook{
		&runbooks.Rollback{},
		&runbooks.Restart{KubeClient: kubeClient},
		&runbooks.Scale{KubeClient: kubeClient},
		&runbooks.RetryBatch{KubeClient: kubeClient},
	}
	m := make(map[string]Runbook, len(rb))
	for _, r := range rb {
		m[r.Name()] = r
	}
	return &Executor{EventStoreURL: eventStoreURL, runbooks: m}
}

// Execute runs the action's runbook, then posts the result and an audit entry
// to the event-store. It always returns a RemediationResult regardless of success.
func (e *Executor) Execute(ctx context.Context, action contracts.RemediationAction) contracts.RemediationResult {
	result := contracts.RemediationResult{
		ActionID:   action.ID,
		IncidentID: action.IncidentID,
		ExecutedAt: time.Now().UTC(),
	}

	rb, ok := e.runbooks[action.RunbookName]
	if !ok {
		result.Success = false
		result.Message = fmt.Sprintf("unknown runbook: %s", action.RunbookName)
		e.persistResult(ctx, result, action)
		return result
	}

	if err := rb.Execute(ctx, action.Params); err != nil {
		log.Printf("orchestrator: runbook %s failed: %v", action.RunbookName, err)
		result.Success = false
		result.Message = err.Error()
	} else {
		result.Success = true
		result.Message = fmt.Sprintf("runbook %s completed successfully", action.RunbookName)
	}

	e.persistResult(ctx, result, action)
	return result
}

// persistResult posts an AuditEntry recording the outcome to the event-store.
func (e *Executor) persistResult(ctx context.Context, result contracts.RemediationResult, action contracts.RemediationAction) {
	event := "action_executed"
	if !result.Success {
		event = "action_failed"
	}
	audit := contracts.AuditEntry{
		ID:         uuid.NewString(),
		IncidentID: action.IncidentID,
		Actor:      "system",
		Event:      event,
		Detail:     fmt.Sprintf("runbook=%s success=%v: %s", action.RunbookName, result.Success, result.Message),
		Timestamp:  time.Now().UTC(),
	}
	if err := e.postJSON(ctx, e.EventStoreURL+"/audit", audit); err != nil {
		log.Printf("orchestrator: persist audit: %v", err)
	}
}

// postJSON marshals v and sends a POST request to url.
func (e *Executor) postJSON(ctx context.Context, url string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("post %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("post %s: unexpected status %d", url, resp.StatusCode)
	}
	return nil
}
