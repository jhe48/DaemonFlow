//! Pet system module.
//!
//! This module contains the pet system implementation including:
//! - Pet state machine that maps to productivity clock state
//! - ASCII art representations for different pet states
//! - Pet struct that combines state and art selection

pub mod art;
pub mod state;

pub use art::get_art_for_state;
pub use state::PetState;

/// Pet struct that combines state machine and art selection.
///
/// The Pet is the visual centerpiece of the TUI, updating its appearance
/// based on the daemon's clock state in real-time.
pub struct Pet {
    state: PetState,
}

impl Pet {
    /// Create a new Pet with default Healthy state.
    pub fn new() -> Self {
        Self { state: PetState::Healthy }
    }

    /// Update the pet's state based on clock state and earned seconds.
    ///
    /// # Arguments
    /// * `clock_state` - One of "working", "break", or "overtime"
    /// * `earned_seconds` - Current balance (positive = earned time, negative = overtime)
    pub fn update(&mut self, clock_state: &str, earned_seconds: i32) {
        self.state = PetState::from_clock_state(clock_state, earned_seconds);
    }

    /// Get a reference to the current pet state.
    pub fn get_state(&self) -> &PetState {
        &self.state
    }

    /// Get the ASCII art for the current pet state.
    pub fn get_art(&self) -> &'static str {
        art::get_art_for_state(&self.state)
    }
}

impl Default for Pet {
    fn default() -> Self {
        Self::new()
    }
}
