---
phase: 08-graveyard
plan: 03
subsystem: daemon
tags: [graveyard, streak, ipc, tui, gamification]

# Dependency graph
requires:
  - phase: 08-01-graveyard-package
    provides: Graveyard package with death tracking, DeathRecord struct
  - phase: 08-02-recovery-mechanics
    provides: Resurrection mechanics, IsDead() state detection
provides:
  - Streak calculation from death records
  - StreakInfo struct with current/longest streak
  - IPC state response extended with streak data
  - TUI streak display in clock status section
affects: [future-gamification-features]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Streak calculation from event records
    - Optional JSON fields with serde defaults

key-files:
  created:
    - internal/graveyard/streak.go
  modified:
    - internal/graveyard/graveyard.go
    - internal/ipc/protocol.go
    - internal/ipc/server.go
    - internal/daemon/daemon.go
    - tui/src/ipc/protocol.rs
    - tui/src/app.rs
    - tui/src/ui.rs

key-decisions:
  - "Current streak = days since last death (0 on death day)"
  - "Longest streak = max gap between consecutive deaths"
  - "Streak defaults to 1 if never died"
  - "TUI uses serde(default) for backwards compatibility"

patterns-established:
  - "Streak calculation from timestamped records"
  - "Optional IPC fields with defaults for backwards compatibility"

issues-created: []

# Metrics
duration: 3 min
completed: 2026-01-21
---

# Phase 8 Plan 03: Streak Tracking Summary

**Streak calculation from death records with IPC exposure and TUI display for gamification motivation.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-01-21T15:01:01Z
- **Completed:** 2026-01-21T15:04:07Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments
- Created StreakInfo struct with CurrentStreak, LongestStreak, LastDeathDate
- Implemented CalculateStreak function computing streak from death records
- Extended IPC StateResponse with current_streak, longest_streak, total_deaths
- Added streak display to TUI clock status with color-coded styling
- Backwards compatible TUI parsing with serde defaults

## Task Commits

Each task was committed atomically:

1. **Task 1: Add streak calculation to graveyard** - `b6de675` (feat)
2. **Task 2: Expose streak in IPC state response** - `832bbab` (feat)
3. **Task 3: Display streak in TUI** - `f96b296` (feat)

## Files Created/Modified
- `internal/graveyard/streak.go` - StreakInfo struct, CalculateStreak function
- `internal/graveyard/graveyard.go` - Added GetStreakInfo() method
- `internal/ipc/protocol.go` - Extended StateResponse with streak fields
- `internal/ipc/server.go` - Added GetStreakInfo to DaemonInterface, updated handleGetState
- `internal/daemon/daemon.go` - Implemented GetStreakInfo() method
- `tui/src/ipc/protocol.rs` - Extended StateResponse with optional streak fields
- `tui/src/app.rs` - Added streak tracking fields to App state
- `tui/src/ui.rs` - Display streak in clock status with color-coded styling

## Decisions Made
- Current streak = days since last death (0 on death day, 1 next day)
- Longest streak tracks maximum gap between consecutive deaths
- If never died, both current and longest streak default to 1
- TUI uses serde(default) for backwards compatibility with older daemons
- Streak display uses green if > 0, dark gray if 0
- Deaths only shown if total_deaths > 0

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness
- Phase 8 (Graveyard) complete - all 3 plans finished
- Full graveyard system operational: death logging, resurrection, streak tracking
- TUI displays all gamification elements: pet state, clock, and streak
- Ready for milestone completion

---
*Phase: 08-graveyard*
*Completed: 2026-01-21*
