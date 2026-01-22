// Package store provides SQLite-based persistent storage for DaemonFlow.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // SQLite driver (pure Go, no CGO)
)

// Store manages the SQLite database connection for DaemonFlow.
type Store struct {
	db *sql.DB
}

// DefaultDBPath returns the default database path (~/.daemonflow/daemonflow.db).
func DefaultDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".daemonflow", "daemonflow.db"), nil
}

// New opens or creates a SQLite database at the specified path.
// It enables WAL mode, sets busy timeout, and enables foreign keys.
// If dbPath is empty, uses the default path (~/.daemonflow/daemonflow.db).
func New(dbPath string) (*Store, error) {
	// Use default path if not specified
	if dbPath == "" {
		var err error
		dbPath, err = DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Build DSN with pragmas
	// DSN format: file:path?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", dbPath)

	// Open database connection
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &Store{db: db}

	// Run migrations
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying *sql.DB for advanced queries.
// Use with caution - prefer using Store methods for common operations.
func (s *Store) DB() *sql.DB {
	return s.db
}
