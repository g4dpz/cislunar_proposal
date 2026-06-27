//! Request handlers for the DTN Management API.
//!
//! Each handler function corresponds to one API endpoint and operates
//! on the shared `AppState` via axum's `State` extractor.

use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::Json;
use serde::Serialize;
use std::path::PathBuf;

use crate::adapter::traits::{ContactRef, NodeRef};
use crate::error::{AbstractionError, ErrorCategory, ValidationDetail};
use crate::lifecycle::state::EngineState;
use crate::model::{Contact, Neighbor, NetworkConfiguration};
use crate::validation;

use super::AppState;

/// Convert validation details into an AbstractionError for API response.
fn validation_error(details: Vec<ValidationDetail>) -> AbstractionError {
    let messages: Vec<String> = details.iter().map(|d| d.message.clone()).collect();
    let mut err = AbstractionError::new(
        ErrorCategory::ValidationError,
        format!("Configuration validation failed: {}", messages.join("; ")),
        "validate",
    );
    err.backend_code = Some(serde_json::to_string(&details).unwrap_or_default());
    err
}

// ─── Error Response Mapping ────────────────────────────────────────────────

/// Maps an `ErrorCategory` to the appropriate HTTP status code.
///
/// - ValidationError → 400
/// - UnsupportedOperation → 400
/// - ConfigurationError (not-found resource) → 404
/// - ConfigurationError (conflict) → 409
/// - LifecycleError → 500
/// - RuntimeError → 500
/// - CommunicationError → 502
pub fn error_to_status(error: &AbstractionError) -> StatusCode {
    match &error.category {
        ErrorCategory::ValidationError => StatusCode::BAD_REQUEST,
        ErrorCategory::UnsupportedOperation => StatusCode::BAD_REQUEST,
        ErrorCategory::ConfigurationError => {
            // If the error message indicates "not found", return 404
            if error.message.to_lowercase().contains("not found") {
                StatusCode::NOT_FOUND
            } else {
                StatusCode::CONFLICT
            }
        }
        ErrorCategory::LifecycleError => StatusCode::INTERNAL_SERVER_ERROR,
        ErrorCategory::RuntimeError => StatusCode::INTERNAL_SERVER_ERROR,
        ErrorCategory::CommunicationError => StatusCode::BAD_GATEWAY,
    }
}

/// Convert an AbstractionError into an axum HTTP response.
impl IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        let status = error_to_status(&self.0);
        let body = Json(&self.0);
        (status, body).into_response()
    }
}

/// Wrapper to make AbstractionError work with axum's IntoResponse.
pub struct ApiError(pub AbstractionError);

impl From<AbstractionError> for ApiError {
    fn from(err: AbstractionError) -> Self {
        ApiError(err)
    }
}

// ─── Configuration Endpoints ───────────────────────────────────────────────

/// POST /config — validate and store canonical configuration.
///
/// Accepts a JSON NetworkConfiguration, validates it, and stores it as the
/// current active configuration.
pub async fn post_config(
    State(state): State<AppState>,
    Json(config): Json<NetworkConfiguration>,
) -> Result<impl IntoResponse, ApiError> {
    // Run structural validation
    validation::validate(&config).map_err(|details| ApiError(validation_error(details)))?;

    // Store the configuration
    let mut current = state.config.write().await;
    *current = Some(config.clone());

    Ok((StatusCode::OK, Json(config)))
}

/// GET /config — retrieve the current canonical configuration.
///
/// Returns 404 if no configuration has been stored yet.
pub async fn get_config(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let current = state.config.read().await;
    match current.as_ref() {
        Some(config) => Ok(Json(config.clone())),
        None => Err(ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored. Use POST /config first.",
            "get_config",
        ))),
    }
}

/// POST /config/preview — generate backend config without deploying.
///
/// Validates the submitted config, looks up the target adapter, and generates
/// backend-specific configuration files. Returns the generated file contents
/// without writing them to disk.
pub async fn post_config_preview(
    State(state): State<AppState>,
    Json(config): Json<NetworkConfiguration>,
) -> Result<impl IntoResponse, ApiError> {
    // Validate
    validation::validate(&config).map_err(|details| ApiError(validation_error(details)))?;

    // Look up the target adapter
    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    // Backend-specific validation
    adapter.validate(&config).await.map_err(ApiError)?;

    // Generate config to a temporary path (preview only)
    let output_dir = PathBuf::from("/tmp/radiant-preview");
    let generated = adapter
        .generate_config(&config, &output_dir)
        .await
        .map_err(ApiError)?;

    Ok(Json(generated))
}

/// POST /config/deploy — validate, generate, and deploy backend configuration.
///
/// Performs full validation, generates backend-specific configs, and deploys
/// them to the target location.
pub async fn post_config_deploy(
    State(state): State<AppState>,
    Json(config): Json<NetworkConfiguration>,
) -> Result<impl IntoResponse, ApiError> {
    // Validate
    validation::validate(&config).map_err(|details| ApiError(validation_error(details)))?;

    // Look up the target adapter
    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    // Backend-specific validation
    adapter.validate(&config).await.map_err(ApiError)?;

    // Deploy
    let output_dir = PathBuf::from("/var/radiant/config");
    adapter
        .deploy(&config, &output_dir)
        .await
        .map_err(ApiError)?;

    // Store as current config
    let mut current = state.config.write().await;
    *current = Some(config);

    Ok(StatusCode::OK)
}

// ─── Lifecycle Endpoints ───────────────────────────────────────────────────

/// POST /lifecycle/start — start DTN engine.
pub async fn post_lifecycle_start(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored. Use POST /config first.",
            "start",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    // Transition state machine: Stopped → Starting
    {
        let mut sm = state.state_machine.write().await;
        sm.transition_to(EngineState::Starting)
            .map_err(ApiError)?;
    }

    // Start the engine
    let config_dir = PathBuf::from("/var/radiant/config");
    match adapter.start(&config_dir).await {
        Ok(()) => {
            let mut sm = state.state_machine.write().await;
            let _ = sm.transition_to(EngineState::Running);
            Ok(Json(StateResponse {
                state: EngineState::Running,
            }))
        }
        Err(e) => {
            let mut sm = state.state_machine.write().await;
            let _ = sm.transition_to_failed(e.message.clone());
            Err(ApiError(e))
        }
    }
}

/// POST /lifecycle/stop — stop DTN engine.
pub async fn post_lifecycle_stop(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored.",
            "stop",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    // Transition: Running → Stopping
    {
        let mut sm = state.state_machine.write().await;
        sm.transition_to(EngineState::Stopping)
            .map_err(ApiError)?;
    }

    match adapter.stop().await {
        Ok(()) => {
            let mut sm = state.state_machine.write().await;
            let _ = sm.transition_to(EngineState::Stopped);
            Ok(Json(StateResponse {
                state: EngineState::Stopped,
            }))
        }
        Err(e) => {
            // Can't transition to Failed from Stopping per state machine rules,
            // so just return the error
            Err(ApiError(e))
        }
    }
}

/// POST /lifecycle/restart — restart DTN engine.
pub async fn post_lifecycle_restart(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored.",
            "restart",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    let config_dir = PathBuf::from("/var/radiant/config");
    adapter
        .restart(&config_dir)
        .await
        .map_err(ApiError)?;

    Ok(Json(StateResponse {
        state: EngineState::Running,
    }))
}

/// GET /lifecycle/state — get current engine state.
pub async fn get_lifecycle_state(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let sm = state.state_machine.read().await;
    Json(StateResponse {
        state: sm.current(),
    })
}

/// GET /lifecycle/health — health check via backend adapter.
pub async fn get_lifecycle_health(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored.",
            "health",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    let health = adapter.health().await.map_err(ApiError)?;
    Ok(Json(health))
}

/// GET /lifecycle/version — engine version query.
pub async fn get_lifecycle_version(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored.",
            "version",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    let version = adapter.version().await.map_err(ApiError)?;
    Ok(Json(VersionResponse { version }))
}

// ─── Runtime Administration Endpoints ──────────────────────────────────────

/// POST /runtime/contacts — add a contact to the running engine.
pub async fn post_runtime_contact(
    State(state): State<AppState>,
    Json(contact): Json<Contact>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    adapter.add_contact(&contact).await.map_err(ApiError)?;
    Ok(StatusCode::OK)
}

/// DELETE /runtime/contacts — remove a contact from the running engine.
pub async fn delete_runtime_contact(
    State(state): State<AppState>,
    Json(contact_ref): Json<ContactRef>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    adapter
        .remove_contact(&contact_ref)
        .await
        .map_err(ApiError)?;
    Ok(StatusCode::OK)
}

/// POST /runtime/neighbors — add a neighbor to the running engine.
pub async fn post_runtime_neighbor(
    State(state): State<AppState>,
    Json(neighbor): Json<Neighbor>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    adapter
        .add_neighbor(&neighbor)
        .await
        .map_err(ApiError)?;
    Ok(StatusCode::OK)
}

/// DELETE /runtime/neighbors — remove a neighbor from the running engine.
pub async fn delete_runtime_neighbor(
    State(state): State<AppState>,
    Json(node_ref): Json<NodeRef>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    adapter
        .remove_neighbor(&node_ref)
        .await
        .map_err(ApiError)?;
    Ok(StatusCode::OK)
}

/// POST /runtime/links/:id/enable — enable a convergence layer link.
pub async fn post_link_enable(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    adapter.enable_link(&id).await.map_err(ApiError)?;
    Ok(StatusCode::OK)
}

/// POST /runtime/links/:id/disable — disable a convergence layer link.
pub async fn post_link_disable(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    adapter.disable_link(&id).await.map_err(ApiError)?;
    Ok(StatusCode::OK)
}

// ─── Monitoring Endpoints ──────────────────────────────────────────────────

/// GET /stats — bundle statistics from the running engine.
pub async fn get_stats(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    let stats = adapter.collect_stats().await.map_err(ApiError)?;
    Ok(Json(stats))
}

/// GET /stats/links — per-neighbor link state information.
pub async fn get_stats_links(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let adapter = get_active_adapter(&state).await?;
    let links = adapter.link_states().await.map_err(ApiError)?;
    Ok(Json(links))
}

/// GET /capabilities — backend capability set.
pub async fn get_capabilities(
    State(state): State<AppState>,
) -> Result<impl IntoResponse, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored. Use POST /config to set backend.",
            "capabilities",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    Ok(Json(adapter.capabilities().clone()))
}

/// GET /adapters — list all registered adapter names.
pub async fn get_adapters(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let names = state.registry.list().await;
    Json(AdaptersResponse { adapters: names })
}

// ─── Helper Functions ──────────────────────────────────────────────────────

/// Get the active backend adapter based on current configuration.
///
/// Returns an error if no configuration is stored.
async fn get_active_adapter(
    state: &AppState,
) -> Result<std::sync::Arc<dyn crate::adapter::traits::BackendAdapter>, ApiError> {
    let config_guard = state.config.read().await;
    let config = config_guard.as_ref().ok_or_else(|| {
        ApiError(AbstractionError::new(
            ErrorCategory::ConfigurationError,
            "No configuration stored.",
            "get_active_adapter",
        ))
    })?;

    let adapter = state
        .registry
        .get(&config.backend)
        .await
        .map_err(ApiError)?;

    Ok(adapter)
}

// ─── Response Types ────────────────────────────────────────────────────────

#[derive(Debug, Serialize)]
pub struct StateResponse {
    pub state: EngineState,
}

#[derive(Debug, Serialize)]
pub struct VersionResponse {
    pub version: String,
}

#[derive(Debug, Serialize)]
pub struct AdaptersResponse {
    pub adapters: Vec<String>,
}
