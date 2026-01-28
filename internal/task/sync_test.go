package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/store"
)

// createTestStore creates an in-memory SQLite store for testing.
func createTestStore(t *testing.T) *store.Store {
	t.Helper()

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	t.Cleanup(func() {
		s.Close()
	})

	return s
}

// createTestTasksFile creates a TASKS.md file in the given directory.
func createTestTasksFile(t *testing.T, dir string, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, "TASKS.md")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test TASKS.md: %v", err)
	}

	return filePath
}

func TestSyncDirectory_NewTasks(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	projectDir := t.TempDir()
	createTestTasksFile(t, projectDir, `# Project Tasks

- [ ] First task
- [ ] Second task
- [x] Third task (done)
`)

	// Execute
	count, err := ts.SyncDirectory(projectDir)

	// Verify
	if err != nil {
		t.Fatalf("SyncDirectory failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 tasks synced, got %d", count)
	}

	// Verify tasks in database
	tasks, err := s.GetTasksByProject(projectDir)
	if err != nil {
		t.Fatalf("GetTasksByProject failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks in DB, got %d", len(tasks))
	}

	// Verify task content
	foundTasks := make(map[string]store.Task)
	for _, task := range tasks {
		foundTasks[task.Text] = task
	}

	// Check First task
	if task, ok := foundTasks["First task"]; !ok {
		t.Error("First task not found in DB")
	} else if task.Completed {
		t.Error("First task should not be completed")
	}

	// Check Second task
	if task, ok := foundTasks["Second task"]; !ok {
		t.Error("Second task not found in DB")
	} else if task.Completed {
		t.Error("Second task should not be completed")
	}

	// Check Third task (done)
	if task, ok := foundTasks["Third task (done)"]; !ok {
		t.Error("Third task not found in DB")
	} else if !task.Completed {
		t.Error("Third task should be completed")
	}
}

func TestSyncDirectory_UpdateTasks(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	projectDir := t.TempDir()

	// First sync with incomplete task
	createTestTasksFile(t, projectDir, `- [ ] Task to complete
`)

	_, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("First SyncDirectory failed: %v", err)
	}

	// Verify task is incomplete
	tasks, _ := s.GetTasksByProject(projectDir)
	if len(tasks) != 1 || tasks[0].Completed {
		t.Error("Task should be incomplete after first sync")
	}

	// Update file - mark task as complete
	createTestTasksFile(t, projectDir, `- [x] Task to complete
`)

	// Execute second sync
	count, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("Second SyncDirectory failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 task synced, got %d", count)
	}

	// Verify task is now complete
	tasks, _ = s.GetTasksByProject(projectDir)
	if len(tasks) != 1 {
		t.Errorf("expected 1 task in DB, got %d", len(tasks))
	}

	if !tasks[0].Completed {
		t.Error("Task should be completed after second sync")
	}
}

func TestSyncDirectory_DeleteStaleTasks(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	projectDir := t.TempDir()

	// First sync with 3 tasks
	createTestTasksFile(t, projectDir, `- [ ] Keep me
- [ ] Delete me
- [ ] Also keep me
`)

	_, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("First SyncDirectory failed: %v", err)
	}

	// Verify 3 tasks in DB
	tasks, _ := s.GetTasksByProject(projectDir)
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks after first sync, got %d", len(tasks))
	}

	// Update file - remove "Delete me" task
	createTestTasksFile(t, projectDir, `- [ ] Keep me
- [ ] Also keep me
`)

	// Execute second sync
	_, err = ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("Second SyncDirectory failed: %v", err)
	}

	// Verify only 2 tasks remain
	tasks, _ = s.GetTasksByProject(projectDir)
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks after second sync, got %d", len(tasks))
	}

	// Verify "Delete me" is gone
	for _, task := range tasks {
		if task.Text == "Delete me" {
			t.Error("Task 'Delete me' should have been removed")
		}
	}

	// Verify remaining tasks
	foundKeep := false
	foundAlso := false
	for _, task := range tasks {
		if task.Text == "Keep me" {
			foundKeep = true
		}
		if task.Text == "Also keep me" {
			foundAlso = true
		}
	}
	if !foundKeep {
		t.Error("Task 'Keep me' should still exist")
	}
	if !foundAlso {
		t.Error("Task 'Also keep me' should still exist")
	}
}

func TestSyncDirectory_PriorityScoring(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	projectDir := t.TempDir()

	// Create tasks with different due date urgencies
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	nextWeek := time.Now().Add(6 * 24 * time.Hour).Format("2006-01-02")
	nextMonth := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")

	createTestTasksFile(t, projectDir, `- [ ] No due date task
- [ ] Urgent task @due(`+tomorrow+`)
- [ ] Soon task @due(`+nextWeek+`)
- [ ] Later task @due(`+nextMonth+`)
`)

	// Execute
	_, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("SyncDirectory failed: %v", err)
	}

	// Verify priority scores
	tasks, _ := s.GetTasksByProject(projectDir)
	taskScores := make(map[string]int)
	for _, task := range tasks {
		taskScores[task.Text] = task.PriorityScore
	}

	// No due date should have lowest priority (0)
	if taskScores["No due date task"] != 0 {
		t.Errorf("No due date task should have priority 0, got %d", taskScores["No due date task"])
	}

	// Urgent task (within 24h) should have priority 100
	urgentScore := taskScores["Urgent task @due("+tomorrow+")"]
	if urgentScore != 100 {
		t.Errorf("Urgent task should have priority 100, got %d", urgentScore)
	}

	// Soon task (within 7d) should have priority 50
	soonScore := taskScores["Soon task @due("+nextWeek+")"]
	if soonScore != 50 {
		t.Errorf("Soon task should have priority 50, got %d", soonScore)
	}

	// Later task (>7d) should have priority 25
	laterScore := taskScores["Later task @due("+nextMonth+")"]
	if laterScore != 25 {
		t.Errorf("Later task should have priority 25, got %d", laterScore)
	}

	// Verify order: urgent > soon > later > no-due
	if urgentScore <= soonScore || soonScore <= laterScore || laterScore <= taskScores["No due date task"] {
		t.Error("Priority scores should be: urgent > soon > later > no-due")
	}
}

func TestSyncDirectory_FileNotFound(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	projectDir := t.TempDir()

	// First, create some tasks in DB
	createTestTasksFile(t, projectDir, `- [ ] Task one
- [ ] Task two
`)

	_, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("First SyncDirectory failed: %v", err)
	}

	// Verify tasks in DB
	tasks, _ := s.GetTasksByProject(projectDir)
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks after first sync, got %d", len(tasks))
	}

	// Delete the TASKS.md file
	os.Remove(filepath.Join(projectDir, "TASKS.md"))

	// Execute sync on directory with no TASKS.md
	count, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("Second SyncDirectory failed: %v", err)
	}

	// Count should be 2 (the tasks that were deleted)
	if count != 2 {
		t.Errorf("expected count to be 2 (deleted tasks), got %d", count)
	}

	// Verify all tasks removed from DB
	tasks, _ = s.GetTasksByProject(projectDir)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after file deletion, got %d", len(tasks))
	}
}

func TestSyncDirectory_MultipleProjects(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	// Create two project directories
	projectA := t.TempDir()
	projectB := t.TempDir()

	createTestTasksFile(t, projectA, `- [ ] Project A task 1
- [ ] Project A task 2
`)

	createTestTasksFile(t, projectB, `- [ ] Project B task 1
- [x] Project B task 2
- [ ] Project B task 3
`)

	// Sync both projects
	countA, _ := ts.SyncDirectory(projectA)
	countB, _ := ts.SyncDirectory(projectB)

	if countA != 2 {
		t.Errorf("expected 2 tasks for project A, got %d", countA)
	}
	if countB != 3 {
		t.Errorf("expected 3 tasks for project B, got %d", countB)
	}

	// Verify tasks are separated by project
	tasksA, _ := s.GetTasksByProject(projectA)
	tasksB, _ := s.GetTasksByProject(projectB)

	if len(tasksA) != 2 {
		t.Errorf("expected 2 tasks in DB for project A, got %d", len(tasksA))
	}
	if len(tasksB) != 3 {
		t.Errorf("expected 3 tasks in DB for project B, got %d", len(tasksB))
	}

	// Verify project paths are correct
	for _, task := range tasksA {
		if task.ProjectPath != projectA {
			t.Errorf("task has wrong project path: %s", task.ProjectPath)
		}
	}
	for _, task := range tasksB {
		if task.ProjectPath != projectB {
			t.Errorf("task has wrong project path: %s", task.ProjectPath)
		}
	}

	// Verify project names are set
	for _, task := range tasksA {
		if task.ProjectName == "" {
			t.Error("project name should be set")
		}
	}
}

func TestExtractDueDate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantDate string // empty if no date expected
	}{
		{
			name:     "date only",
			text:     "Task with @due(2024-06-15)",
			wantDate: "2024-06-15",
		},
		{
			name:     "date with time",
			text:     "Task with @due(2024-06-15T14:30:00)",
			wantDate: "2024-06-15",
		},
		{
			name:     "no due date",
			text:     "Task without due date",
			wantDate: "",
		},
		{
			name:     "invalid format",
			text:     "Task with @due(invalid)",
			wantDate: "",
		},
		{
			name:     "due date at end",
			text:     "Important task @due(2024-12-25)",
			wantDate: "2024-12-25",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractDueDateFromText(tc.text)

			if tc.wantDate == "" {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected date, got nil")
			}

			gotDate := result.Format("2006-01-02")
			if gotDate != tc.wantDate {
				t.Errorf("expected date %s, got %s", tc.wantDate, gotDate)
			}
		})
	}
}

func TestDueDateUrgency(t *testing.T) {
	tests := []struct {
		name       string
		daysOffset int // days from now
		want       string
	}{
		{"due in 12 hours", 0, "urgent"},    // within 24h
		{"due tomorrow", 1, "urgent"},       // within 24h (edge case)
		{"due in 3 days", 3, "soon"},        // within 7d
		{"due in 7 days", 7, "soon"},        // within 7d (edge case)
		{"due in 14 days", 14, "scheduled"}, // >7d
		{"due in 30 days", 30, "scheduled"}, // >7d
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dueDate := time.Now().Add(time.Duration(tc.daysOffset*24) * time.Hour).Format("2006-01-02")
			text := "Task @due(" + dueDate + ")"

			got := DueDateUrgency(text)
			if got != tc.want {
				t.Errorf("DueDateUrgency(%q) = %q, want %q", text, got, tc.want)
			}
		})
	}

	// Test no due date
	t.Run("no due date", func(t *testing.T) {
		got := DueDateUrgency("Task without due date")
		if got != "" {
			t.Errorf("expected empty string for no due date, got %q", got)
		}
	})
}

func TestGetProjectStats(t *testing.T) {
	// Setup
	s := createTestStore(t)
	ts := NewTaskSync(s)

	projectDir := t.TempDir()

	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	nextWeek := time.Now().Add(5 * 24 * time.Hour).Format("2006-01-02")

	createTestTasksFile(t, projectDir, `- [ ] Normal task
- [x] Completed task
- [ ] Urgent task @due(`+tomorrow+`)
- [ ] Soon task @due(`+nextWeek+`)
- [x] Another completed
`)

	_, err := ts.SyncDirectory(projectDir)
	if err != nil {
		t.Fatalf("SyncDirectory failed: %v", err)
	}

	// Execute
	stats, err := ts.GetProjectStats(projectDir)
	if err != nil {
		t.Fatalf("GetProjectStats failed: %v", err)
	}

	// Verify
	if stats.Total != 5 {
		t.Errorf("expected 5 total, got %d", stats.Total)
	}
	if stats.Completed != 2 {
		t.Errorf("expected 2 completed, got %d", stats.Completed)
	}
	if stats.Incomplete != 3 {
		t.Errorf("expected 3 incomplete, got %d", stats.Incomplete)
	}
	if stats.DueWithin24 != 1 {
		t.Errorf("expected 1 due within 24h, got %d", stats.DueWithin24)
	}
	if stats.DueWithin7d != 1 {
		t.Errorf("expected 1 due within 7d, got %d", stats.DueWithin7d)
	}
}
