package graveyard

import (
	"time"
)

// StreakInfo contains streak statistics calculated from death records
type StreakInfo struct {
	// CurrentStreak is the number of consecutive days without a pet death
	CurrentStreak int
	// LongestStreak is the maximum consecutive days without death across all time
	LongestStreak int
	// LastDeathDate is the date of the most recent death (nil if never died)
	LastDeathDate *time.Time
}

// CalculateStreak calculates streak information from death records
// Logic:
// - If no deaths, streak = 1 (first day alive)
// - If deaths exist, current streak = days since last death (0 on death day, 1 next day, etc.)
// - Longest streak tracks max consecutive deathless days across all time
func CalculateStreak(deaths []DeathRecord) StreakInfo {
	now := time.Now().Truncate(24 * time.Hour)

	if len(deaths) == 0 {
		// No deaths ever - streak is 1 (at least first day alive)
		return StreakInfo{
			CurrentStreak: 1,
			LongestStreak: 1,
			LastDeathDate: nil,
		}
	}

	// Sort deaths by timestamp (they should already be sorted, but be safe)
	// Deaths are appended in order, so Records[0] is earliest

	// Calculate current streak: days since last death
	lastDeath := deaths[len(deaths)-1]
	lastDeathDay := lastDeath.Timestamp.Truncate(24 * time.Hour)
	lastDeathCopy := lastDeathDay // Copy for pointer

	// Current streak: days since last death (0 on death day, 1 next day)
	daysSinceLastDeath := int(now.Sub(lastDeathDay).Hours() / 24)
	currentStreak := daysSinceLastDeath

	// Calculate longest streak by examining gaps between deaths
	// A streak is the number of days between two consecutive deaths (or from start to first death)
	longestStreak := currentStreak // Start with current streak as candidate

	// Check gap from "beginning of time" (first day) to first death
	// We assume first death marks the end of the initial streak
	// Initial streak is days from start (we don't track start, so assume minimal contribution)

	// For simplicity: if only one death, longest = max(current, 0)
	// If multiple deaths, check gaps between deaths
	for i := 1; i < len(deaths); i++ {
		prevDeathDay := deaths[i-1].Timestamp.Truncate(24 * time.Hour)
		currDeathDay := deaths[i].Timestamp.Truncate(24 * time.Hour)

		// Gap is days between deaths, minus 1 because death day itself breaks the streak
		// Example: death on day 1, death on day 5 = gap of 3 days (days 2, 3, 4)
		gapDays := int(currDeathDay.Sub(prevDeathDay).Hours()/24) - 1
		if gapDays < 0 {
			gapDays = 0 // Multiple deaths on same day
		}
		if gapDays > longestStreak {
			longestStreak = gapDays
		}
	}

	return StreakInfo{
		CurrentStreak: currentStreak,
		LongestStreak: longestStreak,
		LastDeathDate: &lastDeathCopy,
	}
}
