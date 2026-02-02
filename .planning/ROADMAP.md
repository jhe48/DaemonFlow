# Roadmap: DaemonFlow

## Overview

Build a two-process productivity system: a Go daemon ("Heart") that invisibly monitors Git activity, file changes, and task completions to earn break time, paired with a Rust TUI ("Face") displaying an ASCII pet whose health reflects your productivity. Work earns freedom; slack kills pets.

## Milestones

- ✅ **v1.0 MVP** — Phases 1-8 (shipped 2026-01-21) — [Archive](milestones/v1.0-ROADMAP.md)
- ✅ **v2.0 Smart Harness** — Phases 9-14 (shipped 2026-01-28) — [Archive](milestones/v2.0-ROADMAP.md)
- 🚧 **v3.0 Clarity** — Phases 15-20 (in progress)

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

<details>
<summary>✅ v2.0 Smart Harness (Phases 9-14) — SHIPPED 2026-01-28</summary>

- [x] Phase 9: SQLite Foundation (2/2 plans) — completed 2026-01-22
- [x] Phase 10: Python Brain Setup (2/2 plans) — completed 2026-01-22
- [x] Phase 11: Recurring Task Parser (2/2 plans) — completed 2026-01-28
- [x] Phase 12: Global Task Sync (5/5 plans) — completed 2026-01-28
- [x] Phase 13: Pet Evolution (4/4 plans) — completed 2026-01-28
- [x] Phase 14: Conflict Resolution & Polish (3/3 plans) — completed 2026-01-28

Full details: [milestones/v2.0-ROADMAP.md](milestones/v2.0-ROADMAP.md)

</details>

### 🚧 v3.0 Clarity (In Progress)

**Milestone Goal:** Simple, clear, easy to follow — improve daily interaction and provide visibility into productivity.

#### Phase 15: Quick-Add CLI

**Goal**: Add tasks from terminal without opening TUI (`dflow add "fix the bug"`)
**Depends on**: Previous milestone complete
**Research**: Unlikely (established Go CLI patterns)
**Plans**: 1

Plans:
- [x] 15-01: Add `add` subcommand to CLI (completed 2026-02-02)

#### Phase 16: TUI Task Input

**Goal**: Press `a` to add task inline with natural language parsing
**Depends on**: Phase 15
**Research**: Unlikely (ratatui patterns established)
**Plans**: TBD

Plans:
- [ ] 16-01: TBD

#### Phase 17: Analytics Foundation

**Goal**: Track daily stats (tasks completed, time earned/spent, XP gained) in SQLite
**Depends on**: Phase 16
**Research**: Unlikely (SQLite already integrated)
**Plans**: TBD

Plans:
- [ ] 17-01: TBD

#### Phase 18: Streaks & Summaries

**Goal**: Track productivity streaks and generate daily/weekly summaries
**Depends on**: Phase 17
**Research**: Unlikely (internal patterns)
**Plans**: TBD

Plans:
- [ ] 18-01: TBD

#### Phase 19: TUI Dashboard

**Goal**: Stats panel in TUI showing productivity at a glance
**Depends on**: Phase 18
**Research**: Unlikely (ratatui patterns)
**Plans**: TBD

Plans:
- [ ] 19-01: TBD

#### Phase 20: Notifications

**Goal**: Desktop alerts for breaks earned, pet status, streak milestones
**Depends on**: Phase 19
**Research**: Likely (cross-platform notification libraries)
**Research topics**: Go notification libraries (beeep, notify), Linux/macOS compatibility
**Plans**: TBD

Plans:
- [ ] 20-01: TBD

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
| 9. SQLite Foundation | v2.0 | 2/2 | Complete | 2026-01-22 |
| 10. Python Brain Setup | v2.0 | 2/2 | Complete | 2026-01-22 |
| 11. Recurring Task Parser | v2.0 | 2/2 | Complete | 2026-01-28 |
| 12. Global Task Sync | v2.0 | 5/5 | Complete | 2026-01-28 |
| 13. Pet Evolution | v2.0 | 4/4 | Complete | 2026-01-28 |
| 14. Conflict Resolution & Polish | v2.0 | 3/3 | Complete | 2026-01-28 |
| 15. Quick-Add CLI | v3.0 | 1/1 | Complete | 2026-02-02 |
| 16. TUI Task Input | v3.0 | 0/? | Not started | - |
| 17. Analytics Foundation | v3.0 | 0/? | Not started | - |
| 18. Streaks & Summaries | v3.0 | 0/? | Not started | - |
| 19. TUI Dashboard | v3.0 | 0/? | Not started | - |
| 20. Notifications | v3.0 | 0/? | Not started | - |
