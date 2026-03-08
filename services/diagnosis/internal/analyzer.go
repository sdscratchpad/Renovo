// Package internal contains the core analysis logic for the diagnosis service.
package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ravi-poc/contracts"
	"github.com/ravi-poc/diagnosis/internal/llm"
)

const systemPrompt = `You are an SRE AI. Given incident signals, identify the root cause and recommend one remediation action. Return ONLY valid JSON matching the schema below — no markdown, no extra text.

Available runbooks and their required params:
- rollback-deployment: {"namespace": "<ns>", "deployment": "<name>"}
- scale-deployment:    {"namespace": "<ns>", "deployment": "<name>", "replicas": "<n>"}
- restart-pods:        {"namespace": "<ns>", "selector": "<label=value>"}
- retry-batch-job:     {"namespace": "<ns>", "deployment": "<name>"}

Schema: {"summary": "string", "root_cause": "string", "confidence": 0.0-1.0, "action": {"runbook": "string", "description": "string", "params": {"key": "value"}, "risk": "low|medium|high"}}`

// Analyzer fetches evidence and calls GPT-4o to produce an RCA + remediation recommendation.
type Analyzer struct {
	LLM           *llm.Client
	EventStoreURL string
	PrometheusURL string
	httpClient    *http.Client
}

// NewAnalyzer constructs an Analyzer with the given dependencies.
func NewAnalyzer(llmClient *llm.Client, eventStoreURL, prometheusURL string) *Analyzer {
	return &Analyzer{
		LLM:           llmClient,
		EventStoreURL: eventStoreURL,
		PrometheusURL: prometheusURL,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// llmAction is the remediation sub-object returned by the LLM.
type llmAction struct {
	Runbook     string            `json:"runbook"`
	Description string            `json:"description"`
	Params      map[string]string `json:"params"`
	Risk        string            `json:"risk"`
}

// llmResponse is the JSON the LLM is instructed to return.
type llmResponse struct {
	Summary    string    `json:"summary"`
	RootCause  string    `json:"root_cause"`
	Confidence float64   `json:"confidence"`
	Action     llmAction `json:"action"`
}

// Analyze runs full RCA for the given incident and persists results to event-store.
// Returns the RCAPayload and the proposed RemediationAction.
func (a *Analyzer) Analyze(ctx context.Context, inc contracts.IncidentEvent) (contracts.RCAPayload, contracts.RemediationAction, error) {
	evidence := inc.Evidence

	// Signal that AI analysis has started.
	a.persistStatus(ctx, inc.ID, contracts.StatusAnalyzing)

	// Augment with Prometheus metrics when available.
	promEvidence := a.fetchPrometheusEvidence(ctx, inc)
	evidence = append(evidence, promEvidence...)

	// Build prompt.
	userMsg := a.buildPrompt(inc, evidence)

	// Call LLM.
	raw, err := a.LLM.Complete(ctx, systemPrompt, userMsg)
	if err != nil {
		return contracts.RCAPayload{}, contracts.RemediationAction{}, fmt.Errorf("analyzer: llm: %w", err)
	}

	// Strip markdown fences if the model wraps the JSON.
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}

	var llmResp llmResponse
	if err := json.Unmarshal([]byte(raw), &llmResp); err != nil {
		return contracts.RCAPayload{}, contracts.RemediationAction{}, fmt.Errorf("analyzer: parse llm response: %w\nraw: %s", err, raw)
	}

	now := time.Now().UTC()

	rca := contracts.RCAPayload{
		IncidentID:         inc.ID,
		Summary:            llmResp.Summary,
		RootCause:          llmResp.RootCause,
		ConfidenceScore:    llmResp.Confidence,
		SupportingEvidence: evidence,
		GeneratedAt:        now,
	}

	risk := contracts.RiskLow
	switch contracts.RiskLevel(llmResp.Action.Risk) {
	case contracts.RiskMedium:
		risk = contracts.RiskMedium
	case contracts.RiskHigh:
		risk = contracts.RiskHigh
	}

	action := contracts.RemediationAction{
		ID:          fmt.Sprintf("%s-action-%d", inc.ID, now.UnixNano()),
		IncidentID:  inc.ID,
		RunbookName: llmResp.Action.Runbook,
		Description: llmResp.Action.Description,
		Risk:        risk,
		Params:      llmResp.Action.Params,
		CreatedAt:   now,
	}
	if action.Params == nil {
		action.Params = map[string]string{}
	}
	// Normalise: some LLM responses use "service" instead of the required "deployment" key.
	if action.Params["deployment"] == "" && action.Params["service"] != "" {
		action.Params["deployment"] = action.Params["service"]
	}
	// Ensure namespace is present; fall back to the incident namespace.
	if action.Params["namespace"] == "" {
		action.Params["namespace"] = inc.Namespace
	}

	// Persist to event-store (best-effort; do not fail the RCA on store errors).
	a.persistRCA(ctx, rca)
	a.persistRemediation(ctx, action)
	a.persistAudit(ctx, inc.ID, "rca_generated", fmt.Sprintf("confidence=%.2f root_cause=%s", rca.ConfidenceScore, rca.RootCause))
	a.persistKPI(ctx, inc, rca)
	a.persistStatus(ctx, inc.ID, contracts.StatusAwaitingApproval)

	return rca, action, nil
}

// buildPrompt constructs the user message from the incident and evidence.
func (a *Analyzer) buildPrompt(inc contracts.IncidentEvent, evidence []contracts.EvidenceItem) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Incident ID: %s\n", inc.ID))
	sb.WriteString(fmt.Sprintf("Scenario: %s\n", inc.Scenario))
	sb.WriteString(fmt.Sprintf("Service: %s (namespace: %s)\n", inc.Service, inc.Namespace))
	sb.WriteString(fmt.Sprintf("Severity: %s\n", inc.Severity))
	sb.WriteString(fmt.Sprintf("Detected at: %s\n", inc.DetectedAt.Format(time.RFC3339)))
	sb.WriteString("\nEvidence:\n")
	for _, e := range evidence {
		sb.WriteString(fmt.Sprintf("  - [%s] %s = %s (at %s)\n", e.Source, e.Signal, e.Value, e.Timestamp.Format(time.RFC3339)))
	}
	sb.WriteString("\nReturn the JSON response as instructed.")
	return sb.String()
}

// fetchPrometheusEvidence queries Prometheus for key signals and returns evidence items.
func (a *Analyzer) fetchPrometheusEvidence(ctx context.Context, inc contracts.IncidentEvent) []contracts.EvidenceItem {
	if a.PrometheusURL == "" {
		return nil
	}
	queries := map[string]string{
		"http_error_rate_5xx": `sum(rate(http_requests_total{status=~"5.."}[2m]))`,
		"cpu_usage_percent":   `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[2m])) * 100)`,
		"pod_restart_count":   `sum(kube_pod_container_status_restarts_total)`,
	}
	var items []contracts.EvidenceItem
	for signal, query := range queries {
		val, err := a.queryPrometheus(ctx, query)
		if err != nil {
			log.Printf("analyzer: prometheus [%s]: %v", signal, err)
			continue
		}
		items = append(items, contracts.EvidenceItem{
			Source:    "metrics",
			Signal:    signal,
			Value:     val,
			Timestamp: time.Now().UTC(),
		})
	}
	return items
}

// queryPrometheus runs an instant query against Prometheus and returns the first result value.
func (a *Analyzer) queryPrometheus(ctx context.Context, query string) (string, error) {
	endpoint := strings.TrimRight(a.PrometheusURL, "/") + "/api/v1/query"
	params := url.Values{}
	params.Set("query", query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return "N/A", nil
	}
	return fmt.Sprintf("%v", result.Data.Result[0].Value[1]), nil
}

// persistRCA posts the RCAPayload to event-store.
func (a *Analyzer) persistRCA(ctx context.Context, rca contracts.RCAPayload) {
	if a.EventStoreURL == "" {
		return
	}
	if err := a.postJSON(ctx, a.EventStoreURL+"/rca", rca); err != nil {
		log.Printf("analyzer: persist rca: %v", err)
	}
}

// persistRemediation posts the RemediationRequest to event-store.
func (a *Analyzer) persistRemediation(ctx context.Context, action contracts.RemediationAction) {
	if a.EventStoreURL == "" {
		return
	}
	req := contracts.RemediationRequest{
		Action:    action,
		Approval:  contracts.ApprovalPending,
		UpdatedAt: time.Now().UTC(),
	}
	if err := a.postJSON(ctx, a.EventStoreURL+"/remediations", req); err != nil {
		log.Printf("analyzer: persist remediation: %v", err)
	}
}

// persistAudit posts an AuditEntry to event-store.
func (a *Analyzer) persistAudit(ctx context.Context, incidentID, event, detail string) {
	if a.EventStoreURL == "" {
		return
	}
	entry := contracts.AuditEntry{
		ID:         fmt.Sprintf("%s-%s-%d", incidentID, event, time.Now().UnixNano()),
		IncidentID: incidentID,
		Actor:      "system",
		Event:      event,
		Detail:     detail,
		Timestamp:  time.Now().UTC(),
	}
	if err := a.postJSON(ctx, a.EventStoreURL+"/audit", entry); err != nil {
		log.Printf("analyzer: persist audit: %v", err)
	}
}

// persistKPI writes a KPISnapshot with MTTD (detection-to-RCA duration) to the event-store.
func (a *Analyzer) persistKPI(ctx context.Context, inc contracts.IncidentEvent, rca contracts.RCAPayload) {
	if a.EventStoreURL == "" {
		return
	}
	snap := contracts.KPISnapshot{
		IncidentID: inc.ID,
		MTTD:       rca.GeneratedAt.Sub(inc.DetectedAt),
	}
	if err := a.postJSON(ctx, a.EventStoreURL+"/kpi", snap); err != nil {
		log.Printf("analyzer: persist kpi: %v", err)
	}
}

// persistStatus posts an IncidentStatusUpdate to the event-store.
func (a *Analyzer) persistStatus(ctx context.Context, incidentID string, status contracts.IncidentStatus) {
	if a.EventStoreURL == "" {
		return
	}
	su := contracts.IncidentStatusUpdate{
		IncidentID: incidentID,
		Status:     status,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := a.postJSON(ctx, a.EventStoreURL+"/status", su); err != nil {
		log.Printf("analyzer: persist status [%s]: %v", status, err)
	}
}

// postJSON marshals v and POSTs it to url.
func (a *Analyzer) postJSON(ctx context.Context, endpoint string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
