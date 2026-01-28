// Package task provides recurring task detection and unrolling helpers.
package task

import (
	"context"
	"encoding/json"

	"github.com/jackyhe0402/daemonflow/internal/brain"
)

// ParseRecurringResponse represents the response from parse_recurring action.
type ParseRecurringResponse struct {
	Recurring    bool   `json:"recurring"`
	RRule        string `json:"rrule,omitempty"`
	CleanText    string `json:"clean_text,omitempty"`
	OriginalText string `json:"original_text,omitempty"`
}

// Occurrence represents a single dated occurrence of a recurring task.
type Occurrence struct {
	Date       string `json:"date"`
	InstanceID string `json:"instance_id"`
}

// UnrollResponse represents the response from unroll_recurring action.
type UnrollResponse struct {
	Occurrences []Occurrence `json:"occurrences"`
}

// ParseRecurring calls the Python brain to detect recurring patterns in text.
//
// Parameters:
//   - ctx: Context for cancellation/timeout support
//   - executor: Brain executor for Python calls (can be nil)
//   - text: Task text to analyze for recurring patterns
//
// Returns:
//   - *ParseRecurringResponse: Parsed result with RRULE if recurring
//   - error: If the action failed
//
// If executor is nil, returns (nil, nil) to support graceful degradation
// when the brain is unavailable.
func ParseRecurring(ctx context.Context, executor *brain.Executor, text string) (*ParseRecurringResponse, error) {
	// Graceful degradation when brain unavailable
	if executor == nil {
		return nil, nil
	}

	// Call parse_recurring action
	input := map[string]interface{}{
		"text": text,
	}

	result, err := executor.Execute(ctx, "parse_recurring", input)
	if err != nil {
		return nil, err
	}

	// Unmarshal response
	var resp ParseRecurringResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UnrollRecurring calls the Python brain to expand an RRULE to concrete dates.
//
// Parameters:
//   - ctx: Context for cancellation/timeout support
//   - executor: Brain executor for Python calls (can be nil)
//   - rrule: RFC 5545 RRULE string to expand
//   - taskText: Task text for instance ID generation
//   - horizonDays: Number of days to look ahead
//
// Returns:
//   - *UnrollResponse: List of occurrences with dates and instance IDs
//   - error: If the action failed
//
// If executor is nil, returns (nil, nil) to support graceful degradation
// when the brain is unavailable.
func UnrollRecurring(ctx context.Context, executor *brain.Executor, rrule, taskText string, horizonDays int) (*UnrollResponse, error) {
	// Graceful degradation when brain unavailable
	if executor == nil {
		return nil, nil
	}

	// Call unroll_recurring action
	input := map[string]interface{}{
		"rrule":        rrule,
		"task_text":    taskText,
		"horizon_days": horizonDays,
	}

	result, err := executor.Execute(ctx, "unroll_recurring", input)
	if err != nil {
		return nil, err
	}

	// Unmarshal response
	var resp UnrollResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
