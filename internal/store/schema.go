package store

import (
	"database/sql"
	"fmt"
)

// SchemaVersion is the current database schema version.
const SchemaVersion = 2

// Migration represents a database migration.
type migration struct {
	version int
	up      string
}

// migrations contains all database migrations in order.
var migrations = []migration{
	{
		version: 1,
		up: `
-- Schema version table
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL
);

-- Tasks table for global task storage
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_path TEXT NOT NULL,
	text TEXT NOT NULL,
	completed BOOLEAN NOT NULL DEFAULT 0,
	frequency TEXT,
	due_date TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	UNIQUE(project_path, text)
);

-- Activities table for historical logging
CREATE TABLE IF NOT EXISTS activities (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL,
	project_path TEXT NOT NULL,
	details TEXT,
	timestamp TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_activities_project_timestamp
ON activities(project_path, timestamp);

-- Pet state table for streak/evolution tracking (singleton row)
CREATE TABLE IF NOT EXISTS pet_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	current_streak INTEGER NOT NULL DEFAULT 1,
	longest_streak INTEGER NOT NULL DEFAULT 1,
	total_deaths INTEGER NOT NULL DEFAULT 0,
	total_resurrections INTEGER NOT NULL DEFAULT 0,
	level INTEGER NOT NULL DEFAULT 1,
	experience INTEGER NOT NULL DEFAULT 0,
	last_activity_date TEXT,
	updated_at TEXT NOT NULL
);

-- Insert initial schema version
INSERT INTO schema_version (version) VALUES (1);

-- Insert default pet state
INSERT INTO pet_state (id, current_streak, longest_streak, total_deaths, total_resurrections, level, experience, last_activity_date, updated_at)
VALUES (1, 1, 1, 0, 0, 1, 0, NULL, datetime('now'));
`,
	},
	{
		version: 2,
		up: `
-- Add project_name and priority_score columns to tasks table
ALTER TABLE tasks ADD COLUMN project_name TEXT;
ALTER TABLE tasks ADD COLUMN priority_score INTEGER NOT NULL DEFAULT 0;

-- Populate project_name from existing project_path for existing tasks
-- Extract the last component of the path (e.g., "daemonflow" from "/home/user/daemonflow")
UPDATE tasks SET project_name = (
	CASE
		WHEN project_path LIKE '%/%' THEN
			SUBSTR(project_path, LENGTH(project_path) - LENGTH(REPLACE(project_path, '/', '')) + 1)
		ELSE
			project_path
	END
);

-- Create index on priority_score for efficient sorting
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority_score DESC, due_date, created_at);

-- Update schema version
UPDATE schema_version SET version = 2;
`,
	},
}

// Migrate runs database migrations to bring the schema to the current version.
// It checks the current schema version and runs any pending migrations.
// Migrations are run inside transactions for safety.
func Migrate(db *sql.DB) error {
	// Get current schema version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Run pending migrations
	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}

		// Run migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", m.version, err)
		}

		// Execute migration SQL
		if _, err := tx.Exec(m.up); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %w", m.version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.version, err)
		}
	}

	return nil
}

// getCurrentVersion returns the current schema version.
// Returns 0 if the schema_version table doesn't exist.
func getCurrentVersion(db *sql.DB) (int, error) {
	// Check if schema_version table exists
	var tableName string
	err := db.QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='schema_version'
	`).Scan(&tableName)

	if err == sql.ErrNoRows {
		// Table doesn't exist, schema version is 0
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	// Get current version
	var version int
	err = db.QueryRow("SELECT version FROM schema_version").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return version, nil
}

// GetSchemaVersion returns the current schema version of the database.
func GetSchemaVersion(db *sql.DB) (int, error) {
	return getCurrentVersion(db)
}
