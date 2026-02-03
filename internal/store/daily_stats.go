package store

import (
	"database/sql"
	"fmt"
	"time"
)

// DailyStats represents a daily statistics record.
type DailyStats struct {
	ID                int64
	Date              string // YYYY-MM-DD format
	TasksCompleted    int
	TasksCreated      int
	Commits           int
	FilesChanged      int
	XPEarned          int
	TimeEarnedSeconds int
	TimeSpentSeconds  int
	UpdatedAt         time.Time
}

// validStatColumns is a whitelist of columns that can be incremented.
// This prevents SQL injection via the column name parameter.
var validStatColumns = map[string]bool{
	"tasks_completed":     true,
	"tasks_created":       true,
	"commits":             true,
	"files_changed":       true,
	"xp_earned":           true,
	"time_earned_seconds": true,
	"time_spent_seconds":  true,
}

// GetOrCreateDailyStats returns stats for date, creating a row with zeros if not exists.
// Uses INSERT OR IGNORE + SELECT pattern for atomic upsert.
func (s *Store) GetOrCreateDailyStats(date string) (*DailyStats, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Insert or ignore if exists
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO daily_stats (date, tasks_completed, tasks_created, commits, files_changed, xp_earned, time_earned_seconds, time_spent_seconds, updated_at)
		VALUES (?, 0, 0, 0, 0, 0, 0, 0, ?)
	`, date, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create daily stats: %w", err)
	}

	// Now select the row
	return s.GetDailyStats(date)
}

// IncrementDailyStat atomically increments a specific stat column by amount.
// Column name is validated against a whitelist to prevent SQL injection.
// Creates the row for the date if it doesn't exist.
func (s *Store) IncrementDailyStat(date string, column string, amount int) error {
	// Validate column name against whitelist
	if !validStatColumns[column] {
		return fmt.Errorf("invalid stat column: %s", column)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Ensure row exists first
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO daily_stats (date, tasks_completed, tasks_created, commits, files_changed, xp_earned, time_earned_seconds, time_spent_seconds, updated_at)
		VALUES (?, 0, 0, 0, 0, 0, 0, 0, ?)
	`, date, now)
	if err != nil {
		return fmt.Errorf("failed to ensure daily stats row: %w", err)
	}

	// Update the specific column atomically
	// Note: Using fmt.Sprintf for column name is safe because we validated against whitelist
	query := fmt.Sprintf(`
		UPDATE daily_stats SET %s = %s + ?, updated_at = ? WHERE date = ?
	`, column, column)

	_, err = s.db.Exec(query, amount, now, date)
	if err != nil {
		return fmt.Errorf("failed to increment %s: %w", column, err)
	}

	return nil
}

// GetDailyStats gets stats for a specific date.
// Returns nil, nil if not found.
func (s *Store) GetDailyStats(date string) (*DailyStats, error) {
	var stats DailyStats
	var updatedAt string

	err := s.db.QueryRow(`
		SELECT id, date, tasks_completed, tasks_created, commits, files_changed, xp_earned, time_earned_seconds, time_spent_seconds, updated_at
		FROM daily_stats
		WHERE date = ?
	`, date).Scan(
		&stats.ID,
		&stats.Date,
		&stats.TasksCompleted,
		&stats.TasksCreated,
		&stats.Commits,
		&stats.FilesChanged,
		&stats.XPEarned,
		&stats.TimeEarnedSeconds,
		&stats.TimeSpentSeconds,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	stats.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &stats, nil
}

// GetDailyStatsRange gets stats for a date range (inclusive), ordered by date DESC.
func (s *Store) GetDailyStatsRange(startDate, endDate string) ([]DailyStats, error) {
	rows, err := s.db.Query(`
		SELECT id, date, tasks_completed, tasks_created, commits, files_changed, xp_earned, time_earned_seconds, time_spent_seconds, updated_at
		FROM daily_stats
		WHERE date >= ? AND date <= ?
		ORDER BY date DESC
	`, startDate, endDate)

	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats range: %w", err)
	}
	defer rows.Close()

	return scanDailyStats(rows)
}

// GetRecentDailyStats gets the last N days of stats, ordered by date DESC.
func (s *Store) GetRecentDailyStats(days int) ([]DailyStats, error) {
	rows, err := s.db.Query(`
		SELECT id, date, tasks_completed, tasks_created, commits, files_changed, xp_earned, time_earned_seconds, time_spent_seconds, updated_at
		FROM daily_stats
		ORDER BY date DESC
		LIMIT ?
	`, days)

	if err != nil {
		return nil, fmt.Errorf("failed to get recent daily stats: %w", err)
	}
	defer rows.Close()

	return scanDailyStats(rows)
}

// scanDailyStats scans rows into a slice of DailyStats.
func scanDailyStats(rows *sql.Rows) ([]DailyStats, error) {
	var stats []DailyStats

	for rows.Next() {
		var s DailyStats
		var updatedAt string

		err := rows.Scan(
			&s.ID,
			&s.Date,
			&s.TasksCompleted,
			&s.TasksCreated,
			&s.Commits,
			&s.FilesChanged,
			&s.XPEarned,
			&s.TimeEarnedSeconds,
			&s.TimeSpentSeconds,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}

		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily stats rows: %w", err)
	}

	return stats, nil
}
