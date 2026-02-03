package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// DefaultTimeout is the default timeout for IPC operations
const DefaultTimeout = 5 * time.Second

// Client handles IPC communication with the daemon
type Client struct {
	socketPath string
	timeout    time.Duration
}

// NewClient creates a new IPC client
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    DefaultTimeout,
	}
}

// SetTimeout sets the timeout for IPC operations
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Send sends a request to the daemon and returns the response
func (c *Client) Send(req Request) (*Response, error) {
	// Dial Unix socket
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Set deadline for the entire operation
	deadline := time.Now().Add(c.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Write request
	if err := WriteMessage(conn, req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var resp Response
	if err := ReadMessage(conn, &resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &resp, nil
}

// Ping checks if the daemon is responding
func (c *Client) Ping() error {
	req := Request{Type: RequestTypePing}
	resp, err := c.Send(req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("ping failed: %s", resp.Error)
	}
	return nil
}

// Status retrieves the daemon status
func (c *Client) Status() (*StatusResponse, error) {
	req := Request{Type: RequestTypeStatus}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("status request failed: %s", resp.Error)
	}

	var status StatusResponse
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}
	return &status, nil
}

// Stop requests the daemon to shut down
func (c *Client) Stop() error {
	req := Request{Type: RequestTypeStop}
	resp, err := c.Send(req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("stop request failed: %s", resp.Error)
	}
	return nil
}

// GetState retrieves the current productivity state (for TUI)
func (c *Client) GetState() (*StateResponse, error) {
	req := Request{Type: RequestTypeGetState}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("get_state request failed: %s", resp.Error)
	}

	var state StateResponse
	if err := json.Unmarshal(resp.Data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state response: %w", err)
	}
	return &state, nil
}

// SocketPath returns the socket path for this client
func (c *Client) SocketPath() string {
	return c.socketPath
}

// IsConnected checks if the daemon socket exists and is responsive
func (c *Client) IsConnected() bool {
	return c.Ping() == nil
}

// GetActivities retrieves recent activities from the daemon
func (c *Client) GetActivities(limit int) (*GetActivitiesResponse, error) {
	payload, _ := json.Marshal(GetActivitiesRequest{Limit: limit})
	req := Request{
		Type:    RequestTypeGetActivities,
		Payload: payload,
	}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("get_activities request failed: %s", resp.Error)
	}

	var activities GetActivitiesResponse
	if err := json.Unmarshal(resp.Data, &activities); err != nil {
		return nil, fmt.Errorf("failed to parse activities response: %w", err)
	}
	return &activities, nil
}

// GetTaskCount retrieves the count of pending (incomplete) tasks
func (c *Client) GetTaskCount() (int, error) {
	req := Request{Type: RequestTypeGetTaskCount}
	resp, err := c.Send(req)
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf("get_task_count request failed: %s", resp.Error)
	}

	var taskCount GetTaskCountResponse
	if err := json.Unmarshal(resp.Data, &taskCount); err != nil {
		return 0, fmt.Errorf("failed to parse task count response: %w", err)
	}
	return taskCount.Count, nil
}

// AddTask adds a new task to the specified project's TASKS.md
// If projectPath is empty, uses the daemon's watch directory
func (c *Client) AddTask(projectPath, text string) (*AddTaskResponse, error) {
	payload, _ := json.Marshal(AddTaskRequest{
		ProjectPath: projectPath,
		Text:        text,
	})
	req := Request{
		Type:    RequestTypeAddTask,
		Payload: payload,
	}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("add_task request failed: %s", resp.Error)
	}

	var addResp AddTaskResponse
	if err := json.Unmarshal(resp.Data, &addResp); err != nil {
		return nil, fmt.Errorf("failed to parse add task response: %w", err)
	}
	return &addResp, nil
}

// GetDailyStats retrieves daily productivity stats from the daemon.
// If days > 0, returns last N days. If date is specified, returns that date.
// If both empty, returns today's stats.
func (c *Client) GetDailyStats(date string, days int) (*GetDailyStatsResponse, error) {
	payload, _ := json.Marshal(GetDailyStatsRequest{
		Date: date,
		Days: days,
	})
	req := Request{
		Type:    RequestTypeGetDailyStats,
		Payload: payload,
	}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("get_daily_stats request failed: %s", resp.Error)
	}

	var statsResp GetDailyStatsResponse
	if err := json.Unmarshal(resp.Data, &statsResp); err != nil {
		return nil, fmt.Errorf("failed to parse daily stats response: %w", err)
	}
	return &statsResp, nil
}

// GetProductivityStreak retrieves the productivity streak data from the daemon.
func (c *Client) GetProductivityStreak() (*GetProductivityStreakResponse, error) {
	req := Request{Type: RequestTypeGetProductivityStreak}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("get_productivity_streak request failed: %s", resp.Error)
	}

	var streakResp GetProductivityStreakResponse
	if err := json.Unmarshal(resp.Data, &streakResp); err != nil {
		return nil, fmt.Errorf("failed to parse productivity streak response: %w", err)
	}
	return &streakResp, nil
}

// GetWeeklySummary retrieves the weekly summary from the daemon.
// If date is empty, returns current week's summary.
// Date should be any date in the desired week (YYYY-MM-DD format).
func (c *Client) GetWeeklySummary(date string) (*GetWeeklySummaryResponse, error) {
	payload, _ := json.Marshal(GetWeeklySummaryRequest{
		Date: date,
	})
	req := Request{
		Type:    RequestTypeGetWeeklySummary,
		Payload: payload,
	}
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("get_weekly_summary request failed: %s", resp.Error)
	}

	var summaryResp GetWeeklySummaryResponse
	if err := json.Unmarshal(resp.Data, &summaryResp); err != nil {
		return nil, fmt.Errorf("failed to parse weekly summary response: %w", err)
	}
	return &summaryResp, nil
}
