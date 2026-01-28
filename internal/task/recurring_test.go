package task

import (
	"context"
	"testing"
)

// TestParseRecurring_NilExecutor tests that ParseRecurring handles nil executor gracefully.
func TestParseRecurring_NilExecutor(t *testing.T) {
	ctx := context.Background()

	resp, err := ParseRecurring(ctx, nil, "test task every friday")

	if err != nil {
		t.Errorf("expected nil error for nil executor, got: %v", err)
	}

	if resp != nil {
		t.Errorf("expected nil response for nil executor, got: %+v", resp)
	}
}

// TestUnrollRecurring_NilExecutor tests that UnrollRecurring handles nil executor gracefully.
func TestUnrollRecurring_NilExecutor(t *testing.T) {
	ctx := context.Background()

	resp, err := UnrollRecurring(ctx, nil, "RRULE:FREQ=WEEKLY;BYDAY=FR", "test task", 7)

	if err != nil {
		t.Errorf("expected nil error for nil executor, got: %v", err)
	}

	if resp != nil {
		t.Errorf("expected nil response for nil executor, got: %+v", resp)
	}
}

// Note: Integration tests that actually call Python brain are skipped here.
// They require Python environment setup and would be flaky in CI without
// proper environment configuration. Integration tests can be added with
// build tags like //go:build integration when needed.
