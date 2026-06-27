//! Contact plan types.
//!
//! Defines the schedule of communication opportunities (contacts) and
//! propagation delays (ranges) between DTN nodes.

use serde::{Deserialize, Serialize};

/// A complete contact plan containing scheduled contacts and range data.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ContactPlan {
    pub contacts: Vec<Contact>,
    pub ranges: Vec<Range>,
}

/// A scheduled communication opportunity between two nodes.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Contact {
    pub source_node: u64,
    pub dest_node: u64,
    /// Unix timestamp (seconds)
    pub start_time: i64,
    /// Unix timestamp (seconds)
    pub end_time: i64,
    /// Data rate in bits per second
    pub rate_bps: u64,
    /// Confidence value [0.0, 1.0]
    #[serde(default = "default_confidence")]
    pub confidence: f64,
}

/// Propagation delay information between two nodes.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Range {
    pub source_node: u64,
    pub dest_node: u64,
    /// One-way light time in seconds
    pub owlt_secs: f64,
}

/// Default confidence value for contacts (1.0 = certain).
fn default_confidence() -> f64 {
    1.0
}
