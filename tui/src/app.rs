use std::time::{Duration, Instant};
use anyhow::Result;
use crossterm::event::{self, Event, KeyCode, KeyEventKind};
use ratatui::prelude::*;

use crate::ipc::client::IpcClient;
use crate::ipc::protocol::{StateResponse, TaskData};
use crate::pet::Pet;

const UPDATE_INTERVAL: Duration = Duration::from_millis(500);

/// Input mode for the TUI
#[derive(PartialEq, Clone, Copy)]
pub enum InputMode {
    Normal,
    AddingTask,
}

pub struct App {
    pub ipc_client: IpcClient,
    pub daemon_connected: bool,
    pub clock_state: Option<StateResponse>,
    pub pet: Pet,
    pub last_error: Option<String>,
    pub last_update: Instant,
    pub should_quit: bool,
    // Streak tracking
    pub current_streak: i32,
    pub longest_streak: i32,
    pub total_deaths: i32,
    // Pet evolution
    pub level: i32,
    pub experience: i32,
    pub experience_to_next: i32,
    // Task list
    pub tasks: Vec<TaskData>,
    pub task_scroll: usize,
    // Input mode for task entry
    pub input_mode: InputMode,
    pub input_buffer: String,
}

impl App {
    pub fn new() -> Self {
        let ipc_client = IpcClient::new();
        let daemon_connected = ipc_client.is_daemon_running();

        let mut app = Self {
            ipc_client,
            daemon_connected,
            clock_state: None,
            pet: Pet::new(),
            last_error: None,
            last_update: Instant::now(),
            should_quit: false,
            current_streak: 0,
            longest_streak: 0,
            total_deaths: 0,
            level: 1,
            experience: 0,
            experience_to_next: 100,
            tasks: Vec::new(),
            task_scroll: 0,
            input_mode: InputMode::Normal,
            input_buffer: String::new(),
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
                // Update pet based on clock state
                self.pet.update(&state.clock_state, state.earned_seconds);
                // Extract streak info
                self.current_streak = state.current_streak;
                self.longest_streak = state.longest_streak;
                self.total_deaths = state.total_deaths;
                // Extract level/experience info
                self.level = state.level;
                self.experience = state.experience;
                self.experience_to_next = state.experience_to_next;
                // Sync pet's internal level for ASCII art progression
                self.pet.set_level(state.level);
                self.clock_state = Some(state);
                self.last_error = None;
                self.daemon_connected = true;
            }
            Err(e) => {
                self.last_error = Some(e.to_string());
                self.daemon_connected = false;
            }
        }
        // Also update tasks
        self.update_tasks();
        self.last_update = Instant::now();
    }

    pub fn update_tasks(&mut self) {
        if self.daemon_connected {
            match self.ipc_client.get_tasks(20) {
                Ok(response) => {
                    self.tasks = response.tasks;
                }
                Err(_) => {
                    // Don't clear tasks on error, just log
                    // Tasks are less critical than pet state
                }
            }
        }
    }

    pub fn toggle_break(&mut self) {
        if !self.daemon_connected {
            return;
        }

        // Don't allow ending break if pet is dead - must resurrect first
        if self.pet.get_state().is_dead() {
            self.last_error = Some("Pet is dead! Press 'x' to resurrect.".to_string());
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

    /// Submit the task in the input buffer via IPC
    fn submit_task(&mut self) {
        if self.input_buffer.is_empty() {
            return;
        }

        let text = self.input_buffer.clone();
        self.input_buffer.clear();

        // Use empty project_path to use daemon's watch directory
        match self.ipc_client.add_task("", &text) {
            Ok(resp) => {
                if resp.success {
                    // Trigger immediate task refresh to show new task
                    self.update_tasks();
                    self.last_error = None;
                } else {
                    self.last_error = Some(format!("Failed to add task: {}", resp.message));
                }
            }
            Err(e) => {
                self.last_error = Some(format!("IPC error: {}", e));
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
                        match self.input_mode {
                            InputMode::Normal => {
                                match key.code {
                                    KeyCode::Char('q') | KeyCode::Esc => {
                                        self.should_quit = true;
                                    }
                                    KeyCode::Char('a') => {
                                        // Enter task adding mode
                                        self.input_mode = InputMode::AddingTask;
                                        self.input_buffer.clear();
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
                                    KeyCode::Char('x') => {
                                        // Resurrect dead pet
                                        if self.daemon_connected && self.pet.get_state().is_dead() {
                                            match self.ipc_client.resurrect() {
                                                Ok(res) => {
                                                    if res.success {
                                                        self.last_error = Some("Pet resurrected! All earned time forfeited.".to_string());
                                                        self.update_state();
                                                    } else {
                                                        self.last_error = Some(res.message);
                                                    }
                                                }
                                                Err(e) => {
                                                    self.last_error = Some(e.to_string());
                                                }
                                            }
                                        }
                                    }
                                    KeyCode::Char('j') | KeyCode::Down => {
                                        // Scroll tasks down
                                        if !self.tasks.is_empty() && self.task_scroll < self.tasks.len().saturating_sub(1) {
                                            self.task_scroll += 1;
                                        }
                                    }
                                    KeyCode::Char('k') | KeyCode::Up => {
                                        // Scroll tasks up
                                        if self.task_scroll > 0 {
                                            self.task_scroll = self.task_scroll.saturating_sub(1);
                                        }
                                    }
                                    _ => {}
                                }
                            }
                            InputMode::AddingTask => {
                                match key.code {
                                    KeyCode::Esc => {
                                        // Cancel adding task
                                        self.input_mode = InputMode::Normal;
                                        self.input_buffer.clear();
                                    }
                                    KeyCode::Enter => {
                                        // Submit task if buffer is not empty
                                        if !self.input_buffer.is_empty() {
                                            self.submit_task();
                                        }
                                        self.input_mode = InputMode::Normal;
                                    }
                                    KeyCode::Backspace => {
                                        // Remove last character
                                        self.input_buffer.pop();
                                    }
                                    KeyCode::Char(c) => {
                                        // Add character to buffer (printable only)
                                        if !c.is_control() {
                                            self.input_buffer.push(c);
                                        }
                                    }
                                    _ => {}
                                }
                            }
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

#[cfg(test)]
mod tests {
    #[test]
    fn test_pet_level_matches_app_level() {
        // After update_state(), pet.level should match app.level
        // Note: Can't fully test without mock IPC, but verify Pet API works
        let mut pet = crate::pet::Pet::new();
        pet.set_level(3);
        assert_eq!(pet.get_level(), 3);

        // Verify level affects art selection
        let art_l1 = {
            let mut p = crate::pet::Pet::new();
            p.set_level(1);
            p.get_art()
        };
        let art_l3 = {
            let mut p = crate::pet::Pet::new();
            p.set_level(3);
            p.get_art()
        };
        assert_ne!(art_l1, art_l3, "Different levels should have different art");
    }
}

impl Default for App {
    fn default() -> Self {
        Self::new()
    }
}
