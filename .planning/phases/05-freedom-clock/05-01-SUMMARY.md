---
phase: 05-freedom-clock
plan: 01
subsystem: clock
tags: [go, config, earning, activity]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: config package structure
  - phase: 02-git-monitoring
    provides: GitCommit, GitStage activity types
  - phase: 03-file-watcher
    provides: FileChange activity type
  - phase: 04-task-tracking
    provides: TaskComplete activity type
provides:
  - EarningCalculator for converting activities to earned seconds
  - EarningConfig with per-activity-type weights
affects: [05-02, 05-03, 06-03]

# Tech tracking
tech-stack:
  added: []
  patterns: [switch on activity type, config-driven weights]

key-files:
  created: [internal/clock/earning.go]
  modified: [internal/config/config.go]

key-decisions:
  - "Per-activity integer seconds instead of BaseRate/Multipliers for simplicity"
  - "Switch statement over map for activity type lookup (explicit, no reflection)"
  - "Default weights: commit=5min, stage=1min, file_change=30sec, task=3min"

patterns-established:
  - "Activity-to-seconds conversion via EarningCalculator"

issues-created: []

# Metrics
duration: 1min
completed: 2026-01-20
---

# Phase 5 Plan 01: Earning Formula Summary

**EarningCalculator that converts activity events to earned break seconds with configurable per-activity-type weights**

## Performance

- **Duration:** 1 min
- **Started:** 2026-01-20T22:29:18Z
- **Completed:** 2026-01-20T22:30:11Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Extended EarningConfig with CommitSeconds, StageSeconds, FileChangeSeconds, TaskCompleteSeconds fields
- Removed placeholder BaseRate and Multipliers fields from config
- Created internal/clock package with EarningCalculator
- EarningCalculator.CalculateEarned and GetWeight return correct seconds for all activity types

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend EarningConfig with activity weights** - `373fca3` (feat)
2. **Task 2: Create EarningCalculator** - `ab86df0` (feat)

## Files Created/Modified

- `internal/config/config.go` - EarningConfig with per-activity-type seconds fields and defaults
- `internal/clock/earning.go` - EarningCalculator with CalculateEarned and GetWeight methods

## Decisions Made

- Per-activity integer seconds instead of BaseRate/Multipliers: Simpler, more intuitive config, direct mapping to earned time
- Switch statement over map: Explicit type handling, no reflection, compile-time safety
- Default weights (commit=300s, stage=60s, file_change=30s, task=180s): Balance productivity incentives with reasonable break accumulation

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- EarningCalculator ready for use by clock state machine in 05-02
- Config schema complete with earning weights
- All activity types from Phases 2-4 are mapped to earning seconds

---
*Phase: 05-freedom-clock*
*Completed: 2026-01-20*
