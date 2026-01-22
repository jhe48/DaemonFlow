package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAndClose(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "daemonflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open store
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}

	// Close store
	if err := store.Close(); err != nil {
		t.Errorf("failed to close store: %v", err)
	}
}

func TestSchemaCreation(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "daemonflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open store (triggers migration)
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	db := store.DB()

	// Verify schema_version table
	t.Run("schema_version table", func(t *testing.T) {
		var version int
		err := db.QueryRow("SELECT version FROM schema_version").Scan(&version)
		if err != nil {
			t.Fatalf("failed to query schema_version: %v", err)
		}
		if version != SchemaVersion {
			t.Errorf("schema version mismatch: got %d, want %d", version, SchemaVersion)
		}
	})

	// Verify tasks table structure
	t.Run("tasks table", func(t *testing.T) {
		rows, err := db.Query("PRAGMA table_info(tasks)")
		if err != nil {
			t.Fatalf("failed to get tasks table info: %v", err)
		}
		defer rows.Close()

		columns := make(map[string]string)
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				t.Fatalf("failed to scan column info: %v", err)
			}
			columns[name] = ctype
		}

		expectedColumns := []string{"id", "project_path", "text", "completed", "frequency", "due_date", "created_at", "updated_at"}
		for _, col := range expectedColumns {
			if _, ok := columns[col]; !ok {
				t.Errorf("tasks table missing column: %s", col)
			}
		}
	})

	// Verify activities table structure
	t.Run("activities table", func(t *testing.T) {
		rows, err := db.Query("PRAGMA table_info(activities)")
		if err != nil {
			t.Fatalf("failed to get activities table info: %v", err)
		}
		defer rows.Close()

		columns := make(map[string]string)
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				t.Fatalf("failed to scan column info: %v", err)
			}
			columns[name] = ctype
		}

		expectedColumns := []string{"id", "type", "project_path", "details", "timestamp"}
		for _, col := range expectedColumns {
			if _, ok := columns[col]; !ok {
				t.Errorf("activities table missing column: %s", col)
			}
		}
	})

	// Verify activities index
	t.Run("activities index", func(t *testing.T) {
		rows, err := db.Query("PRAGMA index_list(activities)")
		if err != nil {
			t.Fatalf("failed to get activities index list: %v", err)
		}
		defer rows.Close()

		foundIndex := false
		for rows.Next() {
			var seq int
			var name, unique, origin, partial string
			if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
				t.Fatalf("failed to scan index info: %v", err)
			}
			if name == "idx_activities_project_timestamp" {
				foundIndex = true
				break
			}
		}
		if !foundIndex {
			t.Error("activities table missing index idx_activities_project_timestamp")
		}
	})

	// Verify pet_state table structure
	t.Run("pet_state table", func(t *testing.T) {
		rows, err := db.Query("PRAGMA table_info(pet_state)")
		if err != nil {
			t.Fatalf("failed to get pet_state table info: %v", err)
		}
		defer rows.Close()

		columns := make(map[string]string)
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				t.Fatalf("failed to scan column info: %v", err)
			}
			columns[name] = ctype
		}

		expectedColumns := []string{"id", "current_streak", "longest_streak", "total_deaths", "total_resurrections", "level", "experience", "last_activity_date", "updated_at"}
		for _, col := range expectedColumns {
			if _, ok := columns[col]; !ok {
				t.Errorf("pet_state table missing column: %s", col)
			}
		}
	})

	// Verify pet_state singleton constraint (can only have id=1)
	t.Run("pet_state singleton", func(t *testing.T) {
		// Try to insert a second row with id=2 - should fail
		_, err := db.Exec("INSERT INTO pet_state (id, current_streak, longest_streak, total_deaths, total_resurrections, level, experience, updated_at) VALUES (2, 1, 1, 0, 0, 1, 0, datetime('now'))")
		if err == nil {
			t.Error("pet_state should enforce singleton constraint (id=1 only)")
		}
	})

	// Verify default pet_state row exists
	t.Run("pet_state default row", func(t *testing.T) {
		var id, currentStreak, longestStreak, totalDeaths, totalResurrections, level, experience int
		err := db.QueryRow("SELECT id, current_streak, longest_streak, total_deaths, total_resurrections, level, experience FROM pet_state WHERE id = 1").
			Scan(&id, &currentStreak, &longestStreak, &totalDeaths, &totalResurrections, &level, &experience)
		if err != nil {
			t.Fatalf("failed to query pet_state: %v", err)
		}
		if id != 1 || currentStreak != 1 || longestStreak != 1 || totalDeaths != 0 || totalResurrections != 0 || level != 1 || experience != 0 {
			t.Errorf("pet_state default values incorrect: id=%d, current_streak=%d, longest_streak=%d, total_deaths=%d, total_resurrections=%d, level=%d, experience=%d",
				id, currentStreak, longestStreak, totalDeaths, totalResurrections, level, experience)
		}
	})
}

func TestMigrationIdempotency(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "daemonflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open store first time
	store1, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store first time: %v", err)
	}
	store1.Close()

	// Open store second time (should not error, migration should be idempotent)
	store2, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store second time: %v", err)
	}
	defer store2.Close()

	// Verify schema version is still correct
	version, err := GetSchemaVersion(store2.DB())
	if err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}
	if version != SchemaVersion {
		t.Errorf("schema version mismatch after reopen: got %d, want %d", version, SchemaVersion)
	}
}

func TestWALMode(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "daemonflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open store
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Verify WAL mode is enabled
	var journalMode string
	err = store.DB().QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected journal_mode=wal, got %s", journalMode)
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "daemonflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open store
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Verify foreign keys are enabled
	var foreignKeys int
	err = store.DB().QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("failed to query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("expected foreign_keys=1, got %d", foreignKeys)
	}
}
