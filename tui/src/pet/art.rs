//! ASCII art representations for pet states.
//!
//! Each pet state has a distinct visual representation:
//! - Healthy: Eyes open, upright, happy
//! - Resting: Eyes closed, relaxed, "zzz"
//! - Tired: Eyes droopy, worried
//! - Decaying: X eyes, slumped, concerning
//! - Dead: Gravestone

/// ASCII art for the healthy pet state.
/// Eyes open, upright posture, happy expression.
pub const HEALTHY: &str = r#"
   /\_/\
  ( o.o )
   > ^ <
  /|   |\
 (_|   |_)
"#;

/// ASCII art for the resting pet state.
/// Eyes closed, relaxed, with "zzz" indicator.
pub const RESTING: &str = r#"
   /\_/\   zzz
  ( -.- )
   > ~ <
  /|   |\
 (_|   |_)
"#;

/// Returns the ASCII art for a given pet state.
///
/// # Arguments
/// * `state` - The pet state name (case-insensitive)
///
/// # Returns
/// The ASCII art string for the state, or HEALTHY as default for unknown states.
pub fn get_art(state: &str) -> &'static str {
    match state.to_lowercase().as_str() {
        "healthy" => HEALTHY,
        "resting" => RESTING,
        _ => HEALTHY, // Default to healthy for unknown states
    }
}
