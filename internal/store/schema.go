package store

import "database/sql"

// SchemaVersion is the current database schema version.
const SchemaVersion = 1

// Migrate runs database migrations to bring the schema to the current version.
// This is a placeholder that will be fully implemented in Task 2.
func Migrate(db *sql.DB) error {
	// Placeholder - will be implemented in Task 2
	return nil
}
