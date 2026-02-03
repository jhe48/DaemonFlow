package store

import (
	"time"
)

// ProductivityStreak contains streak info based on daily activity.
// A "productive day" is defined as a day with at least one of:
// - tasks_completed > 0
// - commits > 0
// - xp_earned > 0
type ProductivityStreak struct {
	CurrentStreak   int    // Consecutive productive days ending today (or yesterday if today not productive yet)
	LongestStreak   int    // Longest productivity streak ever
	LastActiveDate  string // Last date with productivity (YYYY-MM-DD)
	TotalActiveDays int    // Total days with any activity
}

// isProductiveDay returns true if the given stats indicate productivity.
func isProductiveDay(stats *DailyStats) bool {
	if stats == nil {
		return false
	}
	return stats.TasksCompleted > 0 || stats.Commits > 0 || stats.XPEarned > 0
}

// GetProductivityStreak calculates streak from daily_stats.
// It queries recent daily stats and walks backwards from today counting consecutive productive days.
func (s *Store) GetProductivityStreak() (*ProductivityStreak, error) {
	// Get all daily stats, ordered by date DESC
	// Use a large limit to get full history for calculating longest streak
	allStats, err := s.GetRecentDailyStats(365) // Last year of data
	if err != nil {
		return nil, err
	}

	result := &ProductivityStreak{
		CurrentStreak:   0,
		LongestStreak:   0,
		LastActiveDate:  "",
		TotalActiveDays: 0,
	}

	if len(allStats) == 0 {
		return result, nil
	}

	// Build a map of date -> productive status for efficient lookup
	productiveMap := make(map[string]bool)
	for _, stat := range allStats {
		productive := isProductiveDay(&stat)
		productiveMap[stat.Date] = productive
		if productive {
			result.TotalActiveDays++
			if result.LastActiveDate == "" || stat.Date > result.LastActiveDate {
				result.LastActiveDate = stat.Date
			}
		}
	}

	// Calculate current streak
	// Start from today and walk backwards
	today := time.Now().UTC().Format("2006-01-02")
	currentDate := today

	// If today is not productive, check if yesterday was (streak might still be active)
	todayProductive := productiveMap[today]
	if !todayProductive {
		// Check yesterday - if productive, count from yesterday
		yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
		if productiveMap[yesterday] {
			currentDate = yesterday
		} else {
			// No current streak
			result.CurrentStreak = 0
		}
	}

	// Walk backwards counting consecutive productive days
	if currentDate == today || productiveMap[currentDate] {
		date, _ := time.Parse("2006-01-02", currentDate)
		for {
			dateStr := date.Format("2006-01-02")
			if productiveMap[dateStr] {
				result.CurrentStreak++
				date = date.AddDate(0, 0, -1)
			} else {
				// Check if we have data for this date - if not, might be a gap
				// but if we do have data and it's not productive, streak ends
				break
			}
		}
	}

	// Calculate longest streak by examining all dates
	// Sort dates and find longest consecutive productive run
	result.LongestStreak = calculateLongestStreak(allStats)

	// Ensure current streak doesn't exceed longest
	if result.CurrentStreak > result.LongestStreak {
		result.LongestStreak = result.CurrentStreak
	}

	return result, nil
}

// calculateLongestStreak finds the longest run of consecutive productive days.
func calculateLongestStreak(stats []DailyStats) int {
	if len(stats) == 0 {
		return 0
	}

	// Create a map of productive dates
	productiveDates := make(map[string]bool)
	var minDate, maxDate string
	for _, stat := range stats {
		if isProductiveDay(&stat) {
			productiveDates[stat.Date] = true
			if minDate == "" || stat.Date < minDate {
				minDate = stat.Date
			}
			if maxDate == "" || stat.Date > maxDate {
				maxDate = stat.Date
			}
		}
	}

	if len(productiveDates) == 0 {
		return 0
	}

	// Walk through date range and find longest consecutive run
	longest := 0
	current := 0

	startDate, _ := time.Parse("2006-01-02", minDate)
	endDate, _ := time.Parse("2006-01-02", maxDate)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if productiveDates[dateStr] {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 0
		}
	}

	return longest
}
