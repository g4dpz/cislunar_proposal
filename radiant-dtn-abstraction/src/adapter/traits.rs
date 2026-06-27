//! Backend adapter trait definition.
//!
//! The `BackendAdapter` trait is the central abstraction point for the DTN
//! abstraction layer. Each DTN engine (ION-DTN, Hardy, etc.) implements this
//! trait to provide configuration generation, lifecycle management,
//! hot reconfiguration, and telemetry collection.

use std::collections::HashMap;
use std::path::Path;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};

use crate::error::AbstractionError;
use crate::model::{Contact, Neighbor, NetworkConfiguration};

use super::capability::CapabilitySet;

/// Reference to a contact for removal operations.
///
/// Identifies a specific contact by its source node, destination node,
/// and start time — the minimal unique key for a contact.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ContactRef {
    pub source_node: u64,
    pub dest_node: u64,
    /// Unix timestamp (seconds) identifying the contact start
    pub start_time: i64,
}

/// Reference to a node for neighbor removal operations.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct NodeRef {
    pub node_number: u64,
}

/// Represents the generated backend-specific configuration files.
///
/// Maps filenames to their content strings (e.g., "node10.ionrc" -> ionadmin commands).
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct GeneratedConfig {
    /// Mapping of filename to file content
    pub files: HashMap<String, String>,
}

/// Health status of a running DTN engine.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct HealthStatus {
    /// Whether the engine process is running
    pub running: bool,
    /// Uptime in seconds (None if not running)
    pub uptime_secs: Option<u64>,
    /// Optional human-readable status message
    pub message: Option<String>,
}

/// Aggregate bundle statistics collected from a DTN engine.
#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
pub struct BundleStatistics {
    pub bundles_sourced: u64,
    pub bundles_forwarded: u64,
    pub bundles_delivered: u64,
    pub bundles_expired: u64,
    pub bundles_queued: u64,
}

/// Per-neighbor link state information.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct LinkState {
    pub neighbor_node: u64,
    pub link_id: String,
    pub active: bool,
    pub bytes_sent: u64,
    pub bytes_received: u64,
}

/// The central backend adapter trait.
///
/// Each DTN engine implements this trait to provide:
/// - Configuration validation and generation
/// - Lifecycle management (start/stop/restart/health)
/// - Hot reconfiguration (add/remove contacts, neighbors, links)
/// - Telemetry collection (bundle stats, link states)
///
/// All methods are async and the trait is object-safe (`Send + Sync`).
#[async_trait]
pub trait BackendAdapter: Send + Sync {
    /// Human-readable adapter name (e.g., "ion-dtn", "hardy").
    fn name(&self) -> &str;

    /// Query supported capabilities without requiring a running engine.
    fn capabilities(&self) -> &CapabilitySet;

    /// Validate canonical config against backend-specific constraints.
    ///
    /// Returns Ok(()) if the configuration is valid for this backend,
    /// or an error describing what is incompatible.
    async fn validate(&self, config: &NetworkConfiguration) -> Result<(), AbstractionError>;

    /// Generate backend-specific configuration artifacts.
    ///
    /// Produces files (e.g., .ionrc, .bprc for ION; YAML for Hardy) from
    /// the canonical configuration. Files are written to `output_dir`.
    async fn generate_config(
        &self,
        config: &NetworkConfiguration,
        output_dir: &Path,
    ) -> Result<GeneratedConfig, AbstractionError>;

    /// Deploy generated configuration (write files, set permissions).
    ///
    /// This writes the generated config files to the target location
    /// and prepares the engine for startup.
    async fn deploy(
        &self,
        config: &NetworkConfiguration,
        output_dir: &Path,
    ) -> Result<(), AbstractionError>;

    // ─── Lifecycle operations ───────────────────────────────────────────

    /// Start the DTN engine using configuration from `config_dir`.
    async fn start(&self, config_dir: &Path) -> Result<(), AbstractionError>;

    /// Stop the DTN engine gracefully.
    async fn stop(&self) -> Result<(), AbstractionError>;

    /// Restart the DTN engine (stop then start with config from `config_dir`).
    async fn restart(&self, config_dir: &Path) -> Result<(), AbstractionError>;

    /// Query the health status of the DTN engine.
    async fn health(&self) -> Result<HealthStatus, AbstractionError>;

    /// Query the version string of the DTN engine.
    async fn version(&self) -> Result<String, AbstractionError>;

    // ─── Hot reconfiguration operations ─────────────────────────────────

    /// Add a contact to the running engine's contact plan.
    async fn add_contact(&self, contact: &Contact) -> Result<(), AbstractionError>;

    /// Remove a contact from the running engine's contact plan.
    async fn remove_contact(&self, contact: &ContactRef) -> Result<(), AbstractionError>;

    /// Add a neighbor to the running engine.
    async fn add_neighbor(&self, neighbor: &Neighbor) -> Result<(), AbstractionError>;

    /// Remove a neighbor from the running engine.
    async fn remove_neighbor(&self, node_ref: &NodeRef) -> Result<(), AbstractionError>;

    /// Enable a convergence layer link by its identifier.
    async fn enable_link(&self, link_id: &str) -> Result<(), AbstractionError>;

    /// Disable a convergence layer link by its identifier.
    async fn disable_link(&self, link_id: &str) -> Result<(), AbstractionError>;

    // ─── Telemetry ──────────────────────────────────────────────────────

    /// Collect aggregate bundle statistics from the engine.
    async fn collect_stats(&self) -> Result<BundleStatistics, AbstractionError>;

    /// Query per-neighbor link states from the engine.
    async fn link_states(&self) -> Result<Vec<LinkState>, AbstractionError>;
}
