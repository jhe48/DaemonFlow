use std::time::{Duration, Instant};
use anyhow::Result;
use crossterm::event::{self, Event, KeyCode, KeyEventKind};
use ratatui::prelude::*;

use crate::ipc::client::IpcClient;
use crate::ipc::protocol::StateResponse;

const UPDATE_INTERVAL: Duration = Duration::from_millis(500);

pub struct App {
    pub ipc_client: IpcClient,
    pub daemon_connected: bool,
    pub clock_state: Option<StateResponse>,
    pub last_error: Option<String>,
    pub last_update: Instant,
    pub should_quit: bool,
}

impl App {
    pub fn new() -> Self {
        let ipc_client = IpcClient::new();
        let daemon_connected = ipc_client.is_daemon_running();

        let mut app = Self {
            ipc_client,
            daemon_connected,
            clock_state: None,
            last_error: None,
            last_update: Instant::now(),
            should_quit: false,
        };

        // If connected, get initial state
        if app.daemon_connected {
            app.update_state();
        }

        app
    }

    pub fn update_state(&mut self) {
        match self.ipc_client.get_state() {
            Ok(state) => {
                self.clock_state = Some(state);
                self.last_error = None;
                self.daemon_connected = true;
            }
            Err(e) => {
                self.last_error = Some(e.to_string());
                self.daemon_connected = false;
            }
        }
        self.last_update = Instant::now();
    }

    pub fn toggle_break(&mut self) {
        if !self.daemon_connected {
            return;
        }

        if let Some(state) = &self.clock_state {
            let result = if state.clock_state == "working" {
                self.ipc_client.start_break()
            } else {
                self.ipc_client.end_break()
            };

            match result {
                Ok(_) => {
                    // Immediately refresh to show new state
                    self.update_state();
                }
                Err(e) => {
                    self.last_error = Some(e.to_string());
                }
            }
        }
    }

    pub fn run(&mut self, terminal: &mut Terminal<impl Backend>) -> Result<()> {
        while !self.should_quit {
            // Check if we need to update state
            if self.last_update.elapsed() >= UPDATE_INTERVAL && self.daemon_connected {
                self.update_state();
            }

            // Draw UI
            terminal.draw(|frame| self.render(frame))?;

            // Handle events with timeout
            if event::poll(Duration::from_millis(100))? {
                if let Event::Key(key) = event::read()? {
                    if key.kind == KeyEventKind::Press {
                        match key.code {
                            KeyCode::Char('q') | KeyCode::Esc => {
                                self.should_quit = true;
                            }
                            KeyCode::Char('b') => {
                                if self.daemon_connected {
                                    self.toggle_break();
                                }
                            }
                            KeyCode::Char('r') => {
                                // Force refresh - also try to reconnect if disconnected
                                if !self.daemon_connected {
                                    self.daemon_connected = self.ipc_client.is_daemon_running();
                                }
                                if self.daemon_connected {
                                    self.update_state();
                                }
                            }
                            _ => {}
                        }
                    }
                }
            }
        }
        Ok(())
    }

    fn render(&self, frame: &mut Frame) {
        crate::ui::render(self, frame);
    }
}

impl Default for App {
    fn default() -> Self {
        Self::new()
    }
}
