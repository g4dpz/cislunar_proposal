//! Canonical configuration model types.
//!
//! Defines the vendor-neutral representation of DTN network configuration
//! including nodes, neighbors, contact plans, convergence layers, and routing.

pub mod config;
pub mod contact_plan;
pub mod convergence;
pub mod neighbor;
pub mod node;
pub mod routing;

// Re-export all model types at the module level for convenience.
pub use config::{NetworkConfiguration, SecurityConfig, StorageConfig};
pub use contact_plan::{Contact, ContactPlan, Range};
pub use convergence::ConvergenceLayerType;
pub use neighbor::{ConvergenceLayerLink, Neighbor};
pub use node::{EndpointId, NodeDefinition, ServiceDemux};
pub use routing::{RoutingConfig, RoutingStrategy, StaticRoute};
