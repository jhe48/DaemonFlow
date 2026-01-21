//! Pet system module.
//!
//! This module contains the pet system implementation including:
//! - Pet state machine that maps to productivity clock state
//! - ASCII art representations for different pet states

pub mod art;
pub mod state;

pub use art::get_art;
pub use state::PetState;
