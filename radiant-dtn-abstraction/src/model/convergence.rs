//! Convergence layer type enumeration.
//!
//! Defines the supported convergence layer transport types for
//! capability discovery and configuration validation.

use serde::{Deserialize, Serialize};

/// Supported convergence layer transport types.
///
/// Used in capability sets to indicate which CL protocols a
/// backend adapter supports.
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub enum ConvergenceLayerType {
    /// LTP over UDP
    LtpUdp,
    /// TCP Convergence Layer (TCPCLv4)
    TcpCl,
    /// KISS framing over TNC (amateur radio)
    Kiss,
    /// Plain UDP
    Udp,
}
