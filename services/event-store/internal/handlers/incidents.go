package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ravi-poc/contracts"
)

// IncidentHandler handles /incidents routes.
type IncidentHandler struct {
	DB *sql.DB
}

// Create handles POST /incidents — stores a new IncidentEvent.
func (h *IncidentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var inc contracts.IncidentEvent
	if err := json.NewDecoder(r.Body).Decode(&inc); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if inc.ID == "" {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(inc)
	if err != nil {
		log.Printf("incidents: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(`INSERT INTO incidents (id, data) VALUES (?, ?)`, inc.ID, string(data)); err != nil {
		log.Printf("incidents: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inc)
}

// List handles GET /incidents — returns all incidents, newest first.
func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT data FROM incidents ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("incidents: list: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]contracts.IncidentEvent, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			log.Printf("incidents: scan: %v", err)
			jsonError(w, "db error", http.StatusInternalServerError)
			return
		}
		var inc contracts.IncidentEvent
		if err := json.Unmarshal([]byte(raw), &inc); err != nil {
			log.Printf("incidents: unmarshal: %v", err)
			continue
		}
		result = append(result, inc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Get handles GET /incidents/{id} — returns a single incident.
func (h *IncidentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var raw string
	err := h.DB.QueryRow(`SELECT data FROM incidents WHERE id = ?`, id).Scan(&raw)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("incidents: get: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(raw))
}

// jsonError writes a JSON error response.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
