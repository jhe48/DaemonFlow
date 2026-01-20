# Roadmap: DaemonFlow

## Overview

Build a two-process productivity system: a Go daemon ("Heart") that invisibly monitors Git activity, file changes, and task completions to earn break time, paired with a Rust TUI ("Face") displaying an ASCII pet whose health reflects your productivity. Work earns freedom; slack kills pets.

## Domain Expertise

None

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [x] **Phase 1: Foundation** - Go daemon skeleton with IPC socket (Complete)
- [x] **Phase 2: Git Monitoring** - Track commits, staged changes, branch switches (Complete)
- [x] **Phase 3: File Watcher** - Monitor file changes in project directory (Complete)
- [ ] **Phase 4: Task Tracking** - Checkbox-style task completion monitoring
- [ ] **Phase 5: Freedom Clock** - Earning formula and countdown logic
- [ ] **Phase 6: TUI Foundation** - Rust TUI with ratatui, connect to daemon
- [ ] **Phase 7: Pet System** - ASCII pet states reflecting productivity
- [ ] **Phase 8: Graveyard** - Death logging, recovery mechanics, streak management

## Phase Details

### Phase 1: Foundation
**Goal**: Working Go daemon that starts, runs persistently, and accepts IPC connections
**Depends on**: Nothing (first phase)
**Research**: Unlikely (standard Go patterns)
**Plans**: 2 plans (complete)

Plans:
- [x] 01-01: Core daemon structure (CLI, lifecycle, config) - Completed 2026-01-17
- [x] 01-02: Unix socket IPC server - Completed 2026-01-17

### Phase 2: Git Monitoring
**Goal**: Daemon detects and records Git activity in watched directory
**Depends on**: Phase 1
**Research**: Unlikely (using git CLI - simpler than go-git, decided in planning)
**Plans**: 2 plans

Plans:
- [x] 02-01: Git repo detection + commit/branch monitoring + daemon integration - Completed 2026-01-17
- [x] 02-02: Staged changes detection - Completed 2026-01-17

### Phase 3: File Watcher
**Goal**: Daemon detects meaningful file changes (not noise like .git or node_modules)
**Depends on**: Phase 1
**Research**: Unlikely (fsnotify is standard Go library, patterns established in Phase 2)
**Plans**: 2 plans

Plans:
- [x] 03-01: fsnotify watcher + ignore patterns + FileChange events - Completed 2026-01-20
- [x] 03-02: Debouncing + daemon integration - Completed 2026-01-20

### Phase 4: Task Tracking
**Goal**: Daemon monitors task file for checkbox completions
**Depends on**: Phase 1
**Research**: Unlikely (markdown parsing, file watching already covered)
**Plans**: TBD

Plans:
- [ ] 04-01: Task file detection and parsing
- [ ] 04-02: Completion detection and scoring

### Phase 5: Freedom Clock
**Goal**: Configurable formula converts activity into earned break time, countdown during breaks
**Depends on**: Phases 2, 3, 4 (all activity sources)
**Research**: Unlikely (internal business logic)
**Plans**: TBD

Plans:
- [ ] 05-01: Earning formula with configurable weights
- [ ] 05-02: Clock state machine (working, break, overtime)
- [ ] 05-03: Consequences at zero

### Phase 6: TUI Foundation
**Goal**: Rust TUI binary that connects to daemon and displays basic status
**Depends on**: Phase 1 (IPC), Phase 5 (clock data to display)
**Research**: Likely (ratatui ecosystem, Rust IPC client)
**Research topics**: ratatui setup, Unix socket client in Rust, message serialization
**Plans**: TBD

Plans:
- [ ] 06-01: Rust project setup with ratatui
- [ ] 06-02: IPC client connecting to daemon
- [ ] 06-03: Basic status display (clock, activity)

### Phase 7: Pet System
**Goal**: ASCII pet with states reflecting productivity (healthy → decaying → dead)
**Depends on**: Phase 6 (TUI infrastructure)
**Research**: Unlikely (ASCII art, state machine)
**Plans**: TBD

Plans:
- [ ] 07-01: Pet state machine and transitions
- [ ] 07-02: ASCII art for each state
- [ ] 07-03: Animation and display integration

### Phase 8: Graveyard
**Goal**: Death logging to GRAVEYARD.md, recovery mechanics, streak tracking
**Depends on**: Phase 7 (pet death trigger)
**Research**: Unlikely (file format, recovery logic)
**Plans**: TBD

Plans:
- [ ] 08-01: Death logging and GRAVEYARD.md format
- [ ] 08-02: Recovery mechanics and cost system
- [ ] 08-03: Streak tracking and display

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 2/2 | Complete | 2026-01-17 |
| 2. Git Monitoring | 2/2 | Complete | 2026-01-17 |
| 3. File Watcher | 2/2 | Complete | 2026-01-20 |
| 4. Task Tracking | 0/2 | Not started | - |
| 5. Freedom Clock | 0/3 | Not started | - |
| 6. TUI Foundation | 0/3 | Not started | - |
| 7. Pet System | 0/3 | Not started | - |
| 8. Graveyard | 0/3 | Not started | - |
