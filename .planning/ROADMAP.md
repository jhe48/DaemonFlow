# Roadmap: DaemonFlow

## Overview

Build a two-process productivity system: a Go daemon ("Heart") that invisibly monitors Git activity, file changes, and task completions to earn break time, paired with a Rust TUI ("Face") displaying an ASCII pet whose health reflects your productivity. Work earns freedom; slack kills pets.

## Domain Expertise

None

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [ ] **Phase 1: Foundation** - Go daemon skeleton with IPC socket (In progress: 2/3 plans)
- [ ] **Phase 2: Git Monitoring** - Track commits, staged changes, branch switches
- [ ] **Phase 3: File Watcher** - Monitor file changes in project directory
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
**Plans**: TBD

Plans:
- [x] 01-01: Daemon process management (start, stop, daemonize) - Completed 2026-01-17
- [x] 01-02: Unix socket IPC server - Completed 2026-01-17
- [ ] 01-03: Configuration loading and state persistence

### Phase 2: Git Monitoring
**Goal**: Daemon detects and records Git activity in watched directory
**Depends on**: Phase 1
**Research**: Likely (go-git library or git CLI integration)
**Research topics**: go-git vs shelling out to git, watching .git directory, parsing git status
**Plans**: TBD

Plans:
- [ ] 02-01: Git repository detection and validation
- [ ] 02-02: Commit and branch change detection
- [ ] 02-03: Activity scoring for Git events

### Phase 3: File Watcher
**Goal**: Daemon detects meaningful file changes (not noise like .git or node_modules)
**Depends on**: Phase 1
**Research**: Likely (fsnotify patterns, debouncing)
**Research topics**: fsnotify library, ignore patterns, debouncing rapid changes
**Plans**: TBD

Plans:
- [ ] 03-01: File system watcher setup with ignore patterns
- [ ] 03-02: Change aggregation and debouncing
- [ ] 03-03: Activity scoring for file changes

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
| 1. Foundation | 2/3 | In progress | - |
| 2. Git Monitoring | 0/3 | Not started | - |
| 3. File Watcher | 0/3 | Not started | - |
| 4. Task Tracking | 0/2 | Not started | - |
| 5. Freedom Clock | 0/3 | Not started | - |
| 6. TUI Foundation | 0/3 | Not started | - |
| 7. Pet System | 0/3 | Not started | - |
| 8. Graveyard | 0/3 | Not started | - |
