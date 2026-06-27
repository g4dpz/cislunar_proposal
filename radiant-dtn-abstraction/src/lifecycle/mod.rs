//! Engine lifecycle state machine.
//!
//! Manages DTN engine state transitions: Stopped → Starting → Running → Stopping → Stopped,
//! with a Failed state for unexpected exits.

pub mod state;

pub use state::{EngineState, StateMachine};
