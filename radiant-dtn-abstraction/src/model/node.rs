//! Node definition and identity types.
//!
//! Defines DTN node identity, endpoint identifiers (ipn:// and dtn:// schemes),
//! and service demultiplexing tokens.

use serde::{Deserialize, Serialize};

/// A DTN node definition in the canonical configuration model.
///
/// Each node has a unique numeric identifier used for ipn:// routing,
/// optional endpoint IDs, and a list of services it responds to.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct NodeDefinition {
    /// Unique numeric node identifier (used for ipn:// routing)
    pub node_number: u64,

    /// Primary endpoint ID (derived from node_number if omitted)
    #[serde(default)]
    pub endpoint_id: Option<EndpointId>,

    /// Callsign-based EID for amateur radio compliance
    /// Format: dtn://callsign-ssid/service
    #[serde(default)]
    pub callsign_eid: Option<EndpointId>,

    /// Human-readable name
    pub name: String,

    /// Service demux tokens this node responds to
    #[serde(default)]
    pub services: Vec<ServiceDemux>,
}

/// A DTN endpoint identifier supporting both ipn and dtn URI schemes.
///
/// Uses serde untagged representation so that:
/// - `{ "node_number": 10, "service_number": 0 }` deserializes as Ipn
/// - `{ "authority": "g4dpz-1", "path": "gs" }` deserializes as Dtn
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(untagged)]
pub enum EndpointId {
    /// ipn:node_number.service_number
    Ipn {
        node_number: u64,
        service_number: u64,
    },
    /// dtn://authority/path
    Dtn {
        authority: String,
        path: String,
    },
}

/// A service demultiplexing token that a node responds to.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ServiceDemux {
    pub service_number: u64,
    pub description: Option<String>,
}
