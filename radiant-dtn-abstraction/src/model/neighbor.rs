//! Neighbor and convergence layer link types.
//!
//! Defines the canonical representation of DTN neighbors and their
//! transport-layer link configurations (LTP/UDP, TCP-CL, KISS, UDP).

use serde::{Deserialize, Serialize};

/// A directly reachable DTN neighbor node.
///
/// Each neighbor is identified by the remote node's node_number and
/// connected via one or more convergence layer links.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Neighbor {
    /// Reference to the remote node (by node_number)
    pub node_number: u64,

    /// Human-readable name for the neighbor
    pub name: Option<String>,

    /// Convergence layer links to this neighbor
    pub links: Vec<ConvergenceLayerLink>,

    /// Optional rate limiting (bits per second)
    pub rate_limit_bps: Option<u64>,
}

/// A convergence layer link definition.
///
/// Tagged enum (serde tag="type") where each variant carries the
/// transport-specific parameters for a particular CL protocol.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(tag = "type")]
pub enum ConvergenceLayerLink {
    /// LTP over UDP transport
    #[serde(rename = "ltp_udp")]
    LtpUdp {
        id: String,
        local_engine_id: u64,
        remote_engine_id: u64,
        remote_host: String,
        remote_port: u16,
        local_port: u16,
        mtu: Option<u32>,
        segment_rate: Option<u32>,
    },
    /// TCP Convergence Layer (TCPCLv4)
    #[serde(rename = "tcpcl")]
    TcpCl {
        id: String,
        remote_host: String,
        remote_port: u16,
        local_port: Option<u16>,
        keepalive_interval_secs: Option<u32>,
    },
    /// KISS framing over TNC for amateur radio
    #[serde(rename = "kiss")]
    Kiss {
        id: String,
        tnc_device: String,
        baud_rate: u32,
        local_engine_id: u64,
        remote_engine_id: u64,
        frame_size: Option<u32>,
    },
    /// Plain UDP transport
    #[serde(rename = "udp")]
    Udp {
        id: String,
        remote_host: String,
        remote_port: u16,
        local_port: Option<u16>,
    },
}
