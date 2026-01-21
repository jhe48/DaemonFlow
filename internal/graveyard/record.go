package graveyard

import (
	"fmt"
	"time"
)

// DeathRecord represents a single pet death event
type DeathRecord struct {
	// Timestamp is when the death occurred
	Timestamp time.Time

	// OvertimeSeconds is how long the pet was in overtime before death (positive value)
	OvertimeSeconds int

	// SessionEarned is the total seconds earned during the session before death
	SessionEarned int

	// Cause is the reason for death (always "overtime" for now)
	Cause string
}

// String returns a human-readable format of the death record
func (r DeathRecord) String() string {
	return fmt.Sprintf("Pet died on %s at %s after %s of overtime (%s)",
		r.Timestamp.Format("2006-01-02"),
		r.Timestamp.Format("15:04"),
		formatDuration(r.OvertimeSeconds),
		r.Cause,
	)
}

// ToMarkdownRow returns the death record as a markdown table row
func (r DeathRecord) ToMarkdownRow() string {
	return fmt.Sprintf("| %s | %s | %s | %s |",
		r.Timestamp.Format("2006-01-02"),
		r.Timestamp.Format("15:04"),
		formatDuration(r.OvertimeSeconds),
		r.Cause,
	)
}

// formatDuration converts seconds to a human-readable duration string
func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	secs := seconds % 60
	if secs == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, secs)
}
