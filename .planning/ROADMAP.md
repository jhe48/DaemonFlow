# Roadmap: DaemonFlow

## Overview

Build a two-process productivity system: a Go daemon ("Heart") that invisibly monitors Git activity, file changes, and task completions to earn break time, paired with a Rust TUI ("Face") displaying an ASCII pet whose health reflects your productivity. Work earns freedom; slack kills pets.

## Milestones

- ✅ **v1.0 MVP** — Phases 1-8 (shipped 2026-01-21) — [Archive](milestones/v1.0-ROADMAP.md)
- 🚧 **v2.0 Smart Harness** — Phases 9-14 (in progress)

## Completed Milestones

<details>
<summary>✅ v1.0 MVP (Phases 1-8) — SHIPPED 2026-01-21</summary>

- [x] Phase 1: Foundation (2/2 plans) — completed 2026-01-17
- [x] Phase 2: Git Monitoring (2/2 plans) — completed 2026-01-17
- [x] Phase 3: File Watcher (2/2 plans) — completed 2026-01-20
- [x] Phase 4: Task Tracking (2/2 plans) — completed 2026-01-20
- [x] Phase 5: Freedom Clock (3/3 plans) — completed 2026-01-20
- [x] Phase 6: TUI Foundation (3/3 plans) — completed 2026-01-20
- [x] Phase 7: Pet System (3/3 plans) — completed 2026-01-21
- [x] Phase 8: Graveyard (3/3 plans) — completed 2026-01-21

Full details: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)

</details>

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation | v1.0 | 2/2 | Complete | 2026-01-17 |
| 2. Git Monitoring | v1.0 | 2/2 | Complete | 2026-01-17 |
| 3. File Watcher | v1.0 | 2/2 | Complete | 2026-01-20 |
| 4. Task Tracking | v1.0 | 2/2 | Complete | 2026-01-20 |
| 5. Freedom Clock | v1.0 | 3/3 | Complete | 2026-01-20 |
| 6. TUI Foundation | v1.0 | 3/3 | Complete | 2026-01-20 |
| 7. Pet System | v1.0 | 3/3 | Complete | 2026-01-21 |
| 8. Graveyard | v1.0 | 3/3 | Complete | 2026-01-21 |

### 🚧 v2.0 Smart Harness (In Progress)

**Milestone Goal:** Add Python "Brain" layer for intelligent task handling, global SQLite sync, and streak-based pet evolution

#### Phase 9: SQLite Foundation

**Goal**: Schema design, Go integration, basic CRUD operations
**Depends on**: Previous milestone complete
**Research**: Unlikely (established patterns)
**Plans**: 2

Plans:
- [x] 09-01: SQLite integration and schema (completed 2026-01-22)
- [x] 09-02: CRUD operations (completed 2026-01-22)

#### Phase 10: Python Brain Setup

**Goal**: Cross-language bridge, os/exec integration, JSON data passing
**Depends on**: Phase 9
**Research**: Unlikely (standard subprocess patterns)
**Plans**: 2

Plans:
- [x] 10-01: Python brain package and Go executor (completed 2026-01-22)
- [x] 10-02: Brain executor daemon integration (completed 2026-01-22)

#### Phase 11: Recurring Task Parser

**Goal**: Natural language recurring pattern detection and RRULE generation
**Depends on**: Phase 10
**Research**: Complete (recurrent library, RFC 5545 RRULEs)
**Plans**: 2

Plans:
- [x] 11-01: Recurring patterns module and parse_recurring action (completed 2026-01-27)
- [x] 11-02: Recurring task unroller (completed 2026-01-28)

#### Phase 12: Global Task Sync

**Goal**: Multi-directory monitoring, upsert logic, global view in TUI
**Depends on**: Phase 11
**Research**: Unlikely (extends existing patterns)
**Plans**: 5

Plans:
- [x] 12-01: Config & schema foundation (completed 2026-01-28)
- [x] 12-02: Task sync engine (completed 2026-01-28)
- [x] 12-03: Daemon multi-tracker integration (completed 2026-01-28)
- [x] 12-04: TUI global view (completed 2026-01-28)
- [ ] 12-05: Integration testing

#### Phase 13: Pet Evolution

**Goal**: Streak calculation, level system, ASCII art progression in Rust TUI
**Depends on**: Phase 12
**Research**: Unlikely (internal UI work)
**Plans**: TBD

Plans:
- [ ] 13-01: TBD

#### Phase 14: Conflict Resolution & Polish

**Goal**: Duplicate prevention, edge cases, integration testing
**Depends on**: Phase 13
**Research**: Unlikely (internal logic)
**Plans**: TBD

Plans:
- [ ] 14-01: TBD

## Progress (v2.0)

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 9. SQLite Foundation | v2.0 | 2/2 | Complete | 2026-01-22 |
| 10. Python Brain Setup | v2.0 | 2/2 | Complete | 2026-01-22 |
| 11. Recurring Task Parser | v2.0 | 2/2 | Complete | 2026-01-28 |
| 12. Global Task Sync | v2.0 | 4/5 | In progress | - |
| 13. Pet Evolution | v2.0 | 0/? | Not started | - |
| 14. Conflict Resolution & Polish | v2.0 | 0/? | Not started | - |
