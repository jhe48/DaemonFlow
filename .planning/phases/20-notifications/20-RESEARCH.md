# Phase 20: Notifications - Research

**Researched:** 2026-02-03
**Domain:** Go cross-platform desktop notifications for daemon-to-user communication
**Confidence:** HIGH

<research_summary>
## Summary

Researched cross-platform desktop notification libraries for Go to implement alerts for breaks earned, pet status, and streak milestones. The standard approach uses **gen2brain/beeep**, the most actively maintained Go notification library with 530+ importers.

Key finding: Desktop notifications from daemons are straightforward on Linux (D-Bus/notify-send) and historically simple on macOS (osascript/terminal-notifier), but **macOS Sequoia (15.x) has introduced breaking changes** where `osascript display notification` from Terminal may fail silently. The workaround is to ensure terminal-notifier is installed as the primary method on macOS.

Notification rate limiting is essential - users don't want notification spam. Implement debouncing (e.g., one notification per event type per 5 minutes) and respect system Do Not Disturb settings.

**Primary recommendation:** Use gen2brain/beeep with `nodbus` build tag (simpler, uses notify-send fallback on Linux). Implement notification queuing with debouncing. On macOS, recommend users install terminal-notifier via Homebrew for reliable notifications.
</research_summary>

<standard_stack>
## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gen2brain/beeep | v0.11.1 | Cross-platform notifications | Most maintained, 530+ importers, handles platform differences |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| golang.org/x/time/rate | latest | Rate limiting | Notification throttling |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| beeep | godbus/dbus direct | More control but requires manual D-Bus code for Linux only |
| beeep | 0xAX/notificator | Less maintained, fewer platform fallbacks |
| beeep | keybase/go-notifier | Requires CGO on macOS, more complex setup |

### Platform Fallbacks (handled by beeep)
- **Linux**: D-Bus org.freedesktop.Notifications → notify-send command
- **macOS**: terminal-notifier → osascript display notification
- **Windows**: Windows Runtime COM API → PowerShell

**Installation:**
```bash
go get github.com/gen2brain/beeep
```

**Build tag option:**
```bash
# Disable D-Bus dependency on Linux (use notify-send only)
go build -tags nodbus
```
</standard_stack>

<architecture_patterns>
## Architecture Patterns

### Recommended Project Structure
```
internal/
├── notify/              # Notification module
│   ├── notifier.go      # Core notifier with rate limiting
│   ├── types.go         # Notification types (break, streak, pet)
│   └── config.go        # Notification preferences
└── daemon/
    └── daemon.go        # Existing - wire up notifier callbacks
```

### Pattern 1: Rate-Limited Notifier
**What:** Wrap beeep calls with rate limiting to prevent notification spam
**When to use:** Always - users hate notification floods
**Example:**
```go
// Source: golang.org/x/time/rate docs + beeep
package notify

import (
    "sync"
    "time"
    "github.com/gen2brain/beeep"
    "golang.org/x/time/rate"
)

type Notifier struct {
    enabled    bool
    limiters   map[string]*rate.Limiter
    mu         sync.Mutex
    minGap     time.Duration  // Minimum gap between same-type notifications
}

func NewNotifier(minGap time.Duration) *Notifier {
    return &Notifier{
        enabled:  true,
        limiters: make(map[string]*rate.Limiter),
        minGap:   minGap,
    }
}

// Send sends a notification if rate limit allows
func (n *Notifier) Send(notifType, title, message string) error {
    if !n.enabled {
        return nil
    }

    n.mu.Lock()
    limiter, exists := n.limiters[notifType]
    if !exists {
        // Allow 1 notification per minGap for this type
        limiter = rate.NewLimiter(rate.Every(n.minGap), 1)
        n.limiters[notifType] = limiter
    }
    n.mu.Unlock()

    if !limiter.Allow() {
        return nil // Rate limited, skip silently
    }

    return beeep.Notify(title, message, "")
}
```

### Pattern 2: Event-Based Notification Triggers
**What:** Define clear notification events in daemon, let notifier decide when to fire
**When to use:** When daemon state changes should trigger notifications
**Example:**
```go
// Source: DaemonFlow architecture pattern
type NotificationEvent int

const (
    EventBreakEarned NotificationEvent = iota
    EventBreakEnding      // 1 minute warning
    EventPetDying         // Overtime warning
    EventPetDied
    EventStreakMilestone  // 7, 14, 30, 100 days
    EventLevelUp
)

// In daemon callbacks:
func (d *Daemon) onBreakEarned(minutes int) {
    if d.notifier != nil {
        d.notifier.SendEvent(EventBreakEarned,
            "Break Time!",
            fmt.Sprintf("You earned %d minutes of break time", minutes))
    }
}
```

### Pattern 3: Configurable Notification Preferences
**What:** Let users disable specific notification types
**When to use:** Always - some users want minimal notifications
**Example:**
```yaml
# In config file
notifications:
  enabled: true
  break_earned: true
  break_ending_warning: true
  pet_status: false        # Some users find this annoying
  streak_milestones: true
  level_up: true
  sound: true              # Whether to use Alert() vs Notify()
  min_gap_seconds: 300     # 5 minutes between same-type notifications
```

### Anti-Patterns to Avoid
- **Sending every event as notification:** Users will disable all notifications
- **No rate limiting:** Can flood notification center
- **Blocking on notification send:** beeep.Notify can hang on some systems; use goroutine with timeout
- **Hardcoded notification text:** Make messages configurable for i18n
</architecture_patterns>

<dont_hand_roll>
## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Platform-specific notification APIs | Direct D-Bus, NSUserNotificationCenter, WinRT calls | gen2brain/beeep | Platform edge cases are numerous, beeep has fallbacks |
| Rate limiting logic | Custom time tracking | golang.org/x/time/rate | Token bucket is correct algorithm, pre-tested |
| Cross-platform sound alerts | Direct audio APIs | beeep.Alert() or beeep.Beep() | Handles platform beep differences |
| Notification queuing | Custom channel+timer | Rate limiter per notification type | Simpler, proven pattern |

**Key insight:** Notification APIs are deceptively simple-looking but have many platform quirks. On macOS, osascript may fail silently on Sequoia. On Linux, D-Bus session bus may not be available to daemon processes. On Windows, different APIs for different versions. Let beeep handle all this.
</dont_hand_roll>

<common_pitfalls>
## Common Pitfalls

### Pitfall 1: macOS Sequoia osascript Silent Failure
**What goes wrong:** `osascript -e 'display notification ...'` silently fails on macOS Sequoia
**Why it happens:** Apple changed security model; Terminal needs explicit notification permissions
**How to avoid:**
1. Recommend users install terminal-notifier (`brew install terminal-notifier`)
2. beeep will use terminal-notifier first if available
3. Document that users should check System Settings → Notifications → Terminal
**Warning signs:** Notifications work in development (Script Editor) but not in production

### Pitfall 2: D-Bus Session Bus Not Available to Daemons
**What goes wrong:** Linux notifications fail when daemon runs as systemd service
**Why it happens:** systemd services don't automatically have access to user's D-Bus session
**How to avoid:**
1. For user-spawned daemons (like DaemonFlow), this is fine - inherits session
2. For system services, need DBUS_SESSION_BUS_ADDRESS (usually `/run/user/<UID>/bus`)
3. Use `nodbus` build tag and rely on notify-send command instead
**Warning signs:** Works when run manually, fails when run as service

### Pitfall 3: Notification Spam Kills User Trust
**What goes wrong:** Users disable all app notifications because too many alerts
**Why it happens:** No rate limiting, notifying on every small event
**How to avoid:**
1. Rate limit per notification type (e.g., 5 minutes between "break earned")
2. Only notify on significant events (earned 5+ minutes, not every 30 seconds)
3. Let users configure which notifications they want
**Warning signs:** User feedback about "too many notifications"

### Pitfall 4: beeep.Notify() Can Hang
**What goes wrong:** Notification call blocks indefinitely on some systems
**Why it happens:** Underlying OS notification call may wait for user interaction or system response
**How to avoid:**
```go
// Call with timeout
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

done := make(chan error, 1)
go func() {
    done <- beeep.Notify(title, message, "")
}()

select {
case err := <-done:
    return err
case <-ctx.Done():
    return nil // Timeout, but don't block daemon
}
```
**Warning signs:** Daemon becomes unresponsive after notification

### Pitfall 5: Icons Not Displaying on macOS
**What goes wrong:** Custom icons don't appear in notifications
**Why it happens:** On macOS, beeep ignores the icon parameter and uses Script Editor icon
**How to avoid:** Accept default icon on macOS, or bundle app with proper Info.plist
**Warning signs:** Icon parameter works on Linux/Windows but not macOS
</common_pitfalls>

<code_examples>
## Code Examples

Verified patterns from official sources:

### Basic Notification with beeep
```go
// Source: gen2brain/beeep README
import "github.com/gen2brain/beeep"

// Simple notification
err := beeep.Notify("Title", "Message body, bla bla", "")
if err != nil {
    log.Printf("notification failed: %v", err)
}

// Alert with sound
err = beeep.Alert("Warning!", "Pet health is critical", "")

// Beep sound only
err = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
```

### Rate-Limited Notifier Integration
```go
// Source: Pattern from golang.org/x/time/rate
package notify

import (
    "context"
    "log"
    "time"

    "github.com/gen2brain/beeep"
    "golang.org/x/time/rate"
)

type Config struct {
    Enabled          bool          `yaml:"enabled"`
    BreakEarned      bool          `yaml:"break_earned"`
    BreakWarning     bool          `yaml:"break_ending_warning"`
    PetStatus        bool          `yaml:"pet_status"`
    StreakMilestone  bool          `yaml:"streak_milestones"`
    LevelUp          bool          `yaml:"level_up"`
    Sound            bool          `yaml:"sound"`
    MinGapSeconds    int           `yaml:"min_gap_seconds"`
}

type Notifier struct {
    config   Config
    limiters map[string]*rate.Limiter
}

func New(cfg Config) *Notifier {
    if cfg.MinGapSeconds <= 0 {
        cfg.MinGapSeconds = 300 // Default 5 minutes
    }
    return &Notifier{
        config:   cfg,
        limiters: make(map[string]*rate.Limiter),
    }
}

func (n *Notifier) getLimiter(notifType string) *rate.Limiter {
    if limiter, ok := n.limiters[notifType]; ok {
        return limiter
    }
    gap := time.Duration(n.config.MinGapSeconds) * time.Second
    limiter := rate.NewLimiter(rate.Every(gap), 1)
    n.limiters[notifType] = limiter
    return limiter
}

func (n *Notifier) BreakEarned(minutes int) {
    if !n.config.Enabled || !n.config.BreakEarned {
        return
    }
    if !n.getLimiter("break_earned").Allow() {
        return
    }
    n.send("Break Time!", fmt.Sprintf("You earned %d minutes", minutes))
}

func (n *Notifier) StreakMilestone(days int) {
    if !n.config.Enabled || !n.config.StreakMilestone {
        return
    }
    // No rate limit for milestones - they're rare
    title := fmt.Sprintf("%d Day Streak!", days)
    n.send(title, "Keep up the great work!")
}

func (n *Notifier) send(title, message string) {
    // Non-blocking send with timeout
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
            log.Printf("notification timed out")
        }
    }()
}
```

### Check Notification Prerequisites
```go
// Helper to check if notifications will likely work
func CheckNotificationSupport() (supported bool, message string) {
    // Try sending a test notification
    err := beeep.Notify("Test", "Testing notifications", "")
    if err != nil {
        return false, fmt.Sprintf("Notifications may not work: %v", err)
    }
    return true, "Notifications supported"
}
```
</code_examples>

<sota_updates>
## State of the Art (2025-2026)

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| osascript on macOS | terminal-notifier preferred | macOS Sequoia (late 2024) | osascript may fail silently from Terminal |
| martinlindhe/notify | gen2brain/beeep | 2020+ | notify deprecated, points to beeep |
| Custom D-Bus code | beeep with nodbus tag | 2023+ | Simpler to just use notify-send fallback |

**New tools/patterns to consider:**
- **beeep v0.11.1**: Latest version with nodbus fixes (Jun 2025)
- **terminal-notifier**: Still the best macOS CLI notification tool
- **mako/dunst**: Modern notification daemons for Wayland/X11 on Linux

**Deprecated/outdated:**
- **martinlindhe/notify**: Officially deprecated, use beeep
- **Direct osascript**: Unreliable on macOS Sequoia without workarounds
- **keybase/go-notifier**: Requires CGO, last updated 2020
</sota_updates>

<open_questions>
## Open Questions

Things that couldn't be fully resolved:

1. **macOS Sequoia specific fix status**
   - What we know: osascript display notification fails from Terminal in Sequoia
   - What's unclear: Whether recent macOS updates have fixed this, or if it's permanent
   - Recommendation: Test on actual macOS Sequoia; use terminal-notifier as primary

2. **beeep hanging issue (#70)**
   - What we know: Open issue about Notify() hanging on some systems
   - What's unclear: Specific conditions that trigger hang
   - Recommendation: Implement timeout wrapper around all beeep calls

3. **Do Not Disturb detection**
   - What we know: beeep can ignore DnD with terminal-notifier's -ignoreDnD flag
   - What's unclear: Whether we should respect or ignore DnD for critical alerts
   - Recommendation: Respect DnD by default; only override for "pet dying" alerts
</open_questions>

<sources>
## Sources

### Primary (HIGH confidence)
- [gen2brain/beeep GitHub](https://github.com/gen2brain/beeep) - README, issues, examples
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) - Rate limiting API
- [Freedesktop Notifications Spec](https://specifications.freedesktop.org/notification/latest-single/) - D-Bus notification protocol

### Secondary (MEDIUM confidence)
- [julienXX/terminal-notifier](https://github.com/julienXX/terminal-notifier) - macOS notification tool
- [Arch Wiki Desktop Notifications](https://wiki.archlinux.org/title/Desktop_notifications) - Linux notification architecture
- [MacScripter forum](https://www.macscripter.net/t/trying-to-use-terminal-for-display-notification/76593) - macOS Sequoia osascript issues

### Tertiary (LOW confidence - needs validation)
- beeep hanging issue (#70) - Reported but not reproduced
- macOS Sequoia workarounds - User-reported, may vary by system
</sources>

<metadata>
## Metadata

**Research scope:**
- Core technology: gen2brain/beeep for Go desktop notifications
- Ecosystem: terminal-notifier (macOS), notify-send (Linux), WinRT (Windows)
- Patterns: Rate limiting, event-based triggers, configurable preferences
- Pitfalls: macOS Sequoia, D-Bus session, notification spam, hanging calls

**Confidence breakdown:**
- Standard stack: HIGH - beeep is clearly the standard, well-documented
- Architecture: HIGH - patterns derived from official docs and Go best practices
- Pitfalls: MEDIUM - macOS Sequoia issues are recent, may evolve
- Code examples: HIGH - from official beeep docs and golang.org/x/time/rate

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (30 days - notification APIs relatively stable)
</metadata>

---

*Phase: 20-notifications*
*Research completed: 2026-02-03*
*Ready for planning: yes*
