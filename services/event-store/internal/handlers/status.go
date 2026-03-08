package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ravi-poc/contracts"
)

// StatusHandler handles /status routes.
type StatusHandler struct {
	DB *sql.DB
}

// Upsert handles POST /status — inserts or updates the pipeline status for an incident.
func (h *StatusHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var su contracts.IncidentStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&su); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if su.IncidentID == "" || su.Status == "" {
		jsonError(w, "incident_id and status are required", http.StatusBadRequest)
		return
	}
	if su.UpdatedAt.IsZero() {
		su.UpdatedAt = time.Now().UTC()
	}
	_, err := h.DB.Exec(
		`INSERT INTO incident_statuses (incident_id, status, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(incident_id) DO UPDATE SET status=excluded.status, updated_at=excluded.updated_at`,
		su.IncidentID, string(su.Status), su.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		jsonError(w, fmt.Sprintf("db: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(su)
}

// Get handles GET /status/{incident_id} — returns the current pipeline status for an incident.
// Returns "detected" as the default when no explicit status has been recorded.
func (h *StatusHandler) Get(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	if incidentID == "" {
		jsonError(w, "incident_id is required", http.StatusBadRequest)
		return
	}
	var su contracts.IncidentStatusUpdate
	var updatedAt string
	err := h.DB.QueryRow(
		`SELECT incident_id, status, updated_at FROM incident_statuses WHERE incident_id = ?`,
		incidentID,
	).Scan(&su.IncidentID, &su.Status, &updatedAt)
	if err == sql.ErrNoRows {
		su = contracts.IncidentStatusUpdate{
			IncidentID: incidentID,
			Status:     contracts.StatusDetected,
			UpdatedAt:  time.Now().UTC(),
		}
	} else if err != nil {
		jsonError(w, fmt.Sprintf("db: %v", err), http.StatusInternalServerError)
		return
	} else {
		su.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(su)
}
