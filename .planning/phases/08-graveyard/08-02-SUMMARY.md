---
phase: 08-graveyard
plan: 02
subsystem: daemon
tags: [graveyard, resurrection, recovery, ipc, clock-reset]

# Dependency graph
requires:
  - phase: 08-01-graveyard-package
    provides: Graveyard package with death tracking, DeathRecord struct
  - phase: 05-freedom-clock
    provides: Clock state machine with death detection
provides:
  - Pet resurrection mechanics with cost system
  - ResurrectionRecord struct and GRAVEYARD.md format extension
  - IsDead() state detection for death/resurrection balance
  - Clock.Reset() for recovery cost (forfeit earned time)
  - IPC resurrect command for TUI integration
affects: [08-03-streak-tracking, tui-resurrection-ui]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Death/resurrection balance tracking (deaths > resurrections = dead)
    - Recovery cost pattern (forfeit earned time on resurrection)

key-files:
  modified:
    - internal/graveyard/record.go
    - internal/graveyard/graveyard.go
    - internal/clock/clock.go
    - internal/ipc/protocol.go
    - internal/ipc/server.go
    - internal/daemon/daemon.go

key-decisions:
  - "Resurrection cost forfeits all earned time (reset to 0)"
  - "IsDead() uses deaths > resurrections balance check"
  - "GRAVEYARD.md extended with Resurrections section"
  - "Can only resurrect if pet is actually dead"

patterns-established:
  - "Recovery cost pattern via Clock.Reset()"
  - "State balance tracking (count comparison for state detection)"

issues-created: []

# Metrics
duration: 3 min
completed: 2026-01-21
---

# Phase 8 Plan 02: Recovery Mechanics Summary

**Pet resurrection with cost system - forfeit all earned time to revive dead pet via IPC resurrect command.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-01-21T14:56:05Z
- **Completed:** 2026-01-21T14:58:46Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Added ResurrectionRecord struct with death index tracking for GRAVEYARD.md
- Extended graveyard with IsDead(), LogResurrection(), GetResurrectionCount()
- Added Clock.Reset() method to forfeit earned time as resurrection cost
- Implemented IPC resurrect command with daemon integration
- GRAVEYARD.md format extended with separate Deaths and Resurrections sections

## Task Commits

Each task was committed atomically:

1. **Task 1: Add resurrection state to graveyard** - `d4ff3fc` (feat)
2. **Task 2: Add clock reset for recovery** - `3b37032` (feat)
3. **Task 3: Add resurrect IPC command** - `2ddbb7e` (feat)

## Files Created/Modified
- `internal/graveyard/record.go` - Added ResurrectionRecord struct with ToMarkdownRow()
- `internal/graveyard/graveyard.go` - Added Resurrections slice, IsDead(), LogResurrection(), GetResurrectionCount(), updated LoadRecords() for both sections
- `internal/clock/clock.go` - Added Reset() method for recovery cost
- `internal/ipc/protocol.go` - Added RequestTypeResurrect constant and ResurrectResponse struct
- `internal/ipc/server.go` - Added Resurrect() to DaemonInterface, handleResurrect() handler
- `internal/daemon/daemon.go` - Added Resurrect() method with graveyard/clock integration

## Decisions Made
- Resurrection cost = forfeit all earned time (reset to 0) - maintains stakes while allowing recovery
- IsDead() checks if deaths > resurrections (balance-based state detection)
- GRAVEYARD.md format extended with separate Deaths and Resurrections sections
- Resurrect IPC returns success=false (not error) if pet not dead - graceful handling

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness
- Recovery mechanics complete and functional
- Ready for Plan 08-03 (streak tracking)
- TUI can now offer resurrection when detecting dead state via IPC

---
*Phase: 08-graveyard*
*Completed: 2026-01-21*
