package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ravi-poc/sample-app/internal/handlers"
)

func TestHello_OK(t *testing.T) {
	t.Setenv("FORCE_ERROR_RATE", "0")

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rr := httptest.NewRecorder()

	handlers.Hello(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["message"] == "" {
		t.Error("expected non-empty message field")
	}
}

func TestHello_ContentType(t *testing.T) {
	t.Setenv("FORCE_ERROR_RATE", "0")

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rr := httptest.NewRecorder()

	handlers.Hello(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHello_ForceError(t *testing.T) {
	t.Setenv("FORCE_ERROR_RATE", "1.0")

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rr := httptest.NewRecorder()

	handlers.Hello(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected non-empty error field")
	}
}
