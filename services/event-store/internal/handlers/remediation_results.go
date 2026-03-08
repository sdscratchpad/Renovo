package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ravi-poc/contracts"
)

// RemediationResultHandler handles /remediation-results routes.
type RemediationResultHandler struct {
	DB *sql.DB
}

// Create handles POST /remediation-results — stores or replaces a RemediationResult.
func (h *RemediationResultHandler) Create(w http.ResponseWriter, r *http.Request) {
	var result contracts.RemediationResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if result.ActionID == "" || result.IncidentID == "" {
		jsonError(w, "action_id and incident_id are required", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("remediation_results: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(
		`INSERT OR REPLACE INTO remediation_results (action_id, incident_id, data) VALUES (?, ?, ?)`,
		result.ActionID, result.IncidentID, string(data),
	); err != nil {
		log.Printf("remediation_results: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// GetByIncident handles GET /remediation-results/{incident_id} — returns the result for one incident.
func (h *RemediationResultHandler) GetByIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	var raw string
	err := h.DB.QueryRow(
		`SELECT data FROM remediation_results WHERE incident_id = ? ORDER BY created_at DESC LIMIT 1`,
		incidentID,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("remediation_results: get: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(raw))
}
