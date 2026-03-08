// event-store persists IncidentEvents, RCAPayloads, RemediationRequests, AuditEntries, and KPISnapshots
// to a local SQLite database. It exposes a REST API consumed by the diagnosis service, orchestrator,
// and web console.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/ravi-poc/event-store/internal"
	"github.com/ravi-poc/event-store/internal/handlers"
	"github.com/ravi-poc/event-store/internal/hub"
)

var wsUpgrader = websocket.Upgrader{
	// Allow all origins for local development.
	CheckOrigin: func(r *http.Request) bool { return true },
}

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

	h := hub.New()

	inc := &handlers.IncidentHandler{DB: db}
	rca := &handlers.RCAHandler{DB: db}
	rem := &handlers.RemediationHandler{DB: db}
	aud := &handlers.AuditHandler{DB: db}
	kpi := &handlers.KPIHandler{DB: db}
	res := &handlers.RemediationResultHandler{DB: db}
	sta := &handlers.StatusHandler{DB: db}
	llmLog := &handlers.LLMInteractionHandler{DB: db}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// WebSocket — clients connect here to receive real-time change events.
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("event-store: ws upgrade: %v", err)
			return
		}
		h.Register(conn)
		defer h.Unregister(conn)
		// Drain messages from client (keep-alive pings); exit on close/error.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	})

	// Incidents
	mux.HandleFunc("POST /incidents", notifyAfter(inc.Create, h, "incident_created"))
	mux.HandleFunc("GET /incidents", inc.List)
	mux.HandleFunc("GET /incidents/{id}", inc.Get)

	// RCA
	mux.HandleFunc("POST /rca", notifyAfter(rca.Create, h, "rca_created"))
	mux.HandleFunc("GET /rca/{incident_id}", rca.GetByIncident)

	// Remediations
	mux.HandleFunc("POST /remediations", notifyAfter(rem.Create, h, "remediation_updated"))
	mux.HandleFunc("GET /remediations/{incident_id}", rem.GetByIncident)
	mux.HandleFunc("PATCH /remediations/{id}/approve", notifyAfter(rem.Approve, h, "remediation_updated"))

	// Audit
	mux.HandleFunc("POST /audit", aud.Create)
	mux.HandleFunc("GET /audit/{incident_id}", aud.GetByIncident)

	// KPI
	mux.HandleFunc("POST /kpi", kpi.Create)
	mux.HandleFunc("GET /kpi/{incident_id}", kpi.GetByIncident)

	// Remediation results
	mux.HandleFunc("POST /remediation-results", notifyAfter(res.Create, h, "result_created"))
	mux.HandleFunc("GET /remediation-results/{incident_id}", res.GetByIncident)

	// Incident pipeline status
	mux.HandleFunc("POST /status", notifyAfter(sta.Upsert, h, "status_updated"))
	mux.HandleFunc("GET /status/{incident_id}", sta.Get)

	// LLM interaction log
	mux.HandleFunc("POST /llm-log", llmLog.Create)
	mux.HandleFunc("GET /llm-log", llmLog.List)
	mux.HandleFunc("GET /llm-log/{incident_id}", llmLog.GetByIncident)

	log.Printf("event-store listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("event-store exited: %v", err)
	}
}

// captureWriter wraps http.ResponseWriter to capture the status code written
// by a handler so notifyAfter can decide whether to broadcast.
type captureWriter struct {
	http.ResponseWriter
	status int
}

func (c *captureWriter) WriteHeader(code int) {
	c.status = code
	c.ResponseWriter.WriteHeader(code)
}

// notifyAfter wraps handler and calls h.Broadcast(eventType) after any 2xx response.
func notifyAfter(handler http.HandlerFunc, h *hub.Hub, eventType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cw := &captureWriter{ResponseWriter: w, status: http.StatusOK}
		handler(cw, r)
		if cw.status < 300 {
			go h.Broadcast(eventType)
		}
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
