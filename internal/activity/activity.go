package activity

import "time"

// ActivityType represents the type of development activity
type ActivityType string

const (
	// GitCommit indicates a new git commit was made
	GitCommit ActivityType = "git_commit"
	// GitBranchSwitch indicates the git branch was changed
	GitBranchSwitch ActivityType = "git_branch_switch"
	// GitStage indicates files were staged for commit
	GitStage ActivityType = "git_stage"
	// FileChange indicates files were modified (future use)
	FileChange ActivityType = "file_change"
	// TaskComplete indicates a task was completed (future use)
	TaskComplete ActivityType = "task_complete"
)

// Activity represents a development activity event
type Activity struct {
	// Type indicates what kind of activity this is
	Type ActivityType `json:"type"`

	// Timestamp when the activity was detected
	Timestamp time.Time `json:"timestamp"`

	// Details contains type-specific information about the activity
	// For GitCommit: "hash", "message"
	// For GitBranchSwitch: "old_branch", "new_branch"
	Details map[string]string `json:"details"`
}

// NewActivity creates a new Activity with the current timestamp
func NewActivity(actType ActivityType, details map[string]string) Activity {
	return Activity{
		Type:      actType,
		Timestamp: time.Now(),
		Details:   details,
	}
}

// ActivityListener is an interface for components that want to receive activity events
type ActivityListener interface {
	// OnActivity is called when a new activity is detected
	OnActivity(activity Activity)
}
