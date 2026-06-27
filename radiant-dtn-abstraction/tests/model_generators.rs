//! Proptest `Arbitrary` generators for all canonical model types.
//!
//! These generators produce valid NetworkConfiguration values suitable
//! for round-trip serialization testing (YAML and JSON).

use proptest::prelude::*;
use std::collections::HashMap;

use radiant_dtn_abstraction::model::{
    Contact, ContactPlan, ConvergenceLayerLink, EndpointId, Neighbor, NetworkConfiguration,
    NodeDefinition, Range, RoutingConfig, RoutingStrategy, SecurityConfig, ServiceDemux,
    StaticRoute, StorageConfig,
};

/// Generate a valid node number (1..=1000).
pub fn arb_node_number() -> impl Strategy<Value = u64> {
    1u64..=1000
}

/// Generate a valid service number (0..=100).
pub fn arb_service_number() -> impl Strategy<Value = u64> {
    0u64..=100
}

/// Generate a valid EndpointId.
///
/// For the untagged enum to round-trip correctly through YAML,
/// Ipn uses numeric fields and Dtn uses string fields.
pub fn arb_endpoint_id() -> impl Strategy<Value = EndpointId> {
    prop_oneof![
        (arb_node_number(), arb_service_number()).prop_map(|(n, s)| EndpointId::Ipn {
            node_number: n,
            service_number: s,
        }),
        ("[a-z][a-z0-9]{1,5}-[0-9]{1,2}", "[a-z]{2,6}").prop_map(|(auth, path)| {
            EndpointId::Dtn {
                authority: auth,
                path: path,
            }
        }),
    ]
}

/// Generate a ServiceDemux.
pub fn arb_service_demux() -> impl Strategy<Value = ServiceDemux> {
    (arb_service_number(), proptest::option::of("[a-z ]{3,20}")).prop_map(|(sn, desc)| {
        ServiceDemux {
            service_number: sn,
            description: desc,
        }
    })
}

/// Generate a NodeDefinition.
pub fn arb_node_definition() -> impl Strategy<Value = NodeDefinition> {
    (
        arb_node_number(),
        proptest::option::of(arb_endpoint_id()),
        proptest::option::of(arb_endpoint_id()),
        "[a-zA-Z][a-zA-Z0-9 ]{2,20}",
        proptest::collection::vec(arb_service_demux(), 0..=3),
    )
        .prop_map(|(node_number, endpoint_id, callsign_eid, name, services)| {
            NodeDefinition {
                node_number,
                endpoint_id,
                callsign_eid,
                name,
                services,
            }
        })
}

/// Generate an LtpUdp convergence layer link.
fn arb_ltp_udp_link() -> impl Strategy<Value = ConvergenceLayerLink> {
    (
        "[a-z]{3,8}-[0-9]{1,3}",
        arb_node_number(),
        arb_node_number(),
        "192\\.168\\.[0-9]{1,3}\\.[0-9]{1,3}",
        1024u16..=65000,
        1024u16..=65000,
        proptest::option::of(512u32..=9000),
        proptest::option::of(1u32..=1000),
    )
        .prop_map(
            |(id, local_engine_id, remote_engine_id, remote_host, remote_port, local_port, mtu, segment_rate)| {
                ConvergenceLayerLink::LtpUdp {
                    id,
                    local_engine_id,
                    remote_engine_id,
                    remote_host,
                    remote_port,
                    local_port,
                    mtu,
                    segment_rate,
                }
            },
        )
}

/// Generate a TcpCl convergence layer link.
fn arb_tcpcl_link() -> impl Strategy<Value = ConvergenceLayerLink> {
    (
        "[a-z]{3,8}-[0-9]{1,3}",
        "192\\.168\\.[0-9]{1,3}\\.[0-9]{1,3}",
        1024u16..=65000,
        proptest::option::of(1024u16..=65000),
        proptest::option::of(10u32..=300),
    )
        .prop_map(
            |(id, remote_host, remote_port, local_port, keepalive_interval_secs)| {
                ConvergenceLayerLink::TcpCl {
                    id,
                    remote_host,
                    remote_port,
                    local_port,
                    keepalive_interval_secs,
                }
            },
        )
}

/// Generate a Kiss convergence layer link.
fn arb_kiss_link() -> impl Strategy<Value = ConvergenceLayerLink> {
    (
        "[a-z]{3,8}-[0-9]{1,3}",
        "/dev/tty[A-Z]{3}[0-9]",
        prop_oneof![Just(1200u32), Just(9600u32), Just(19200u32)],
        arb_node_number(),
        arb_node_number(),
        proptest::option::of(64u32..=512),
    )
        .prop_map(
            |(id, tnc_device, baud_rate, local_engine_id, remote_engine_id, frame_size)| {
                ConvergenceLayerLink::Kiss {
                    id,
                    tnc_device,
                    baud_rate,
                    local_engine_id,
                    remote_engine_id,
                    frame_size,
                }
            },
        )
}

/// Generate a Udp convergence layer link.
fn arb_udp_link() -> impl Strategy<Value = ConvergenceLayerLink> {
    (
        "[a-z]{3,8}-[0-9]{1,3}",
        "192\\.168\\.[0-9]{1,3}\\.[0-9]{1,3}",
        1024u16..=65000,
        proptest::option::of(1024u16..=65000),
    )
        .prop_map(|(id, remote_host, remote_port, local_port)| {
            ConvergenceLayerLink::Udp {
                id,
                remote_host,
                remote_port,
                local_port,
            }
        })
}

/// Generate a ConvergenceLayerLink (any variant).
pub fn arb_convergence_layer_link() -> impl Strategy<Value = ConvergenceLayerLink> {
    prop_oneof![
        arb_ltp_udp_link(),
        arb_tcpcl_link(),
        arb_kiss_link(),
        arb_udp_link(),
    ]
}

/// Generate a Neighbor with 1-3 links.
pub fn arb_neighbor() -> impl Strategy<Value = Neighbor> {
    (
        arb_node_number(),
        proptest::option::of("[a-zA-Z][a-zA-Z0-9 ]{2,15}"),
        proptest::collection::vec(arb_convergence_layer_link(), 1..=3),
        proptest::option::of(1000u64..=1_000_000_000),
    )
        .prop_map(|(node_number, name, links, rate_limit_bps)| Neighbor {
            node_number,
            name,
            links,
            rate_limit_bps,
        })
}

/// Generate a Contact with start_time < end_time and valid confidence.
pub fn arb_contact() -> impl Strategy<Value = Contact> {
    (
        arb_node_number(),
        arb_node_number(),
        0i64..=1_700_000_000,
        1i64..=86400,
        1u64..=1_000_000_000,
        0.0f64..=1.0,
    )
        .prop_map(
            |(source_node, dest_node, start_time, duration, rate_bps, confidence)| Contact {
                source_node,
                dest_node,
                start_time,
                end_time: start_time + duration,
                rate_bps,
                confidence,
            },
        )
}

/// Generate a Range.
pub fn arb_range() -> impl Strategy<Value = Range> {
    (arb_node_number(), arb_node_number(), 0.0f64..=10.0).prop_map(
        |(source_node, dest_node, owlt_secs)| Range {
            source_node,
            dest_node,
            owlt_secs,
        },
    )
}

/// Generate a ContactPlan with 0-10 contacts and 0-5 ranges.
pub fn arb_contact_plan() -> impl Strategy<Value = ContactPlan> {
    (
        proptest::collection::vec(arb_contact(), 0..=10),
        proptest::collection::vec(arb_range(), 0..=5),
    )
        .prop_map(|(contacts, ranges)| ContactPlan { contacts, ranges })
}

/// Generate a RoutingStrategy.
pub fn arb_routing_strategy() -> impl Strategy<Value = RoutingStrategy> {
    prop_oneof![
        Just(RoutingStrategy::Cgr),
        Just(RoutingStrategy::Static),
        Just(RoutingStrategy::Default),
    ]
}

/// Generate a StaticRoute.
pub fn arb_static_route() -> impl Strategy<Value = StaticRoute> {
    (arb_node_number(), arb_node_number()).prop_map(|(destination_node, next_hop_node)| {
        StaticRoute {
            destination_node,
            next_hop_node,
        }
    })
}

/// Generate a RoutingConfig.
pub fn arb_routing_config() -> impl Strategy<Value = RoutingConfig> {
    (
        arb_routing_strategy(),
        proptest::collection::vec(arb_static_route(), 0..=5),
    )
        .prop_map(|(strategy, static_routes)| RoutingConfig {
            strategy,
            static_routes,
        })
}

/// Generate a SecurityConfig.
pub fn arb_security_config() -> impl Strategy<Value = SecurityConfig> {
    // Per amateur radio regulations, security is typically disabled,
    // but we test both states for round-trip correctness.
    proptest::bool::ANY.prop_map(|enabled| SecurityConfig { enabled })
}

/// Generate a StorageConfig.
pub fn arb_storage_config() -> impl Strategy<Value = StorageConfig> {
    (
        "/[a-z]{3,8}/[a-z]{3,8}",
        proptest::option::of(1024u64..=10_000_000_000),
    )
        .prop_map(|(path, max_bytes)| StorageConfig { path, max_bytes })
}

/// Generate a complete NetworkConfiguration.
///
/// Constraints:
/// - 0-5 neighbors
/// - 0-10 contacts
/// - 0-5 ranges
/// - backend_options is kept empty for round-trip testing (serde_yaml::Value
///   equality is well-defined for simple cases but we avoid complexity here)
pub fn arb_network_configuration() -> impl Strategy<Value = NetworkConfiguration> {
    (
        arb_node_definition(),
        proptest::collection::vec(arb_neighbor(), 0..=5),
        arb_contact_plan(),
        arb_routing_config(),
        proptest::option::of(arb_security_config()),
        proptest::option::of(arb_storage_config()),
    )
        .prop_map(
            |(local_node, neighbors, contact_plan, routing, security, storage)| {
                NetworkConfiguration {
                    version: "1.0".to_string(),
                    backend: "ion-dtn".to_string(),
                    local_node,
                    neighbors,
                    contact_plan,
                    routing,
                    security,
                    storage,
                    backend_options: HashMap::new(),
                }
            },
        )
}
