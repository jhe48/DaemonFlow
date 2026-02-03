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

// DeathThresholdSeconds is the overtime threshold at which the pet dies
// Matches TUI's DEATH_THRESHOLD_SECONDS (-300 = 5 minutes of overtime)
const DeathThresholdSeconds = -300

// Clock tracks earned break time and manages state transitions
type Clock struct {
	state          ClockState
	earnedSeconds  int // Total earned seconds (can go negative in overtime)
	sessionEarned  int // Earned this session (resets on daemon start)
	earning        *EarningCalculator
	deathTriggered bool // Prevents multiple death events per overtime period
	mu             sync.RWMutex
	ticker         *time.Ticker
	stopChan       chan struct{}

	// OnDeath is called when the pet dies from extended overtime
	// Parameters: overtimeSeconds (negative), sessionEarned (positive)
	OnDeath func(overtimeSeconds int, sessionEarned int)

	// OnBreakEnding is called when 1 minute of break time remaining
	// Parameter: secondsLeft (should be 60)
	OnBreakEnding func(secondsLeft int)

	// OnOvertimeWarning is called at -180 seconds (3 minutes into overtime, 2 min before death)
	// Parameter: overtimeSeconds (negative)
	OnOvertimeWarning func(overtimeSeconds int)
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
		// Notify when 1 minute remaining
		if c.earnedSeconds == 60 && c.OnBreakEnding != nil {
			go c.OnBreakEnding(60)
		}
		// Transition to overtime when time runs out
		if c.earnedSeconds <= 0 {
			c.state = StateOvertime
		}
	case StateOvertime:
		// In overtime, continue decrementing (goes negative)
		c.earnedSeconds--
		// Warning at -180 seconds (3 minutes into overtime, 2 min before death)
		if c.earnedSeconds == -180 && c.OnOvertimeWarning != nil {
			overtimeSeconds := c.earnedSeconds
			go c.OnOvertimeWarning(overtimeSeconds)
		}
		// Check for death threshold
		if c.earnedSeconds <= DeathThresholdSeconds && !c.deathTriggered {
			c.deathTriggered = true
			if c.OnDeath != nil {
				// Call outside the lock to avoid deadlocks
				overtimeSeconds := c.earnedSeconds
				sessionEarned := c.sessionEarned
				go c.OnDeath(overtimeSeconds, sessionEarned)
			}
		}
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
		// Reset death trigger so death can occur again if user goes back to overtime
		c.deathTriggered = false
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

// Reset resets the clock to initial working state
// This is called during resurrection - forfeits all earned time as recovery cost
func (c *Clock) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.earnedSeconds = 0  // Forfeit all earned time (recovery cost)
	c.state = StateWorking
	c.deathTriggered = false
}
