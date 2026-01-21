use ratatui::prelude::*;
use ratatui::widgets::{Block, Borders, Paragraph};
use crate::app::App;
use crate::pet::{Pet, PetState};

pub fn render(app: &App, frame: &mut Frame) {
    let area = frame.area();

    // Main layout: vertical split
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),   // Header
            Constraint::Min(12),     // Pet display (primary focus)
            Constraint::Length(5),   // Clock status
            Constraint::Length(3),   // Controls help
        ])
        .split(area);

    // Header
    let header = Paragraph::new("DaemonFlow TUI")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
        .alignment(Alignment::Center)
        .block(Block::default().borders(Borders::BOTTOM));
    frame.render_widget(header, chunks[0]);

    // Pet display (new, primary visual focus)
    render_pet(&app.pet, frame, chunks[1]);

    // Clock status
    render_clock_status(app, frame, chunks[2]);

    // Controls
    let controls = if app.daemon_connected {
        "q: quit | b: toggle break | r: refresh"
    } else {
        "q: quit | Daemon not running - start with: daemonflow start"
    };
    let controls_widget = Paragraph::new(controls)
        .style(Style::default().fg(Color::DarkGray))
        .alignment(Alignment::Center);
    frame.render_widget(controls_widget, chunks[3]);
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

/// Render the pet display with state-based coloring.
fn render_pet(pet: &Pet, frame: &mut Frame, area: Rect) {
    let art = pet.get_art();
    let color = match pet.get_state() {
        PetState::Healthy => Color::Green,
        PetState::Resting => Color::Cyan,
        PetState::Tired => Color::Yellow,
        PetState::Decaying => Color::Red,
        PetState::Dead => Color::DarkGray,
    };

    let title = format!("Pet ({})", pet.get_state().display_name());
    let widget = Paragraph::new(art)
        .style(Style::default().fg(color))
        .alignment(Alignment::Center)
        .block(Block::default().borders(Borders::ALL).title(title));
    frame.render_widget(widget, area);
}
