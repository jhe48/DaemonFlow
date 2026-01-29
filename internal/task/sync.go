package task

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/store"
)

// TaskSync synchronizes TASKS.md files to SQLite database.
type TaskSync struct {
	store *store.Store
}

// NewTaskSync creates a new TaskSync with the given store.
func NewTaskSync(s *store.Store) *TaskSync {
	return &TaskSync{store: s}
}

// SyncDirectory parses TASKS.md from directory and syncs all tasks to SQLite.
// Returns number of tasks synced and any error.
// If TASKS.md doesn't exist, deletes all tasks for that project from DB.
func (ts *TaskSync) SyncDirectory(dirPath string) (int, error) {
	log.Printf("TaskSync: Syncing directory %s", dirPath)

	taskFilePath := filepath.Join(dirPath, "TASKS.md")

	// Parse the TASKS.md file
	tasks, err := ParseFile(taskFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("TaskSync: No TASKS.md in %s, removing project tasks", dirPath)
			// File doesn't exist - remove all tasks for this project
			return ts.removeAllProjectTasks(dirPath)
		}
		log.Printf("TaskSync: Parse error in %s: %v", dirPath, err)
		return 0, err
	}

	log.Printf("TaskSync: Found %d tasks in %s", len(tasks), dirPath)

	// Track which task texts we've seen (for stale detection)
	seenTexts := make(map[string]bool)
	syncedCount := 0

	// Sync each parsed task
	for _, task := range tasks {
		seenTexts[task.Text] = true

		storeTask := &store.Task{
			ProjectPath:   dirPath,
			Text:          task.Text,
			Completed:     task.Completed,
			PriorityScore: computePriority(task, dirPath),
		}

		_, err := ts.store.InsertTask(storeTask)
		if err != nil {
			log.Printf("TaskSync: Failed to insert task '%s': %v", task.Text, err)
			return 0, err
		}
		syncedCount++
	}

	// Delete stale tasks (in DB but not in file)
	keepTexts := make([]string, 0, len(seenTexts))
	for text := range seenTexts {
		keepTexts = append(keepTexts, text)
	}

	err = ts.store.DeleteTasksByProjectExcept(dirPath, keepTexts)
	if err != nil {
		log.Printf("TaskSync: Failed to delete stale tasks in %s: %v", dirPath, err)
		return 0, err
	}

	log.Printf("TaskSync: Successfully synced %d tasks from %s", syncedCount, dirPath)
	return len(tasks), nil
}

// removeAllProjectTasks removes all tasks for a project when TASKS.md is deleted.
func (ts *TaskSync) removeAllProjectTasks(dirPath string) (int, error) {
	// Get current tasks to count them
	texts, err := ts.store.GetTaskTextsByProject(dirPath)
	if err != nil {
		return 0, err
	}

	// Delete all tasks for this project
	err = ts.store.DeleteTasksByProjectExcept(dirPath, []string{})
	if err != nil {
		return 0, err
	}

	return len(texts), nil
}

// dueDateRegex matches due date patterns like @due(2024-01-15) or @due(2024-01-15T10:00:00)
var dueDateRegex = regexp.MustCompile(`@due\((\d{4}-\d{2}-\d{2}(?:T\d{2}:\d{2}:\d{2})?)\)`)

// computePriority calculates priority score for a task.
// Priority factors:
// - Has due date: +100 if due within 24h, +50 if due within 7d, +25 otherwise
// - Task age: +1 for each day since creation (up to +30)
func computePriority(task Task, dirPath string) int {
	score := 0

	// Check for due date in task text
	dueDate := extractDueDate(task.Text)
	if dueDate != nil {
		now := time.Now()
		hoursUntilDue := dueDate.Sub(now).Hours()

		switch {
		case hoursUntilDue <= 24:
			score += 100 // Due within 24 hours
		case hoursUntilDue <= 24*7:
			score += 50 // Due within 7 days
		default:
			score += 25 // Has due date but not urgent
		}
	}

	// Task age bonus: For now, we don't have creation time in the parsed task,
	// so we'll rely on the database to track this in the future.
	// The InsertTask uses INSERT OR REPLACE which preserves created_at on existing tasks,
	// so older tasks will naturally accumulate priority over time through future enhancements.

	return score
}

// extractDueDate parses a due date from task text.
// Supports formats: @due(2024-01-15) or @due(2024-01-15T10:00:00)
func extractDueDate(text string) *time.Time {
	matches := dueDateRegex.FindStringSubmatch(text)
	if matches == nil {
		return nil
	}

	dateStr := matches[1]

	// Try parsing with time
	if strings.Contains(dateStr, "T") {
		t, err := time.Parse("2006-01-02T15:04:05", dateStr)
		if err == nil {
			return &t
		}
	}

	// Try parsing date only (assume start of day)
	t, err := time.Parse("2006-01-02", dateStr)
	if err == nil {
		return &t
	}

	return nil
}

// extractAge calculates days since a reference time.
// Used for age-based priority scoring.
func extractAge(createdAt time.Time) int {
	days := int(time.Since(createdAt).Hours() / 24)
	if days > 30 {
		return 30 // Cap at 30 days
	}
	return days
}

// ParseDueDate is exported for use by other packages that need to extract due dates.
// Returns nil if no due date found in text.
func ParseDueDate(text string) *time.Time {
	return extractDueDate(text)
}

// FormatDueDate formats a due date for display.
// Returns empty string if date is nil.
func FormatDueDate(d *time.Time) string {
	if d == nil {
		return ""
	}
	return d.Format("2006-01-02")
}

// ComputeTaskPriority calculates priority for a task with creation time.
// This is the full priority algorithm including age-based scoring.
func ComputeTaskPriority(text string, createdAt time.Time) int {
	score := 0

	// Due date scoring
	dueDate := extractDueDate(text)
	if dueDate != nil {
		now := time.Now()
		hoursUntilDue := dueDate.Sub(now).Hours()

		switch {
		case hoursUntilDue <= 24:
			score += 100
		case hoursUntilDue <= 24*7:
			score += 50
		default:
			score += 25
		}
	}

	// Age scoring: +1 per day up to +30
	ageDays := extractAge(createdAt)
	score += ageDays

	return score
}

// PriorityScoreFromTask computes priority for an existing store.Task.
// Uses task's creation time for age-based scoring.
func PriorityScoreFromTask(task *store.Task) int {
	// Check if task has valid created_at
	if task.CreatedAt.IsZero() {
		// Use current time as fallback (no age bonus)
		return extractPriorityFromText(task.Text)
	}

	return ComputeTaskPriority(task.Text, task.CreatedAt)
}

// extractPriorityFromText computes priority based only on text (no age bonus).
func extractPriorityFromText(text string) int {
	score := 0

	dueDate := extractDueDate(text)
	if dueDate != nil {
		now := time.Now()
		hoursUntilDue := dueDate.Sub(now).Hours()

		switch {
		case hoursUntilDue <= 24:
			score += 100
		case hoursUntilDue <= 24*7:
			score += 50
		default:
			score += 25
		}
	}

	return score
}

// UpdatePrioritiesForProject recalculates priority scores for all tasks in a project.
// Useful for periodic refresh when time-based priorities change.
func (ts *TaskSync) UpdatePrioritiesForProject(dirPath string) error {
	tasks, err := ts.store.GetTasksByProject(dirPath)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		newScore := PriorityScoreFromTask(&task)
		if newScore != task.PriorityScore {
			err := ts.store.UpdatePriorityScore(task.ID, newScore)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateAllPriorities recalculates priority scores for all tasks.
// Useful for daily priority refresh (age changes, due dates become more urgent).
func (ts *TaskSync) UpdateAllPriorities() error {
	tasks, err := ts.store.GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		newScore := PriorityScoreFromTask(&task)
		if newScore != task.PriorityScore {
			err := ts.store.UpdatePriorityScore(task.ID, newScore)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ExtractDueDateFromText extracts due date from task text for external use.
// Returns nil if no due date found.
func ExtractDueDateFromText(text string) *time.Time {
	return extractDueDate(text)
}

// AgeDays returns the number of days since a time, capped at 30.
func AgeDays(t time.Time) int {
	return extractAge(t)
}

// DueDateUrgency returns urgency level for a due date.
// Returns: "urgent" (<=24h), "soon" (<=7d), "scheduled" (>7d), or "" (no due date)
func DueDateUrgency(text string) string {
	dueDate := extractDueDate(text)
	if dueDate == nil {
		return ""
	}

	now := time.Now()
	hoursUntilDue := dueDate.Sub(now).Hours()

	switch {
	case hoursUntilDue <= 24:
		return "urgent"
	case hoursUntilDue <= 24*7:
		return "soon"
	default:
		return "scheduled"
	}
}

// TaskStats provides summary statistics for synced tasks.
type TaskStats struct {
	Total       int
	Completed   int
	Incomplete  int
	DueWithin24 int
	DueWithin7d int
}

// GetProjectStats returns task statistics for a project.
func (ts *TaskSync) GetProjectStats(dirPath string) (TaskStats, error) {
	tasks, err := ts.store.GetTasksByProject(dirPath)
	if err != nil {
		return TaskStats{}, err
	}

	stats := TaskStats{Total: len(tasks)}
	now := time.Now()

	for _, task := range tasks {
		if task.Completed {
			stats.Completed++
		} else {
			stats.Incomplete++
		}

		dueDate := extractDueDate(task.Text)
		if dueDate != nil {
			hoursUntilDue := dueDate.Sub(now).Hours()
			if hoursUntilDue <= 24 {
				stats.DueWithin24++
			} else if hoursUntilDue <= 24*7 {
				stats.DueWithin7d++
			}
		}
	}

	return stats, nil
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	ProjectPath string
	TasksCount  int
	Added       int
	Updated     int
	Deleted     int
	Errors      []string
}

// String returns a human-readable summary of the sync result.
func (r SyncResult) String() string {
	return "Synced " + strconv.Itoa(r.TasksCount) + " tasks from " + filepath.Base(r.ProjectPath)
}
