//! Capability set definitions for backend adapters.
//!
//! Declares what features each backend adapter supports, enabling the
//! abstraction layer to reject unsupported operations before dispatching
//! them to the engine.

use serde::{Deserialize, Serialize};

use crate::model::convergence::ConvergenceLayerType;
use crate::model::routing::RoutingStrategy;

/// Declarative feature support discovery for a backend adapter.
///
/// Returned by `BackendAdapter::capabilities()` to allow the abstraction
/// layer (and operators) to query what a backend supports without requiring
/// a running engine instance.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct CapabilitySet {
    /// Which hot reconfiguration operations are supported at runtime.
    pub hot_reconfig: HotReconfigCapabilities,
    /// Which convergence layer transport types the backend supports.
    pub convergence_layers: Vec<ConvergenceLayerType>,
    /// Which routing strategies the backend supports.
    pub routing_strategies: Vec<RoutingStrategy>,
    /// Security-related capabilities.
    pub security: SecurityCapabilities,
}

/// Hot reconfiguration capabilities — which runtime mutation operations
/// the backend supports without requiring a full restart.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct HotReconfigCapabilities {
    /// Can add contacts to the running contact plan.
    pub add_contact: bool,
    /// Can remove contacts from the running contact plan.
    pub remove_contact: bool,
    /// Can add neighbors at runtime.
    pub add_neighbor: bool,
    /// Can remove neighbors at runtime.
    pub remove_neighbor: bool,
    /// Can enable a convergence layer link at runtime.
    pub enable_link: bool,
    /// Can disable a convergence layer link at runtime.
    pub disable_link: bool,
}

/// Security capabilities of a backend adapter.
///
/// For amateur radio deployments, BPSec must be disabled. This struct
/// declares whether the backend supports security features (which would
/// only be used on non-amateur links).
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct SecurityCapabilities {
    /// Whether the backend supports BPSec integrity blocks (BIB).
    pub bpsec_bib: bool,
    /// Whether the backend supports BPSec confidentiality blocks (BCB).
    pub bpsec_bcb: bool,
}

impl HotReconfigCapabilities {
    /// Returns a capability set where all hot reconfiguration is supported.
    pub fn all() -> Self {
        Self {
            add_contact: true,
            remove_contact: true,
            add_neighbor: true,
            remove_neighbor: true,
            enable_link: true,
            disable_link: true,
        }
    }

    /// Returns a capability set where no hot reconfiguration is supported.
    pub fn none() -> Self {
        Self {
            add_contact: false,
            remove_contact: false,
            add_neighbor: false,
            remove_neighbor: false,
            enable_link: false,
            disable_link: false,
        }
    }
}

impl SecurityCapabilities {
    /// Returns a capability set with no security features (amateur radio compliant).
    pub fn none() -> Self {
        Self {
            bpsec_bib: false,
            bpsec_bcb: false,
        }
    }

    /// Returns a capability set with full BPSec support.
    pub fn full() -> Self {
        Self {
            bpsec_bib: true,
            bpsec_bcb: true,
        }
    }
}
