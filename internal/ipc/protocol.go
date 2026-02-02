package ipc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// Request types for IPC communication
const (
	RequestTypePing          = "ping"           // Health check
	RequestTypeStatus        = "status"         // Get daemon status
	RequestTypeStop          = "stop"           // Request graceful shutdown
	RequestTypeGetState      = "get_state"      // Get current productivity state (for TUI)
	RequestTypeGetActivities = "get_activities" // Get recent activity events
	RequestTypeStartBreak    = "start_break"    // Start break mode
	RequestTypeEndBreak      = "end_break"      // End break mode
	RequestTypeResurrect     = "resurrect"      // Resurrect the pet after death
	RequestTypeGetTasks      = "get_tasks"      // Get global task list
	RequestTypeGetTaskCount  = "get_task_count" // Get pending task count
	RequestTypeAddTask       = "add_task"       // Add a new task
)

// Request represents an IPC request message
type Request struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Response represents an IPC response message
type Response struct {
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// StatusResponse contains daemon status information
type StatusResponse struct {
	Running  bool   `json:"running"`
	Uptime   int64  `json:"uptime"`   // Seconds since daemon started
	WatchDir string `json:"watch_dir"`
}

// StateResponse contains productivity state for TUI
type StateResponse struct {
	ClockState       string `json:"clock_state"`        // "working", "break", "overtime"
	EarnedSeconds    int    `json:"earned_seconds"`     // Current balance (can be negative)
	SessionEarned    int    `json:"session_earned"`     // Earned this session
	CurrentStreak    int    `json:"current_streak"`     // Days without death
	LongestStreak    int    `json:"longest_streak"`     // Best streak ever
	TotalDeaths      int    `json:"total_deaths"`       // Total deaths
	Level            int    `json:"level"`              // Current pet level (1-5+)
	Experience       int    `json:"experience"`         // Current XP
	ExperienceToNext int    `json:"experience_to_next"` // XP needed for next level
}

// ClockEventResponse contains the result of a break state transition
type ClockEventResponse struct {
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
}

// ResurrectResponse contains the result of a resurrection attempt
type ResurrectResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	NewState string `json:"new_state"`
}

// GetActivitiesRequest contains parameters for getting activities
type GetActivitiesRequest struct {
	Limit int `json:"limit"` // Max number of activities to return (default 10)
}

// ActivityData represents an activity for IPC transfer
type ActivityData struct {
	Type      string            `json:"type"`
	Timestamp string            `json:"timestamp"` // ISO 8601 format
	Details   map[string]string `json:"details"`
}

// GetActivitiesResponse contains recent activities
type GetActivitiesResponse struct {
	Activities []ActivityData `json:"activities"`
}

// GetTasksRequest contains parameters for getting tasks
type GetTasksRequest struct {
	Limit            int    `json:"limit"`              // Max tasks to return (default 50)
	IncludeCompleted bool   `json:"include_completed"`  // Include completed tasks (default false)
	ProjectPath      string `json:"project_path"`       // Filter by project (optional, empty = all)
}

// TaskData represents a task for IPC transfer
type TaskData struct {
	ID            int64  `json:"id"`
	ProjectPath   string `json:"project_path"`
	ProjectName   string `json:"project_name"`
	Text          string `json:"text"`
	Completed     bool   `json:"completed"`
	DueDate       string `json:"due_date,omitempty"`   // ISO 8601 or empty
	PriorityScore int    `json:"priority_score"`
	CreatedAt     string `json:"created_at"`           // ISO 8601
}

// GetTasksResponse contains the task list
type GetTasksResponse struct {
	Tasks []TaskData `json:"tasks"`
	Total int        `json:"total"` // Total count (may be more than returned if limit applied)
}

// GetTaskCountResponse contains the pending task count
type GetTaskCountResponse struct {
	Count int `json:"count"` // Number of incomplete tasks
}

// AddTaskRequest contains parameters for adding a new task
type AddTaskRequest struct {
	ProjectPath string `json:"project_path"` // Project directory (empty = use daemon's watch dir)
	Text        string `json:"text"`         // Task text content
}

// AddTaskResponse contains the result of adding a task
type AddTaskResponse struct {
	Success bool   `json:"success"` // Whether task was added successfully
	Message string `json:"message"` // Success or error message
}

// WriteMessage writes a length-prefixed JSON message to the writer
// Format: 4-byte big-endian length prefix + JSON payload
func WriteMessage(w io.Writer, msg interface{}) error {
	// Marshal to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write 4-byte length prefix (big-endian)
	length := uint32(len(data))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	// Write JSON payload
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// ReadMessage reads a length-prefixed JSON message from the reader
func ReadMessage(r io.Reader, msg interface{}) error {
	// Read 4-byte length prefix (big-endian)
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Sanity check on message size (max 1MB)
	if length > 1024*1024 {
		return fmt.Errorf("message too large: %d bytes", length)
	}

	// Read JSON payload
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

// NewSuccessResponse creates a successful response with optional data
func NewSuccessResponse(data interface{}) (*Response, error) {
	resp := &Response{Success: true}

	if data != nil {
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response data: %w", err)
		}
		resp.Data = dataJSON
	}

	return resp, nil
}

// NewErrorResponse creates an error response
func NewErrorResponse(errMsg string) *Response {
	return &Response{
		Success: false,
		Error:   errMsg,
	}
}
