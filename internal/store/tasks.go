package store

import (
	"database/sql"
	"path/filepath"
	"time"
)

// Task represents a task stored in the database.
type Task struct {
	ID            int64
	ProjectPath   string
	ProjectName   string     // Extracted from project_path (e.g., "daemonflow" from "/home/user/daemonflow")
	Text          string
	Completed     bool
	Frequency     *string    // nil for one-time tasks
	DueDate       *time.Time // nil if no due date
	PriorityScore int        // Computed score for ordering (higher = more important)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// InsertTask inserts or updates a task in the database.
// Uses INSERT OR REPLACE to handle upsert based on UNIQUE(project_path, text) constraint.
// Automatically extracts project_name from project_path using filepath.Base().
// Returns the inserted row ID.
func (s *Store) InsertTask(task *Task) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Extract project_name from project_path
	projectName := filepath.Base(task.ProjectPath)

	// Handle nullable fields
	var frequency sql.NullString
	if task.Frequency != nil {
		frequency = sql.NullString{String: *task.Frequency, Valid: true}
	}

	var dueDate sql.NullString
	if task.DueDate != nil {
		dueDate = sql.NullString{String: task.DueDate.UTC().Format(time.RFC3339), Valid: true}
	}

	result, err := s.db.Exec(`
		INSERT OR REPLACE INTO tasks (project_path, project_name, text, completed, frequency, due_date, priority_score, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ProjectPath, projectName, task.Text, task.Completed, frequency, dueDate, task.PriorityScore, now, now)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateTask updates an existing task by ID.
// Sets updated_at to the current time.
// Automatically extracts project_name from project_path using filepath.Base().
func (s *Store) UpdateTask(task *Task) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Extract project_name from project_path
	projectName := filepath.Base(task.ProjectPath)

	// Handle nullable fields
	var frequency sql.NullString
	if task.Frequency != nil {
		frequency = sql.NullString{String: *task.Frequency, Valid: true}
	}

	var dueDate sql.NullString
	if task.DueDate != nil {
		dueDate = sql.NullString{String: task.DueDate.UTC().Format(time.RFC3339), Valid: true}
	}

	_, err := s.db.Exec(`
		UPDATE tasks SET
			project_path = ?,
			project_name = ?,
			text = ?,
			completed = ?,
			frequency = ?,
			due_date = ?,
			priority_score = ?,
			updated_at = ?
		WHERE id = ?
	`, task.ProjectPath, projectName, task.Text, task.Completed, frequency, dueDate, task.PriorityScore, now, task.ID)

	return err
}

// GetTasksByProject retrieves all tasks for a given project path.
// Returns tasks ordered by created_at.
func (s *Store) GetTasksByProject(projectPath string) ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT id, project_path, project_name, text, completed, frequency, due_date, priority_score, created_at, updated_at
		FROM tasks
		WHERE project_path = ?
		ORDER BY created_at
	`, projectPath)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetAllTasks retrieves all tasks from the database.
// Returns tasks ordered by project_path, then created_at.
func (s *Store) GetAllTasks() ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT id, project_path, project_name, text, completed, frequency, due_date, priority_score, created_at, updated_at
		FROM tasks
		ORDER BY project_path, created_at
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetIncompleteTasks retrieves all incomplete tasks.
// Returns tasks ordered by due_date (NULLs last), then created_at.
func (s *Store) GetIncompleteTasks() ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT id, project_path, project_name, text, completed, frequency, due_date, priority_score, created_at, updated_at
		FROM tasks
		WHERE completed = 0
		ORDER BY due_date IS NULL, due_date, created_at
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetTasksSortedByPriority retrieves all incomplete tasks ordered by priority.
// Order: priority_score DESC, due_date ASC (NULLs last), created_at ASC.
// Use this for the global task view showing what to work on next.
func (s *Store) GetTasksSortedByPriority() ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT id, project_path, project_name, text, completed, frequency, due_date, priority_score, created_at, updated_at
		FROM tasks
		WHERE completed = 0
		ORDER BY priority_score DESC, due_date IS NULL, due_date ASC, created_at ASC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

// DeleteTask deletes a task by ID.
func (s *Store) DeleteTask(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

// MarkTaskComplete marks a task as complete by ID.
// Sets completed = 1 and updates updated_at.
func (s *Store) MarkTaskComplete(id int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		UPDATE tasks SET completed = 1, updated_at = ? WHERE id = ?
	`, now, id)
	return err
}

// scanTasks scans rows into a slice of Task structs.
func scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task

	for rows.Next() {
		var task Task
		var projectName sql.NullString
		var frequency sql.NullString
		var dueDate sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(
			&task.ID,
			&task.ProjectPath,
			&projectName,
			&task.Text,
			&task.Completed,
			&frequency,
			&dueDate,
			&task.PriorityScore,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse nullable fields
		if projectName.Valid {
			task.ProjectName = projectName.String
		}

		if frequency.Valid {
			task.Frequency = &frequency.String
		}

		if dueDate.Valid {
			t, err := time.Parse(time.RFC3339, dueDate.String)
			if err == nil {
				task.DueDate = &t
			}
		}

		// Parse timestamps
		task.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		task.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
