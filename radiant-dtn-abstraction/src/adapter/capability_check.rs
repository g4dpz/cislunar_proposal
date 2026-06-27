//! Capability enforcement logic for the DTN Abstraction Layer.
//!
//! Provides functions to check whether a hot-reconfiguration operation,
//! convergence layer type, or routing strategy is supported by a given
//! adapter's CapabilitySet before dispatching the operation. Returns
//! `UnsupportedOperation` errors with identifying messages when the
//! capability is not supported.
//!
//! These checks work without a running engine — they query the adapter's
//! static capability declaration only.

use crate::error::{AbstractionError, ErrorCategory};
use crate::model::convergence::ConvergenceLayerType;
use crate::model::routing::RoutingStrategy;

use super::capability::CapabilitySet;

/// Enumerates the hot-reconfiguration operations that can be checked
/// against a backend's capability set.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum HotReconfigOp {
    AddContact,
    RemoveContact,
    AddNeighbor,
    RemoveNeighbor,
    EnableLink,
    DisableLink,
}

impl HotReconfigOp {
    /// Returns a human-readable operation name for error messages.
    pub fn operation_name(&self) -> &'static str {
        match self {
            HotReconfigOp::AddContact => "add_contact",
            HotReconfigOp::RemoveContact => "remove_contact",
            HotReconfigOp::AddNeighbor => "add_neighbor",
            HotReconfigOp::RemoveNeighbor => "remove_neighbor",
            HotReconfigOp::EnableLink => "enable_link",
            HotReconfigOp::DisableLink => "disable_link",
        }
    }
}

/// Check whether a hot-reconfiguration operation is supported by the given adapter.
///
/// Returns `Ok(())` if the operation is supported, or an `UnsupportedOperation`
/// error with an identifying message if not.
///
/// This function works without a running engine — it only inspects the
/// adapter's declared capability set.
///
/// # Arguments
/// * `capabilities` — The adapter's declared capability set.
/// * `operation` — The hot-reconfiguration operation to check.
/// * `backend_name` — Name of the backend adapter (for error context).
///
/// # Errors
/// Returns `AbstractionError` with category `UnsupportedOperation` when the
/// operation is not supported by the backend.
pub fn check_hot_reconfig_capability(
    capabilities: &CapabilitySet,
    operation: HotReconfigOp,
    backend_name: &str,
) -> Result<(), AbstractionError> {
    let supported = match operation {
        HotReconfigOp::AddContact => capabilities.hot_reconfig.add_contact,
        HotReconfigOp::RemoveContact => capabilities.hot_reconfig.remove_contact,
        HotReconfigOp::AddNeighbor => capabilities.hot_reconfig.add_neighbor,
        HotReconfigOp::RemoveNeighbor => capabilities.hot_reconfig.remove_neighbor,
        HotReconfigOp::EnableLink => capabilities.hot_reconfig.enable_link,
        HotReconfigOp::DisableLink => capabilities.hot_reconfig.disable_link,
    };

    if supported {
        Ok(())
    } else {
        Err(AbstractionError::new(
            ErrorCategory::UnsupportedOperation,
            format!(
                "Operation '{}' is not supported by backend '{}'",
                operation.operation_name(),
                backend_name
            ),
            operation.operation_name(),
        )
        .with_backend(backend_name))
    }
}

/// Check whether a convergence layer type is supported by the given adapter.
///
/// Returns `Ok(())` if the convergence layer is supported, or an
/// `UnsupportedOperation` error if not.
///
/// This function works without a running engine.
///
/// # Arguments
/// * `capabilities` — The adapter's declared capability set.
/// * `cl_type` — The convergence layer type to check.
/// * `backend_name` — Name of the backend adapter (for error context).
///
/// # Errors
/// Returns `AbstractionError` with category `UnsupportedOperation` when the
/// convergence layer type is not supported by the backend.
pub fn check_convergence_layer_support(
    capabilities: &CapabilitySet,
    cl_type: ConvergenceLayerType,
    backend_name: &str,
) -> Result<(), AbstractionError> {
    if capabilities.convergence_layers.contains(&cl_type) {
        Ok(())
    } else {
        Err(AbstractionError::new(
            ErrorCategory::UnsupportedOperation,
            format!(
                "Convergence layer '{:?}' is not supported by backend '{}'",
                cl_type, backend_name
            ),
            "check_convergence_layer",
        )
        .with_backend(backend_name))
    }
}

/// Check whether a routing strategy is supported by the given adapter.
///
/// Returns `Ok(())` if the routing strategy is supported, or an
/// `UnsupportedOperation` error if not.
///
/// This function works without a running engine.
///
/// # Arguments
/// * `capabilities` — The adapter's declared capability set.
/// * `strategy` — The routing strategy to check.
/// * `backend_name` — Name of the backend adapter (for error context).
///
/// # Errors
/// Returns `AbstractionError` with category `UnsupportedOperation` when the
/// routing strategy is not supported by the backend.
pub fn check_routing_strategy_support(
    capabilities: &CapabilitySet,
    strategy: RoutingStrategy,
    backend_name: &str,
) -> Result<(), AbstractionError> {
    if capabilities.routing_strategies.contains(&strategy) {
        Ok(())
    } else {
        Err(AbstractionError::new(
            ErrorCategory::UnsupportedOperation,
            format!(
                "Routing strategy '{:?}' is not supported by backend '{}'",
                strategy, backend_name
            ),
            "check_routing_strategy",
        )
        .with_backend(backend_name))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::adapter::capability::{HotReconfigCapabilities, SecurityCapabilities};

    /// Helper: build a capability set where all hot-reconfig ops are supported.
    fn all_caps() -> CapabilitySet {
        CapabilitySet {
            hot_reconfig: HotReconfigCapabilities::all(),
            convergence_layers: vec![
                ConvergenceLayerType::LtpUdp,
                ConvergenceLayerType::TcpCl,
                ConvergenceLayerType::Kiss,
                ConvergenceLayerType::Udp,
            ],
            routing_strategies: vec![
                RoutingStrategy::Cgr,
                RoutingStrategy::Static,
                RoutingStrategy::Default,
            ],
            security: SecurityCapabilities::none(),
        }
    }

    /// Helper: build a capability set where no hot-reconfig ops are supported.
    fn no_caps() -> CapabilitySet {
        CapabilitySet {
            hot_reconfig: HotReconfigCapabilities::none(),
            convergence_layers: vec![],
            routing_strategies: vec![],
            security: SecurityCapabilities::none(),
        }
    }

    /// Helper: build Hardy-like capabilities (partial hot-reconfig support).
    fn hardy_like_caps() -> CapabilitySet {
        CapabilitySet {
            hot_reconfig: HotReconfigCapabilities {
                add_contact: true,
                remove_contact: true,
                add_neighbor: false,
                remove_neighbor: false,
                enable_link: true,
                disable_link: true,
            },
            convergence_layers: vec![
                ConvergenceLayerType::LtpUdp,
                ConvergenceLayerType::TcpCl,
            ],
            routing_strategies: vec![RoutingStrategy::Cgr, RoutingStrategy::Static],
            security: SecurityCapabilities::none(),
        }
    }

    #[test]
    fn test_all_hot_reconfig_ops_supported() {
        let caps = all_caps();
        let ops = [
            HotReconfigOp::AddContact,
            HotReconfigOp::RemoveContact,
            HotReconfigOp::AddNeighbor,
            HotReconfigOp::RemoveNeighbor,
            HotReconfigOp::EnableLink,
            HotReconfigOp::DisableLink,
        ];
        for op in &ops {
            assert!(
                check_hot_reconfig_capability(&caps, *op, "ion-dtn").is_ok(),
                "Expected {:?} to be supported",
                op
            );
        }
    }

    #[test]
    fn test_no_hot_reconfig_ops_supported() {
        let caps = no_caps();
        let ops = [
            HotReconfigOp::AddContact,
            HotReconfigOp::RemoveContact,
            HotReconfigOp::AddNeighbor,
            HotReconfigOp::RemoveNeighbor,
            HotReconfigOp::EnableLink,
            HotReconfigOp::DisableLink,
        ];
        for op in &ops {
            let result = check_hot_reconfig_capability(&caps, *op, "no-engine");
            assert!(result.is_err(), "Expected {:?} to be unsupported", op);
            let err = result.unwrap_err();
            assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
            assert!(err.message.contains(op.operation_name()));
            assert!(err.message.contains("no-engine"));
        }
    }

    #[test]
    fn test_hardy_partial_hot_reconfig() {
        let caps = hardy_like_caps();

        // Supported
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::AddContact, "hardy").is_ok());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::RemoveContact, "hardy").is_ok());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::EnableLink, "hardy").is_ok());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::DisableLink, "hardy").is_ok());

        // Not supported
        let err = check_hot_reconfig_capability(&caps, HotReconfigOp::AddNeighbor, "hardy")
            .unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("add_neighbor"));
        assert!(err.message.contains("hardy"));

        let err = check_hot_reconfig_capability(&caps, HotReconfigOp::RemoveNeighbor, "hardy")
            .unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("remove_neighbor"));
        assert!(err.message.contains("hardy"));
    }

    #[test]
    fn test_convergence_layer_support() {
        let caps = hardy_like_caps();

        // Supported
        assert!(check_convergence_layer_support(&caps, ConvergenceLayerType::LtpUdp, "hardy").is_ok());
        assert!(check_convergence_layer_support(&caps, ConvergenceLayerType::TcpCl, "hardy").is_ok());

        // Not supported
        let err =
            check_convergence_layer_support(&caps, ConvergenceLayerType::Kiss, "hardy").unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("Kiss"));
        assert!(err.message.contains("hardy"));

        let err =
            check_convergence_layer_support(&caps, ConvergenceLayerType::Udp, "hardy").unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("Udp"));
        assert!(err.message.contains("hardy"));
    }

    #[test]
    fn test_routing_strategy_support() {
        let caps = hardy_like_caps();

        // Supported
        assert!(check_routing_strategy_support(&caps, RoutingStrategy::Cgr, "hardy").is_ok());
        assert!(check_routing_strategy_support(&caps, RoutingStrategy::Static, "hardy").is_ok());

        // Not supported
        let err =
            check_routing_strategy_support(&caps, RoutingStrategy::Default, "hardy").unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("Default"));
        assert!(err.message.contains("hardy"));
    }

    #[test]
    fn test_error_context_has_backend() {
        let caps = no_caps();
        let err = check_hot_reconfig_capability(&caps, HotReconfigOp::AddContact, "test-backend")
            .unwrap_err();
        assert_eq!(err.context.backend.as_deref(), Some("test-backend"));
        assert_eq!(err.context.operation, "add_contact");
    }

    #[test]
    fn test_capability_check_works_without_engine() {
        // This test demonstrates that capability checks don't require
        // any engine to be running — they work purely on the static
        // CapabilitySet declaration.
        let caps = CapabilitySet {
            hot_reconfig: HotReconfigCapabilities {
                add_contact: true,
                remove_contact: false,
                add_neighbor: true,
                remove_neighbor: false,
                enable_link: false,
                disable_link: true,
            },
            convergence_layers: vec![ConvergenceLayerType::Kiss],
            routing_strategies: vec![RoutingStrategy::Default],
            security: SecurityCapabilities::none(),
        };

        // These work purely from the CapabilitySet struct — no engine needed
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::AddContact, "x").is_ok());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::RemoveContact, "x").is_err());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::AddNeighbor, "x").is_ok());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::RemoveNeighbor, "x").is_err());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::EnableLink, "x").is_err());
        assert!(check_hot_reconfig_capability(&caps, HotReconfigOp::DisableLink, "x").is_ok());

        assert!(check_convergence_layer_support(&caps, ConvergenceLayerType::Kiss, "x").is_ok());
        assert!(check_convergence_layer_support(&caps, ConvergenceLayerType::LtpUdp, "x").is_err());

        assert!(check_routing_strategy_support(&caps, RoutingStrategy::Default, "x").is_ok());
        assert!(check_routing_strategy_support(&caps, RoutingStrategy::Cgr, "x").is_err());
    }

    #[test]
    fn test_operation_name_mapping() {
        assert_eq!(HotReconfigOp::AddContact.operation_name(), "add_contact");
        assert_eq!(HotReconfigOp::RemoveContact.operation_name(), "remove_contact");
        assert_eq!(HotReconfigOp::AddNeighbor.operation_name(), "add_neighbor");
        assert_eq!(HotReconfigOp::RemoveNeighbor.operation_name(), "remove_neighbor");
        assert_eq!(HotReconfigOp::EnableLink.operation_name(), "enable_link");
        assert_eq!(HotReconfigOp::DisableLink.operation_name(), "disable_link");
    }
}
