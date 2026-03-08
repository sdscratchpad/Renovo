package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ravi-poc/contracts"
)

// LLMInteractionHandler handles /llm-log routes.
type LLMInteractionHandler struct {
	DB *sql.DB
}

// Create handles POST /llm-log — stores an LLMInteraction record.
func (h *LLMInteractionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var rec contracts.LLMInteraction
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if rec.IncidentID == "" {
		jsonError(w, "incident_id is required", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(rec)
	if err != nil {
		log.Printf("llm-log: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(
		`INSERT INTO llm_interactions (incident_id, data) VALUES (?, ?)
		 ON CONFLICT(incident_id) DO UPDATE SET data = excluded.data`,
		rec.IncidentID, string(data),
	); err != nil {
		log.Printf("llm-log: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rec)
}

// GetByIncident handles GET /llm-log/{incident_id} — returns the stored LLM interaction.
func (h *LLMInteractionHandler) GetByIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	var raw string
	err := h.DB.QueryRow(`SELECT data FROM llm_interactions WHERE incident_id = ?`, incidentID).Scan(&raw)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("llm-log: get: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(raw))
}

// List handles GET /llm-log — returns all stored LLM interactions.
func (h *LLMInteractionHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT data FROM llm_interactions ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("llm-log: list: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var results []json.RawMessage
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			continue
		}
		results = append(results, json.RawMessage(raw))
	}
	if results == nil {
		results = []json.RawMessage{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
