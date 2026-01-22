package brain

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// getProjectRoot returns the project root directory for testing.
// Assumes tests are run from within the project.
func getProjectRoot() string {
	// Walk up from current directory to find go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return ""
		}
		dir = parent
	}
}

func TestExecute_EchoAction(t *testing.T) {
	projectRoot := getProjectRoot()
	if projectRoot == "" {
		t.Skip("Could not find project root (go.mod)")
	}

	executor := NewExecutor()
	executor.SetBrainDir(projectRoot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := map[string]interface{}{
		"hello": "world",
		"count": 42,
	}

	result, err := executor.Execute(ctx, "echo", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output matches input
	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if output["hello"] != "world" {
		t.Errorf("Expected hello=world, got %v", output["hello"])
	}

	// JSON numbers are float64 by default
	if output["count"] != float64(42) {
		t.Errorf("Expected count=42, got %v", output["count"])
	}
}

func TestExecute_InvalidAction(t *testing.T) {
	projectRoot := getProjectRoot()
	if projectRoot == "" {
		t.Skip("Could not find project root (go.mod)")
	}

	executor := NewExecutor()
	executor.SetBrainDir(projectRoot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := executor.Execute(ctx, "nonexistent_action", map[string]string{"test": "data"})
	if err == nil {
		t.Error("Expected error for invalid action, got nil")
	}
}

func TestExecute_MalformedInput(t *testing.T) {
	projectRoot := getProjectRoot()
	if projectRoot == "" {
		t.Skip("Could not find project root (go.mod)")
	}

	executor := NewExecutor()
	executor.SetBrainDir(projectRoot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use a channel as input - cannot be marshaled to JSON
	badInput := make(chan int)

	_, err := executor.Execute(ctx, "echo", badInput)
	if err == nil {
		t.Error("Expected error for unmarshallable input, got nil")
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	projectRoot := getProjectRoot()
	if projectRoot == "" {
		t.Skip("Could not find project root (go.mod)")
	}

	executor := NewExecutor()
	executor.SetBrainDir(projectRoot)

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executor.Execute(ctx, "echo", map[string]string{"test": "data"})
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}

func TestNewExecutor_Defaults(t *testing.T) {
	executor := NewExecutor()

	if executor.pythonPath != "python3" {
		t.Errorf("Expected default pythonPath=python3, got %s", executor.pythonPath)
	}

	if executor.brainDir != "" {
		t.Errorf("Expected default brainDir='', got %s", executor.brainDir)
	}
}

func TestSetPythonPath(t *testing.T) {
	executor := NewExecutor()
	executor.SetPythonPath("/usr/local/bin/python3.11")

	if executor.pythonPath != "/usr/local/bin/python3.11" {
		t.Errorf("Expected pythonPath=/usr/local/bin/python3.11, got %s", executor.pythonPath)
	}
}

func TestSetBrainDir(t *testing.T) {
	executor := NewExecutor()
	executor.SetBrainDir("/path/to/project")

	if executor.brainDir != "/path/to/project" {
		t.Errorf("Expected brainDir=/path/to/project, got %s", executor.brainDir)
	}
}

func TestExecute_EmptyInput(t *testing.T) {
	projectRoot := getProjectRoot()
	if projectRoot == "" {
		t.Skip("Could not find project root (go.mod)")
	}

	executor := NewExecutor()
	executor.SetBrainDir(projectRoot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Empty map is valid JSON
	result, err := executor.Execute(ctx, "echo", map[string]string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should return empty object
	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if len(output) != 0 {
		t.Errorf("Expected empty output, got %v", output)
	}
}

func TestExecute_ComplexNestedInput(t *testing.T) {
	projectRoot := getProjectRoot()
	if projectRoot == "" {
		t.Skip("Could not find project root (go.mod)")
	}

	executor := NewExecutor()
	executor.SetBrainDir(projectRoot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := map[string]interface{}{
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": 123,
			},
		},
		"array": []int{1, 2, 3},
		"null":  nil,
	}

	result, err := executor.Execute(ctx, "echo", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify nested structure preserved
	nested, ok := output["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("nested field not preserved")
	}

	deep, ok := nested["deep"].(map[string]interface{})
	if !ok {
		t.Fatal("nested.deep field not preserved")
	}

	if deep["value"] != float64(123) {
		t.Errorf("Expected nested.deep.value=123, got %v", deep["value"])
	}
}
