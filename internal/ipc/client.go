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
