//! Engine state machine and transitions.
//!
//! Tracks the lifecycle state of a DTN engine (Stopped, Starting, Running, Stopping, Failed)
//! and enforces legal state transitions.

use crate::error::{AbstractionError, ErrorCategory};

/// The possible states of a DTN engine.
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum EngineState {
    /// Engine is not running. Initial state.
    Stopped,
    /// Engine is in the process of starting up.
    Starting,
    /// Engine is running and healthy.
    Running,
    /// Engine is in the process of shutting down.
    Stopping,
    /// Engine has encountered a fatal error.
    Failed,
}

impl std::fmt::Display for EngineState {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EngineState::Stopped => write!(f, "Stopped"),
            EngineState::Starting => write!(f, "Starting"),
            EngineState::Running => write!(f, "Running"),
            EngineState::Stopping => write!(f, "Stopping"),
            EngineState::Failed => write!(f, "Failed"),
        }
    }
}

/// Manages the lifecycle state of a DTN engine, enforcing legal transitions
/// and recording failure reasons.
#[derive(Debug, Clone)]
pub struct StateMachine {
    state: EngineState,
    failure_reason: Option<String>,
}

impl StateMachine {
    /// Create a new state machine in the Stopped state.
    pub fn new() -> Self {
        Self {
            state: EngineState::Stopped,
            failure_reason: None,
        }
    }

    /// Returns the current engine state.
    pub fn current(&self) -> EngineState {
        self.state
    }

    /// Returns the recorded failure reason, if the engine is in the Failed state.
    pub fn failure_reason(&self) -> Option<&str> {
        self.failure_reason.as_deref()
    }

    /// Attempt to transition to a new state.
    ///
    /// Returns `Ok(())` if the transition is legal, or an error if the transition
    /// is not permitted from the current state.
    ///
    /// When transitioning to `Failed`, a reason must be provided via
    /// [`transition_to_failed`](Self::transition_to_failed) instead.
    pub fn transition_to(&mut self, new_state: EngineState) -> Result<(), AbstractionError> {
        if new_state == EngineState::Failed {
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                "Use transition_to_failed() to transition to the Failed state with a reason",
                "state_transition",
            ));
        }

        if !Self::is_valid_transition(self.state, new_state) {
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!(
                    "Illegal state transition: {} → {}",
                    self.state, new_state
                ),
                "state_transition",
            ));
        }

        self.state = new_state;

        // Clear failure reason when leaving Failed state
        if self.failure_reason.is_some() {
            self.failure_reason = None;
        }

        Ok(())
    }

    /// Transition to the Failed state with a reason describing the failure.
    ///
    /// Only legal from Starting or Running states.
    pub fn transition_to_failed(&mut self, reason: impl Into<String>) -> Result<(), AbstractionError> {
        if !Self::is_valid_transition(self.state, EngineState::Failed) {
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!(
                    "Illegal state transition: {} → Failed",
                    self.state
                ),
                "state_transition",
            ));
        }

        self.state = EngineState::Failed;
        self.failure_reason = Some(reason.into());
        Ok(())
    }

    /// Check whether a transition from `from` to `to` is legal.
    ///
    /// Legal transitions:
    /// - Stopped → Starting
    /// - Starting → Running
    /// - Starting → Failed
    /// - Running → Stopping
    /// - Running → Failed
    /// - Stopping → Stopped
    /// - Failed → Starting (retry)
    /// - Failed → Stopped (acknowledge/reset)
    pub fn is_valid_transition(from: EngineState, to: EngineState) -> bool {
        matches!(
            (from, to),
            (EngineState::Stopped, EngineState::Starting)
                | (EngineState::Starting, EngineState::Running)
                | (EngineState::Starting, EngineState::Failed)
                | (EngineState::Running, EngineState::Stopping)
                | (EngineState::Running, EngineState::Failed)
                | (EngineState::Stopping, EngineState::Stopped)
                | (EngineState::Failed, EngineState::Starting)
                | (EngineState::Failed, EngineState::Stopped)
        )
    }
}

impl Default for StateMachine {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // ─── Valid transition tests ───────────────────────────────────────────────

    #[test]
    fn test_stopped_to_starting() {
        let mut sm = StateMachine::new();
        assert_eq!(sm.current(), EngineState::Stopped);
        assert!(sm.transition_to(EngineState::Starting).is_ok());
        assert_eq!(sm.current(), EngineState::Starting);
    }

    #[test]
    fn test_starting_to_running() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        assert!(sm.transition_to(EngineState::Running).is_ok());
        assert_eq!(sm.current(), EngineState::Running);
    }

    #[test]
    fn test_running_to_stopping() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        assert!(sm.transition_to(EngineState::Stopping).is_ok());
        assert_eq!(sm.current(), EngineState::Stopping);
    }

    #[test]
    fn test_stopping_to_stopped() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        sm.transition_to(EngineState::Stopping).unwrap();
        assert!(sm.transition_to(EngineState::Stopped).is_ok());
        assert_eq!(sm.current(), EngineState::Stopped);
    }

    #[test]
    fn test_full_lifecycle_stopped_starting_running_stopping_stopped() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        sm.transition_to(EngineState::Stopping).unwrap();
        sm.transition_to(EngineState::Stopped).unwrap();
        assert_eq!(sm.current(), EngineState::Stopped);
    }

    // ─── Failed state transitions ────────────────────────────────────────────

    #[test]
    fn test_starting_to_failed() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        assert!(sm.transition_to_failed("startup timeout").is_ok());
        assert_eq!(sm.current(), EngineState::Failed);
        assert_eq!(sm.failure_reason(), Some("startup timeout"));
    }

    #[test]
    fn test_running_to_failed() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        assert!(sm.transition_to_failed("unexpected process exit").is_ok());
        assert_eq!(sm.current(), EngineState::Failed);
        assert_eq!(sm.failure_reason(), Some("unexpected process exit"));
    }

    #[test]
    fn test_failed_to_starting_retry() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to_failed("error").unwrap();
        assert!(sm.transition_to(EngineState::Starting).is_ok());
        assert_eq!(sm.current(), EngineState::Starting);
        // Failure reason should be cleared
        assert_eq!(sm.failure_reason(), None);
    }

    #[test]
    fn test_failed_to_stopped_acknowledge() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to_failed("error").unwrap();
        assert!(sm.transition_to(EngineState::Stopped).is_ok());
        assert_eq!(sm.current(), EngineState::Stopped);
        // Failure reason should be cleared
        assert_eq!(sm.failure_reason(), None);
    }

    #[test]
    fn test_failure_reason_recorded() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to_failed("ION daemon exited with code 1").unwrap();
        assert_eq!(sm.failure_reason(), Some("ION daemon exited with code 1"));
    }

    // ─── Illegal transition tests ────────────────────────────────────────────

    #[test]
    fn test_stopped_to_running_illegal() {
        let mut sm = StateMachine::new();
        let result = sm.transition_to(EngineState::Running);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Stopped);
    }

    #[test]
    fn test_stopped_to_stopping_illegal() {
        let mut sm = StateMachine::new();
        let result = sm.transition_to(EngineState::Stopping);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Stopped);
    }

    #[test]
    fn test_stopped_to_failed_illegal() {
        let mut sm = StateMachine::new();
        let result = sm.transition_to_failed("should not work");
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Stopped);
        assert_eq!(sm.failure_reason(), None);
    }

    #[test]
    fn test_running_to_starting_illegal() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        let result = sm.transition_to(EngineState::Starting);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Running);
    }

    #[test]
    fn test_starting_to_stopping_illegal() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        let result = sm.transition_to(EngineState::Stopping);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Starting);
    }

    #[test]
    fn test_stopping_to_running_illegal() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        sm.transition_to(EngineState::Stopping).unwrap();
        let result = sm.transition_to(EngineState::Running);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Stopping);
    }

    #[test]
    fn test_stopping_to_failed_illegal() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to(EngineState::Running).unwrap();
        sm.transition_to(EngineState::Stopping).unwrap();
        let result = sm.transition_to_failed("should not work");
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Stopping);
    }

    #[test]
    fn test_failed_to_running_illegal() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to_failed("error").unwrap();
        let result = sm.transition_to(EngineState::Running);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Failed);
    }

    #[test]
    fn test_failed_to_stopping_illegal() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        sm.transition_to_failed("error").unwrap();
        let result = sm.transition_to(EngineState::Stopping);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Failed);
    }

    // ─── Edge cases ──────────────────────────────────────────────────────────

    #[test]
    fn test_transition_to_failed_via_transition_to_rejected() {
        let mut sm = StateMachine::new();
        sm.transition_to(EngineState::Starting).unwrap();
        // Must use transition_to_failed() to go to Failed
        let result = sm.transition_to(EngineState::Failed);
        assert!(result.is_err());
        assert_eq!(sm.current(), EngineState::Starting);
    }

    #[test]
    fn test_default_impl() {
        let sm = StateMachine::default();
        assert_eq!(sm.current(), EngineState::Stopped);
        assert_eq!(sm.failure_reason(), None);
    }

    #[test]
    fn test_no_failure_reason_when_not_failed() {
        let mut sm = StateMachine::new();
        assert_eq!(sm.failure_reason(), None);
        sm.transition_to(EngineState::Starting).unwrap();
        assert_eq!(sm.failure_reason(), None);
        sm.transition_to(EngineState::Running).unwrap();
        assert_eq!(sm.failure_reason(), None);
    }

    #[test]
    fn test_error_contains_transition_info() {
        let mut sm = StateMachine::new();
        let err = sm.transition_to(EngineState::Running).unwrap_err();
        assert_eq!(err.category, ErrorCategory::LifecycleError);
        assert!(err.message.contains("Stopped"));
        assert!(err.message.contains("Running"));
    }

    #[test]
    fn test_is_valid_transition_static() {
        // Valid transitions
        assert!(StateMachine::is_valid_transition(EngineState::Stopped, EngineState::Starting));
        assert!(StateMachine::is_valid_transition(EngineState::Starting, EngineState::Running));
        assert!(StateMachine::is_valid_transition(EngineState::Starting, EngineState::Failed));
        assert!(StateMachine::is_valid_transition(EngineState::Running, EngineState::Stopping));
        assert!(StateMachine::is_valid_transition(EngineState::Running, EngineState::Failed));
        assert!(StateMachine::is_valid_transition(EngineState::Stopping, EngineState::Stopped));
        assert!(StateMachine::is_valid_transition(EngineState::Failed, EngineState::Starting));
        assert!(StateMachine::is_valid_transition(EngineState::Failed, EngineState::Stopped));

        // Invalid transitions
        assert!(!StateMachine::is_valid_transition(EngineState::Stopped, EngineState::Running));
        assert!(!StateMachine::is_valid_transition(EngineState::Stopped, EngineState::Stopping));
        assert!(!StateMachine::is_valid_transition(EngineState::Stopped, EngineState::Failed));
        assert!(!StateMachine::is_valid_transition(EngineState::Running, EngineState::Starting));
        assert!(!StateMachine::is_valid_transition(EngineState::Starting, EngineState::Stopping));
        assert!(!StateMachine::is_valid_transition(EngineState::Stopping, EngineState::Running));
        assert!(!StateMachine::is_valid_transition(EngineState::Stopping, EngineState::Failed));
        assert!(!StateMachine::is_valid_transition(EngineState::Failed, EngineState::Running));
        assert!(!StateMachine::is_valid_transition(EngineState::Failed, EngineState::Stopping));
    }
}
