//! Pet system module.
//!
//! This module contains the pet system implementation including:
//! - Pet state machine that maps to productivity clock state
//! - ASCII art representations for different pet states with level progression
//! - Pet struct that combines state, level, and art selection

pub mod art;
pub mod state;

pub use art::get_art_for_state;
pub use art::get_art_for_level_and_state;
pub use state::PetState;

/// Pet struct that combines state machine, level, and art selection.
///
/// The Pet is the visual centerpiece of the TUI, updating its appearance
/// based on the daemon's clock state in real-time. As the pet levels up,
/// it visually evolves with more detailed ASCII art.
pub struct Pet {
    state: PetState,
    level: i32,
}

impl Pet {
    /// Create a new Pet with default Healthy state at level 1.
    pub fn new() -> Self {
        Self {
            state: PetState::Healthy,
            level: 1,
        }
    }

    /// Update the pet's state based on clock state and earned seconds.
    ///
    /// # Arguments
    /// * `clock_state` - One of "working", "break", or "overtime"
    /// * `earned_seconds` - Current balance (positive = earned time, negative = overtime)
    pub fn update(&mut self, clock_state: &str, earned_seconds: i32) {
        self.state = PetState::from_clock_state(clock_state, earned_seconds);
    }

    /// Set the pet's level (1-5).
    ///
    /// The level is automatically clamped to the valid range.
    /// Higher levels result in more detailed ASCII art.
    ///
    /// # Arguments
    /// * `level` - The new level (will be clamped to 1-5)
    pub fn set_level(&mut self, level: i32) {
        self.level = level.clamp(1, 5);
    }

    /// Get the pet's current level.
    pub fn get_level(&self) -> i32 {
        self.level
    }

    /// Get a reference to the current pet state.
    pub fn get_state(&self) -> &PetState {
        &self.state
    }

    /// Get the ASCII art for the current pet state and level.
    pub fn get_art(&self) -> &'static str {
        art::get_art_for_level_and_state(self.level, &self.state)
    }
}

impl Default for Pet {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_pet_starts_at_level_1() {
        let pet = Pet::new();
        assert_eq!(pet.get_level(), 1);
    }

    #[test]
    fn test_set_level_clamps_to_valid_range() {
        let mut pet = Pet::new();

        pet.set_level(3);
        assert_eq!(pet.get_level(), 3);

        pet.set_level(0);
        assert_eq!(pet.get_level(), 1);

        pet.set_level(10);
        assert_eq!(pet.get_level(), 5);

        pet.set_level(-5);
        assert_eq!(pet.get_level(), 1);
    }

    #[test]
    fn test_get_art_uses_level() {
        let mut pet = Pet::new();

        // Level 1 healthy art
        let art_l1 = pet.get_art();

        pet.set_level(5);
        let art_l5 = pet.get_art();

        // Different levels should produce different art
        assert_ne!(art_l1, art_l5);
    }

    #[test]
    fn test_dead_state_same_across_levels() {
        let mut pet = Pet::new();
        pet.update("overtime", -500); // Dead state

        let art_l1 = pet.get_art();

        pet.set_level(5);
        let art_l5 = pet.get_art();

        // Dead art should be the same regardless of level
        assert_eq!(art_l1, art_l5);
    }
}
