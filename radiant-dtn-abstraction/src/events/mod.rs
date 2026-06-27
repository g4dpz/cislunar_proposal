//! Event bus and subscription management.
//!
//! Provides a tokio::broadcast-based pub/sub system for operational events
//! including link state changes, contact plan updates, and engine state transitions.

pub mod bus;

pub use bus::{ContactAction, ContactSummary, DtnEvent, EventBus, LinkActivity};
