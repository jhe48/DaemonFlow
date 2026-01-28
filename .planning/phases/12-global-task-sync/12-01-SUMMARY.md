---
phase: 12-global-task-sync
plan: 01
subsystem: config, database
tags: [multi-directory, sqlite, schema-migration, priority-scoring]

# Dependency graph
requires:
  - phase: 09-sqlite-foundation
    provides: SQLite store with tasks table
provides:
  - Multi-directory watch_dirs config support
  - Task priority_score and project_name columns
  - GetTasksSortedByPriority() query for global view
affects: [12-global-task-sync, 13-pet-evolution]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Config backward compatibility with normalizeWatchDirs()
    - Database migration pattern for adding columns

key-files:
  created: []
  modified:
    - internal/config/config.go
    - internal/store/schema.go
    - internal/store/tasks.go

key-decisions:
  - "WatchDirs takes precedence over WatchDir when both are set"
  - "project_name extracted via filepath.Base() at insert time"
  - "priority_score defaults to 0, computed by separate sync function"

patterns-established:
  - "Config normalization pattern: normalizeWatchDirs() called in all LoadConfig paths"
  - "Migration pattern: ALTER TABLE ADD COLUMN with UPDATE for backfill"

issues-created: []

# Metrics
duration: 3min
completed: 2026-01-28
---

# Phase 12 Plan 01: Config & Schema Foundation Summary

**Multi-directory config support (watch_dirs) with backward compatibility, task schema extended with project_name and priority_score columns**

## Performance

- **Duration:** 3 min
- **Started:** 2026-01-28T02:57:12Z
- **Completed:** 2026-01-28T03:00:27Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Config supports both `watch_dir` (single string, backward compat) and `watch_dirs` (array) with automatic normalization
- Task schema v2 adds `project_name` (extracted from path) and `priority_score` (for global view ordering)
- New `GetTasksSortedByPriority()` returns incomplete tasks ordered by priority_score DESC, due_date ASC, created_at ASC

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend config for multi-directory support** - `7c4aab1` (feat)
2. **Task 2: Add priority scoring and project name to task schema** - `6c2e701` (feat)

## Files Created/Modified

- `internal/config/config.go` - Added WatchDirs field, normalizeWatchDirs(), GetWatchDirs() helper
- `internal/store/schema.go` - Added migration 2 with project_name, priority_score columns and index
- `internal/store/tasks.go` - Extended Task struct, updated all CRUD operations, added GetTasksSortedByPriority()

## Decisions Made

- WatchDirs takes precedence when both WatchDir and WatchDirs are configured
- project_name is extracted using filepath.Base() automatically on insert/update
- priority_score defaults to 0 and will be computed by the sync function (Plan 02)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- Config ready for multi-tracker implementation
- Schema ready for priority sync function
- Ready for 12-02-PLAN.md (multi-tracker implementation)

---
*Phase: 12-global-task-sync*
*Completed: 2026-01-28*
