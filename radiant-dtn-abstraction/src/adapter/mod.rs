//! Backend adapter trait and registry.
//!
//! Each DTN engine (ION-DTN, Hardy, etc.) implements the `BackendAdapter` trait
//! to provide configuration generation, lifecycle management, and telemetry.

pub mod capability;
pub mod capability_check;
pub mod hardy;
pub mod ion;
pub mod registry;
pub mod traits;

// Re-export primary types at the adapter module level.
pub use capability::{CapabilitySet, HotReconfigCapabilities, SecurityCapabilities};
pub use capability_check::{
    check_convergence_layer_support, check_hot_reconfig_capability, check_routing_strategy_support,
    HotReconfigOp,
};
pub use registry::AdapterRegistry;
pub use traits::{
    BackendAdapter, BundleStatistics, ContactRef, GeneratedConfig, HealthStatus, LinkState, NodeRef,
};
