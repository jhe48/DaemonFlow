# DaemonFlow

## What This Is

A high-stakes productivity system for terminal-heavy developers. A Go-based background daemon ("the Heart") monitors your work — Git commits, task completions, and file changes — while a Rust-based TUI ("the Face") displays a digital ASCII pet. You earn guilt-free break time through productive work; if you slack off, the clock hits zero, the pet decays into a Graveyard, and your streak is wiped.

## Core Value

The daemon must work flawlessly in the background without intervention. Everything else depends on reliable, invisible monitoring.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Go daemon ("Heart") that runs persistently in background
- [ ] Monitor Git activity (commits, staged changes, branch switches)
- [ ] Monitor file changes in watched project directory
- [ ] Task completion tracking (checkbox-style)
- [ ] Freedom Clock: configurable formula for earning break time
- [ ] Clock countdown during breaks, consequences at zero
- [ ] Rust TUI ("Face") displaying ASCII pet
- [ ] Pet state reflects productivity (healthy → decaying → dead)
- [ ] Graveyard system: log deaths to GRAVEYARD.md, recovery possible at cost
- [ ] IPC between daemon and TUI
- [ ] User configuration for earning formula weights

### Out of Scope

- Multi-project tracking — v1 monitors one project at a time
- Team/social features — no leaderboards, sharing, or multiplayer
- Mobile/web companion — terminal only, no external interfaces
- Cloud sync — offline-first, no network dependencies
- Windows support — focus on Unix-like systems first

## Context

Target users are developers who live in the terminal and struggle with productivity guilt — the feeling that any break is unearned. DaemonFlow replaces vague guilt with a concrete, measurable system: work earns time, time gets spent, consequences are real.

The two-process architecture (Go daemon + Rust TUI) allows the monitoring to be completely invisible while the pet interface remains optional and lightweight. Users can run just the daemon and check in occasionally, or keep the TUI open as a constant companion.

## Constraints

- **Minimal dependencies**: No heavy runtimes. Go and Rust compile to single binaries.
- **Offline-first**: Core functionality requires no network. All data stays local.
- **Low resource usage**: Daemon must be invisible in terms of CPU/memory impact.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go for daemon, Rust for TUI | Go excels at concurrent background services; Rust excels at terminal UIs (ratatui ecosystem) | — Pending |
| Configurable earning formula | Different workflows need different weights (some commit often, some in bursts) | — Pending |
| Recovery + permanent shame model | Deaths logged forever but revival possible — maintains stakes without being punitive | — Pending |
| File changes + Git + tasks as signals | Multiple input signals give more accurate picture of actual work | — Pending |

---
*Last updated: 2026-01-17 after initialization*
