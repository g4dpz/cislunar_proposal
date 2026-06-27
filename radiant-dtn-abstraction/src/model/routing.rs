//! Routing configuration types.
//!
//! Defines routing strategy selection and static route entries
//! for the canonical DTN configuration model.

use serde::{Deserialize, Serialize};

/// Routing configuration for the DTN network.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct RoutingConfig {
    pub strategy: RoutingStrategy,
    #[serde(default)]
    pub static_routes: Vec<StaticRoute>,
}

/// Available routing strategies.
///
/// Serialized as lowercase strings (cgr, static, default).
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum RoutingStrategy {
    /// Contact Graph Routing using contact plan data
    Cgr,
    /// Static routing with explicit next-hop entries
    Static,
    /// Backend default routing behavior
    Default,
}

/// A static route entry mapping a destination to a next-hop node.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct StaticRoute {
    pub destination_node: u64,
    pub next_hop_node: u64,
}
