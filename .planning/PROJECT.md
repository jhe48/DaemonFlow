# DaemonFlow

## What This Is

A high-stakes productivity system for terminal-heavy developers. A Go-based background daemon ("the Heart") monitors your work — Git commits, task completions, and file changes — while a Rust-based TUI ("the Face") displays a digital ASCII pet. You earn guilt-free break time through productive work; if you slack off, the clock hits zero, the pet decays into a Graveyard, and your streak is wiped.

## Core Value

The daemon must work flawlessly in the background without intervention. Everything else depends on reliable, invisible monitoring.

## Requirements

### Validated

- ✓ Go daemon ("Heart") that runs persistently in background — v1.0
- ✓ Monitor Git activity (commits, staged changes, branch switches) — v1.0
- ✓ Monitor file changes in watched project directory — v1.0
- ✓ Task completion tracking (checkbox-style) — v1.0
- ✓ Freedom Clock: configurable formula for earning break time — v1.0
- ✓ Clock countdown during breaks, consequences at zero — v1.0
- ✓ Rust TUI ("Face") displaying ASCII pet — v1.0
- ✓ Pet state reflects productivity (healthy → decaying → dead) — v1.0
- ✓ Graveyard system: log deaths to GRAVEYARD.md, recovery possible at cost — v1.0
- ✓ IPC between daemon and TUI — v1.0
- ✓ User configuration for earning formula weights — v1.0

### Active

(All v1.0 requirements shipped — define next milestone to add new requirements)

### Out of Scope

- Multi-project tracking — v1 monitors one project at a time
- Team/social features — no leaderboards, sharing, or multiplayer
- Mobile/web companion — terminal only, no external interfaces
- Cloud sync — offline-first, no network dependencies
- Windows support — focus on Unix-like systems first

## Context

Shipped v1.0 with 4,332 LOC (3,510 Go + 822 Rust).
Tech stack: Go daemon (cobra, fsnotify), Rust TUI (ratatui, crossterm).
Two-process architecture working as designed: daemon monitors invisibly, TUI provides optional emotional feedback.

Target users are developers who live in the terminal and struggle with productivity guilt — the feeling that any break is unearned. DaemonFlow replaces vague guilt with a concrete, measurable system: work earns time, time gets spent, consequences are real.

## Constraints

- **Minimal dependencies**: No heavy runtimes. Go and Rust compile to single binaries.
- **Offline-first**: Core functionality requires no network. All data stays local.
- **Low resource usage**: Daemon must be invisible in terms of CPU/memory impact.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go for daemon, Rust for TUI | Go excels at concurrent background services; Rust excels at terminal UIs (ratatui ecosystem) | ✓ Good — Clean separation, both languages played to strengths |
| Configurable earning formula | Different workflows need different weights (some commit often, some in bursts) | ✓ Good — Per-activity seconds (commit=5m, stage=1m, file=30s, task=3m) |
| Recovery + permanent shame model | Deaths logged forever but revival possible — maintains stakes without being punitive | ✓ Good — GRAVEYARD.md with resurrection at cost of earned time |
| File changes + Git + tasks as signals | Multiple input signals give more accurate picture of actual work | ✓ Good — Three monitors (git, file, task) all contribute to earned time |
| Git CLI over go-git | Shell out to git CLI instead of embedding go-git library | ✓ Good — Simpler, lighter, no complex dependency |
| fsnotify for file watching | Cross-platform file system notifications with recursive directory support | ✓ Good — Standard Go library, worked well |
| 500ms debounce for file events | Balance responsiveness with noise reduction from rapid saves | ✓ Good — Reduces spam without losing events |
| Length-prefixed JSON IPC | 4-byte big-endian prefix for message framing, JSON for debuggability | ✓ Good — Simple, debuggable, sufficient performance |
| Cat-based ASCII pet | Simple, emotionally expressive design (~5 lines tall) | ✓ Good — Clear emotional states, fits terminal well |
| 5-minute overtime death threshold | -300 seconds before pet dies | ✓ Good — Enough warning time while maintaining stakes |

---
*Last updated: 2026-01-21 after v1.0 milestone*
