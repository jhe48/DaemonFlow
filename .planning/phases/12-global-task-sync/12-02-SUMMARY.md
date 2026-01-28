---
phase: 12-global-task-sync
plan: 02
subsystem: sync
tags: [sqlite, task-sync, priority-scoring, due-dates]

# Dependency graph
requires:
  - phase: 12-01
    provides: task schema with priority_score column
provides:
  - TaskSync type with SyncDirectory() method
  - Priority scoring algorithm (due date based)
  - Store methods for efficient sync operations
  - Due date extraction utilities
affects: [12-03-daemon-integration, 12-04-tui-global-view]

# Tech tracking
tech-stack:
  added: []
  patterns: ["upsert-based sync", "stale detection via text comparison"]

key-files:
  created:
    - internal/task/sync.go
    - internal/task/sync_test.go
  modified:
    - internal/store/tasks.go

key-decisions:
  - "Due date format: @due(YYYY-MM-DD) or @due(YYYY-MM-DDTHH:MM:SS)"
  - "Priority levels: +100 (24h), +50 (7d), +25 (scheduled), 0 (no due date)"
  - "Stale detection: compare task texts, delete tasks in DB but not in file"

patterns-established:
  - "Sync pattern: parse file → upsert tasks → delete stale"
  - "Priority scoring: time-based urgency levels"

issues-created: []

# Metrics
duration: 3min
completed: 2026-01-28
---

# Phase 12 Plan 02: Task Sync Engine Summary

**TaskSync with SyncDirectory() parses TASKS.md files, computes priority scores from due dates, and syncs to SQLite with stale task deletion**

## Performance

- **Duration:** 3 min
- **Started:** 2026-01-28T03:02:19Z
- **Completed:** 2026-01-28T03:05:01Z
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments

- TaskSync type with SyncDirectory() method for reliable file-to-DB sync
- Priority scoring algorithm: +100 (due within 24h), +50 (due within 7d), +25 (has due date)
- Stale task detection and deletion when tasks removed from TASKS.md
- Due date extraction with @due(YYYY-MM-DD) format support
- Store helper methods for efficient sync operations

## Task Commits

Each task was committed atomically:

1. **Task 2: Add store methods for sync operations** - `d2be6d1` (feat)
2. **Task 1: Create task sync engine with priority scoring** - `38cf6bf` (feat)
3. **Task 3: Add sync tests** - `05391c7` (test)

Note: Tasks 1 and 2 were reordered during execution since Task 1 depended on Task 2's store methods.

## Files Created/Modified

- `internal/task/sync.go` - TaskSync type, SyncDirectory(), priority computation, due date parsing
- `internal/task/sync_test.go` - Comprehensive test coverage for sync scenarios
- `internal/store/tasks.go` - GetTaskTextsByProject, DeleteTasksByProjectExcept, UpdatePriorityScore methods

## Decisions Made

- **Due date format:** Chose @due(YYYY-MM-DD) format that's easy to type and parse
- **Priority levels:** Three-tier urgency (+100/+50/+25) provides meaningful differentiation
- **Stale detection:** Text-based comparison rather than ID-based, since TASKS.md has no IDs

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- TaskSync ready to be integrated with daemon's directory watcher
- Priority-sorted queries available via GetTasksSortedByPriority()
- Ready for 12-03-PLAN.md (daemon integration)

---
*Phase: 12-global-task-sync*
*Completed: 2026-01-28*
