// Package brain provides Go→Python executor for the DaemonFlow brain layer.
package brain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Executor manages execution of Python brain actions.
type Executor struct {
	pythonPath string
	brainDir   string
}

// NewExecutor creates a new Executor with default settings.
// Uses "python3" as the default Python binary.
func NewExecutor() *Executor {
	return &Executor{
		pythonPath: "python3",
		brainDir:   "", // Empty means use current working directory
	}
}

// SetPythonPath overrides the Python binary path.
func (e *Executor) SetPythonPath(path string) {
	e.pythonPath = path
}

// SetBrainDir sets the directory containing the brain package.
// This is the directory from which "python -m brain" will be run.
func (e *Executor) SetBrainDir(dir string) {
	e.brainDir = dir
}

// Execute runs a Python brain action with the given input.
//
// Parameters:
//   - ctx: Context for cancellation/timeout support
//   - action: The action to perform (e.g., "echo")
//   - input: Data to pass to the action (will be marshaled to JSON)
//
// Returns:
//   - json.RawMessage: The raw JSON response from the brain
//   - error: If the action failed or returned an error
func (e *Executor) Execute(ctx context.Context, action string, input interface{}) (json.RawMessage, error) {
	// Marshal input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Build command: python3 -m brain --action <action> --input <json>
	args := []string{"-m", "brain", "--action", action, "--input", string(inputJSON)}
	cmd := exec.CommandContext(ctx, e.pythonPath, args...)

	// Set working directory if configured
	if e.brainDir != "" {
		cmd.Dir = e.brainDir
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err = cmd.Run()
	if err != nil {
		// Include stderr in error message for debugging
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = stdout.String() // Error might be in stdout as JSON
		}
		return nil, fmt.Errorf("brain execution failed: %w (output: %s)", err, errMsg)
	}

	// Parse stdout as JSON
	output := stdout.Bytes()
	if len(output) == 0 {
		return nil, fmt.Errorf("brain returned empty output")
	}

	// Validate it's valid JSON
	if !json.Valid(output) {
		return nil, fmt.Errorf("brain returned invalid JSON: %s", string(output))
	}

	// Check for error response
	var errResp struct {
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(output, &errResp); err == nil && errResp.Error != "" {
		return nil, fmt.Errorf("brain action error: %s", errResp.Error)
	}

	return json.RawMessage(output), nil
}

// PingResponse represents the response from a ping health check.
type PingResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Ping checks if the Python brain is reachable and returns its version.
// Returns version string on success, error if Python or brain unavailable.
func (e *Executor) Ping(ctx context.Context) (string, error) {
	// Call Execute with action="ping" and empty input
	result, err := e.Execute(ctx, "ping", map[string]interface{}{})
	if err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}

	// Parse response for status and version
	var resp PingResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", fmt.Errorf("failed to parse ping response: %w", err)
	}

	// Verify status is "ok"
	if resp.Status != "ok" {
		return "", fmt.Errorf("ping returned non-ok status: %s", resp.Status)
	}

	return resp.Version, nil
}
