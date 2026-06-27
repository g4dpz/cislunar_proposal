//! Top-level network configuration type.
//!
//! Defines the complete canonical configuration document that encompasses
//! all network, operational, and backend-specific settings.

use std::collections::HashMap;

use serde::{Deserialize, Serialize};

use super::contact_plan::ContactPlan;
use super::neighbor::Neighbor;
use super::node::NodeDefinition;
use super::routing::RoutingConfig;

/// Top-level canonical DTN network configuration.
///
/// This is the primary configuration document that operators provide.
/// It is serializable to/from YAML and JSON, and is used by backend
/// adapters to generate implementation-specific configuration artifacts.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct NetworkConfiguration {
    /// Schema version for forward compatibility
    pub version: String,

    /// Target backend adapter name (e.g., "ion-dtn", "hardy")
    pub backend: String,

    /// Local node definition
    pub local_node: NodeDefinition,

    /// Neighbor nodes and their convergence layer links
    pub neighbors: Vec<Neighbor>,

    /// Contact plan (scheduled contacts and ranges)
    pub contact_plan: ContactPlan,

    /// Routing configuration
    pub routing: RoutingConfig,

    /// Security configuration (optional)
    /// Per amateur radio regulations: no BPSec over amateur links.
    pub security: Option<SecurityConfig>,

    /// Storage configuration (optional)
    pub storage: Option<StorageConfig>,

    /// Backend-specific overrides (opaque key-value pairs)
    #[serde(default)]
    pub backend_options: HashMap<String, serde_yaml::Value>,
}

/// Security configuration.
///
/// For amateur radio deployments, security (BPSec) is disabled.
/// This field exists to support non-amateur deployments where
/// encryption and integrity blocks may be used.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct SecurityConfig {
    /// Whether BPSec is enabled (must be false for amateur radio links)
    pub enabled: bool,
}

/// Storage configuration for bundle persistence.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct StorageConfig {
    /// Path to bundle storage directory
    pub path: String,
    /// Maximum storage in bytes
    pub max_bytes: Option<u64>,
}
