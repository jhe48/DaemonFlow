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
  Level: 3 (450 XP)
```

### Pet Evolution

Your pet evolves as you gain XP through productive work:

```
   Level 1          Level 2          Level 3          Level 4          Level 5
   Kitten           Young Cat        Adult Cat        Wise Cat         Royal Cat
                                                                         /\
    /\_/\            /\_/\            /\_/\            /\_/\            /||\
   ( o.o )          ( o.o )         =( o.o )=        =( o.o )=          /\_/\
    > ^ <            > ^ <           >[   ]<          >[:::]<         =( o.o )=
      |             /|   |\         /|     |\        /|     |\         [=*=*=]
                   (_|   |_)       (_|     |_)      (_|     |_)        /|   |\
```

Your pet has 5 states at every stage:

```
   HEALTHY:        RESTING:        TIRED:          DECAYING:       DEAD:
                      zzz            !             
     /\_/\          /\_/\          /\_/\          ' /\_/\ '        _____
    ( o.o )        ( -.- )        ( o_o )          ( x_x )        |     | 
     > ^ <          > ~ <          > ~ <            > ~ <         | RIP |
    /|   |\        /|   |\        /|   |\          /|   |\        |     |
   (_|   |_)      (_|   |_)      (_|   |_)        (_|___|_)       |_____|
```

## How It Works

1. **Work** - The daemon detects Git commits, staged changes, file saves, and task completions
2. **Earn** - Each activity earns break time and XP (configurable: commits = 5 min, file changes = 30 sec, etc.)
3. **Break** - Start a break and the clock counts down. Your pet rests.
4. **Level Up** - Gain XP to evolve your pet through 5 levels
5. **Consequences** - Run out of time? 5 minutes of overtime kills your pet. It's logged forever, but you can resurrect and get back to work.

## Installation

### Prerequisites

- Go 1.21+
- Rust 1.70+
- Python 3.10+ (optional, for recurring task parsing)

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

# Start in foreground (see logs)
./daemonflow start -f

# Check daemon status
./daemonflow status

# View recent activity
./daemonflow activities

# Stop the daemon
./daemonflow stop
```

### Quick-Add Tasks

Add tasks directly from your terminal without opening the TUI:

```bash
# Add a simple task
./daemonflow add "Fix the authentication bug"

# Tasks are added to your TASKS.md and synced to SQLite
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
| `a` | Add new task inline |
| `d` | Toggle dashboard (stats view) |
| `x` | Resurrect dead pet |
| `r` | Refresh/reconnect to daemon |
| `q` | Quit |

## Configuration

Create `~/.daemonflow/config.yaml`:

```yaml
# Directories to monitor (supports multiple)
watch_dirs:
  - "/path/to/project1"
  - "/path/to/project2"

# Or single directory (backward compatible)
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

# Desktop notifications
notifications:
  enabled: true
  break_earned: true      # Notify when break time earned
  break_ending: true      # Warn when 1 minute left in break
  pet_warning: true       # Alert when pet is in overtime danger
  streak_milestone: true  # Celebrate streak milestones (7, 14, 30, 100 days)
  level_up: true          # Notify on pet level up
  sound: false            # Play system sound with notifications
  min_gap_seconds: 300    # Minimum gap between same-type notifications
```

## Task Tracking

Create a markdown file with checkbox tasks in your `watch_dir`:

```markdown
# Today's Tasks

- [x] Fix authentication bug
- [ ] Write unit tests
- [ ] Update documentation
```

When you check off a task (`[ ]` to `[x]`), you earn break time and XP.

### Recurring Tasks

DaemonFlow supports natural language recurring tasks:

```markdown
- [ ] Review PRs every weekday
- [ ] Weekly team sync every Monday at 10am
- [ ] Monthly report on the 1st
```

## Dashboard

Press `d` in the TUI to view your productivity dashboard:

- **Today's Stats**: Tasks completed, commits, XP earned
- **Current Streak**: Days without pet death
- **Weekly Summary**: Productivity overview
- **Time Balance**: Break time earned vs spent

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

Deaths are permanent records, and they stay in the Graveyard forever. Your streak resets to 0 on death.

## Data Files

All DaemonFlow data is stored in `~/.daemonflow/`:

| File | Purpose |
|------|---------|
| `config.yaml` | Your configuration (create this) |
| `daemonflow.db` | SQLite database (tasks, stats, streaks) |
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
| - Freedom clock |                           | - Dashboard     |
| - Graveyard     |                           | - Task input    |
| - Notifications |                           |                 |
| - SQLite store  |                           |                 |
+-----------------+                           +-----------------+
        |
        v
  ~/.daemonflow/
  +-- config.yaml
  +-- daemonflow.db
  +-- daemonflow.pid
  +-- daemonflow.sock
  +-- GRAVEYARD.md
```

## Version History


- **v1.0** - Core daemon, TUI, pet system, graveyard
- **v2.0** - SQLite storage, recurring tasks, pet evolution, multi-project sync
- **v3.0** - Quick-add CLI, TUI task input, analytics dashboard, streaks, notifications
