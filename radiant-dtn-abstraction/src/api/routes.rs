//! Axum route definitions for the DTN Management API.
//!
//! Defines the complete set of HTTP endpoints for configuration, lifecycle,
//! runtime administration, monitoring, and event streaming.

use axum::routing::{delete, get, post};
use axum::Router;

use super::handlers;
use super::sse;
use super::AppState;

/// Build the complete axum Router with all management API routes.
///
/// # Endpoints
///
/// ## Configuration
/// - `POST /config` — validate and store canonical config
/// - `GET /config` — retrieve current canonical config
/// - `POST /config/preview` — generate backend config without deploying
/// - `POST /config/deploy` — generate and deploy backend config
///
/// ## Lifecycle
/// - `POST /lifecycle/start` — start DTN engine
/// - `POST /lifecycle/stop` — stop DTN engine
/// - `POST /lifecycle/restart` — restart DTN engine
/// - `GET /lifecycle/state` — get current engine state
/// - `GET /lifecycle/health` — health check
/// - `GET /lifecycle/version` — engine version query
///
/// ## Runtime Administration (hot reconfiguration)
/// - `POST /runtime/contacts` — add contact
/// - `DELETE /runtime/contacts` — remove contact
/// - `POST /runtime/neighbors` — add neighbor
/// - `DELETE /runtime/neighbors` — remove neighbor
/// - `POST /runtime/links/:id/enable` — enable link
/// - `POST /runtime/links/:id/disable` — disable link
///
/// ## Monitoring
/// - `GET /stats` — bundle statistics
/// - `GET /stats/links` — per-neighbor link state
/// - `GET /capabilities` — backend capability set
/// - `GET /adapters` — list registered adapters
///
/// ## Events
/// - `GET /events` — Server-Sent Events stream
pub fn build_router(state: AppState) -> Router {
    Router::new()
        // Configuration endpoints
        .route("/config", post(handlers::post_config))
        .route("/config", get(handlers::get_config))
        .route("/config/preview", post(handlers::post_config_preview))
        .route("/config/deploy", post(handlers::post_config_deploy))
        // Lifecycle endpoints
        .route("/lifecycle/start", post(handlers::post_lifecycle_start))
        .route("/lifecycle/stop", post(handlers::post_lifecycle_stop))
        .route("/lifecycle/restart", post(handlers::post_lifecycle_restart))
        .route("/lifecycle/state", get(handlers::get_lifecycle_state))
        .route("/lifecycle/health", get(handlers::get_lifecycle_health))
        .route("/lifecycle/version", get(handlers::get_lifecycle_version))
        // Runtime administration endpoints
        .route("/runtime/contacts", post(handlers::post_runtime_contact))
        .route(
            "/runtime/contacts",
            delete(handlers::delete_runtime_contact),
        )
        .route("/runtime/neighbors", post(handlers::post_runtime_neighbor))
        .route(
            "/runtime/neighbors",
            delete(handlers::delete_runtime_neighbor),
        )
        .route(
            "/runtime/links/:id/enable",
            post(handlers::post_link_enable),
        )
        .route(
            "/runtime/links/:id/disable",
            post(handlers::post_link_disable),
        )
        // Monitoring endpoints
        .route("/stats", get(handlers::get_stats))
        .route("/stats/links", get(handlers::get_stats_links))
        .route("/capabilities", get(handlers::get_capabilities))
        .route("/adapters", get(handlers::get_adapters))
        // SSE event streaming
        .route("/events", get(sse::get_events))
        .with_state(state)
}
