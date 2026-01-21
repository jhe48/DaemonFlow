use ratatui::prelude::*;
use ratatui::widgets::{Block, Borders, Paragraph};
use crate::app::App;

pub fn render(app: &App, frame: &mut Frame) {
    let area = frame.area();

    // Main layout: vertical split
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),  // Header
            Constraint::Length(5),  // Clock status
            Constraint::Length(3),  // Controls help
            Constraint::Min(0),     // Spacer
        ])
        .split(area);

    // Header
    let header = Paragraph::new("DaemonFlow TUI")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
        .alignment(Alignment::Center)
        .block(Block::default().borders(Borders::BOTTOM));
    frame.render_widget(header, chunks[0]);

    // Clock status
    render_clock_status(app, frame, chunks[1]);

    // Controls
    let controls = if app.daemon_connected {
        "q: quit | b: toggle break | r: refresh"
    } else {
        "q: quit | Daemon not running - start with: daemonflow start"
    };
    let controls_widget = Paragraph::new(controls)
        .style(Style::default().fg(Color::DarkGray))
        .alignment(Alignment::Center);
    frame.render_widget(controls_widget, chunks[2]);
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

            let status_text = format!(
                "State: {} | Earned: {} | Session: {}s",
                state.clock_state.to_uppercase(),
                time_str,
                state.session_earned
            );

            let status = Paragraph::new(status_text)
                .style(Style::default().fg(state_color))
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
