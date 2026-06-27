//! HTTP/JSON Management API.
//!
//! Exposes axum-based HTTP endpoints for configuration, lifecycle management,
//! runtime administration, monitoring, and Server-Sent Events for notifications.

pub mod handlers;
pub mod routes;
pub mod sse;

use std::sync::Arc;

use tokio::sync::RwLock;

use crate::adapter::registry::AdapterRegistry;
use crate::events::bus::EventBus;
use crate::lifecycle::state::StateMachine;
use crate::model::NetworkConfiguration;

/// Shared application state for all API handlers.
///
/// Wrapped in `Arc<RwLock<_>>` and passed to axum handlers via the `State` extractor.
#[derive(Clone)]
pub struct AppState {
    /// Registry of available backend adapters.
    pub registry: Arc<AdapterRegistry>,
    /// Currently stored canonical configuration (None until first POST /config).
    pub config: Arc<RwLock<Option<NetworkConfiguration>>>,
    /// Event bus for SSE streaming.
    pub event_bus: Arc<EventBus>,
    /// Engine lifecycle state machine.
    pub state_machine: Arc<RwLock<StateMachine>>,
}

impl AppState {
    /// Create a new AppState with empty configuration and stopped engine.
    pub fn new(registry: Arc<AdapterRegistry>, event_bus: Arc<EventBus>) -> Self {
        Self {
            registry,
            config: Arc::new(RwLock::new(None)),
            event_bus,
            state_machine: Arc::new(RwLock::new(StateMachine::new())),
        }
    }
}
