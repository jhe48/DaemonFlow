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

// StateResponse contains productivity state (placeholder for future phases)
type StateResponse struct {
	// Placeholder for clock/activity data from Phase 5
	EarnedToday    float64 `json:"earned_today"`
	CurrentSession float64 `json:"current_session"`
	ActiveTask     string  `json:"active_task,omitempty"`
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
