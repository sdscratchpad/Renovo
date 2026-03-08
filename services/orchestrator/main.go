// orchestrator executes approved RemediationActions by calling the Kubernetes API.
// Supported runbooks: rollback-deployment, restart-pod, scale-deployment, retry-batch-job.
// Policy checks and blast-radius controls are enforced before any action is executed.
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
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ravi-poc/contracts"
	"github.com/ravi-poc/orchestrator/internal"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildKubeClient() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig.
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("kubeconfig: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	eventStoreURL := os.Getenv("EVENT_STORE_URL")
	if eventStoreURL == "" {
		eventStoreURL = "http://localhost:8085"
	}

	kubeClient, err := buildKubeClient()
	if err != nil {
		log.Printf("orchestrator: kube client unavailable (dry-run mode): %v", err)
		kubeClient = nil
	}

	exec := internal.NewExecutor(eventStoreURL, kubeClient)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// POST /remediate — validate policy and execute the remediation runbook.
	mux.HandleFunc("POST /remediate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			contracts.RemediationAction
			Approval contracts.ApprovalStatus `json:"approval"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.ID == "" {
			req.ID = uuid.NewString()
		}
		if req.CreatedAt.IsZero() {
			req.CreatedAt = time.Now().UTC()
		}

		if err := internal.CheckPolicy(req.RemediationAction, req.Approval); err != nil {
			switch err {
			case internal.ErrNamespaceNotAllowed:
				jsonError(w, err.Error(), http.StatusForbidden)
			case internal.ErrApprovalRequired:
				jsonError(w, err.Error(), http.StatusForbidden)
			default:
				jsonError(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		// Store in event-store before executing.
		remReq := contracts.RemediationRequest{
			Action:    req.RemediationAction,
			Approval:  req.Approval,
			UpdatedAt: time.Now().UTC(),
		}
		postToEventStore(eventStoreURL+"/remediations", remReq)
		postToEventStore(eventStoreURL+"/status", contracts.IncidentStatusUpdate{
			IncidentID: req.IncidentID,
			Status:     contracts.StatusRemediating,
			UpdatedAt:  time.Now().UTC(),
		})

		result := exec.Execute(r.Context(), req.RemediationAction)
		go persistKPI(eventStoreURL, result)
		finalStatus := contracts.StatusFailed
		if result.Success {
			finalStatus = contracts.StatusResolved
		}
		go postToEventStore(eventStoreURL+"/status", contracts.IncidentStatusUpdate{
			IncidentID: result.IncidentID,
			Status:     finalStatus,
			UpdatedAt:  time.Now().UTC(),
		})
		w.Header().Set("Content-Type", "application/json")
		if result.Success {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(result)
	})

	// GET /remediations — list remediations for an incident (incident_id query param required).
	mux.HandleFunc("GET /remediations", func(w http.ResponseWriter, r *http.Request) {
		incidentID := r.URL.Query().Get("incident_id")
		if incidentID == "" {
			jsonError(w, "incident_id query param is required", http.StatusBadRequest)
			return
		}
		resp, err := http.Get(fmt.Sprintf("%s/remediations/%s", eventStoreURL, incidentID))
		if err != nil {
			log.Printf("orchestrator: event-store list: %v", err)
			jsonError(w, "event-store unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// PATCH /remediations/{id}/approve — update approval in event-store, then execute if approved.
	mux.HandleFunc("PATCH /remediations/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			jsonError(w, "cannot read body", http.StatusBadRequest)
			return
		}

		var approveReq struct {
			Approval contracts.ApprovalStatus `json:"approval"`
		}
		if err := json.Unmarshal(body, &approveReq); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Forward to event-store to persist the approval change.
		esURL := fmt.Sprintf("%s/remediations/%s/approve", eventStoreURL, id)
		patchReq, err := http.NewRequestWithContext(r.Context(), http.MethodPatch, esURL, bytes.NewReader(body))
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}
		patchReq.Header.Set("Content-Type", "application/json")
		esResp, err := http.DefaultClient.Do(patchReq)
		if err != nil {
			log.Printf("orchestrator: event-store approve: %v", err)
			jsonError(w, "event-store unavailable", http.StatusBadGateway)
			return
		}
		defer esResp.Body.Close()

		if esResp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(esResp.StatusCode)
			io.Copy(w, esResp.Body)
			return
		}

		// Decode the updated RemediationRequest returned by the event-store.
		var remReq contracts.RemediationRequest
		if err := json.NewDecoder(esResp.Body).Decode(&remReq); err != nil {
			jsonError(w, "event-store response invalid", http.StatusBadGateway)
			return
		}

		// If approved, execute the runbook asynchronously so the caller gets a
		// "remediating" response immediately. The UI polls /remediation-results
		// to pick up the outcome once the runbook finishes.
		if approveReq.Approval == contracts.ApprovalApproved {
			if err := internal.CheckPolicy(remReq.Action, contracts.ApprovalApproved); err != nil {
				jsonError(w, err.Error(), http.StatusForbidden)
				return
			}
			postToEventStore(eventStoreURL+"/status", contracts.IncidentStatusUpdate{
				IncidentID: remReq.Action.IncidentID,
				Status:     contracts.StatusRemediating,
				UpdatedAt:  time.Now().UTC(),
			})
			go func(action contracts.RemediationAction) {
				result := exec.Execute(context.Background(), action)
				if result.Success {
					// Runbook API call succeeded. Pods are provisioning.
					// Enter "verifying" state and wait for Kubernetes to confirm
					// the service is actually healthy before declaring resolved.
					postToEventStore(eventStoreURL+"/status", contracts.IncidentStatusUpdate{
						IncidentID: action.IncidentID,
						Status:     contracts.StatusVerifying,
						UpdatedAt:  time.Now().UTC(),
					})
					healthy, reason := internal.WaitUntilHealthy(kubeClient, action.RunbookName, action.Params)
					recoveredAt := time.Now().UTC()
					result.RecoveredAt = &recoveredAt
					if !healthy {
						result.Success = false
						result.Message = "runbook executed but verification timed out: " + reason
						log.Printf("orchestrator: verification failed for incident %s: %s", action.IncidentID, reason)
					} else {
						result.Message = fmt.Sprintf("runbook %s completed and service verified healthy", action.RunbookName)
						log.Printf("orchestrator: incident %s verified healthy", action.IncidentID)
					}
				}
				persistKPI(eventStoreURL, result)
				postToEventStore(eventStoreURL+"/remediation-results", result)
				finalStatus := contracts.StatusFailed
				if result.Success {
					finalStatus = contracts.StatusResolved
				}
				postToEventStore(eventStoreURL+"/status", contracts.IncidentStatusUpdate{
					IncidentID: result.IncidentID,
					Status:     finalStatus,
					UpdatedAt:  time.Now().UTC(),
				})
			}(remReq.Action)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{
				"status":      "remediating",
				"incident_id": remReq.Action.IncidentID,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(remReq)
	})

	log.Printf("orchestrator listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("orchestrator exited: %v", err)
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

// postToEventStore fires-and-forgets a JSON POST; logs on error.
func postToEventStore(url string, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("orchestrator: marshal for %s: %v", url, err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("orchestrator: post %s: %v", url, err)
		return
	}
	resp.Body.Close()
}

// persistKPI fetches the incident DetectedAt, merges MTTR into any existing
// KPISnapshot written by the diagnosis service, and upserts it.
func persistKPI(eventStoreURL string, result contracts.RemediationResult) {
	if result.IncidentID == "" || !result.Success {
		return
	}
	incResp, err := http.Get(eventStoreURL + "/incidents/" + result.IncidentID)
	if err != nil {
		log.Printf("orchestrator: persistKPI: fetch incident: %v", err)
		return
	}
	defer incResp.Body.Close()
	var inc contracts.IncidentEvent
	if err := json.NewDecoder(incResp.Body).Decode(&inc); err != nil || inc.DetectedAt.IsZero() {
		log.Printf("orchestrator: persistKPI: decode incident: %v", err)
		return
	}
	// Read existing KPI to preserve MTTD written by diagnosis.
	snap := contracts.KPISnapshot{IncidentID: result.IncidentID}
	if kpiResp, err := http.Get(eventStoreURL + "/kpi/" + result.IncidentID); err == nil {
		if kpiResp.StatusCode == http.StatusOK {
			json.NewDecoder(kpiResp.Body).Decode(&snap)
		}
		kpiResp.Body.Close()
	}
	snap.IncidentID = result.IncidentID
	snap.MTTR = result.ExecutedAt.Sub(inc.DetectedAt)
	data, err := json.Marshal(snap)
	if err != nil {
		log.Printf("orchestrator: persistKPI: marshal: %v", err)
		return
	}
	resp, err := http.Post(eventStoreURL+"/kpi", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("orchestrator: persistKPI: post: %v", err)
		return
	}
	resp.Body.Close()
}
