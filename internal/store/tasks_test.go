package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTaskCRUD(t *testing.T) {
	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "store-tasks-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Test InsertTask
	t.Run("InsertTask", func(t *testing.T) {
		task := &Task{
			ProjectPath: "/path/to/project",
			Text:        "Test task 1",
			Completed:   false,
		}

		id, err := store.InsertTask(task)
		if err != nil {
			t.Fatalf("InsertTask failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive ID, got %d", id)
		}
	})

	// Test GetTasksByProject
	t.Run("GetTasksByProject", func(t *testing.T) {
		tasks, err := store.GetTasksByProject("/path/to/project")
		if err != nil {
			t.Fatalf("GetTasksByProject failed: %v", err)
		}
		if len(tasks) != 1 {
			t.Errorf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Text != "Test task 1" {
			t.Errorf("expected 'Test task 1', got '%s'", tasks[0].Text)
		}
		if tasks[0].Completed {
			t.Error("expected task to be incomplete")
		}
	})

	// Test InsertTask with frequency and due_date
	t.Run("InsertTaskWithOptionalFields", func(t *testing.T) {
		freq := "daily"
		dueDate := time.Now().Add(24 * time.Hour).UTC()
		task := &Task{
			ProjectPath: "/path/to/project",
			Text:        "Recurring task",
			Completed:   false,
			Frequency:   &freq,
			DueDate:     &dueDate,
		}

		id, err := store.InsertTask(task)
		if err != nil {
			t.Fatalf("InsertTask failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive ID, got %d", id)
		}

		// Verify the task was stored with optional fields
		tasks, err := store.GetTasksByProject("/path/to/project")
		if err != nil {
			t.Fatalf("GetTasksByProject failed: %v", err)
		}

		var foundTask *Task
		for _, tk := range tasks {
			if tk.Text == "Recurring task" {
				foundTask = &tk
				break
			}
		}

		if foundTask == nil {
			t.Fatal("recurring task not found")
		}

		if foundTask.Frequency == nil || *foundTask.Frequency != "daily" {
			t.Errorf("expected frequency 'daily', got %v", foundTask.Frequency)
		}

		if foundTask.DueDate == nil {
			t.Error("expected due date to be set")
		}
	})

	// Test MarkTaskComplete
	t.Run("MarkTaskComplete", func(t *testing.T) {
		tasks, _ := store.GetTasksByProject("/path/to/project")

		// Find the "Test task 1" task specifically
		var testTask1 *Task
		for _, tk := range tasks {
			if tk.Text == "Test task 1" {
				testTask1 = &tk
				break
			}
		}
		if testTask1 == nil {
			t.Fatal("Test task 1 not found")
		}
		taskID := testTask1.ID

		err := store.MarkTaskComplete(taskID)
		if err != nil {
			t.Fatalf("MarkTaskComplete failed: %v", err)
		}

		// Verify task is now complete
		tasks, _ = store.GetTasksByProject("/path/to/project")
		var foundTask *Task
		for _, tk := range tasks {
			if tk.ID == taskID {
				foundTask = &tk
				break
			}
		}

		if foundTask == nil {
			t.Fatal("task not found")
		}
		if !foundTask.Completed {
			t.Error("expected task to be completed")
		}
	})

	// Test GetIncompleteTasks
	t.Run("GetIncompleteTasks", func(t *testing.T) {
		tasks, err := store.GetIncompleteTasks()
		if err != nil {
			t.Fatalf("GetIncompleteTasks failed: %v", err)
		}

		// Only the recurring task should be incomplete now
		if len(tasks) != 1 {
			t.Errorf("expected 1 incomplete task, got %d", len(tasks))
		}
		if tasks[0].Text != "Recurring task" {
			t.Errorf("expected 'Recurring task', got '%s'", tasks[0].Text)
		}
	})

	// Test GetAllTasks
	t.Run("GetAllTasks", func(t *testing.T) {
		tasks, err := store.GetAllTasks()
		if err != nil {
			t.Fatalf("GetAllTasks failed: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(tasks))
		}
	})

	// Test UpdateTask
	t.Run("UpdateTask", func(t *testing.T) {
		tasks, _ := store.GetTasksByProject("/path/to/project")
		task := tasks[0]
		task.Text = "Updated task text"

		err := store.UpdateTask(&task)
		if err != nil {
			t.Fatalf("UpdateTask failed: %v", err)
		}

		// Verify task was updated
		tasks, _ = store.GetTasksByProject("/path/to/project")
		var foundTask *Task
		for _, tk := range tasks {
			if tk.ID == task.ID {
				foundTask = &tk
				break
			}
		}

		if foundTask == nil {
			t.Fatal("task not found")
		}
		if foundTask.Text != "Updated task text" {
			t.Errorf("expected 'Updated task text', got '%s'", foundTask.Text)
		}
	})

	// Test DeleteTask
	t.Run("DeleteTask", func(t *testing.T) {
		tasks, _ := store.GetTasksByProject("/path/to/project")
		taskID := tasks[0].ID
		initialCount := len(tasks)

		err := store.DeleteTask(taskID)
		if err != nil {
			t.Fatalf("DeleteTask failed: %v", err)
		}

		// Verify task was deleted
		tasks, _ = store.GetTasksByProject("/path/to/project")
		if len(tasks) != initialCount-1 {
			t.Errorf("expected %d tasks after delete, got %d", initialCount-1, len(tasks))
		}
	})

	// Test upsert behavior (INSERT OR REPLACE)
	t.Run("Upsert", func(t *testing.T) {
		// Insert initial task
		task := &Task{
			ProjectPath: "/upsert/test",
			Text:        "Same text",
			Completed:   false,
		}
		id1, err := store.InsertTask(task)
		if err != nil {
			t.Fatalf("first InsertTask failed: %v", err)
		}

		// Insert same project_path + text combination (should replace)
		task2 := &Task{
			ProjectPath: "/upsert/test",
			Text:        "Same text",
			Completed:   true, // Different value
		}
		id2, err := store.InsertTask(task2)
		if err != nil {
			t.Fatalf("second InsertTask failed: %v", err)
		}

		// Verify only one task exists
		tasks, _ := store.GetTasksByProject("/upsert/test")
		if len(tasks) != 1 {
			t.Errorf("expected 1 task after upsert, got %d", len(tasks))
		}

		// The ID might change with INSERT OR REPLACE
		if tasks[0].Completed != true {
			t.Error("expected task to be completed after upsert")
		}

		// Just verify we got valid IDs (they may be different due to REPLACE behavior)
		if id1 <= 0 || id2 <= 0 {
			t.Errorf("expected positive IDs, got id1=%d, id2=%d", id1, id2)
		}
	})
}

func TestTaskNullFields(t *testing.T) {
	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "store-tasks-null-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Insert task with nil frequency and due_date
	task := &Task{
		ProjectPath: "/null/test",
		Text:        "Task with nulls",
		Completed:   false,
		Frequency:   nil, // Explicitly nil
		DueDate:     nil, // Explicitly nil
	}

	_, err = store.InsertTask(task)
	if err != nil {
		t.Fatalf("InsertTask failed: %v", err)
	}

	// Retrieve and verify NULL handling
	tasks, err := store.GetTasksByProject("/null/test")
	if err != nil {
		t.Fatalf("GetTasksByProject failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Frequency != nil {
		t.Errorf("expected nil frequency, got %v", tasks[0].Frequency)
	}

	if tasks[0].DueDate != nil {
		t.Errorf("expected nil due_date, got %v", tasks[0].DueDate)
	}
}
