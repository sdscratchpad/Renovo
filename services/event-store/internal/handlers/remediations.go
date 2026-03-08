package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ravi-poc/contracts"
)

// RemediationHandler handles /remediations routes.
type RemediationHandler struct {
	DB *sql.DB
}

// Create handles POST /remediations — stores a RemediationRequest.
func (h *RemediationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req contracts.RemediationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Action.ID == "" || req.Action.IncidentID == "" {
		jsonError(w, "action.id and action.incident_id are required", http.StatusBadRequest)
		return
	}
	if req.Approval == "" {
		req.Approval = contracts.ApprovalPending
	}
	data, err := json.Marshal(req)
	if err != nil {
		log.Printf("remediations: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(
		`INSERT INTO remediations (id, incident_id, approval, data) VALUES (?, ?, ?, ?)`,
		req.Action.ID, req.Action.IncidentID, string(req.Approval), string(data),
	); err != nil {
		log.Printf("remediations: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

// GetByIncident handles GET /remediations/{incident_id} — returns remediations for an incident.
func (h *RemediationHandler) GetByIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	rows, err := h.DB.Query(
		`SELECT data FROM remediations WHERE incident_id = ? ORDER BY created_at DESC`,
		incidentID,
	)
	if err != nil {
		log.Printf("remediations: list: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]contracts.RemediationRequest, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			log.Printf("remediations: scan: %v", err)
			jsonError(w, "db error", http.StatusInternalServerError)
			return
		}
		var req contracts.RemediationRequest
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			log.Printf("remediations: unmarshal: %v", err)
			continue
		}
		result = append(result, req)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// approveBody is the request body for PATCH /remediations/{id}/approve.
type approveBody struct {
	Approval contracts.ApprovalStatus `json:"approval"`
}

// Approve handles PATCH /remediations/{id}/approve — approves or rejects a remediation.
func (h *RemediationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body approveBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Approval != contracts.ApprovalApproved && body.Approval != contracts.ApprovalRejected {
		jsonError(w, "approval must be 'approved' or 'rejected'", http.StatusBadRequest)
		return
	}

	var raw string
	err := h.DB.QueryRow(`SELECT data FROM remediations WHERE id = ?`, id).Scan(&raw)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("remediations: approve fetch: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}

	var req contracts.RemediationRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		log.Printf("remediations: approve unmarshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	req.Approval = body.Approval
	req.UpdatedAt = time.Now().UTC()

	updated, err := json.Marshal(req)
	if err != nil {
		log.Printf("remediations: approve marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err := h.DB.Exec(
		`UPDATE remediations SET approval = ?, data = ? WHERE id = ?`,
		string(req.Approval), string(updated), id,
	); err != nil {
		log.Printf("remediations: approve update: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}
