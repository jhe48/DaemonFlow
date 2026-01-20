package clock

import (
	"sync"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/activity"
	"github.com/jackyhe0402/daemonflow/internal/config"
)

// ClockState represents the current state of the Freedom Clock
type ClockState string

const (
	// StateWorking means the user is actively working and earning time
	StateWorking ClockState = "working"
	// StateBreak means the user is on break and time is counting down
	StateBreak ClockState = "break"
	// StateOvertime means break time is exhausted and user is in debt
	StateOvertime ClockState = "overtime"
)

// Clock tracks earned break time and manages state transitions
type Clock struct {
	state         ClockState
	earnedSeconds int // Total earned seconds (can go negative in overtime)
	sessionEarned int // Earned this session (resets on daemon start)
	earning       *EarningCalculator
	mu            sync.RWMutex
	ticker        *time.Ticker
	stopChan      chan struct{}
}

// NewClock creates a new Clock with the given earning configuration
func NewClock(earningCfg *config.EarningConfig) *Clock {
	return &Clock{
		state:         StateWorking,
		earnedSeconds: 0,
		earning:       NewEarningCalculator(earningCfg),
		stopChan:      make(chan struct{}),
	}
}

// Start begins the clock ticker (1-second resolution)
func (c *Clock) Start() {
	c.mu.Lock()
	c.ticker = time.NewTicker(1 * time.Second)
	c.mu.Unlock()

	go c.runTicker()
}

// runTicker handles the ticker loop
func (c *Clock) runTicker() {
	for {
		select {
		case <-c.stopChan:
			return
		case <-c.ticker.C:
			c.tick()
		}
	}
}

// tick processes one second of time
func (c *Clock) tick() {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case StateWorking:
		// In working state, ticker does nothing - time is earned via OnActivity
	case StateBreak:
		// In break state, decrement earned seconds
		c.earnedSeconds--
		// Transition to overtime when time runs out
		if c.earnedSeconds <= 0 {
			c.state = StateOvertime
		}
	case StateOvertime:
		// In overtime, continue decrementing (goes negative)
		c.earnedSeconds--
	}
}

// Stop halts the clock ticker
func (c *Clock) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.stopChan)
}

// OnActivity processes an activity and adds earned time
// Implements activity.ActivityListener
func (c *Clock) OnActivity(act activity.Activity) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Only earn time when in working state
	if c.state == StateWorking {
		earned := c.earning.CalculateEarned(act)
		c.earnedSeconds += earned
		c.sessionEarned += earned
	}
}

// StartBreak transitions to break state and returns previous/new state
func (c *Clock) StartBreak() (previousState, newState string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	previousState = string(c.state)
	// Can only start break if in working state
	if c.state == StateWorking {
		c.state = StateBreak
	}
	newState = string(c.state)
	return
}

// EndBreak transitions back to working state and returns previous/new state
func (c *Clock) EndBreak() (previousState, newState string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	previousState = string(c.state)
	// EndBreak always returns to working, regardless of current state (break or overtime)
	if c.state == StateBreak || c.state == StateOvertime {
		c.state = StateWorking
	}
	newState = string(c.state)
	return
}

// GetState returns current clock state
func (c *Clock) GetState() ClockState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// GetEarnedSeconds returns current earned seconds balance
func (c *Clock) GetEarnedSeconds() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.earnedSeconds
}

// GetSessionEarned returns total seconds earned this session
func (c *Clock) GetSessionEarned() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionEarned
}
