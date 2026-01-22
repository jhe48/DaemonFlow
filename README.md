# DaemonFlow

A high-stakes productivity system for terminal-heavy developers. Work earns freedom; slack kills pets.

## Overview

DaemonFlow is a two-process productivity system:

- **Heart** (Go daemon) - Invisibly monitors your work: Git commits, file changes, and task completions
- **Face** (Rust TUI) - Displays an ASCII pet whose health reflects your productivity

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

1. **Work** - The daemon detects Git commits, staged changes, file saves, and task completions
2. **Earn** - Each activity earns break time (configurable: commits = 5 min, file changes = 30 sec, etc.)
3. **Break** - Start a break and the clock counts down. Your pet rests.
4. **Consequences** - Run out of time? 5 minutes of overtime kills your pet. It's logged forever, but you can resurrect and get back to work.

## Installation

### Prerequisites

- Go 1.21+
- Rust 1.70+

### Build

```bash
# Build the daemon (Go)
go build -o daemonflow ./cmd/daemonflow

# Build the TUI (Rust)
cd tui && cargo build --release
```

## Usage

### Start the Daemon

```bash
# Start monitoring in background
./daemonflow start

# Check daemon status
./daemonflow status

# View recent activity
./daemonflow activities

# Stop the daemon
./daemonflow stop
```

### Launch the TUI

```bash
cd tui && cargo run --release
# Or run the built binary directly:
./tui/target/release/daemonflow-tui
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `b` | Toggle break mode (blocked when pet is dead) |
| `x` | Resurrect dead pet |
| `r` | Refresh/reconnect to daemon |
| `q` | Quit |

## Configuration

Create `~/.daemonflow/config.yaml`:

```yaml
# Directory to monitor (defaults to current directory)
watch_dir: "/path/to/your/project"

# Logging level (debug, info, warn, error)
log_level: "info"

# Earning weights (seconds earned per activity)
earning:
  commit_seconds: 300        # 5 minutes per git commit
  stage_seconds: 60          # 1 minute per git stage
  file_change_seconds: 30    # 30 seconds per file change
  task_complete_seconds: 180 # 3 minutes per task completion

# Task tracking
task:
  enabled: true
  file_path: "TASKS.md"      # Relative to watch_dir
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

Create a markdown file with checkbox tasks in your `watch_dir`:

```markdown
# Today's Tasks

- [x] Fix authentication bug
- [ ] Write unit tests
- [ ] Update documentation
```

When you check off a task (`[ ]` to `[x]`), you earn break time.

## The Graveyard

When your pet dies, it's logged permanently to `~/.daemonflow/GRAVEYARD.md`.

### Viewing the Graveyard

```bash
cat ~/.daemonflow/GRAVEYARD.md
```

### Graveyard Format

```markdown
## Deaths

| # | Date | Time in Overtime | Session Earned | Cause |
|---|------|------------------|----------------|-------|
| 1 | 2026-01-21 10:30 | 5m 23s | 45m | Extended overtime |

## Resurrections

| # | Death # | Date |
|---|---------|------|
| 1 | 1 | 2026-01-21 10:35 |
```

### Resurrection

When your pet dies, press `x` in the TUI to resurrect. Your pet revives and you must immediately start working to earn break time again.

Deaths are permanent records - they stay in the Graveyard forever. Your streak resets to 0 on death.

## Data Files

All DaemonFlow data is stored in `~/.daemonflow/`:

| File | Purpose |
|------|---------|
| `config.yaml` | Your configuration (create this) |
| `daemonflow.pid` | Process ID of running daemon |
| `daemonflow.sock` | Unix socket for IPC |
| `GRAVEYARD.md` | Death and resurrection records |

## Architecture

```
+-----------------+     IPC (Unix Socket)     +-----------------+
|   Heart (Go)    | <-----------------------> |   Face (Rust)   |
|                 |      JSON messages        |                 |
| - Git monitor   |                           | - ASCII pet     |
| - File watcher  |                           | - Clock display |
| - Task tracker  |                           | - Streak info   |
| - Freedom clock |                           |                 |
| - Graveyard     |                           |                 |
+-----------------+                           +-----------------+
        |
        v
  ~/.daemonflow/
  +-- config.yaml
  +-- daemonflow.pid
  +-- daemonflow.sock
  +-- GRAVEYARD.md
```
