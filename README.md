# DaemonFlow

A high-stakes productivity system for terminal-heavy developers. Work earns freedom; slack kills pets.

## Overview

DaemonFlow is a two-process productivity system:

- **Heart** (Go daemon) — Invisibly monitors your work: Git commits, file changes, and task completions
- **Face** (Rust TUI) — Displays an ASCII pet whose health reflects your productivity

You earn guilt-free break time through productive work. If you slack off, the clock hits zero, your pet dies, and it's logged to the Graveyard forever.

```
    /\_/\
   ( o.o )    <- Your productivity buddy
    > ^ <
   /|   |\
  (_|   |_)

  Status: HEALTHY
  Earned: 15:30
  Streak: 3 days
```

## How It Works

1. **Work** — The daemon detects Git commits, staged changes, file saves, and task completions
2. **Earn** — Each activity earns break time (configurable: commits = 5 min, file changes = 30 sec, etc.)
3. **Break** — Start a break and the clock counts down. Your pet rests.
4. **Consequences** — Run out of time? 5 minutes of overtime kills your pet. It's logged forever, but you can resurrect (at a cost).

## Installation

### Prerequisites

- Go 1.21+
- Rust 1.70+

### Build

```bash
# Build the daemon (Go)
go build -o bin/daemonflow ./cmd/daemonflow

# Build the TUI (Rust)
cd tui && cargo build --release
```

## Usage

### Start the Daemon

```bash
# Start monitoring in background
./bin/daemonflow start

# Check status
./bin/daemonflow status

# View recent activity
./bin/daemonflow activities

# Stop the daemon
./bin/daemonflow stop
```

### Launch the TUI

```bash
cd tui && cargo run --release
```

**Keyboard shortcuts:**
- `b` — Toggle break mode
- `r` — Refresh/reconnect to daemon
- `q` — Quit

## Configuration

Create `~/.daemonflow/config.yaml`:

```yaml
# Directory to monitor
watch_dir: "/path/to/your/project"

# Logging level (debug, info, warn, error)
log_level: "info"

# Earning weights (seconds earned per activity)
earning:
  commit_seconds: 300      # 5 minutes per commit
  stage_seconds: 60        # 1 minute per staging
  file_change_seconds: 30  # 30 seconds per file change
  task_complete_seconds: 180  # 3 minutes per task

# Task tracking
task:
  enabled: true
  file_path: "TASKS.md"    # Relative to watch_dir
  poll_interval: "2s"

# File watcher
watcher:
  enabled: true
  debounce_window: "500ms"
  ignore_patterns:
    - ".git/**"
    - "node_modules/**"
    - "*.log"
```

## Task Tracking

Create a markdown file with checkbox tasks:

```markdown
# Today's Tasks

- [x] Fix authentication bug
- [ ] Write unit tests
- [ ] Update documentation
```

When you check off a task, you earn break time.

## The Graveyard

When your pet dies, it's logged to `~/.daemonflow/GRAVEYARD.md`:

```markdown
## Deaths

| # | Date | Time in Overtime | Session Earned | Cause |
|---|------|------------------|----------------|-------|
| 1 | 2026-01-21 10:30 | 5m 23s | 45m | Extended overtime |

## Resurrections

| # | Death # | Date | Cost |
|---|---------|------|------|
| 1 | 1 | 2026-01-21 10:35 | Forfeited 45m earned |
```

Deaths are permanent records. Resurrection is possible, but costs all your earned time.

## Architecture

```
┌─────────────────┐     IPC (Unix Socket)     ┌─────────────────┐
│   Heart (Go)    │ ◄──────────────────────► │   Face (Rust)   │
│                 │      JSON messages        │                 │
│ • Git monitor   │                           │ • ASCII pet     │
│ • File watcher  │                           │ • Clock display │
│ • Task tracker  │                           │ • Streak info   │
│ • Freedom clock │                           │                 │
│ • Graveyard     │                           │                 │
└─────────────────┘                           └─────────────────┘
        │
        ▼
  ~/.daemonflow/
  ├── config.yaml
  ├── daemonflow.pid
  ├── daemonflow.sock
  └── GRAVEYARD.md
```

## License

MIT
