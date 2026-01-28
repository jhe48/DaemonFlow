package store

import (
	"database/sql"
	"time"
)

// PetState represents the pet's evolution state from the database.
type PetState struct {
	ID                  int64
	CurrentStreak       int
	LongestStreak       int
	TotalDeaths         int
	TotalResurrections  int
	Level               int
	Experience          int
	LastActivityDate    *time.Time
	UpdatedAt           time.Time
}

// LevelInfo contains level and XP progression information.
type LevelInfo struct {
	Level        int // Current level (1-based)
	Experience   int // Total accumulated XP
	XPForCurrent int // XP threshold for current level
	XPForNext    int // XP threshold for next level
	XPProgress   int // XP earned toward next level
}

// XP thresholds for each level (cumulative XP needed to reach that level)
// Index is level, value is cumulative XP needed.
// Level 1: 0 XP (starting level)
// Level 2: 100 XP
// Level 3: 300 XP
// Level 4: 600 XP
// Level 5: 1000 XP
// Level 6+: 500 XP per level beyond 5
var levelThresholds = []int{0, 0, 100, 300, 600, 1000, 1500, 2000, 2500, 3000, 3500}

// GetPetState retrieves the pet state from the database (singleton row).
func (s *Store) GetPetState() (*PetState, error) {
	var state PetState
	var lastActivityDate sql.NullString
	var updatedAt string

	err := s.db.QueryRow(`
		SELECT id, current_streak, longest_streak, total_deaths, total_resurrections,
		       level, experience, last_activity_date, updated_at
		FROM pet_state
		WHERE id = 1
	`).Scan(
		&state.ID,
		&state.CurrentStreak,
		&state.LongestStreak,
		&state.TotalDeaths,
		&state.TotalResurrections,
		&state.Level,
		&state.Experience,
		&lastActivityDate,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse nullable last_activity_date
	if lastActivityDate.Valid {
		t, err := time.Parse(time.RFC3339, lastActivityDate.String)
		if err == nil {
			state.LastActivityDate = &t
		}
	}

	// Parse updated_at
	state.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &state, nil
}

// UpdatePetState updates the pet state in the database.
func (s *Store) UpdatePetState(state *PetState) error {
	now := time.Now().UTC().Format(time.RFC3339)

	var lastActivityDate sql.NullString
	if state.LastActivityDate != nil {
		lastActivityDate = sql.NullString{
			String: state.LastActivityDate.UTC().Format(time.RFC3339),
			Valid:  true,
		}
	}

	_, err := s.db.Exec(`
		UPDATE pet_state SET
			current_streak = ?,
			longest_streak = ?,
			total_deaths = ?,
			total_resurrections = ?,
			level = ?,
			experience = ?,
			last_activity_date = ?,
			updated_at = ?
		WHERE id = 1
	`, state.CurrentStreak, state.LongestStreak, state.TotalDeaths, state.TotalResurrections,
		state.Level, state.Experience, lastActivityDate, now)

	return err
}

// AddExperience atomically adds XP to the pet and updates level if needed.
// Returns the new state after XP addition.
func (s *Store) AddExperience(amount int) (*PetState, error) {
	// Read current state
	state, err := s.GetPetState()
	if err != nil {
		return nil, err
	}

	// Add XP
	state.Experience += amount

	// Recalculate level from new XP total
	state.Level = LevelFromExperience(state.Experience)

	// Update last activity date
	now := time.Now().UTC()
	state.LastActivityDate = &now

	// Save updated state
	if err := s.UpdatePetState(state); err != nil {
		return nil, err
	}

	return state, nil
}

// SyncGraveyardStats updates the pet state with death/resurrection counts and streak info.
// Called by daemon to sync graveyard data to SQLite.
func (s *Store) SyncGraveyardStats(currentStreak, longestStreak, totalDeaths, totalResurrections int) error {
	state, err := s.GetPetState()
	if err != nil {
		return err
	}

	state.CurrentStreak = currentStreak
	state.LongestStreak = longestStreak
	state.TotalDeaths = totalDeaths
	state.TotalResurrections = totalResurrections

	return s.UpdatePetState(state)
}

// LevelFromExperience calculates the level from total experience points.
func LevelFromExperience(experience int) int {
	if experience < 0 {
		return 1
	}

	level := 1
	for i := 1; i < len(levelThresholds); i++ {
		if experience >= levelThresholds[i] {
			level = i
		} else {
			break
		}
	}

	// Handle levels beyond predefined thresholds (500 XP per level after level 10)
	if level >= len(levelThresholds)-1 && experience >= levelThresholds[len(levelThresholds)-1] {
		extraXP := experience - levelThresholds[len(levelThresholds)-1]
		extraLevels := extraXP / 500
		level = len(levelThresholds) - 1 + extraLevels
	}

	return level
}

// GetLevelInfo returns detailed level and XP progression information.
func GetLevelInfo(experience int) LevelInfo {
	level := LevelFromExperience(experience)

	xpForCurrent := getThresholdForLevel(level)
	xpForNext := getThresholdForLevel(level + 1)
	xpProgress := experience - xpForCurrent

	return LevelInfo{
		Level:        level,
		Experience:   experience,
		XPForCurrent: xpForCurrent,
		XPForNext:    xpForNext,
		XPProgress:   xpProgress,
	}
}

// getThresholdForLevel returns the XP threshold to reach a given level.
func getThresholdForLevel(level int) int {
	if level <= 0 {
		return 0
	}
	if level < len(levelThresholds) {
		return levelThresholds[level]
	}
	// Beyond predefined levels: 500 XP per level after level 10
	// Level 11 = 3500 + 500 = 4000, Level 12 = 4500, etc.
	return levelThresholds[len(levelThresholds)-1] + (level-(len(levelThresholds)-1))*500
}

// ExperienceToNextLevel returns the XP needed to reach the next level from current experience.
func ExperienceToNextLevel(experience int) int {
	currentLevel := LevelFromExperience(experience)
	nextLevelXP := getThresholdForLevel(currentLevel + 1)
	return nextLevelXP - experience
}
