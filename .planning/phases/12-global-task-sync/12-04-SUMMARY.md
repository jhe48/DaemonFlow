---
phase: 12-global-task-sync
plan: 04
subsystem: ipc
tags: [ipc, sqlite, daemon, tasks, go]

# Dependency graph
requires:
  - phase: 12-02
    provides: TaskSync engine with database operations
  - phase: 12-03
    provides: Daemon multi-tracker integration with sync callbacks
provides:
  - get_tasks IPC endpoint for TUI to fetch global task list
  - Priority-sorted task retrieval from SQLite
  - Filtering by project and completion status
affects: [12-05, 13-01, tui-global-view]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - IPC endpoint pattern: protocol types -> server handler -> daemon method
    - Store abstraction: daemon translates store.Task to ipc.TaskData

key-files:
  created: []
  modified:
    - internal/ipc/protocol.go
    - internal/ipc/server.go
    - internal/daemon/daemon.go

key-decisions:
  - "Default limit of 50 tasks for get_tasks endpoint"
  - "TaskData uses string for dates (RFC3339) for JSON serialization"

patterns-established:
  - "IPC endpoint implementation: add constant, types, handler, and daemon method"

issues-created: []

# Metrics
duration: 2min
completed: 2026-01-28
---

# Phase 12 Plan 04: TUI Global View Summary

**Added get_tasks IPC endpoint enabling TUI to fetch priority-sorted tasks from SQLite with optional project filtering and completion status control**

## Performance

- **Duration:** 2 min
- **Started:** 2026-01-28T16:58:21Z
- **Completed:** 2026-01-28T17:00:39Z
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments

- Added RequestTypeGetTasks constant and task-related IPC types (GetTasksRequest, TaskData, GetTasksResponse)
- Implemented handleGetTasks handler in IPC server with request parsing and defaults
- Added GetGlobalTasks method to daemon bridging IPC to store operations
- Get_tasks endpoint returns priority-ordered tasks with project_name for display

## Task Commits

Each task was committed atomically:

1. **Task 1: Add task-related IPC protocol types** - `9f6e74d` (feat)
2. **Task 2: Implement get_tasks handler in IPC server** - `9f43bb4` (feat)
3. **Task 3: Add GetGlobalTasks method to daemon** - `6e173bb` (feat)

## Files Created/Modified

- `internal/ipc/protocol.go` - Added RequestTypeGetTasks constant, GetTasksRequest, TaskData, and GetTasksResponse types
- `internal/ipc/server.go` - Added GetGlobalTasks to DaemonInterface, handleGetTasks handler, and routing case
- `internal/daemon/daemon.go` - Implemented GetGlobalTasks method with filtering and conversion

## Decisions Made

- Default limit of 50 tasks for get_tasks endpoint (reasonable for TUI display)
- TaskData uses RFC3339 string format for dates rather than native time types for clean JSON serialization
- GetGlobalTasks filters completed tasks by default (includeCompleted=false)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- get_tasks endpoint ready for TUI integration
- Ready for 12-05-PLAN.md (Integration testing)

---
*Phase: 12-global-task-sync*
*Completed: 2026-01-28*
