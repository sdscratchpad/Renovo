package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ravi-poc/contracts"
)

// RCAHandler handles /rca routes.
type RCAHandler struct {
	DB *sql.DB
}

// Create handles POST /rca — stores an RCAPayload.
func (h *RCAHandler) Create(w http.ResponseWriter, r *http.Request) {
	var rca contracts.RCAPayload
	if err := json.NewDecoder(r.Body).Decode(&rca); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if rca.IncidentID == "" {
		jsonError(w, "incident_id is required", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(rca)
	if err != nil {
		log.Printf("rca: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(
		`INSERT INTO rca (incident_id, data) VALUES (?, ?) ON CONFLICT(incident_id) DO UPDATE SET data = excluded.data`,
		rca.IncidentID, string(data),
	); err != nil {
		log.Printf("rca: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rca)
}

// GetByIncident handles GET /rca/{incident_id} — returns the RCA for a given incident.
func (h *RCAHandler) GetByIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	var raw string
	err := h.DB.QueryRow(`SELECT data FROM rca WHERE incident_id = ?`, incidentID).Scan(&raw)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("rca: get: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(raw))
}
