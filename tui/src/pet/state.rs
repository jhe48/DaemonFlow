/// Pet state that reflects the user's productivity status.
///
/// The pet's visual state maps to the freedom clock state:
/// - Healthy when actively working
/// - Resting during well-earned breaks
/// - Tired when break time is running low
/// - Decaying when in overtime (break exhausted)
/// - Dead after extended overtime
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PetState {
    /// Working state - pet is healthy and thriving
    Healthy,
    /// Break with plenty of time (>= 60s) - pet is resting comfortably
    Resting,
    /// Break with low time (< 60s) - pet is getting tired, warning to return
    Tired,
    /// Overtime - break exhausted, pet is suffering
    Decaying,
    /// Extended overtime (5+ minutes) - pet has died
    Dead,
}

/// Threshold in seconds for overtime before pet dies (5 minutes = 300 seconds)
pub const DEATH_THRESHOLD_SECONDS: i32 = -300;

/// Threshold in seconds for tired warning during break (60 seconds)
pub const TIRED_THRESHOLD_SECONDS: i32 = 60;

impl PetState {
    /// Determine pet state from clock state and earned seconds balance.
    ///
    /// # Arguments
    /// * `clock_state` - One of "working", "break", or "overtime"
    /// * `earned_seconds` - Current balance (positive = earned time, negative = overtime)
    ///
    /// # Returns
    /// The appropriate PetState based on the clock state and balance
    pub fn from_clock_state(clock_state: &str, earned_seconds: i32) -> Self {
        match clock_state {
            "working" => PetState::Healthy,
            "break" => {
                if earned_seconds >= TIRED_THRESHOLD_SECONDS {
                    PetState::Resting
                } else {
                    PetState::Tired
                }
            }
            "overtime" => {
                if earned_seconds <= DEATH_THRESHOLD_SECONDS {
                    PetState::Dead
                } else {
                    PetState::Decaying
                }
            }
            // Unknown state defaults to Healthy (fail-safe)
            _ => PetState::Healthy,
        }
    }

    /// Returns true if the pet is in a negative state (tired, decaying, or dead)
    pub fn is_warning(&self) -> bool {
        matches!(self, PetState::Tired | PetState::Decaying | PetState::Dead)
    }

    /// Returns true if the pet is dead
    pub fn is_dead(&self) -> bool {
        matches!(self, PetState::Dead)
    }

    /// Get a display name for the pet state
    pub fn display_name(&self) -> &'static str {
        match self {
            PetState::Healthy => "Healthy",
            PetState::Resting => "Resting",
            PetState::Tired => "Tired",
            PetState::Decaying => "Decaying",
            PetState::Dead => "Dead",
        }
    }
}

impl Default for PetState {
    fn default() -> Self {
        PetState::Healthy
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_working_state_is_healthy() {
        // Working state should always be Healthy regardless of balance
        assert_eq!(PetState::from_clock_state("working", 0), PetState::Healthy);
        assert_eq!(PetState::from_clock_state("working", 1000), PetState::Healthy);
        assert_eq!(PetState::from_clock_state("working", -100), PetState::Healthy);
    }

    #[test]
    fn test_break_with_plenty_time_is_resting() {
        // Break with >= 60 seconds should be Resting
        assert_eq!(PetState::from_clock_state("break", 60), PetState::Resting);
        assert_eq!(PetState::from_clock_state("break", 120), PetState::Resting);
        assert_eq!(PetState::from_clock_state("break", 1000), PetState::Resting);
    }

    #[test]
    fn test_break_with_low_time_is_tired() {
        // Break with < 60 seconds should be Tired
        assert_eq!(PetState::from_clock_state("break", 59), PetState::Tired);
        assert_eq!(PetState::from_clock_state("break", 30), PetState::Tired);
        assert_eq!(PetState::from_clock_state("break", 1), PetState::Tired);
        assert_eq!(PetState::from_clock_state("break", 0), PetState::Tired);
    }

    #[test]
    fn test_overtime_is_decaying() {
        // Overtime with > -300 seconds should be Decaying
        assert_eq!(PetState::from_clock_state("overtime", -1), PetState::Decaying);
        assert_eq!(PetState::from_clock_state("overtime", -100), PetState::Decaying);
        assert_eq!(PetState::from_clock_state("overtime", -299), PetState::Decaying);
    }

    #[test]
    fn test_extended_overtime_is_dead() {
        // Overtime with <= -300 seconds should be Dead
        assert_eq!(PetState::from_clock_state("overtime", -300), PetState::Dead);
        assert_eq!(PetState::from_clock_state("overtime", -500), PetState::Dead);
        assert_eq!(PetState::from_clock_state("overtime", -1000), PetState::Dead);
    }

    #[test]
    fn test_unknown_state_defaults_to_healthy() {
        // Unknown clock states should default to Healthy
        assert_eq!(PetState::from_clock_state("unknown", 0), PetState::Healthy);
        assert_eq!(PetState::from_clock_state("", 0), PetState::Healthy);
    }

    #[test]
    fn test_is_warning() {
        assert!(!PetState::Healthy.is_warning());
        assert!(!PetState::Resting.is_warning());
        assert!(PetState::Tired.is_warning());
        assert!(PetState::Decaying.is_warning());
        assert!(PetState::Dead.is_warning());
    }

    #[test]
    fn test_is_dead() {
        assert!(!PetState::Healthy.is_dead());
        assert!(!PetState::Resting.is_dead());
        assert!(!PetState::Tired.is_dead());
        assert!(!PetState::Decaying.is_dead());
        assert!(PetState::Dead.is_dead());
    }

    #[test]
    fn test_display_name() {
        assert_eq!(PetState::Healthy.display_name(), "Healthy");
        assert_eq!(PetState::Resting.display_name(), "Resting");
        assert_eq!(PetState::Tired.display_name(), "Tired");
        assert_eq!(PetState::Decaying.display_name(), "Decaying");
        assert_eq!(PetState::Dead.display_name(), "Dead");
    }

    #[test]
    fn test_default() {
        assert_eq!(PetState::default(), PetState::Healthy);
    }
}
