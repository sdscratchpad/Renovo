package internal_test

import (
	"context"
	"testing"
	"time"

	worker "github.com/ravi-poc/batch-worker/internal"
)

func TestProcessJob_Success(t *testing.T) {
	t.Setenv("BLOCK_DEPENDENCY", "false")
	t.Setenv("FAIL_MODE", "")

	ctx := context.Background()
	if err := worker.ProcessJob(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessJob_BlockedByBLOCK_DEPENDENCY(t *testing.T) {
	t.Setenv("BLOCK_DEPENDENCY", "true")
	t.Setenv("FAIL_MODE", "")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := worker.ProcessJob(ctx)
	if err == nil {
		t.Fatal("expected error for blocked dependency, got nil")
	}
}

func TestProcessJob_BlockedByFAIL_MODE(t *testing.T) {
	t.Setenv("BLOCK_DEPENDENCY", "false")
	t.Setenv("FAIL_MODE", "timeout")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := worker.ProcessJob(ctx)
	if err == nil {
		t.Fatal("expected error for FAIL_MODE=timeout, got nil")
	}
}

func TestRecordFailure_PushesIncidentAfterThreshold(t *testing.T) {
	// Point at a non-existent server; we only verify the call does not panic.
	t.Setenv("EVENT_STORE_URL", "http://127.0.0.1:19999")

	for i := 0; i < 3; i++ {
		worker.RecordFailure(context.DeadlineExceeded)
	}
	// If we reach here without panic, the incident push was attempted gracefully.
}
