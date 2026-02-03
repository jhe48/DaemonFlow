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

// DailySummary contains aggregated stats for a single day with computed fields.
type DailySummary struct {
	Date            string
	TasksCompleted  int
	TasksCreated    int
	Commits         int
	FilesChanged    int
	XPEarned        int
	TimeEarnedMins  int  // time_earned_seconds / 60
	TimeSpentMins   int  // time_spent_seconds / 60
	NetTimeMins     int  // earned - spent
	IsProductive    bool // Had meaningful activity
}

// WeeklySummary contains aggregated stats for a week.
type WeeklySummary struct {
	WeekStart           string         // Monday of the week (YYYY-MM-DD)
	WeekEnd             string         // Sunday of the week
	TotalTasksCompleted int
	TotalCommits        int
	TotalXPEarned       int
	TotalTimeEarnedMins int
	TotalTimeSpentMins  int
	ProductiveDays      int            // Days with IsProductive=true
	DailyBreakdown      []DailySummary
}

// dailyStatsToDailySummary converts a DailyStats to a DailySummary with computed fields.
func dailyStatsToDailySummary(stats *DailyStats) DailySummary {
	if stats == nil {
		return DailySummary{}
	}
	timeEarnedMins := stats.TimeEarnedSeconds / 60
	timeSpentMins := stats.TimeSpentSeconds / 60
	isProductive := stats.TasksCompleted > 0 || stats.Commits > 0 || stats.XPEarned > 0

	return DailySummary{
		Date:           stats.Date,
		TasksCompleted: stats.TasksCompleted,
		TasksCreated:   stats.TasksCreated,
		Commits:        stats.Commits,
		FilesChanged:   stats.FilesChanged,
		XPEarned:       stats.XPEarned,
		TimeEarnedMins: timeEarnedMins,
		TimeSpentMins:  timeSpentMins,
		NetTimeMins:    timeEarnedMins - timeSpentMins,
		IsProductive:   isProductive,
	}
}

// GetDailySummary returns a summary for a specific date.
func (s *Store) GetDailySummary(date string) (*DailySummary, error) {
	stats, err := s.GetDailyStats(date)
	if err != nil {
		return nil, err
	}

	if stats == nil {
		// Return empty summary for date with no data
		return &DailySummary{
			Date:         date,
			IsProductive: false,
		}, nil
	}

	summary := dailyStatsToDailySummary(stats)
	return &summary, nil
}

// GetWeeklySummary returns a summary for the week containing the given date.
// Week starts on Monday. If date is empty, uses current week.
func (s *Store) GetWeeklySummary(date string) (*WeeklySummary, error) {
	var targetDate time.Time
	var err error

	if date == "" {
		targetDate = time.Now().UTC()
	} else {
		targetDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	// Calculate week boundaries (Monday to Sunday)
	weekday := targetDate.Weekday()
	// Go's time.Weekday: Sunday=0, Monday=1, ..., Saturday=6
	// We want Monday as start of week
	var daysToMonday int
	if weekday == time.Sunday {
		daysToMonday = 6 // Sunday is 6 days after Monday
	} else {
		daysToMonday = int(weekday) - 1 // Monday=0, Tuesday=1, etc.
	}

	monday := targetDate.AddDate(0, 0, -daysToMonday)
	sunday := monday.AddDate(0, 0, 6)

	weekStart := monday.Format("2006-01-02")
	weekEnd := sunday.Format("2006-01-02")

	// Query daily_stats for this range (ordered ASC for chronological breakdown)
	rows, err := s.db.Query(`
		SELECT id, date, tasks_completed, tasks_created, commits, files_changed, xp_earned, time_earned_seconds, time_spent_seconds, updated_at
		FROM daily_stats
		WHERE date >= ? AND date <= ?
		ORDER BY date ASC
	`, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly stats: %w", err)
	}
	defer rows.Close()

	stats, err := scanDailyStats(rows)
	if err != nil {
		return nil, err
	}

	// Build daily breakdown and aggregate totals
	// Create summaries for all 7 days of the week (fill gaps with empty days)
	dailyBreakdown := make([]DailySummary, 0, 7)
	statsMap := make(map[string]DailyStats)
	for _, stat := range stats {
		statsMap[stat.Date] = stat
	}

	summary := &WeeklySummary{
		WeekStart:      weekStart,
		WeekEnd:        weekEnd,
		DailyBreakdown: dailyBreakdown,
	}

	// Walk through each day of the week
	for d := monday; !d.After(sunday); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		var daySummary DailySummary

		if stat, exists := statsMap[dateStr]; exists {
			daySummary = dailyStatsToDailySummary(&stat)
			// Aggregate totals
			summary.TotalTasksCompleted += stat.TasksCompleted
			summary.TotalCommits += stat.Commits
			summary.TotalXPEarned += stat.XPEarned
			summary.TotalTimeEarnedMins += stat.TimeEarnedSeconds / 60
			summary.TotalTimeSpentMins += stat.TimeSpentSeconds / 60
			if daySummary.IsProductive {
				summary.ProductiveDays++
			}
		} else {
			daySummary = DailySummary{
				Date:         dateStr,
				IsProductive: false,
			}
		}

		summary.DailyBreakdown = append(summary.DailyBreakdown, daySummary)
	}

	return summary, nil
}
