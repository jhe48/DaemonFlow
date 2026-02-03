package notify

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"golang.org/x/time/rate"
)

// Notifier sends desktop notifications with rate limiting
type Notifier struct {
	config   NotifyConfig
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
}

// NewNotifier creates a new Notifier with the given configuration
func NewNotifier(cfg NotifyConfig) *Notifier {
	if cfg.MinGapSeconds <= 0 {
		cfg.MinGapSeconds = 300 // Default 5 minutes
	}
	return &Notifier{
		config:   cfg,
		limiters: make(map[string]*rate.Limiter),
	}
}

// getLimiter returns the rate limiter for a notification type, creating one if needed
func (n *Notifier) getLimiter(notifType string) *rate.Limiter {
	n.mu.Lock()
	defer n.mu.Unlock()

	if limiter, ok := n.limiters[notifType]; ok {
		return limiter
	}

	gap := time.Duration(n.config.MinGapSeconds) * time.Second
	limiter := rate.NewLimiter(rate.Every(gap), 1)
	n.limiters[notifType] = limiter
	return limiter
}

// send sends a notification with timeout protection
// Non-blocking - runs in goroutine with 3-second timeout to prevent daemon hangs
func (n *Notifier) send(title, message string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			if n.config.Sound {
				done <- beeep.Alert(title, message, "")
			} else {
				done <- beeep.Notify(title, message, "")
			}
		}()

		select {
		case err := <-done:
			if err != nil {
				log.Printf("notification failed: %v", err)
			}
		case <-ctx.Done():
			log.Printf("notification timed out: %s", title)
		}
	}()
}

// BreakEarned sends a "Break Time!" notification
func (n *Notifier) BreakEarned(minutes int) {
	if !n.config.Enabled || !n.config.BreakEarned {
		return
	}
	if !n.getLimiter("break_earned").Allow() {
		return
	}
	n.send("Break Time!", fmt.Sprintf("You've earned %d minutes of break time", minutes))
	log.Printf("Notification: BreakEarned (%d minutes)", minutes)
}

// BreakEnding sends a "Break Ending Soon" warning
func (n *Notifier) BreakEnding(secondsLeft int) {
	if !n.config.Enabled || !n.config.BreakEnding {
		return
	}
	if !n.getLimiter("break_ending").Allow() {
		return
	}
	n.send("Break Ending Soon", fmt.Sprintf("%d seconds remaining", secondsLeft))
	log.Printf("Notification: BreakEnding (%d seconds left)", secondsLeft)
}

// PetWarning sends a warning when pet is in overtime danger
func (n *Notifier) PetWarning(overtimeSeconds int) {
	if !n.config.Enabled || !n.config.PetWarning {
		return
	}
	if !n.getLimiter("pet_warning").Allow() {
		return
	}
	// overtimeSeconds is negative (debt), convert to positive for display
	debtMinutes := (-overtimeSeconds) / 60
	n.send("Pet in Danger!", fmt.Sprintf("Overtime for %d minutes - pet will die soon!", debtMinutes))
	log.Printf("Notification: PetWarning (%d minutes in overtime)", debtMinutes)
}

// PetDied sends a notification when the pet dies
func (n *Notifier) PetDied() {
	if !n.config.Enabled || !n.config.PetWarning {
		return
	}
	// No rate limit for death - it's rare and important
	n.send("Pet Died", "Your pet has died from extended overtime. Resurrect to continue.")
	log.Printf("Notification: PetDied")
}

// StreakMilestone sends a celebration notification for streak milestones
func (n *Notifier) StreakMilestone(days int) {
	if !n.config.Enabled || !n.config.StreakMilestone {
		return
	}
	// No rate limit for milestones - they're rare
	title := fmt.Sprintf("%d Day Streak!", days)
	n.send(title, "Keep up the great work!")
	log.Printf("Notification: StreakMilestone (%d days)", days)
}

// LevelUp sends a notification when the pet levels up
func (n *Notifier) LevelUp(level int) {
	if !n.config.Enabled || !n.config.LevelUp {
		return
	}
	// No rate limit for level ups - they're rare
	n.send("Level Up!", fmt.Sprintf("Your pet reached level %d", level))
	log.Printf("Notification: LevelUp (level %d)", level)
}
