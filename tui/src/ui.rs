use ratatui::prelude::*;
use ratatui::widgets::{Block, Borders, List, ListItem, Paragraph};
use crate::app::{App, InputMode};
use crate::pet::PetState;

pub fn render(app: &App, frame: &mut Frame) {
    let area = frame.area();

    // Determine layout based on input mode
    let in_input_mode = app.input_mode == InputMode::AddingTask;

    // Main layout: vertical split
    let constraints = if in_input_mode {
        vec![
            Constraint::Length(3),   // Header
            Constraint::Min(10),     // Pet display (primary focus)
            Constraint::Length(5),   // Clock status
            Constraint::Length(6),   // Task list (reduced when input shown)
            Constraint::Length(3),   // Input box
            Constraint::Length(1),   // Controls help
        ]
    } else {
        vec![
            Constraint::Length(3),   // Header
            Constraint::Min(10),     // Pet display (primary focus)
            Constraint::Length(5),   // Clock status
            Constraint::Length(8),   // Task list
            Constraint::Length(3),   // Controls help
        ]
    };

    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints(constraints)
        .split(area);

    // Header
    let header = Paragraph::new("DaemonFlow TUI")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
        .alignment(Alignment::Center)
        .block(Block::default().borders(Borders::BOTTOM));
    frame.render_widget(header, chunks[0]);

    // Pet display (new, primary visual focus)
    render_pet(app, frame, chunks[1]);

    // Clock status
    render_clock_status(app, frame, chunks[2]);

    // Task list or Dashboard
    if app.show_dashboard {
        render_dashboard(app, frame, chunks[3]);
    } else {
        render_tasks(app, frame, chunks[3]);
    }

    // Input box (only in AddingTask mode)
    if in_input_mode {
        render_input_box(app, frame, chunks[4]);

        // Controls for input mode
        let controls = "Enter: submit | Esc: cancel";
        let controls_widget = Paragraph::new(controls)
            .style(Style::default().fg(Color::Yellow))
            .alignment(Alignment::Center);
        frame.render_widget(controls_widget, chunks[5]);
    } else {
        // Normal mode controls
        let controls = if app.daemon_connected {
            if app.show_dashboard {
                "q: quit | d: tasks | b: toggle break | r: refresh"
            } else {
                "q: quit | d: dashboard | a: add task | b: break | j/k: scroll"
            }
        } else {
            "q: quit | Daemon not running - start with: daemonflow start"
        };
        let controls_widget = Paragraph::new(controls)
            .style(Style::default().fg(Color::DarkGray))
            .alignment(Alignment::Center);
        frame.render_widget(controls_widget, chunks[4]);
    }
}

fn render_clock_status(app: &App, frame: &mut Frame, area: Rect) {
    if !app.daemon_connected {
        let msg = Paragraph::new("Daemon not connected")
            .style(Style::default().fg(Color::Red))
            .alignment(Alignment::Center)
            .block(Block::default().borders(Borders::ALL).title("Status"));
        frame.render_widget(msg, area);
        return;
    }

    match &app.clock_state {
        Some(state) => {
            // Format earned time as MM:SS or -MM:SS
            let minutes = state.earned_seconds.abs() / 60;
            let seconds = state.earned_seconds.abs() % 60;
            let sign = if state.earned_seconds < 0 { "-" } else { "" };
            let time_str = format!("{}{:02}:{:02}", sign, minutes, seconds);

            // State color
            let state_color = match state.clock_state.as_str() {
                "working" => Color::Green,
                "break" => Color::Yellow,
                "overtime" => Color::Red,
                _ => Color::White,
            };

            // Build status line with streak info
            let status_parts = [
                format!("State: {}", state.clock_state.to_uppercase()),
                format!("Earned: {}", time_str),
            ];

            // Add streak info (subtle styling)
            let streak_color = if app.current_streak > 0 { Color::Green } else { Color::DarkGray };
            let streak_text = if app.current_streak == 1 {
                format!("Streak: {} day", app.current_streak)
            } else {
                format!("Streak: {} days", app.current_streak)
            };

            // Add deaths if any
            let deaths_text = if app.total_deaths > 0 {
                format!("Deaths: {}", app.total_deaths)
            } else {
                String::new()
            };

            // Create multi-line display
            let line1 = status_parts.join(" | ");
            let mut line2_parts = vec![streak_text];
            if app.longest_streak > app.current_streak {
                line2_parts.push(format!("(Best: {})", app.longest_streak));
            }
            if !deaths_text.is_empty() {
                line2_parts.push(deaths_text);
            }
            let line2 = line2_parts.join(" | ");

            // Use Line and Span for multi-color text
            use ratatui::text::{Line, Span};
            let text = vec![
                Line::from(Span::styled(line1, Style::default().fg(state_color))),
                Line::from(Span::styled(line2, Style::default().fg(streak_color))),
            ];

            let status = Paragraph::new(text)
                .alignment(Alignment::Center)
                .block(Block::default().borders(Borders::ALL).title("Clock"));
            frame.render_widget(status, area);
        }
        None => {
            let msg = match &app.last_error {
                Some(err) => format!("Error: {}", err),
                None => "Loading...".to_string(),
            };
            let widget = Paragraph::new(msg)
                .alignment(Alignment::Center)
                .block(Block::default().borders(Borders::ALL).title("Clock"));
            frame.render_widget(widget, area);
        }
    }
}

/// Render the global task list.
fn render_tasks(app: &App, frame: &mut Frame, area: Rect) {
    let block = Block::default()
        .title(" Tasks ")
        .borders(Borders::ALL)
        .border_style(Style::default().fg(Color::DarkGray));

    if app.tasks.is_empty() {
        let text = Paragraph::new("No tasks")
            .style(Style::default().fg(Color::DarkGray))
            .block(block);
        frame.render_widget(text, area);
        return;
    }

    let items: Vec<ListItem> = app.tasks.iter()
        .skip(app.task_scroll)
        .take(area.height.saturating_sub(2) as usize) // Account for borders
        .map(|task| {
            let checkbox = if task.completed { "[x]" } else { "[ ]" };
            let due = task.due_date.as_ref()
                .map(|d| {
                    // Just date part if longer than 10 chars
                    if d.len() >= 10 {
                        format!(" ({})", &d[..10])
                    } else {
                        format!(" ({})", d)
                    }
                })
                .unwrap_or_default();
            let text = format!("{} {} {}{}", checkbox, task.project_name, task.text, due);
            ListItem::new(text)
        })
        .collect();

    let list = List::new(items)
        .block(block)
        .style(Style::default().fg(Color::White));

    frame.render_widget(list, area);
}

/// Render the pet display with state-based coloring, level, and XP progress.
fn render_pet(app: &App, frame: &mut Frame, area: Rect) {
    use ratatui::text::{Line, Span};

    let art = app.pet.get_art();
    let color = match app.pet.get_state() {
        PetState::Healthy => Color::Green,
        PetState::Resting => Color::Cyan,
        PetState::Tired => Color::Yellow,
        PetState::Decaying => Color::Red,
        PetState::Dead => Color::DarkGray,
    };

    // Title with level and state
    let title = format!(" Pet Lv.{} ({}) ", app.level, app.pet.get_state().display_name());

    // Calculate XP progress bar (10 chars)
    let progress_pct = if app.experience_to_next > 0 {
        (app.experience as f64 / app.experience_to_next as f64 * 100.0).min(100.0)
    } else {
        0.0
    };
    let filled = ((progress_pct / 10.0).round() as usize).min(10);
    let empty = 10 - filled;
    let progress_bar = format!(
        "{}{}",
        "█".repeat(filled),
        "░".repeat(empty)
    );

    // Build multi-line content: art + XP bar
    let mut lines: Vec<Line> = art
        .lines()
        .map(|line| Line::from(Span::styled(line.to_string(), Style::default().fg(color))))
        .collect();

    // Add empty line and XP progress line
    lines.push(Line::from(""));
    lines.push(Line::from(vec![
        Span::styled("XP: ", Style::default().fg(Color::Cyan)),
        Span::styled(
            format!("{}/{} ", app.experience, app.experience_to_next),
            Style::default().fg(Color::White),
        ),
        Span::styled(
            format!("[{}]", progress_bar),
            Style::default().fg(Color::Green),
        ),
    ]));

    let widget = Paragraph::new(lines)
        .alignment(Alignment::Center)
        .block(Block::default().borders(Borders::ALL).title(title).title_style(Style::default().fg(Color::Cyan)));
    frame.render_widget(widget, area);
}

/// Render the task input box.
fn render_input_box(app: &App, frame: &mut Frame, area: Rect) {
    // Display input buffer with cursor indicator
    let input_text = format!("{}|", app.input_buffer);

    let block = Block::default()
        .title(" Add Task ")
        .borders(Borders::ALL)
        .border_style(Style::default().fg(Color::Yellow));

    let input_widget = Paragraph::new(input_text)
        .style(Style::default().fg(Color::White))
        .block(block);

    frame.render_widget(input_widget, area);
}

/// Format minutes as human-readable time (e.g., 135 -> "2h 15m")
fn format_time_mins(mins: i32) -> String {
    let hours = mins / 60;
    let minutes = mins % 60;
    if hours > 0 {
        format!("{}h {}m", hours, minutes)
    } else {
        format!("{}m", minutes)
    }
}

/// Render the dashboard panel with streak and weekly summary
fn render_dashboard(app: &App, frame: &mut Frame, area: Rect) {
    use ratatui::text::{Line, Span};

    let block = Block::default()
        .title(" Dashboard (d: tasks) ")
        .borders(Borders::ALL)
        .border_style(Style::default().fg(Color::Cyan));

    // Check if we have data
    let weekly = match &app.dashboard_weekly {
        Some(w) => w,
        None => {
            let text = Paragraph::new("Loading...")
                .style(Style::default().fg(Color::DarkGray))
                .block(block);
            frame.render_widget(text, area);
            return;
        }
    };

    // Get streak data (fallback to app's current_streak if dashboard_streak not available)
    let (current_streak, longest_streak) = match &app.dashboard_streak {
        Some(s) => (s.current_streak, s.longest_streak),
        None => (app.current_streak, app.longest_streak),
    };

    // Format week range (extract just date part from ISO dates)
    let week_start = if weekly.week_start.len() >= 10 {
        &weekly.week_start[5..10] // "MM-DD"
    } else {
        &weekly.week_start
    };
    let week_end = if weekly.week_end.len() >= 10 {
        &weekly.week_end[5..10]
    } else {
        &weekly.week_end
    };

    // Build lines for dashboard
    let mut lines: Vec<Line> = Vec::new();

    // Week range
    lines.push(Line::from(vec![
        Span::styled("Week: ", Style::default().fg(Color::Cyan)),
        Span::styled(format!("{} - {}", week_start, week_end), Style::default().fg(Color::White)),
    ]));

    // Streak info
    let streak_color = if current_streak > 0 { Color::Green } else { Color::DarkGray };
    let streak_text = if longest_streak > current_streak {
        format!("{} days (Best: {})", current_streak, longest_streak)
    } else {
        format!("{} days", current_streak)
    };
    lines.push(Line::from(vec![
        Span::styled("Streak: ", Style::default().fg(Color::Cyan)),
        Span::styled(streak_text, Style::default().fg(streak_color)),
    ]));

    // Empty line
    lines.push(Line::from(""));

    // This Week header
    lines.push(Line::from(Span::styled("This Week:", Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))));

    // Stats line 1: Tasks | Commits | XP
    lines.push(Line::from(vec![
        Span::styled("  Tasks: ", Style::default().fg(Color::DarkGray)),
        Span::styled(format!("{}", weekly.total_tasks_completed), Style::default().fg(Color::White)),
        Span::styled(" | Commits: ", Style::default().fg(Color::DarkGray)),
        Span::styled(format!("{}", weekly.total_commits), Style::default().fg(Color::White)),
        Span::styled(" | XP: ", Style::default().fg(Color::DarkGray)),
        Span::styled(format!("{}", weekly.total_xp_earned), Style::default().fg(Color::Green)),
    ]));

    // Stats line 2: Time earned/spent
    let time_earned = format_time_mins(weekly.total_time_earned_mins);
    let time_spent = format_time_mins(weekly.total_time_spent_mins);
    lines.push(Line::from(vec![
        Span::styled("  Time: +", Style::default().fg(Color::DarkGray)),
        Span::styled(time_earned, Style::default().fg(Color::Green)),
        Span::styled(" earned | ", Style::default().fg(Color::DarkGray)),
        Span::styled(time_spent, Style::default().fg(Color::Yellow)),
        Span::styled(" spent", Style::default().fg(Color::DarkGray)),
    ]));

    // Productive days
    lines.push(Line::from(vec![
        Span::styled("  Productive: ", Style::default().fg(Color::DarkGray)),
        Span::styled(format!("{}/7 days", weekly.productive_days), Style::default().fg(Color::White)),
    ]));

    let widget = Paragraph::new(lines)
        .block(block);

    frame.render_widget(widget, area);
}
