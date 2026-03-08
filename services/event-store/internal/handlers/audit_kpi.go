package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ravi-poc/contracts"
)

// AuditHandler handles /audit routes.
type AuditHandler struct {
	DB *sql.DB
}

// Create handles POST /audit — stores an AuditEntry.
func (h *AuditHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entry contracts.AuditEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if entry.ID == "" || entry.IncidentID == "" {
		jsonError(w, "id and incident_id are required", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("audit: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(
		`INSERT INTO audit (id, incident_id, data) VALUES (?, ?, ?)`,
		entry.ID, entry.IncidentID, string(data),
	); err != nil {
		log.Printf("audit: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// GetByIncident handles GET /audit/{incident_id} — returns the audit trail for an incident.
func (h *AuditHandler) GetByIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	rows, err := h.DB.Query(
		`SELECT data FROM audit WHERE incident_id = ? ORDER BY created_at ASC`,
		incidentID,
	)
	if err != nil {
		log.Printf("audit: list: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]contracts.AuditEntry, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			log.Printf("audit: scan: %v", err)
			jsonError(w, "db error", http.StatusInternalServerError)
			return
		}
		var entry contracts.AuditEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			log.Printf("audit: unmarshal: %v", err)
			continue
		}
		result = append(result, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// KPIHandler handles /kpi routes.
type KPIHandler struct {
	DB *sql.DB
}

// Create handles POST /kpi — stores or replaces a KPISnapshot.
func (h *KPIHandler) Create(w http.ResponseWriter, r *http.Request) {
	var snap contracts.KPISnapshot
	if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if snap.IncidentID == "" {
		jsonError(w, "incident_id is required", http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(snap)
	if err != nil {
		log.Printf("kpi: marshal: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(
		`INSERT INTO kpi (incident_id, data) VALUES (?, ?) ON CONFLICT(incident_id) DO UPDATE SET data = excluded.data`,
		snap.IncidentID, string(data),
	); err != nil {
		log.Printf("kpi: insert: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(snap)
}

// GetByIncident handles GET /kpi/{incident_id} — returns the KPI snapshot for an incident.
func (h *KPIHandler) GetByIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	var raw string
	err := h.DB.QueryRow(`SELECT data FROM kpi WHERE incident_id = ?`, incidentID).Scan(&raw)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("kpi: get: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(raw))
}
