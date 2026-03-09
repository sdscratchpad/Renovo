// diagnosis is the AI engine service. It receives IncidentEvents, queries Prometheus/log data
// for supporting evidence, then calls GPT-4o via GitHub Models to generate an RCA and
// ranked list of RemediationActions. Results are persisted to the event-store.
//
// Auth: set GITHUB_TOKEN env var with a GitHub PAT that has the models:read scope.
// Endpoint: https://models.inference.ai.azure.com (GitHub Models inference endpoint).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ravi-poc/contracts"
	diaginternal "github.com/ravi-poc/diagnosis/internal"
	"github.com/ravi-poc/diagnosis/internal/detector"
	"github.com/ravi-poc/diagnosis/internal/llm"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN env var is required for GitHub Models access")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	eventStoreURL := os.Getenv("EVENT_STORE_URL")
	if eventStoreURL == "" {
		eventStoreURL = "http://localhost:8085"
	}

	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://localhost:9090"
	}

	sampleAppURL := os.Getenv("SAMPLE_APP_URL")
	if sampleAppURL == "" {
		sampleAppURL = "http://localhost:8080"
	}

	llmClient := llm.New(token)
	analyzer := diaginternal.NewAnalyzer(llmClient, eventStoreURL, prometheusURL)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// POST /diagnose — accept an IncidentEvent, run RCA, return RCAPayload.
	mux.HandleFunc("POST /diagnose", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			jsonError(w, "read body failed", http.StatusBadRequest)
			return
		}
		var inc contracts.IncidentEvent
		if err := json.Unmarshal(body, &inc); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if inc.ID == "" || inc.Scenario == "" {
			jsonError(w, "id and scenario are required", http.StatusBadRequest)
			return
		}

		// Persist the incident to event-store before analysis (best-effort).
		persistIncident(eventStoreURL, inc)

		rca, _, err := analyzer.Analyze(r.Context(), inc)
		if err != nil {
			log.Printf("diagnose: %v", err)
			jsonError(w, fmt.Sprintf("analysis failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rca)
	})

	// GET /incidents — proxy list to event-store.
	mux.HandleFunc("GET /incidents", func(w http.ResponseWriter, r *http.Request) {
		proxy(w, r, eventStoreURL+"/incidents")
	})

	// GET /incidents/{id}/rca — proxy to event-store.
	mux.HandleFunc("GET /incidents/{id}/rca", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		proxy(w, r, eventStoreURL+"/rca/"+id)
	})

	// Start background watcher: auto-diagnose new incidents from event-store.
	startWatcher(eventStoreURL, analyzer)

	// Start SRE golden signal detector: autonomously detect faults and create incidents.
	det := detector.New(prometheusURL, eventStoreURL, sampleAppURL)
	go det.Start(context.Background())

	log.Printf("diagnosis service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("diagnosis exited: %v", err)
	}
}

// startWatcher polls the event-store for new incidents that have no pipeline status yet
// (meaning the fault-injector created them but nobody kicked off analysis) and automatically
// runs Analyze on each one. This closes the loop: inject → detect → AI diagnose → remediate.
func startWatcher(eventStoreURL string, analyzer *diaginternal.Analyzer) {
	log.Println("watcher: incident auto-trigger started (polling every 5s)")
	go func() {
		// Brief startup delay so event-store is fully up before first poll.
		time.Sleep(3 * time.Second)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		var inProgress sync.Map // keyed by incident ID
		for range ticker.C {
			processUnanalyzed(eventStoreURL, analyzer, &inProgress)
		}
	}()
}

// processUnanalyzed fetches all incidents, and for any that have no status record (never touched
// by the diagnosis pipeline), kicks off an async Analyze call.
func processUnanalyzed(eventStoreURL string, analyzer *diaginternal.Analyzer, inProgress *sync.Map) {
	resp, err := http.Get(eventStoreURL + "/incidents")
	if err != nil {
		log.Printf("watcher: fetch incidents: %v", err)
		return
	}
	defer resp.Body.Close()

	var incidents []contracts.IncidentEvent
	if err := json.NewDecoder(resp.Body).Decode(&incidents); err != nil {
		log.Printf("watcher: decode incidents: %v", err)
		return
	}

	for _, inc := range incidents {
		// Skip if already being analyzed in this process.
		if _, running := inProgress.Load(inc.ID); running {
			continue
		}

		// Check pipeline status. The event-store returns "detected" (HTTP 200) as the
		// default for incidents that have no status row — i.e. never been touched by
		// the diagnosis pipeline. Only process those.
		statResp, err := http.Get(eventStoreURL + "/status/" + inc.ID)
		if err != nil {
			log.Printf("watcher: check status %s: %v", inc.ID, err)
			continue
		}
		var su contracts.IncidentStatusUpdate
		decodeErr := json.NewDecoder(statResp.Body).Decode(&su)
		statResp.Body.Close()
		if decodeErr != nil {
			continue
		}
		// Process if: never touched ("detected"), OR stuck in "analyzing" for >2 min
		// (service restart mid-analysis leaves incidents permanently orphaned otherwise).
		stuckAnalyzing := su.Status == contracts.StatusAnalyzing &&
			time.Since(su.UpdatedAt) > 2*time.Minute
		if su.Status != contracts.StatusDetected && !stuckAnalyzing {
			continue
		}

		// Mark in-progress before spawning goroutine to prevent double-dispatch.
		inProgress.Store(inc.ID, true)
		go func(incident contracts.IncidentEvent) {
			defer inProgress.Delete(incident.ID)
			log.Printf("watcher: auto-diagnosing %s (scenario=%s service=%s)", incident.ID, incident.Scenario, incident.Service)
			_, _, err := analyzer.Analyze(context.Background(), incident)
			if err != nil {
				log.Printf("watcher: analysis failed for %s: %v", incident.ID, err)
			} else {
				log.Printf("watcher: analysis complete for %s", incident.ID)
			}
		}(inc)
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

// persistIncident posts an IncidentEvent to the event-store (best-effort).
func persistIncident(eventStoreURL string, inc contracts.IncidentEvent) {
	data, err := json.Marshal(inc)
	if err != nil {
		log.Printf("persistIncident: marshal: %v", err)
		return
	}
	resp, err := http.Post(eventStoreURL+"/incidents", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("persistIncident: post: %v", err)
		return
	}
	resp.Body.Close()
}

// proxy forwards a request to targetURL and writes the response back.
func proxy(w http.ResponseWriter, r *http.Request, targetURL string) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		jsonError(w, "proxy build request failed", http.StatusInternalServerError)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		jsonError(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
