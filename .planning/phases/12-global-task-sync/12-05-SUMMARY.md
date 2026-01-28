---
phase: 12-global-task-sync
plan: 05
subsystem: tui
tags: [rust, ratatui, ipc, tasks]

# Dependency graph
requires:
  - phase: 12-04
    provides: IPC get_tasks handler in Go daemon
provides:
  - TUI task list display with project context
  - Task scrolling with j/k keys
  - IPC client get_tasks method
affects: [13-pet-evolution]

# Tech tracking
tech-stack:
  added: []
  patterns: [ratatui List widget for task rendering]

key-files:
  created: []
  modified:
    - tui/src/ipc/protocol.rs
    - tui/src/ipc/client.rs
    - tui/src/app.rs
    - tui/src/ui.rs

key-decisions:
  - "Task list height fixed at 8 lines for visual balance"
  - "Tasks display project_name prefix for global context"

patterns-established:
  - "IPC request/response pattern for new endpoints"

issues-created: []

# Metrics
duration: 2min
completed: 2026-01-28
---

# Phase 12 Plan 05: Global Task List View Summary

**TUI displays unified task list from all projects with project names, due dates, and j/k scrolling**

## Performance

- **Duration:** 2 min
- **Started:** 2026-01-28T17:02:28Z
- **Completed:** 2026-01-28T17:04:34Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- TUI IPC client can fetch tasks via get_tasks() method
- App state tracks tasks and scroll position
- Task list renders in dedicated panel with project context and due dates
- j/k and Up/Down keys scroll through task list

## Task Commits

Each task was committed atomically:

1. **Task 1: Add task types and IPC client method** - `c9a7d20` (feat)
2. **Task 2: Add task state to App** - `897a5c4` (feat)
3. **Task 3: Render task list in UI** - `744bd3a` (feat)

## Files Created/Modified
- `tui/src/ipc/protocol.rs` - Added TaskData, GetTasksResponse structs and REQUEST_TYPE_GET_TASKS constant
- `tui/src/ipc/client.rs` - Added get_tasks() method to IpcClient
- `tui/src/app.rs` - Added tasks/task_scroll fields, update_tasks() method, j/k scroll handlers
- `tui/src/ui.rs` - Added render_tasks() function and layout changes

## Decisions Made
- Task list panel uses 8 lines height for good visual balance
- Tasks show project_name prefix to provide global context
- Tasks show date portion of due_date (first 10 chars)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness
- Phase 12 complete with all 5 plans done
- Ready for Phase 13: Pet Evolution
- TUI now shows global task view enabling user to see all tasks across projects

---
*Phase: 12-global-task-sync*
*Completed: 2026-01-28*
