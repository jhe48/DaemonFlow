package clock

import (
	"github.com/jackyhe0402/daemonflow/internal/activity"
	"github.com/jackyhe0402/daemonflow/internal/config"
)

// EarningCalculator calculates earned break time from activity events
type EarningCalculator struct {
	config *config.EarningConfig
}

// NewEarningCalculator creates a new EarningCalculator with the given config
func NewEarningCalculator(cfg *config.EarningConfig) *EarningCalculator {
	return &EarningCalculator{
		config: cfg,
	}
}

// CalculateEarned returns seconds earned for an activity
func (e *EarningCalculator) CalculateEarned(act activity.Activity) int {
	return e.GetWeight(act.Type)
}

// GetWeight returns configured seconds for an activity type
func (e *EarningCalculator) GetWeight(actType activity.ActivityType) int {
	switch actType {
	case activity.GitCommit:
		return e.config.CommitSeconds
	case activity.GitStage:
		return e.config.StageSeconds
	case activity.FileChange:
		return e.config.FileChangeSeconds
	case activity.TaskComplete:
		return e.config.TaskCompleteSeconds
	default:
		// Unknown types (like GitBranchSwitch) return 0
		return 0
	}
}
