package notify

// NotifyConfig holds notification preferences
type NotifyConfig struct {
	// Enabled controls whether notifications are active at all
	Enabled bool `yaml:"enabled"`

	// BreakEarned enables "Break Time!" notifications when break time is earned
	BreakEarned bool `yaml:"break_earned"`

	// BreakEnding enables warnings when 1 minute of break time remaining
	BreakEnding bool `yaml:"break_ending"`

	// PetWarning enables warnings when pet is in overtime danger
	PetWarning bool `yaml:"pet_warning"`

	// StreakMilestone enables celebration notifications at 7, 14, 30, 100 day milestones
	StreakMilestone bool `yaml:"streak_milestone"`

	// LevelUp enables notifications on pet level up
	LevelUp bool `yaml:"level_up"`

	// Sound uses Alert() instead of Notify() for system sound
	Sound bool `yaml:"sound"`

	// MinGapSeconds is the minimum gap between same-type notifications (rate limiting)
	MinGapSeconds int `yaml:"min_gap_seconds"`
}

// DefaultNotifyConfig returns sensible defaults for notification config
func DefaultNotifyConfig() NotifyConfig {
	return NotifyConfig{
		Enabled:         true,
		BreakEarned:     true,
		BreakEnding:     true,
		PetWarning:      true,
		StreakMilestone: true,
		LevelUp:         true,
		Sound:           false,
		MinGapSeconds:   300, // 5 minutes
	}
}
