//! Configuration validation logic.
//!
//! Validates referential integrity, temporal constraints, and structural correctness
//! of the canonical configuration model. Uses fail-slow validation to collect all
//! errors before returning.

use std::collections::HashSet;

use crate::error::ValidationDetail;
use crate::model::{
    EndpointId, NetworkConfiguration, NodeDefinition, RoutingStrategy,
};

/// Validate a `NetworkConfiguration` for structural and referential correctness.
///
/// Uses fail-slow validation: collects ALL errors before returning rather than
/// stopping at the first violation. Returns `Ok(())` if the configuration is valid,
/// or `Err(Vec<ValidationDetail>)` containing all discovered violations.
pub fn validate(config: &NetworkConfiguration) -> Result<(), Vec<ValidationDetail>> {
    let mut errors: Vec<ValidationDetail> = Vec::new();

    // Build the set of all known/defined node numbers
    let mut defined_nodes: HashSet<u64> = HashSet::new();
    defined_nodes.insert(config.local_node.node_number);
    for neighbor in &config.neighbors {
        defined_nodes.insert(neighbor.node_number);
    }

    // Check: local_node.node_number must not duplicate any neighbor's node_number
    check_duplicate_local_node(config, &mut errors);

    // Check: neighbor node_numbers must be unique among neighbors
    check_duplicate_neighbor_nodes(config, &mut errors);

    // Check: contact node references must resolve to defined nodes
    check_contact_node_references(config, &defined_nodes, &mut errors);

    // Check: range node references must resolve to defined nodes
    check_range_node_references(config, &defined_nodes, &mut errors);

    // Check: temporal constraints on contacts (start_time < end_time)
    check_contact_temporal_constraints(config, &mut errors);

    // Check: static route references (if routing strategy is Static)
    check_static_route_references(config, &defined_nodes, &mut errors);

    if errors.is_empty() {
        Ok(())
    } else {
        Err(errors)
    }
}

/// Derive the default endpoint ID for a node when `endpoint_id` is `None`.
///
/// Sets `endpoint_id` to `EndpointId::Ipn { node_number: N, service_number: 0 }`
/// where N is the node's `node_number`.
pub fn derive_default_endpoint(node: &mut NodeDefinition) {
    if node.endpoint_id.is_none() {
        node.endpoint_id = Some(EndpointId::Ipn {
            node_number: node.node_number,
            service_number: 0,
        });
    }
}

/// Check that local_node.node_number does not appear in any neighbor's node_number.
fn check_duplicate_local_node(config: &NetworkConfiguration, errors: &mut Vec<ValidationDetail>) {
    let local_num = config.local_node.node_number;
    for (i, neighbor) in config.neighbors.iter().enumerate() {
        if neighbor.node_number == local_num {
            errors.push(ValidationDetail {
                path: format!("neighbors[{}].node_number", i),
                rule: "duplicate_node_number".to_string(),
                message: format!(
                    "Neighbor node_number {} duplicates local_node.node_number",
                    neighbor.node_number
                ),
            });
        }
    }
}

/// Check that neighbor node_numbers are unique among neighbors.
fn check_duplicate_neighbor_nodes(
    config: &NetworkConfiguration,
    errors: &mut Vec<ValidationDetail>,
) {
    let mut seen: HashSet<u64> = HashSet::new();
    for (i, neighbor) in config.neighbors.iter().enumerate() {
        if !seen.insert(neighbor.node_number) {
            errors.push(ValidationDetail {
                path: format!("neighbors[{}].node_number", i),
                rule: "duplicate_node_number".to_string(),
                message: format!(
                    "Duplicate neighbor node_number {} (already defined by another neighbor)",
                    neighbor.node_number
                ),
            });
        }
    }
}

/// Check that every contact's source_node and dest_node reference defined nodes.
fn check_contact_node_references(
    config: &NetworkConfiguration,
    defined_nodes: &HashSet<u64>,
    errors: &mut Vec<ValidationDetail>,
) {
    for (i, contact) in config.contact_plan.contacts.iter().enumerate() {
        if !defined_nodes.contains(&contact.source_node) {
            errors.push(ValidationDetail {
                path: format!("contact_plan.contacts[{}].source_node", i),
                rule: "referential_integrity".to_string(),
                message: format!(
                    "Contact source_node {} does not reference a defined node",
                    contact.source_node
                ),
            });
        }
        if !defined_nodes.contains(&contact.dest_node) {
            errors.push(ValidationDetail {
                path: format!("contact_plan.contacts[{}].dest_node", i),
                rule: "referential_integrity".to_string(),
                message: format!(
                    "Contact dest_node {} does not reference a defined node",
                    contact.dest_node
                ),
            });
        }
    }
}

/// Check that every range's source_node and dest_node reference defined nodes.
fn check_range_node_references(
    config: &NetworkConfiguration,
    defined_nodes: &HashSet<u64>,
    errors: &mut Vec<ValidationDetail>,
) {
    for (i, range) in config.contact_plan.ranges.iter().enumerate() {
        if !defined_nodes.contains(&range.source_node) {
            errors.push(ValidationDetail {
                path: format!("contact_plan.ranges[{}].source_node", i),
                rule: "referential_integrity".to_string(),
                message: format!(
                    "Range source_node {} does not reference a defined node",
                    range.source_node
                ),
            });
        }
        if !defined_nodes.contains(&range.dest_node) {
            errors.push(ValidationDetail {
                path: format!("contact_plan.ranges[{}].dest_node", i),
                rule: "referential_integrity".to_string(),
                message: format!(
                    "Range dest_node {} does not reference a defined node",
                    range.dest_node
                ),
            });
        }
    }
}

/// Check that every contact has start_time < end_time.
fn check_contact_temporal_constraints(
    config: &NetworkConfiguration,
    errors: &mut Vec<ValidationDetail>,
) {
    for (i, contact) in config.contact_plan.contacts.iter().enumerate() {
        if contact.start_time >= contact.end_time {
            errors.push(ValidationDetail {
                path: format!("contact_plan.contacts[{}]", i),
                rule: "temporal_order".to_string(),
                message: format!(
                    "Contact start_time ({}) must be less than end_time ({})",
                    contact.start_time, contact.end_time
                ),
            });
        }
    }
}

/// Check that static route destination_node and next_hop_node reference defined nodes.
/// Only applies when routing strategy is `Static`.
fn check_static_route_references(
    config: &NetworkConfiguration,
    defined_nodes: &HashSet<u64>,
    errors: &mut Vec<ValidationDetail>,
) {
    if config.routing.strategy != RoutingStrategy::Static {
        return;
    }

    for (i, route) in config.routing.static_routes.iter().enumerate() {
        if !defined_nodes.contains(&route.destination_node) {
            errors.push(ValidationDetail {
                path: format!("routing.static_routes[{}].destination_node", i),
                rule: "referential_integrity".to_string(),
                message: format!(
                    "Static route destination_node {} does not reference a defined node",
                    route.destination_node
                ),
            });
        }
        if !defined_nodes.contains(&route.next_hop_node) {
            errors.push(ValidationDetail {
                path: format!("routing.static_routes[{}].next_hop_node", i),
                rule: "referential_integrity".to_string(),
                message: format!(
                    "Static route next_hop_node {} does not reference a defined node",
                    route.next_hop_node
                ),
            });
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::model::{
        Contact, ContactPlan, NetworkConfiguration, NodeDefinition, Range,
        RoutingConfig, RoutingStrategy, StaticRoute,
    };
    use crate::model::neighbor::Neighbor;
    use std::collections::HashMap;

    /// Helper: build a minimal valid configuration for testing.
    fn valid_config() -> NetworkConfiguration {
        NetworkConfiguration {
            version: "1.0".to_string(),
            backend: "ion-dtn".to_string(),
            local_node: NodeDefinition {
                node_number: 10,
                endpoint_id: Some(EndpointId::Ipn {
                    node_number: 10,
                    service_number: 0,
                }),
                callsign_eid: None,
                name: "Ground Station".to_string(),
                services: vec![],
            },
            neighbors: vec![Neighbor {
                node_number: 20,
                name: Some("Orbiter".to_string()),
                links: vec![],
                rate_limit_bps: None,
            }],
            contact_plan: ContactPlan {
                contacts: vec![Contact {
                    source_node: 10,
                    dest_node: 20,
                    start_time: 1000,
                    end_time: 2000,
                    rate_bps: 9600,
                    confidence: 1.0,
                }],
                ranges: vec![Range {
                    source_node: 10,
                    dest_node: 20,
                    owlt_secs: 1.3,
                }],
            },
            routing: RoutingConfig {
                strategy: RoutingStrategy::Cgr,
                static_routes: vec![],
            },
            security: None,
            storage: None,
            backend_options: HashMap::new(),
        }
    }

    #[test]
    fn test_valid_config_passes() {
        let config = valid_config();
        assert!(validate(&config).is_ok());
    }

    #[test]
    fn test_duplicate_node_number_local_vs_neighbor() {
        let mut config = valid_config();
        // Set a neighbor's node_number to match local_node
        config.neighbors[0].node_number = 10;
        // Also update contacts/ranges to reference valid nodes
        config.contact_plan.contacts[0].dest_node = 10;
        config.contact_plan.ranges[0].dest_node = 10;

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "duplicate_node_number"
            && e.path.contains("neighbors[0]")));
    }

    #[test]
    fn test_duplicate_neighbor_node_numbers() {
        let mut config = valid_config();
        // Add a second neighbor with the same node_number
        config.neighbors.push(Neighbor {
            node_number: 20,
            name: Some("Duplicate".to_string()),
            links: vec![],
            rate_limit_bps: None,
        });

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "duplicate_node_number"
            && e.path.contains("neighbors[1]")));
    }

    #[test]
    fn test_undefined_node_in_contact() {
        let mut config = valid_config();
        // Reference a node that doesn't exist
        config.contact_plan.contacts[0].dest_node = 999;

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "referential_integrity"
            && e.message.contains("999")));
    }

    #[test]
    fn test_undefined_node_in_range() {
        let mut config = valid_config();
        config.contact_plan.ranges[0].source_node = 888;

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "referential_integrity"
            && e.path.contains("ranges[0].source_node")
            && e.message.contains("888")));
    }

    #[test]
    fn test_contact_start_time_equals_end_time() {
        let mut config = valid_config();
        config.contact_plan.contacts[0].start_time = 5000;
        config.contact_plan.contacts[0].end_time = 5000;

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "temporal_order"));
    }

    #[test]
    fn test_contact_start_time_greater_than_end_time() {
        let mut config = valid_config();
        config.contact_plan.contacts[0].start_time = 9000;
        config.contact_plan.contacts[0].end_time = 1000;

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "temporal_order"));
    }

    #[test]
    fn test_multiple_errors_collected() {
        let mut config = valid_config();
        // Error 1: duplicate node number (neighbor matches local)
        config.neighbors[0].node_number = 10;
        // Error 2: contact references undefined node
        config.contact_plan.contacts[0].dest_node = 999;
        // Error 3: temporal violation
        config.contact_plan.contacts.push(Contact {
            source_node: 10,
            dest_node: 10,
            start_time: 5000,
            end_time: 3000,
            rate_bps: 1200,
            confidence: 1.0,
        });

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        // Should have at least 3 errors (fail-slow collects all)
        assert!(
            errors.len() >= 3,
            "Expected at least 3 errors, got {}: {:?}",
            errors.len(),
            errors
        );
    }

    #[test]
    fn test_static_route_undefined_destination() {
        let mut config = valid_config();
        config.routing.strategy = RoutingStrategy::Static;
        config.routing.static_routes = vec![StaticRoute {
            destination_node: 777,
            next_hop_node: 20,
        }];

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "referential_integrity"
            && e.path.contains("static_routes[0].destination_node")));
    }

    #[test]
    fn test_static_route_undefined_next_hop() {
        let mut config = valid_config();
        config.routing.strategy = RoutingStrategy::Static;
        config.routing.static_routes = vec![StaticRoute {
            destination_node: 20,
            next_hop_node: 555,
        }];

        let result = validate(&config);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| e.rule == "referential_integrity"
            && e.path.contains("static_routes[0].next_hop_node")));
    }

    #[test]
    fn test_static_route_not_checked_for_cgr() {
        let mut config = valid_config();
        // Use CGR strategy with invalid static routes — should NOT produce errors
        config.routing.strategy = RoutingStrategy::Cgr;
        config.routing.static_routes = vec![StaticRoute {
            destination_node: 777,
            next_hop_node: 888,
        }];

        let result = validate(&config);
        assert!(result.is_ok());
    }

    #[test]
    fn test_derive_default_endpoint_when_none() {
        let mut node = NodeDefinition {
            node_number: 42,
            endpoint_id: None,
            callsign_eid: None,
            name: "Test Node".to_string(),
            services: vec![],
        };

        derive_default_endpoint(&mut node);

        assert_eq!(
            node.endpoint_id,
            Some(EndpointId::Ipn {
                node_number: 42,
                service_number: 0,
            })
        );
    }

    #[test]
    fn test_derive_default_endpoint_does_not_overwrite() {
        let existing = EndpointId::Dtn {
            authority: "g4dpz-1".to_string(),
            path: "gs".to_string(),
        };
        let mut node = NodeDefinition {
            node_number: 42,
            endpoint_id: Some(existing.clone()),
            callsign_eid: None,
            name: "Test Node".to_string(),
            services: vec![],
        };

        derive_default_endpoint(&mut node);

        // Should not overwrite the existing endpoint_id
        assert_eq!(node.endpoint_id, Some(existing));
    }
}
