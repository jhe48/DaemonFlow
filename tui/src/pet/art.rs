//! ASCII art representations for pet states with level-based progression.
//!
//! Each pet state has a distinct visual representation:
//! - Healthy: Eyes open, upright, happy
//! - Resting: Eyes closed, relaxed, "zzz"
//! - Tired: Eyes droopy, worried
//! - Decaying: X eyes, slumped, concerning
//! - Dead: Gravestone (same across all levels)
//!
//! Level progression (cat theme, growing bigger/more detailed):
//! - Level 1: Tiny kitten (~4 lines)
//! - Level 2: Small cat (~5 lines, original design)
//! - Level 3: Adult cat with more detail (~6 lines)
//! - Level 4: Majestic cat with collar (~6 lines)
//! - Level 5: Royal cat with crown (~7 lines)

// =============================================================================
// LEVEL 1: Tiny Kitten (~4 lines)
// =============================================================================

/// Level 1 healthy: Tiny kitten, eyes open, happy.
pub const LEVEL_1_HEALTHY: &str = r#"
  /\_/\
 ( o.o )
  > ^ <
"#;

/// Level 1 resting: Tiny kitten sleeping.
pub const LEVEL_1_RESTING: &str = r#"
   zzz
  /\_/\
 ( -.- )
"#;

/// Level 1 tired: Tiny kitten worried.
pub const LEVEL_1_TIRED: &str = r#"
   !
  /\_/\
 ( o_o )
"#;

/// Level 1 decaying: Tiny kitten suffering.
pub const LEVEL_1_DECAYING: &str = r#"
  /\_/\
 ( x_x )
  > ~ <
"#;

// =============================================================================
// LEVEL 2: Small Cat (~5 lines, original design)
// =============================================================================

/// Level 2 healthy: Small cat, classic design.
pub const LEVEL_2_HEALTHY: &str = r#"
   /\_/\
  ( o.o )
   > ^ <
  /|   |\
  (_|   |_)
"#;

/// Level 2 resting: Small cat sleeping.
pub const LEVEL_2_RESTING: &str = r#"
   zzz
   /\_/\
  ( -.- )
   > ~ <
  /|   |\
  (_|   |_)
"#;

/// Level 2 tired: Small cat worried.
pub const LEVEL_2_TIRED: &str = r#"
  !
   /\_/\
  ( o_o )
   > ~ <
  /|   |\
  (_|   |_)
"#;

/// Level 2 decaying: Small cat suffering.
pub const LEVEL_2_DECAYING: &str = r#"
  ' /\_/\ '
  ( x_x )
   > ~ <
  /|   |\
  (_|___|_)
"#;

// =============================================================================
// LEVEL 3: Adult Cat (~6 lines, more detail)
// =============================================================================

/// Level 3 healthy: Adult cat with whiskers.
pub const LEVEL_3_HEALTHY: &str = r#"
    /\_/\
  =( o.o )=
    > ^ <
   /|   |\
  / |   | \
 (___|   |___)
"#;

/// Level 3 resting: Adult cat sleeping peacefully.
pub const LEVEL_3_RESTING: &str = r#"
    zzz
    /\_/\
  =( -.- )=
    > ~ <
   /|   |\
  (___|   |___)
"#;

/// Level 3 tired: Adult cat looking worried.
pub const LEVEL_3_TIRED: &str = r#"
   !
    /\_/\
  =( o_o )=
    > ~ <
   /|   |\
  (___|   |___)
"#;

/// Level 3 decaying: Adult cat suffering.
pub const LEVEL_3_DECAYING: &str = r#"
  '  /\_/\  '
  =( x_x )=
    > ~ <
   /|   |\
  (___|___|___)
"#;

// =============================================================================
// LEVEL 4: Majestic Cat with Collar (~6 lines)
// =============================================================================

/// Level 4 healthy: Majestic cat with collar.
pub const LEVEL_4_HEALTHY: &str = r#"
     /\_/\
   =( o.o )=
    [=^=^=]
    /|   |\
   / |   | \
  (___|   |___)
"#;

/// Level 4 resting: Majestic cat sleeping.
pub const LEVEL_4_RESTING: &str = r#"
     zzz
     /\_/\
   =( -.- )=
    [=^=^=]
   / |   | \
  (___|   |___)
"#;

/// Level 4 tired: Majestic cat worried.
pub const LEVEL_4_TIRED: &str = r#"
    !
     /\_/\
   =( o_o )=
    [=^=^=]
   / |   | \
  (___|   |___)
"#;

/// Level 4 decaying: Majestic cat suffering.
pub const LEVEL_4_DECAYING: &str = r#"
  '   /\_/\   '
   =( x_x )=
    [=~=~=]
   / |   | \
  (___|___|___)
"#;

// =============================================================================
// LEVEL 5: Royal Cat with Crown (~7 lines)
// =============================================================================

/// Level 5 healthy: Royal cat with crown.
pub const LEVEL_5_HEALTHY: &str = r#"
      /\
     /||\
     /\_/\
   =( o.o )=
    [=*=*=]
    /|   |\
   / |   | \
  (___|   |___)
"#;

/// Level 5 resting: Royal cat sleeping.
pub const LEVEL_5_RESTING: &str = r#"
     zzz
      /\
     /||\
     /\_/\
   =( -.- )=
    [=*=*=]
  (___|   |___)
"#;

/// Level 5 tired: Royal cat worried.
pub const LEVEL_5_TIRED: &str = r#"
    !
      /\
     /||\
     /\_/\
   =( o_o )=
    [=*=*=]
  (___|   |___)
"#;

/// Level 5 decaying: Royal cat suffering.
pub const LEVEL_5_DECAYING: &str = r#"
  '    /\    '
     /||\
     /\_/\
   =( x_x )=
    [=~=~=]
  (___|___|___)
"#;

// =============================================================================
// DEAD STATE: Gravestone (shared across all levels)
// =============================================================================

/// Dead state gravestone - same across all levels.
/// Death is death, no matter the level.
pub const DEAD: &str = r#"
   _____
  |     |
  | RIP |
  |     |
  |_____|
"#;

// =============================================================================
// LEGACY CONSTANTS (for backwards compatibility)
// =============================================================================

/// ASCII art for the healthy pet state (Level 2).
pub const HEALTHY: &str = LEVEL_2_HEALTHY;

/// ASCII art for the resting pet state (Level 2).
pub const RESTING: &str = LEVEL_2_RESTING;

/// ASCII art for the tired pet state (Level 2).
pub const TIRED: &str = LEVEL_2_TIRED;

/// ASCII art for the decaying pet state (Level 2).
pub const DECAYING: &str = LEVEL_2_DECAYING;

use crate::pet::state::PetState;

/// Returns the ASCII art for a given pet level and state.
///
/// # Arguments
/// * `level` - The pet level (1-5), clamped to valid range
/// * `state` - The PetState enum value
///
/// # Returns
/// The ASCII art string for the level and state combination.
/// Dead state returns gravestone regardless of level.
pub fn get_art_for_level_and_state(level: i32, state: &PetState) -> &'static str {
    // Dead state is always gravestone regardless of level
    if *state == PetState::Dead {
        return DEAD;
    }

    // Clamp level to 1-5 range
    let clamped_level = level.clamp(1, 5);

    match clamped_level {
        1 => match state {
            PetState::Healthy => LEVEL_1_HEALTHY,
            PetState::Resting => LEVEL_1_RESTING,
            PetState::Tired => LEVEL_1_TIRED,
            PetState::Decaying => LEVEL_1_DECAYING,
            PetState::Dead => DEAD, // Already handled above, but needed for exhaustive match
        },
        2 => match state {
            PetState::Healthy => LEVEL_2_HEALTHY,
            PetState::Resting => LEVEL_2_RESTING,
            PetState::Tired => LEVEL_2_TIRED,
            PetState::Decaying => LEVEL_2_DECAYING,
            PetState::Dead => DEAD,
        },
        3 => match state {
            PetState::Healthy => LEVEL_3_HEALTHY,
            PetState::Resting => LEVEL_3_RESTING,
            PetState::Tired => LEVEL_3_TIRED,
            PetState::Decaying => LEVEL_3_DECAYING,
            PetState::Dead => DEAD,
        },
        4 => match state {
            PetState::Healthy => LEVEL_4_HEALTHY,
            PetState::Resting => LEVEL_4_RESTING,
            PetState::Tired => LEVEL_4_TIRED,
            PetState::Decaying => LEVEL_4_DECAYING,
            PetState::Dead => DEAD,
        },
        5 => match state {
            PetState::Healthy => LEVEL_5_HEALTHY,
            PetState::Resting => LEVEL_5_RESTING,
            PetState::Tired => LEVEL_5_TIRED,
            PetState::Decaying => LEVEL_5_DECAYING,
            PetState::Dead => DEAD,
        },
        // This should never happen due to clamp, but needed for exhaustive match
        _ => match state {
            PetState::Healthy => LEVEL_2_HEALTHY,
            PetState::Resting => LEVEL_2_RESTING,
            PetState::Tired => LEVEL_2_TIRED,
            PetState::Decaying => LEVEL_2_DECAYING,
            PetState::Dead => DEAD,
        },
    }
}

/// Returns the ASCII art for a given pet state enum (Level 2 default).
///
/// # Arguments
/// * `state` - The PetState enum value
///
/// # Returns
/// The ASCII art string for the state at Level 2 (backwards compatible).
pub fn get_art_for_state(state: &PetState) -> &'static str {
    get_art_for_level_and_state(2, state)
}

/// Returns the ASCII art for a given pet state name.
///
/// # Arguments
/// * `state` - The pet state name (case-insensitive)
///
/// # Returns
/// The ASCII art string for the state, or HEALTHY as default for unknown states.
#[allow(dead_code)]
pub fn get_art(state: &str) -> &'static str {
    match state.to_lowercase().as_str() {
        "healthy" => HEALTHY,
        "resting" => RESTING,
        "tired" => TIRED,
        "decaying" => DECAYING,
        "dead" => DEAD,
        _ => HEALTHY, // Default to healthy for unknown states
    }
}
