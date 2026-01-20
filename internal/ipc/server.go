package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

// DaemonInterface defines methods the IPC server can call on the daemon
type DaemonInterface interface {
	GetUptime() int64
	GetWatchDir() string
	RequestShutdown()
	GetRecentActivitiesData(limit int) []ActivityData
	GetClockState() string
	GetEarnedSeconds() int
	GetSessionEarned() int
	StartBreak() (previousState, newState string)
	EndBreak() (previousState, newState string)
}

// Server handles IPC connections from clients
type Server struct {
	socketPath string
	listener   net.Listener
	daemon     DaemonInterface
	running    bool
	mu         sync.Mutex
	wg         sync.WaitGroup
}

// NewServer creates a new IPC server
func NewServer(socketPath string, daemon DaemonInterface) *Server {
	return &Server{
		socketPath: socketPath,
		daemon:     daemon,
	}
}

// Start starts the IPC server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing socket file if present (stale from crash)
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket listener: %w", err)
	}

	s.listener = listener
	s.running = true

	// Accept connections in goroutine
	s.wg.Add(1)
	go s.acceptLoop()

	log.Printf("IPC server listening on %s", s.socketPath)
	return nil
}

// Stop stops the IPC server and cleans up the socket file
func (s *Server) Stop() error {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for accept loop to finish
	s.wg.Wait()

	// Clean up socket file
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	log.Printf("IPC server stopped")
	return nil
}

// acceptLoop accepts incoming connections
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()

			if !running {
				return // Server is stopping
			}
			log.Printf("IPC accept error: %v", err)
			continue
		}

		// Handle each connection in a goroutine
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

// handleConnection handles a single IPC connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read request
	var req Request
	if err := ReadMessage(conn, &req); err != nil {
		log.Printf("IPC read error: %v", err)
		return
	}

	// Route to handler based on type
	var resp *Response
	var err error

	switch req.Type {
	case RequestTypePing:
		resp, err = s.handlePing()
	case RequestTypeStatus:
		resp, err = s.handleStatus()
	case RequestTypeStop:
		resp, err = s.handleStop()
	case RequestTypeGetState:
		resp, err = s.handleGetState()
	case RequestTypeGetActivities:
		resp, err = s.handleGetActivities(req)
	default:
		resp = NewErrorResponse(fmt.Sprintf("unknown request type: %s", req.Type))
	}

	if err != nil {
		resp = NewErrorResponse(err.Error())
	}

	// Send response
	if err := WriteMessage(conn, resp); err != nil {
		log.Printf("IPC write error: %v", err)
	}
}

// handlePing handles ping requests (health check)
func (s *Server) handlePing() (*Response, error) {
	return NewSuccessResponse(nil)
}

// handleStatus handles status requests
func (s *Server) handleStatus() (*Response, error) {
	status := StatusResponse{
		Running:  true,
		Uptime:   s.daemon.GetUptime(),
		WatchDir: s.daemon.GetWatchDir(),
	}
	return NewSuccessResponse(status)
}

// handleStop handles stop requests
func (s *Server) handleStop() (*Response, error) {
	// Request daemon shutdown
	s.daemon.RequestShutdown()
	return NewSuccessResponse(nil)
}

// handleGetState handles get_state requests (placeholder for TUI)
func (s *Server) handleGetState() (*Response, error) {
	// Placeholder - will be populated in Phase 5
	state := StateResponse{
		EarnedToday:    0,
		CurrentSession: 0,
		ActiveTask:     "",
	}
	return NewSuccessResponse(state)
}

// handleGetActivities handles get_activities requests
func (s *Server) handleGetActivities(req Request) (*Response, error) {
	// Parse request payload for limit
	limit := 10 // default
	if req.Payload != nil {
		var actReq GetActivitiesRequest
		if err := json.Unmarshal(req.Payload, &actReq); err == nil && actReq.Limit > 0 {
			limit = actReq.Limit
		}
	}

	activities := s.daemon.GetRecentActivitiesData(limit)
	return NewSuccessResponse(GetActivitiesResponse{Activities: activities})
}

// SocketPath returns the socket path for this server
func (s *Server) SocketPath() string {
	return s.socketPath
}
