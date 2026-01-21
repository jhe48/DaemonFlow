---
phase: 08-graveyard
plan: 01
subsystem: daemon
tags: [graveyard, death-detection, overtime, markdown, persistence]

# Dependency graph
requires:
  - phase: 07-pet-system
    provides: Pet state including death threshold constant
  - phase: 05-freedom-clock
    provides: Clock state machine with overtime tracking
provides:
  - Graveyard package for death record management
  - Death detection callback in Clock
  - GRAVEYARD.md file format with markdown table
  - Automatic death logging on extended overtime
affects: [08-02-recovery-mechanics, tui-death-display]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Callback-based event notification (OnDeath)
    - Markdown file as persistent storage
    - Thread-safe graveyard with RWMutex

key-files:
  created:
    - internal/graveyard/record.go
    - internal/graveyard/graveyard.go
  modified:
    - internal/clock/clock.go
    - internal/daemon/daemon.go

key-decisions:
  - "DeathThresholdSeconds = -300 matches TUI constant"
  - "OnDeath callback fires exactly once per overtime period"
  - "GRAVEYARD.md uses markdown table format for human readability"
  - "Death records stored with timestamp, overtime duration, and cause"

patterns-established:
  - "Event callbacks for cross-package notifications"
  - "Markdown files for user-visible persistent data"

issues-created: []

# Metrics
duration: 2 min
completed: 2026-01-21
---

# Phase 8 Plan 01: Graveyard Package Summary

**Death detection and logging system with GRAVEYARD.md for permanent pet death records.**

## Performance

- **Duration:** 2 min
- **Started:** 2026-01-21T09:51:15Z
- **Completed:** 2026-01-21T09:53:43Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- Created graveyard package with DeathRecord struct and GRAVEYARD.md file management
- Added death detection to Clock with OnDeath callback at -300 seconds threshold
- Integrated graveyard into daemon for automatic death logging
- Death records persist in ~/.daemonflow/GRAVEYARD.md with markdown table format

## Task Commits

Each task was committed atomically:

1. **Task 1: Create graveyard package with DeathRecord struct** - `7f773b9` (feat)
2. **Task 2: Add death detection to Clock** - `4eb26b2` (feat)
3. **Task 3: Integrate graveyard with daemon** - `ad953ec` (feat)

## Files Created/Modified
- `internal/graveyard/record.go` - DeathRecord struct with String() and ToMarkdownRow() methods
- `internal/graveyard/graveyard.go` - Graveyard struct with LogDeath(), LoadRecords(), GetDeathCount()
- `internal/clock/clock.go` - Added DeathThresholdSeconds, OnDeath callback, deathTriggered flag
- `internal/daemon/daemon.go` - Added graveyard field, initialization, and death callback wiring

## Decisions Made
- DeathThresholdSeconds = -300 to match TUI's DEATH_THRESHOLD_SECONDS constant
- OnDeath callback uses goroutine to avoid deadlocks with clock mutex
- GRAVEYARD.md format uses markdown table for human readability
- Death records include timestamp, overtime duration, session earned, and cause
- deathTriggered resets on EndBreak() to allow multiple deaths per session

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness
- Death detection and logging complete
- Ready for Plan 08-02 (recovery mechanics)
- GRAVEYARD.md file format established for TUI display

---
*Phase: 08-graveyard*
*Completed: 2026-01-21*
