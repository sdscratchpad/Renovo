// event-store persists IncidentEvents, RCAPayloads, RemediationRequests, AuditEntries, and KPISnapshots
// to a local SQLite database. It exposes a REST API consumed by the diagnosis service, orchestrator,
// and web console.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ravi-poc/event-store/internal"
	"github.com/ravi-poc/event-store/internal/handlers"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/events.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("event-store using DB at %s", dbPath)

	db, err := internal.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("event-store: open db: %v", err)
	}
	defer db.Close()

	inc := &handlers.IncidentHandler{DB: db}
	rca := &handlers.RCAHandler{DB: db}
	rem := &handlers.RemediationHandler{DB: db}
	aud := &handlers.AuditHandler{DB: db}
	kpi := &handlers.KPIHandler{DB: db}
	res := &handlers.RemediationResultHandler{DB: db}
	sta := &handlers.StatusHandler{DB: db}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Incidents
	mux.HandleFunc("POST /incidents", inc.Create)
	mux.HandleFunc("GET /incidents", inc.List)
	mux.HandleFunc("GET /incidents/{id}", inc.Get)

	// RCA
	mux.HandleFunc("POST /rca", rca.Create)
	mux.HandleFunc("GET /rca/{incident_id}", rca.GetByIncident)

	// Remediations
	mux.HandleFunc("POST /remediations", rem.Create)
	mux.HandleFunc("GET /remediations/{incident_id}", rem.GetByIncident)
	mux.HandleFunc("PATCH /remediations/{id}/approve", rem.Approve)

	// Audit
	mux.HandleFunc("POST /audit", aud.Create)
	mux.HandleFunc("GET /audit/{incident_id}", aud.GetByIncident)

	// KPI
	mux.HandleFunc("POST /kpi", kpi.Create)
	mux.HandleFunc("GET /kpi/{incident_id}", kpi.GetByIncident)

	// Remediation results
	mux.HandleFunc("POST /remediation-results", res.Create)
	mux.HandleFunc("GET /remediation-results/{incident_id}", res.GetByIncident)

	// Incident pipeline status
	mux.HandleFunc("POST /status", sta.Upsert)
	mux.HandleFunc("GET /status/{incident_id}", sta.Get)

	log.Printf("event-store listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("event-store exited: %v", err)
	}
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
