package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetPetState_DefaultValues(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "daemonflow-pet-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open store (triggers migration with default pet_state)
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Get default pet state
	state, err := s.GetPetState()
	if err != nil {
		t.Fatalf("GetPetState failed: %v", err)
	}

	// Verify default values
	if state.ID != 1 {
		t.Errorf("expected ID=1, got %d", state.ID)
	}
	if state.CurrentStreak != 1 {
		t.Errorf("expected CurrentStreak=1, got %d", state.CurrentStreak)
	}
	if state.LongestStreak != 1 {
		t.Errorf("expected LongestStreak=1, got %d", state.LongestStreak)
	}
	if state.TotalDeaths != 0 {
		t.Errorf("expected TotalDeaths=0, got %d", state.TotalDeaths)
	}
	if state.TotalResurrections != 0 {
		t.Errorf("expected TotalResurrections=0, got %d", state.TotalResurrections)
	}
	if state.Level != 1 {
		t.Errorf("expected Level=1, got %d", state.Level)
	}
	if state.Experience != 0 {
		t.Errorf("expected Experience=0, got %d", state.Experience)
	}
	if state.LastActivityDate != nil {
		t.Errorf("expected LastActivityDate=nil, got %v", state.LastActivityDate)
	}
}

func TestUpdatePetState_PersistsChanges(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemonflow-pet-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Get current state
	state, err := s.GetPetState()
	if err != nil {
		t.Fatalf("GetPetState failed: %v", err)
	}

	// Modify state
	now := time.Now().UTC()
	state.CurrentStreak = 5
	state.LongestStreak = 10
	state.TotalDeaths = 3
	state.TotalResurrections = 2
	state.Level = 4
	state.Experience = 750
	state.LastActivityDate = &now

	// Update
	if err := s.UpdatePetState(state); err != nil {
		t.Fatalf("UpdatePetState failed: %v", err)
	}

	// Read back
	updated, err := s.GetPetState()
	if err != nil {
		t.Fatalf("GetPetState after update failed: %v", err)
	}

	// Verify all fields persisted
	if updated.CurrentStreak != 5 {
		t.Errorf("expected CurrentStreak=5, got %d", updated.CurrentStreak)
	}
	if updated.LongestStreak != 10 {
		t.Errorf("expected LongestStreak=10, got %d", updated.LongestStreak)
	}
	if updated.TotalDeaths != 3 {
		t.Errorf("expected TotalDeaths=3, got %d", updated.TotalDeaths)
	}
	if updated.TotalResurrections != 2 {
		t.Errorf("expected TotalResurrections=2, got %d", updated.TotalResurrections)
	}
	if updated.Level != 4 {
		t.Errorf("expected Level=4, got %d", updated.Level)
	}
	if updated.Experience != 750 {
		t.Errorf("expected Experience=750, got %d", updated.Experience)
	}
	if updated.LastActivityDate == nil {
		t.Error("expected LastActivityDate to be set, got nil")
	}
}

func TestAddExperience_IncrementsCorrectly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemonflow-pet-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Start with 0 XP
	initial, err := s.GetPetState()
	if err != nil {
		t.Fatalf("GetPetState failed: %v", err)
	}
	if initial.Experience != 0 {
		t.Fatalf("expected initial experience=0, got %d", initial.Experience)
	}

	// Add 50 XP
	state, err := s.AddExperience(50)
	if err != nil {
		t.Fatalf("AddExperience(50) failed: %v", err)
	}
	if state.Experience != 50 {
		t.Errorf("expected experience=50, got %d", state.Experience)
	}
	if state.Level != 1 {
		t.Errorf("expected level=1 at 50 XP, got %d", state.Level)
	}

	// Add another 60 XP (total 110)
	state, err = s.AddExperience(60)
	if err != nil {
		t.Fatalf("AddExperience(60) failed: %v", err)
	}
	if state.Experience != 110 {
		t.Errorf("expected experience=110, got %d", state.Experience)
	}
	if state.Level != 2 {
		t.Errorf("expected level=2 at 110 XP, got %d", state.Level)
	}

	// Verify last activity date was set
	if state.LastActivityDate == nil {
		t.Error("expected LastActivityDate to be set after AddExperience")
	}
}

func TestAddExperience_LevelUp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemonflow-pet-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Add enough XP to reach level 3 (300 XP threshold)
	state, err := s.AddExperience(300)
	if err != nil {
		t.Fatalf("AddExperience(300) failed: %v", err)
	}
	if state.Experience != 300 {
		t.Errorf("expected experience=300, got %d", state.Experience)
	}
	if state.Level != 3 {
		t.Errorf("expected level=3 at 300 XP, got %d", state.Level)
	}
}

func TestLevelFromExperience_Thresholds(t *testing.T) {
	testCases := []struct {
		xp            int
		expectedLevel int
	}{
		{0, 1},    // Level 1: 0 XP
		{50, 1},   // Still level 1
		{99, 1},   // Still level 1
		{100, 2},  // Level 2: 100 XP
		{150, 2},  // Still level 2
		{299, 2},  // Still level 2
		{300, 3},  // Level 3: 300 XP
		{599, 3},  // Still level 3
		{600, 4},  // Level 4: 600 XP
		{999, 4},  // Still level 4
		{1000, 5}, // Level 5: 1000 XP
		{1499, 5}, // Still level 5
		{1500, 6}, // Level 6: 1500 XP
		{1999, 6}, // Still level 6
		{2000, 7}, // Level 7: 2000 XP
		{2500, 8}, // Level 8: 2500 XP
		{3000, 9}, // Level 9: 3000 XP
		{3500, 10}, // Level 10: 3500 XP
		{3999, 10}, // Still level 10
		{4000, 11}, // Level 11: 4000 XP (500 XP per level beyond 10)
		{4500, 12}, // Level 12: 4500 XP
		{5000, 13}, // Level 13: 5000 XP
	}

	for _, tc := range testCases {
		result := LevelFromExperience(tc.xp)
		if result != tc.expectedLevel {
			t.Errorf("LevelFromExperience(%d): expected level %d, got %d",
				tc.xp, tc.expectedLevel, result)
		}
	}
}

func TestGetLevelInfo(t *testing.T) {
	// Test XP progress calculation for level 2 (at 150 XP)
	info := GetLevelInfo(150)
	if info.Level != 2 {
		t.Errorf("expected level=2, got %d", info.Level)
	}
	if info.Experience != 150 {
		t.Errorf("expected experience=150, got %d", info.Experience)
	}
	if info.XPForCurrent != 100 {
		t.Errorf("expected XPForCurrent=100, got %d", info.XPForCurrent)
	}
	if info.XPForNext != 300 {
		t.Errorf("expected XPForNext=300, got %d", info.XPForNext)
	}
	if info.XPProgress != 50 {
		t.Errorf("expected XPProgress=50, got %d", info.XPProgress)
	}
}

func TestSyncGraveyardStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "daemonflow-pet-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Sync graveyard stats
	err = s.SyncGraveyardStats(7, 14, 5, 4)
	if err != nil {
		t.Fatalf("SyncGraveyardStats failed: %v", err)
	}

	// Verify stats were synced
	state, err := s.GetPetState()
	if err != nil {
		t.Fatalf("GetPetState failed: %v", err)
	}

	if state.CurrentStreak != 7 {
		t.Errorf("expected CurrentStreak=7, got %d", state.CurrentStreak)
	}
	if state.LongestStreak != 14 {
		t.Errorf("expected LongestStreak=14, got %d", state.LongestStreak)
	}
	if state.TotalDeaths != 5 {
		t.Errorf("expected TotalDeaths=5, got %d", state.TotalDeaths)
	}
	if state.TotalResurrections != 4 {
		t.Errorf("expected TotalResurrections=4, got %d", state.TotalResurrections)
	}
}
