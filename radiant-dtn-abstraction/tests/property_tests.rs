//! Property-based tests for the DTN Abstraction Layer.
//!
//! Uses proptest to verify correctness properties defined in the design document.

mod model_generators;

use model_generators::{arb_network_configuration, arb_node_number};
use proptest::prelude::*;

use std::collections::HashMap;
use std::sync::Arc;

use async_trait::async_trait;
use radiant_dtn_abstraction::adapter::capability::{
    CapabilitySet, HotReconfigCapabilities, SecurityCapabilities,
};
use radiant_dtn_abstraction::adapter::registry::AdapterRegistry;
use radiant_dtn_abstraction::adapter::traits::{
    BackendAdapter, BundleStatistics, ContactRef, GeneratedConfig, HealthStatus, LinkState,
    NodeRef,
};
use radiant_dtn_abstraction::error::{AbstractionError, ErrorCategory};
use radiant_dtn_abstraction::model::{
    Contact, ContactPlan, EndpointId, Neighbor, NetworkConfiguration, NodeDefinition, Range,
    RoutingConfig, RoutingStrategy,
};
use radiant_dtn_abstraction::validation::{derive_default_endpoint, validate};

/// A minimal mock adapter for property-based testing of the registry.
struct PropMockAdapter {
    adapter_name: String,
}

impl PropMockAdapter {
    fn new(name: &str) -> Self {
        Self {
            adapter_name: name.to_string(),
        }
    }
}

#[async_trait]
impl BackendAdapter for PropMockAdapter {
    fn name(&self) -> &str {
        &self.adapter_name
    }

    fn capabilities(&self) -> &CapabilitySet {
        Box::leak(Box::new(CapabilitySet {
            hot_reconfig: HotReconfigCapabilities {
                add_contact: false,
                remove_contact: false,
                add_neighbor: false,
                remove_neighbor: false,
                enable_link: false,
                disable_link: false,
            },
            convergence_layers: vec![],
            routing_strategies: vec![],
            security: SecurityCapabilities::none(),
        }))
    }

    async fn validate(&self, _config: &NetworkConfiguration) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn generate_config(
        &self,
        _config: &NetworkConfiguration,
        _output_dir: &std::path::Path,
    ) -> Result<GeneratedConfig, AbstractionError> {
        Ok(GeneratedConfig {
            files: HashMap::new(),
        })
    }

    async fn deploy(
        &self,
        _config: &NetworkConfiguration,
        _output_dir: &std::path::Path,
    ) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn start(&self, _config_dir: &std::path::Path) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn stop(&self) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn restart(&self, _config_dir: &std::path::Path) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn health(&self) -> Result<HealthStatus, AbstractionError> {
        Ok(HealthStatus {
            running: false,
            uptime_secs: None,
            message: None,
        })
    }

    async fn version(&self) -> Result<String, AbstractionError> {
        Ok("mock-1.0".to_string())
    }

    async fn add_contact(&self, _contact: &Contact) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn remove_contact(&self, _contact: &ContactRef) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn add_neighbor(&self, _neighbor: &Neighbor) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn remove_neighbor(&self, _node_ref: &NodeRef) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn enable_link(&self, _link_id: &str) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn disable_link(&self, _link_id: &str) -> Result<(), AbstractionError> {
        Ok(())
    }

    async fn collect_stats(&self) -> Result<BundleStatistics, AbstractionError> {
        Ok(BundleStatistics {
            bundles_sourced: 0,
            bundles_forwarded: 0,
            bundles_delivered: 0,
            bundles_expired: 0,
            bundles_queued: 0,
        })
    }

    async fn link_states(&self) -> Result<Vec<LinkState>, AbstractionError> {
        Ok(vec![])
    }
}

// Feature: dtn-abstraction-layer, Property 4: YAML Serialization Round-Trip
// **Validates: Requirements 7.5, 2.1, 2.2, 2.3, 3.1, 3.3, 4.1, 4.2, 5.1, 5.2, 5.3, 5.4, 6.1, 6.3**
proptest! {
    #[test]
    fn yaml_serialization_round_trip(config in arb_network_configuration()) {
        // Serialize to YAML
        let yaml_str = serde_yaml::to_string(&config)
            .expect("serialization to YAML should not fail for valid NetworkConfiguration");

        // Deserialize from YAML
        let deserialized: radiant_dtn_abstraction::model::NetworkConfiguration =
            serde_yaml::from_str(&yaml_str)
                .expect("deserialization from YAML should not fail for previously-serialized value");

        // Round-trip equality
        prop_assert_eq!(&config, &deserialized,
            "YAML round-trip failed.\nOriginal: {:?}\nYAML:\n{}\nDeserialized: {:?}",
            config, yaml_str, deserialized);
    }
}

// Feature: dtn-abstraction-layer, Property 5: JSON Serialization Round-Trip
// **Validates: Requirements 7.6, 2.1, 2.2, 2.3, 3.1, 3.3, 4.1, 4.2, 5.1, 5.2, 5.3, 5.4, 6.1, 6.3**
//
// JSON uses decimal floating-point representation. Some f64 bit patterns may
// experience up to 1 ULP drift when serialized to decimal and parsed back.
// This test verifies that all structural data round-trips exactly, and that
// float fields (owlt_secs, confidence) are preserved within the inherent
// precision of JSON's decimal float encoding.
proptest! {
    #[test]
    fn json_serialization_round_trip(config in arb_network_configuration()) {
        // Serialize to JSON
        let json_str = serde_json::to_string(&config)
            .expect("serialization to JSON should not fail for valid NetworkConfiguration");

        // Deserialize from JSON
        let deserialized: radiant_dtn_abstraction::model::NetworkConfiguration =
            serde_json::from_str(&json_str)
                .expect("deserialization from JSON should not fail for previously-serialized value");

        // Verify all non-float structural fields are exactly preserved
        prop_assert_eq!(&config.version, &deserialized.version);
        prop_assert_eq!(&config.backend, &deserialized.backend);
        prop_assert_eq!(&config.local_node, &deserialized.local_node);
        prop_assert_eq!(&config.neighbors, &deserialized.neighbors);
        prop_assert_eq!(&config.routing, &deserialized.routing);
        prop_assert_eq!(&config.security, &deserialized.security);
        prop_assert_eq!(&config.storage, &deserialized.storage);
        prop_assert_eq!(&config.backend_options, &deserialized.backend_options);

        // Verify contact plan contacts (integer fields exact, confidence within tolerance)
        prop_assert_eq!(
            config.contact_plan.contacts.len(),
            deserialized.contact_plan.contacts.len(),
            "Contact count mismatch"
        );
        for (i, (orig, deser)) in config.contact_plan.contacts.iter()
            .zip(deserialized.contact_plan.contacts.iter()).enumerate()
        {
            prop_assert_eq!(orig.source_node, deser.source_node,
                "contacts[{}].source_node mismatch", i);
            prop_assert_eq!(orig.dest_node, deser.dest_node,
                "contacts[{}].dest_node mismatch", i);
            prop_assert_eq!(orig.start_time, deser.start_time,
                "contacts[{}].start_time mismatch", i);
            prop_assert_eq!(orig.end_time, deser.end_time,
                "contacts[{}].end_time mismatch", i);
            prop_assert_eq!(orig.rate_bps, deser.rate_bps,
                "contacts[{}].rate_bps mismatch", i);
            // confidence: f64 may drift by at most a few ULPs through JSON decimal representation
            let ulps = (orig.confidence.to_bits() as i64 - deser.confidence.to_bits() as i64).unsigned_abs();
            prop_assert!(ulps <= 2,
                "contacts[{}].confidence ULP drift too large: {} vs {} ({} ULPs)",
                i, orig.confidence, deser.confidence, ulps);
        }

        // Verify contact plan ranges (integer fields exact, owlt_secs within tolerance)
        prop_assert_eq!(
            config.contact_plan.ranges.len(),
            deserialized.contact_plan.ranges.len(),
            "Range count mismatch"
        );
        for (i, (orig, deser)) in config.contact_plan.ranges.iter()
            .zip(deserialized.contact_plan.ranges.iter()).enumerate()
        {
            prop_assert_eq!(orig.source_node, deser.source_node,
                "ranges[{}].source_node mismatch", i);
            prop_assert_eq!(orig.dest_node, deser.dest_node,
                "ranges[{}].dest_node mismatch", i);
            // owlt_secs: f64 may drift by at most a few ULPs through JSON decimal representation
            let ulps = (orig.owlt_secs.to_bits() as i64 - deser.owlt_secs.to_bits() as i64).unsigned_abs();
            prop_assert!(ulps <= 2,
                "ranges[{}].owlt_secs ULP drift too large: {} vs {} ({} ULPs)",
                i, orig.owlt_secs, deser.owlt_secs, ulps);
        }

        // Verify that deserialized value re-serializes stably (idempotent after first trip)
        let json_str2 = serde_json::to_string(&deserialized)
            .expect("re-serialization should not fail");
        let deserialized2: radiant_dtn_abstraction::model::NetworkConfiguration =
            serde_json::from_str(&json_str2)
                .expect("second deserialization should not fail");
        prop_assert_eq!(&deserialized, &deserialized2,
            "JSON not stable after second round-trip: values diverge after initial parse");
    }
}


// Feature: dtn-abstraction-layer, Property 1: Adapter Registry Round-Trip
// **Validates: Requirements 1.1, 1.2**
proptest! {
    #[test]
    fn registry_round_trip(name in "[a-z]{3,10}") {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let registry = AdapterRegistry::new();
            let adapter: Arc<dyn BackendAdapter> = Arc::new(PropMockAdapter::new(&name));
            let adapter_name = adapter.name().to_string();

            registry.register(&name, adapter.clone()).await.unwrap();

            let retrieved = registry.get(&name).await.unwrap();

            // Verify the retrieved adapter is the same instance (same name)
            prop_assert_eq!(retrieved.name(), adapter_name.as_str());
            // Verify we get back the same Arc (pointer equality)
            prop_assert!(Arc::ptr_eq(&adapter, &retrieved));
            Ok(())
        })?;
    }
}

// Feature: dtn-abstraction-layer, Property 2: Adapter Registry Duplicate Rejection
// **Validates: Requirements 1.3**
proptest! {
    #[test]
    fn registry_duplicate_rejection(name in "[a-z]{3,10}") {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let registry = AdapterRegistry::new();
            let adapter1: Arc<dyn BackendAdapter> = Arc::new(PropMockAdapter::new(&name));
            let adapter2: Arc<dyn BackendAdapter> = Arc::new(PropMockAdapter::new(&name));

            // First registration succeeds
            registry.register(&name, adapter1).await.unwrap();

            // Second registration with same name must fail
            let result = registry.register(&name, adapter2).await;
            prop_assert!(result.is_err());

            let err = match result {
                Err(e) => e,
                Ok(_) => unreachable!(),
            };
            prop_assert_eq!(err.category, ErrorCategory::ConfigurationError);
            prop_assert!(
                err.message.contains("already registered"),
                "Error message should indicate duplicate: {}",
                err.message
            );
            Ok(())
        })?;
    }
}

// Feature: dtn-abstraction-layer, Property 3: Adapter Registry Not-Found Error
// **Validates: Requirements 1.4**
proptest! {
    #[test]
    fn registry_not_found(name in "[a-z]{3,10}") {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let registry = AdapterRegistry::new();

            // Looking up a name that was never registered must fail
            let result = registry.get(&name).await;
            prop_assert!(result.is_err());

            let err = match result {
                Err(e) => e,
                Ok(_) => unreachable!(),
            };
            prop_assert_eq!(err.category, ErrorCategory::ConfigurationError);
            prop_assert!(
                err.message.contains("not found") || err.message.contains("Adapter not found"),
                "Error message should indicate not found: {}",
                err.message
            );
            Ok(())
        })?;
    }
}


// Feature: dtn-abstraction-layer, Property 6: Default Endpoint ID Derivation
// **Validates: Requirements 2.4**
proptest! {
    #[test]
    fn default_endpoint_derivation(node_number in arb_node_number()) {
        // Create a NodeDefinition with endpoint_id = None
        let mut node = NodeDefinition {
            node_number,
            endpoint_id: None,
            callsign_eid: None,
            name: "PropTest Node".to_string(),
            services: vec![],
        };

        // Apply default endpoint derivation
        derive_default_endpoint(&mut node);

        // Assert endpoint_id is now Some(Ipn { node_number: N, service_number: 0 })
        let expected = Some(EndpointId::Ipn {
            node_number,
            service_number: 0,
        });
        prop_assert_eq!(
            &node.endpoint_id, &expected,
            "derive_default_endpoint should produce Ipn{{node_number: {}, service_number: 0}} but got {:?}",
            node_number, expected
        );
    }
}

// Feature: dtn-abstraction-layer, Property 7: Validation Rejects Referential and Temporal Inconsistencies
// **Validates: Requirements 2.5, 3.4, 4.4, 4.5**

/// Helper: build a minimal valid base config for validation property tests.
fn base_valid_config(local_node_number: u64, neighbor_node_number: u64) -> NetworkConfiguration {
    NetworkConfiguration {
        version: "1.0".to_string(),
        backend: "ion-dtn".to_string(),
        local_node: NodeDefinition {
            node_number: local_node_number,
            endpoint_id: Some(EndpointId::Ipn {
                node_number: local_node_number,
                service_number: 0,
            }),
            callsign_eid: None,
            name: "Local Node".to_string(),
            services: vec![],
        },
        neighbors: vec![Neighbor {
            node_number: neighbor_node_number,
            name: Some("Neighbor".to_string()),
            links: vec![],
            rate_limit_bps: None,
        }],
        contact_plan: ContactPlan {
            contacts: vec![Contact {
                source_node: local_node_number,
                dest_node: neighbor_node_number,
                start_time: 1000,
                end_time: 2000,
                rate_bps: 9600,
                confidence: 1.0,
            }],
            ranges: vec![Range {
                source_node: local_node_number,
                dest_node: neighbor_node_number,
                owlt_secs: 1.0,
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

proptest! {
    /// Sub-case 1: Duplicate node number — a neighbor has the same node_number as local_node.
    /// Validation must detect and reject this.
    #[test]
    fn validation_rejects_duplicate_node_number(node_number in arb_node_number()) {
        // Create config where a neighbor has the same node_number as local_node
        let mut config = base_valid_config(node_number, node_number + 1);
        config.neighbors[0].node_number = node_number; // duplicate!
        // Fix contact/range references to avoid unrelated referential errors
        config.contact_plan.contacts[0].dest_node = node_number;
        config.contact_plan.ranges[0].dest_node = node_number;

        let result = validate(&config);
        prop_assert!(result.is_err(),
            "Validation should reject duplicate node_number {}", node_number);
        let errors = result.unwrap_err();
        prop_assert!(
            errors.iter().any(|e| e.rule == "duplicate_node_number"),
            "Expected a 'duplicate_node_number' rule error, got: {:?}", errors
        );
    }

    /// Sub-case 2: Unresolved contact reference — a contact references a node not defined
    /// in local_node or neighbors.
    #[test]
    fn validation_rejects_unresolved_contact_reference(
        local_num in 1u64..=500,
        neighbor_num in 501u64..=999,
        undefined_node in 1001u64..=2000
    ) {
        let mut config = base_valid_config(local_num, neighbor_num);
        // Add a contact that references an undefined node
        config.contact_plan.contacts.push(Contact {
            source_node: local_num,
            dest_node: undefined_node,
            start_time: 3000,
            end_time: 4000,
            rate_bps: 1200,
            confidence: 1.0,
        });

        let result = validate(&config);
        prop_assert!(result.is_err(),
            "Validation should reject contact referencing undefined node {}", undefined_node);
        let errors = result.unwrap_err();
        prop_assert!(
            errors.iter().any(|e| e.rule == "referential_integrity"
                && e.path.contains("contacts")
                && e.message.contains(&undefined_node.to_string())),
            "Expected a 'referential_integrity' error for contact referencing node {}, got: {:?}",
            undefined_node, errors
        );
    }

    /// Sub-case 3: Temporal violation — a contact where start_time >= end_time.
    #[test]
    fn validation_rejects_temporal_violation(
        local_num in 1u64..=500,
        neighbor_num in 501u64..=999,
        start_time in 1000i64..=1_000_000_000
    ) {
        let mut config = base_valid_config(local_num, neighbor_num);
        // Add a contact with start_time >= end_time (use start_time == end_time)
        config.contact_plan.contacts.push(Contact {
            source_node: local_num,
            dest_node: neighbor_num,
            start_time,
            end_time: start_time, // equal → temporal violation
            rate_bps: 9600,
            confidence: 1.0,
        });

        let result = validate(&config);
        prop_assert!(result.is_err(),
            "Validation should reject contact with start_time ({}) >= end_time ({})",
            start_time, start_time);
        let errors = result.unwrap_err();
        prop_assert!(
            errors.iter().any(|e| e.rule == "temporal_order"),
            "Expected a 'temporal_order' rule error, got: {:?}", errors
        );
    }

    /// Sub-case 4: Unresolved range reference — a range references a node not defined.
    #[test]
    fn validation_rejects_unresolved_range_reference(
        local_num in 1u64..=500,
        neighbor_num in 501u64..=999,
        undefined_node in 1001u64..=2000
    ) {
        let mut config = base_valid_config(local_num, neighbor_num);
        // Add a range that references an undefined node
        config.contact_plan.ranges.push(Range {
            source_node: undefined_node,
            dest_node: local_num,
            owlt_secs: 2.5,
        });

        let result = validate(&config);
        prop_assert!(result.is_err(),
            "Validation should reject range referencing undefined node {}", undefined_node);
        let errors = result.unwrap_err();
        prop_assert!(
            errors.iter().any(|e| e.rule == "referential_integrity"
                && e.path.contains("ranges")
                && e.message.contains(&undefined_node.to_string())),
            "Expected a 'referential_integrity' error for range referencing node {}, got: {:?}",
            undefined_node, errors
        );
    }
}


// Feature: dtn-abstraction-layer, Property 9: Hardy Config Generation Produces Valid YAML
// **Validates: Requirements 8.2, 3.2, 4.3, 5.5, 6.4**
proptest! {
    #[test]
    fn hardy_config_generation_produces_valid_yaml(config in arb_network_configuration()) {
        use radiant_dtn_abstraction::adapter::hardy::config_gen::generate_hardy_config;

        let generated = generate_hardy_config(&config);

        // Assert: GeneratedConfig contains hardy.yaml
        prop_assert!(generated.files.contains_key("hardy.yaml"),
            "Generated Hardy config should contain hardy.yaml, got keys: {:?}",
            generated.files.keys().collect::<Vec<_>>());

        let yaml_content = &generated.files["hardy.yaml"];

        // Assert: Content is parseable as valid YAML
        let parsed: serde_yaml::Value = serde_yaml::from_str(yaml_content)
            .map_err(|e| proptest::test_runner::TestCaseError::Fail(
                format!("Hardy YAML not parseable: {}. Content:\n{}", e, yaml_content).into()
            ))?;

        // Assert: The YAML root is a mapping with required Hardy BPA keys
        let mapping = parsed.as_mapping()
            .ok_or_else(|| proptest::test_runner::TestCaseError::Fail(
                "Hardy YAML root is not a mapping".into()
            ))?;

        // Assert: Required top-level keys exist (real Hardy BPA format)
        let node_ids_key = serde_yaml::Value::String("node-ids".to_string());
        prop_assert!(mapping.contains_key(&node_ids_key),
            "Hardy YAML missing 'node-ids' section. Keys: {:?}",
            mapping.keys().collect::<Vec<_>>());

        let grpc_key = serde_yaml::Value::String("grpc".to_string());
        prop_assert!(mapping.contains_key(&grpc_key),
            "Hardy YAML missing 'grpc' section. Keys: {:?}",
            mapping.keys().collect::<Vec<_>>());

        let storage_key = serde_yaml::Value::String("storage".to_string());
        prop_assert!(mapping.contains_key(&storage_key),
            "Hardy YAML missing 'storage' section. Keys: {:?}",
            mapping.keys().collect::<Vec<_>>());

        let validity_key = serde_yaml::Value::String("rfc9171-validity".to_string());
        prop_assert!(mapping.contains_key(&validity_key),
            "Hardy YAML missing 'rfc9171-validity' section. Keys: {:?}",
            mapping.keys().collect::<Vec<_>>());

        // Assert: node-ids contains exactly 2 entries (ipn: and dtn://)
        let node_ids = mapping.get(&node_ids_key).unwrap();
        let node_ids_seq = node_ids.as_sequence()
            .ok_or_else(|| proptest::test_runner::TestCaseError::Fail(
                "Hardy YAML 'node-ids' is not a sequence".into()
            ))?;
        prop_assert_eq!(node_ids_seq.len(), 2,
            "Hardy YAML node-ids should have exactly 2 entries (ipn: and dtn://), got {}",
            node_ids_seq.len());

        // First node-id should be ipn: scheme
        let first_id = node_ids_seq[0].as_str()
            .ok_or_else(|| proptest::test_runner::TestCaseError::Fail(
                "First node-id is not a string".into()
            ))?;
        prop_assert!(first_id.starts_with("ipn:"),
            "First node-id should start with 'ipn:', got '{}'", first_id);

        // Second node-id should be dtn:// scheme
        let second_id = node_ids_seq[1].as_str()
            .ok_or_else(|| proptest::test_runner::TestCaseError::Fail(
                "Second node-id is not a string".into()
            ))?;
        prop_assert!(second_id.starts_with("dtn://"),
            "Second node-id should start with 'dtn://', got '{}'", second_id);

        // Assert: Contacts are NOT in config.yaml (Hardy manages via gRPC)
        let contacts_key = serde_yaml::Value::String("contacts".to_string());
        prop_assert!(!mapping.contains_key(&contacts_key),
            "Hardy config.yaml should NOT contain 'contacts' (managed via gRPC runtime API)");

        // Assert: No file content contains BPSec/encryption directive *keys*
        // (amateur radio compliance). We check for YAML keys that would indicate
        // security configuration, not arbitrary values that might contain substrings.
        // Directive patterns are checked as key-like occurrences (followed by colon).
        let forbidden_directives = [
            "bpsec:", "bpsec_policy:", "encryption:", "encrypt:",
            "hmac:", "integrity_block:", "confidentiality_block:",
            "a bpsource", "a bibrule", "a bcbrule",
        ];
        for (filename, content) in &generated.files {
            let non_comment: String = content
                .lines()
                .filter(|l| !l.trim_start().starts_with('#'))
                .collect::<Vec<&str>>()
                .join("\n")
                .to_lowercase();
            for directive in &forbidden_directives {
                prop_assert!(!non_comment.contains(directive),
                    "File '{}' contains forbidden security directive '{}' (amateur radio compliance violation)",
                    filename, directive);
            }
        }
    }
}

// Feature: dtn-abstraction-layer, Property 8: ION Config Generation Produces Valid File Set
// **Validates: Requirements 8.1, 3.2, 4.3, 5.5, 6.4**
proptest! {
    #[test]
    fn ion_config_generation_produces_valid_file_set(config in arb_network_configuration()) {
        use radiant_dtn_abstraction::adapter::ion::config_gen::generate_ion_config;

        let generated = generate_ion_config(&config);
        let node_num = config.local_node.node_number;
        let prefix = format!("node{}", node_num);

        // Assert: The GeneratedConfig contains files with keys ending in .ionrc, .bprc, .ltprc, .ipnrc
        let ionrc_key = format!("{}.ionrc", prefix);
        let bprc_key = format!("{}.bprc", prefix);
        let ltprc_key = format!("{}.ltprc", prefix);
        let ipnrc_key = format!("{}.ipnrc", prefix);

        prop_assert!(generated.files.contains_key(&ionrc_key),
            "Missing .ionrc file: expected key '{}'", ionrc_key);
        prop_assert!(generated.files.contains_key(&bprc_key),
            "Missing .bprc file: expected key '{}'", bprc_key);
        prop_assert!(generated.files.contains_key(&ltprc_key),
            "Missing .ltprc file: expected key '{}'", ltprc_key);
        prop_assert!(generated.files.contains_key(&ipnrc_key),
            "Missing .ipnrc file: expected key '{}'", ipnrc_key);

        // Assert: The .ionrc file contains the initialization command "1 {node_number}"
        let ionrc_content = &generated.files[&ionrc_key];
        let init_cmd = format!("1 {} ''", node_num);
        prop_assert!(ionrc_content.contains(&init_cmd),
            ".ionrc missing initialization command '{}'. Content:\n{}", init_cmd, ionrc_content);

        // Assert: The .bprc file contains at least one "a induct" and one "a outduct" for each
        // neighbor that has links
        let bprc_content = &generated.files[&bprc_key];
        for neighbor in &config.neighbors {
            if !neighbor.links.is_empty() {
                prop_assert!(bprc_content.contains("a induct"),
                    ".bprc missing 'a induct' entry for neighbor node {}. Content:\n{}",
                    neighbor.node_number, bprc_content);
                prop_assert!(bprc_content.contains("a outduct"),
                    ".bprc missing 'a outduct' entry for neighbor node {}. Content:\n{}",
                    neighbor.node_number, bprc_content);
            }
        }

        // Assert: No file content contains BPSec-related directives
        // (amateur radio compliance — no encryption over amateur links)
        let bpsec_directives = ["a bpsource", "a bibrule", "a bcbrule"];
        for (filename, content) in &generated.files {
            for directive in &bpsec_directives {
                prop_assert!(!content.contains(directive),
                    "File '{}' contains forbidden BPSec directive '{}'. Content:\n{}",
                    filename, directive, content);
            }
        }
    }
}


// Feature: dtn-abstraction-layer, Property 10: Unsupported Capability Rejection
// **Validates: Requirements 8.5, 10.7, 13.1, 13.2, 13.3, 13.4**

use radiant_dtn_abstraction::adapter::capability_check::{
    check_convergence_layer_support, check_hot_reconfig_capability,
    check_routing_strategy_support, HotReconfigOp,
};
use radiant_dtn_abstraction::model::convergence::ConvergenceLayerType;

/// Strategy to generate an arbitrary HotReconfigCapabilities with any
/// combination of enabled/disabled operations.
fn arb_hot_reconfig_capabilities() -> impl Strategy<Value = HotReconfigCapabilities> {
    (
        proptest::bool::ANY,
        proptest::bool::ANY,
        proptest::bool::ANY,
        proptest::bool::ANY,
        proptest::bool::ANY,
        proptest::bool::ANY,
    )
        .prop_map(|(ac, rc, an, rn, el, dl)| HotReconfigCapabilities {
            add_contact: ac,
            remove_contact: rc,
            add_neighbor: an,
            remove_neighbor: rn,
            enable_link: el,
            disable_link: dl,
        })
}

/// Strategy to generate an arbitrary subset of convergence layer types.
fn arb_convergence_layer_subset() -> impl Strategy<Value = Vec<ConvergenceLayerType>> {
    proptest::sample::subsequence(
        vec![
            ConvergenceLayerType::LtpUdp,
            ConvergenceLayerType::TcpCl,
            ConvergenceLayerType::Kiss,
            ConvergenceLayerType::Udp,
        ],
        0..=4,
    )
}

/// Strategy to generate an arbitrary subset of routing strategies.
fn arb_routing_strategy_subset() -> impl Strategy<Value = Vec<RoutingStrategy>> {
    proptest::sample::subsequence(
        vec![
            RoutingStrategy::Cgr,
            RoutingStrategy::Static,
            RoutingStrategy::Default,
        ],
        0..=3,
    )
}

/// Strategy to generate a backend name.
fn arb_backend_name() -> impl Strategy<Value = String> {
    "[a-z]{3,10}"
}

proptest! {
    /// Property 10: For any operation not in the adapter's CapabilitySet,
    /// execution returns UnsupportedOperation error. For any operation that IS
    /// in the CapabilitySet, execution returns Ok(()).
    #[test]
    fn unsupported_capability_rejection(
        hot_reconfig in arb_hot_reconfig_capabilities(),
        cls in arb_convergence_layer_subset(),
        strategies in arb_routing_strategy_subset(),
        backend_name in arb_backend_name(),
    ) {
        let caps = CapabilitySet {
            hot_reconfig: hot_reconfig.clone(),
            convergence_layers: cls.clone(),
            routing_strategies: strategies.clone(),
            security: SecurityCapabilities::none(),
        };

        // Check all 6 hot-reconfig operations
        let op_flags = [
            (HotReconfigOp::AddContact, hot_reconfig.add_contact),
            (HotReconfigOp::RemoveContact, hot_reconfig.remove_contact),
            (HotReconfigOp::AddNeighbor, hot_reconfig.add_neighbor),
            (HotReconfigOp::RemoveNeighbor, hot_reconfig.remove_neighbor),
            (HotReconfigOp::EnableLink, hot_reconfig.enable_link),
            (HotReconfigOp::DisableLink, hot_reconfig.disable_link),
        ];

        for (op, supported) in &op_flags {
            let result = check_hot_reconfig_capability(&caps, *op, &backend_name);
            if *supported {
                prop_assert!(result.is_ok(),
                    "Expected {:?} to be supported for backend '{}' but got error: {:?}",
                    op, backend_name, result.unwrap_err());
            } else {
                prop_assert!(result.is_err(),
                    "Expected {:?} to be unsupported for backend '{}' but got Ok(())",
                    op, backend_name);
                let err = result.unwrap_err();
                prop_assert_eq!(&err.category, &ErrorCategory::UnsupportedOperation,
                    "Expected UnsupportedOperation error category, got {:?}", err.category);
                prop_assert!(err.message.contains(op.operation_name()),
                    "Error message '{}' should contain operation name '{}'",
                    err.message, op.operation_name());
                prop_assert!(err.message.contains(&backend_name),
                    "Error message '{}' should contain backend name '{}'",
                    err.message, backend_name);
                prop_assert_eq!(err.context.backend.as_deref(), Some(backend_name.as_str()),
                    "Error context backend should be '{}'", backend_name);
            }
        }

        // Check all 4 convergence layer types
        let all_cl_types = [
            ConvergenceLayerType::LtpUdp,
            ConvergenceLayerType::TcpCl,
            ConvergenceLayerType::Kiss,
            ConvergenceLayerType::Udp,
        ];

        for cl_type in &all_cl_types {
            let supported = cls.contains(cl_type);
            let result = check_convergence_layer_support(&caps, *cl_type, &backend_name);
            if supported {
                prop_assert!(result.is_ok(),
                    "Expected {:?} CL to be supported for backend '{}' but got error: {:?}",
                    cl_type, backend_name, result.unwrap_err());
            } else {
                prop_assert!(result.is_err(),
                    "Expected {:?} CL to be unsupported for backend '{}' but got Ok(())",
                    cl_type, backend_name);
                let err = result.unwrap_err();
                prop_assert_eq!(&err.category, &ErrorCategory::UnsupportedOperation);
                prop_assert!(err.message.contains(&backend_name));
                prop_assert_eq!(err.context.backend.as_deref(), Some(backend_name.as_str()));
            }
        }

        // Check all 3 routing strategies
        let all_strategies = [
            RoutingStrategy::Cgr,
            RoutingStrategy::Static,
            RoutingStrategy::Default,
        ];

        for strategy in &all_strategies {
            let supported = strategies.contains(strategy);
            let result = check_routing_strategy_support(&caps, *strategy, &backend_name);
            if supported {
                prop_assert!(result.is_ok(),
                    "Expected {:?} routing to be supported for backend '{}' but got error: {:?}",
                    strategy, backend_name, result.unwrap_err());
            } else {
                prop_assert!(result.is_err(),
                    "Expected {:?} routing to be unsupported for backend '{}' but got Ok(())",
                    strategy, backend_name);
                let err = result.unwrap_err();
                prop_assert_eq!(&err.category, &ErrorCategory::UnsupportedOperation);
                prop_assert!(err.message.contains(&backend_name));
                prop_assert_eq!(err.context.backend.as_deref(), Some(backend_name.as_str()));
            }
        }
    }
}

// Feature: dtn-abstraction-layer, Property 11: Event Delivery to All Subscribers
// **Validates: Requirements 12.4, 12.5**

use radiant_dtn_abstraction::events::bus::{
    ContactAction, ContactSummary, DtnEvent, EventBus, LinkActivity,
};
use radiant_dtn_abstraction::lifecycle::EngineState;
use chrono::{TimeZone, Utc};

/// Strategy to generate an arbitrary DtnEvent.
fn arb_dtn_event() -> impl Strategy<Value = DtnEvent> {
    // Use a fixed-but-arbitrary timestamp range for reproducibility
    let arb_timestamp = (0i64..2_000_000_000).prop_map(|secs| {
        Utc.timestamp_opt(secs, 0).unwrap()
    });

    prop_oneof![
        // LinkStateChange
        (
            arb_timestamp.clone(),
            1u64..=1000,
            "[a-z]{3,12}",
            proptest::bool::ANY,
        )
            .prop_map(|(ts, node, link_id, active)| {
                DtnEvent::LinkStateChange {
                    timestamp: ts,
                    neighbor_node: node,
                    link_id,
                    new_state: if active {
                        LinkActivity::Active
                    } else {
                        LinkActivity::Inactive
                    },
                }
            }),
        // ContactPlanChange
        (
            arb_timestamp.clone(),
            proptest::bool::ANY,
            1u64..=500,
            501u64..=1000,
            0i64..=1_000_000_000,
            1_000_000_001i64..=2_000_000_000,
        )
            .prop_map(|(ts, added, src, dst, start, end)| {
                DtnEvent::ContactPlanChange {
                    timestamp: ts,
                    action: if added {
                        ContactAction::Added
                    } else {
                        ContactAction::Removed
                    },
                    contact: ContactSummary {
                        source_node: src,
                        dest_node: dst,
                        start_time: start,
                        end_time: end,
                    },
                }
            }),
        // EngineStateChange
        (
            arb_timestamp,
            prop_oneof![
                Just(EngineState::Stopped),
                Just(EngineState::Starting),
                Just(EngineState::Running),
                Just(EngineState::Stopping),
                Just(EngineState::Failed),
            ],
            prop_oneof![
                Just(EngineState::Stopped),
                Just(EngineState::Starting),
                Just(EngineState::Running),
                Just(EngineState::Stopping),
                Just(EngineState::Failed),
            ],
            proptest::option::of("[a-z ]{3,20}"),
        )
            .prop_map(|(ts, old, new, reason)| {
                DtnEvent::EngineStateChange {
                    timestamp: ts,
                    old_state: old,
                    new_state: new,
                    reason,
                }
            }),
    ]
}

proptest! {
    /// Property 11: For any DtnEvent and N subscribers (1..=10),
    /// all subscribers receive the event with correct type, valid timestamp,
    /// and matching details.
    #[test]
    fn event_delivery_to_all_subscribers(
        event in arb_dtn_event(),
        num_subscribers in 1usize..=10,
    ) {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let bus = EventBus::new(16);

            // Create N subscribers before publishing
            let mut receivers: Vec<_> = (0..num_subscribers)
                .map(|_| bus.subscribe())
                .collect();

            // Publish the event
            bus.publish(event.clone());

            // All subscribers must receive the same event
            for (i, rx) in receivers.iter_mut().enumerate() {
                let received = rx.recv().await
                    .map_err(|e| proptest::test_runner::TestCaseError::Fail(
                        format!("Subscriber {} failed to receive: {:?}", i, e).into()
                    ))?;

                // Verify the received event matches the published event
                prop_assert_eq!(&received, &event,
                    "Subscriber {} received different event.\nExpected: {:?}\nGot: {:?}",
                    i, event, received);

                // Verify timestamp is valid (non-zero, within our generated range)
                match &received {
                    DtnEvent::LinkStateChange { timestamp, .. }
                    | DtnEvent::ContactPlanChange { timestamp, .. }
                    | DtnEvent::EngineStateChange { timestamp, .. } => {
                        prop_assert!(timestamp.timestamp() >= 0,
                            "Subscriber {} received event with invalid timestamp: {:?}",
                            i, timestamp);
                    }
                }
            }

            Ok(())
        })?;
    }
}

// Feature: dtn-abstraction-layer, Property 12: Error Mapping Preserves Context
// **Validates: Requirements 14.3, 14.4**

use radiant_dtn_abstraction::error_mapping::{
    enrich_error, BackendErrorMapper, HardyErrorMapper, IonErrorMapper,
};

/// Strategy to select which error mapping method to invoke.
#[derive(Debug, Clone, Copy)]
enum ErrorMappingMethod {
    Lifecycle,
    Runtime,
    Communication,
}

/// Strategy to select which backend mapper to use.
#[derive(Debug, Clone, Copy)]
enum BackendMapper {
    Ion,
    Hardy,
}

fn arb_error_mapping_method() -> impl Strategy<Value = ErrorMappingMethod> {
    prop_oneof![
        Just(ErrorMappingMethod::Lifecycle),
        Just(ErrorMappingMethod::Runtime),
        Just(ErrorMappingMethod::Communication),
    ]
}

fn arb_backend_mapper() -> impl Strategy<Value = BackendMapper> {
    prop_oneof![
        Just(BackendMapper::Ion),
        Just(BackendMapper::Hardy),
    ]
}

proptest! {
    /// Property 12: For any backend error mapped through the abstraction layer,
    /// the resulting AbstractionError contains original error info in backend_code
    /// and complete ErrorContext (operation, backend name). When enriched with a
    /// resource, the resource is also preserved.
    #[test]
    fn error_mapping_preserves_context(
        raw_error in "[a-zA-Z0-9 :_\\-\\.]{1,100}",
        operation in "[a-z_]{3,30}",
        resource in "[a-zA-Z0-9 \\-_/]{3,50}",
        method in arb_error_mapping_method(),
        backend in arb_backend_mapper(),
    ) {
        // Select mapper based on generated backend choice
        let mapper: &dyn BackendErrorMapper = match backend {
            BackendMapper::Ion => &IonErrorMapper,
            BackendMapper::Hardy => &HardyErrorMapper,
        };

        let expected_backend_name = mapper.backend_name().to_string();

        // Map the error using the selected method
        let err = match method {
            ErrorMappingMethod::Lifecycle => mapper.map_lifecycle_error(&raw_error, &operation),
            ErrorMappingMethod::Runtime => mapper.map_runtime_error(&raw_error, &operation),
            ErrorMappingMethod::Communication => mapper.map_communication_error(&raw_error, &operation),
        };

        // Assert: backend_code preserves the original raw error string
        prop_assert_eq!(
            err.backend_code.as_deref(),
            Some(raw_error.as_str()),
            "backend_code should preserve original error '{}', got {:?}",
            raw_error, err.backend_code
        );

        // Assert: context.operation matches the operation name
        prop_assert_eq!(
            &err.context.operation, &operation,
            "context.operation should be '{}', got '{}'",
            operation, err.context.operation
        );

        // Assert: context.backend matches the mapper's backend name
        prop_assert_eq!(
            err.context.backend.as_deref(),
            Some(expected_backend_name.as_str()),
            "context.backend should be '{}', got {:?}",
            expected_backend_name, err.context.backend
        );

        // Assert: category matches the expected category for the method
        let expected_category = match method {
            ErrorMappingMethod::Lifecycle => ErrorCategory::LifecycleError,
            ErrorMappingMethod::Runtime => ErrorCategory::RuntimeError,
            ErrorMappingMethod::Communication => ErrorCategory::CommunicationError,
        };
        prop_assert_eq!(
            &err.category, &expected_category,
            "category should be {:?}, got {:?}",
            expected_category, err.category
        );

        // Enrich with resource context and verify preservation
        let enriched = enrich_error(err, &resource);

        // Assert: resource is set after enrichment
        prop_assert_eq!(
            enriched.context.resource.as_deref(),
            Some(resource.as_str()),
            "context.resource should be '{}' after enrichment, got {:?}",
            resource, enriched.context.resource
        );

        // Assert: all other fields are still preserved after enrichment
        prop_assert_eq!(
            enriched.backend_code.as_deref(),
            Some(raw_error.as_str()),
            "backend_code should still be '{}' after enrichment, got {:?}",
            raw_error, enriched.backend_code
        );
        prop_assert_eq!(
            &enriched.context.operation, &operation,
            "context.operation should still be '{}' after enrichment, got '{}'",
            operation, enriched.context.operation
        );
        prop_assert_eq!(
            enriched.context.backend.as_deref(),
            Some(expected_backend_name.as_str()),
            "context.backend should still be '{}' after enrichment, got {:?}",
            expected_backend_name, enriched.context.backend
        );
        prop_assert_eq!(
            &enriched.category, &expected_category,
            "category should still be {:?} after enrichment, got {:?}",
            expected_category, enriched.category
        );
    }
}


// Feature: dtn-abstraction-layer, Property 13: API Error Response Code Mapping
// **Validates: Requirements 15.4, 15.5**

use radiant_dtn_abstraction::api::handlers::error_to_status;

/// Strategy to generate an arbitrary ErrorCategory.
fn arb_error_category() -> impl Strategy<Value = ErrorCategory> {
    prop_oneof![
        Just(ErrorCategory::ValidationError),
        Just(ErrorCategory::ConfigurationError),
        Just(ErrorCategory::LifecycleError),
        Just(ErrorCategory::RuntimeError),
        Just(ErrorCategory::UnsupportedOperation),
        Just(ErrorCategory::CommunicationError),
    ]
}

proptest! {
    /// Property 13: For any AbstractionError, the API returns the correct HTTP
    /// status code corresponding to the error category, and the response body
    /// deserializes to a valid AbstractionError structure.
    #[test]
    fn api_error_response_code_mapping(
        category in arb_error_category(),
        message in "[a-zA-Z0-9 ]{5,50}",
        operation in "[a-z_]{3,20}",
        resource in proptest::option::of("[a-zA-Z0-9_\\-]{3,20}"),
        backend in proptest::option::of("[a-z]{3,10}"),
        backend_code in proptest::option::of("[A-Z_0-9]{3,10}"),
    ) {
        use axum::http::StatusCode;

        // Construct an AbstractionError with the generated category
        let mut err = AbstractionError::new(category.clone(), message.clone(), operation.clone());
        if let Some(ref r) = resource {
            err = err.with_resource(r.clone());
        }
        if let Some(ref b) = backend {
            err = err.with_backend(b.clone());
        }
        if let Some(ref bc) = backend_code {
            err = err.with_backend_code(bc.clone());
        }

        // Determine expected status code based on category
        let expected_status = match &category {
            ErrorCategory::ValidationError => StatusCode::BAD_REQUEST,
            ErrorCategory::UnsupportedOperation => StatusCode::BAD_REQUEST,
            ErrorCategory::ConfigurationError => {
                // "not found" in message → 404, otherwise → 409
                if message.to_lowercase().contains("not found") {
                    StatusCode::NOT_FOUND
                } else {
                    StatusCode::CONFLICT
                }
            }
            ErrorCategory::LifecycleError => StatusCode::INTERNAL_SERVER_ERROR,
            ErrorCategory::RuntimeError => StatusCode::INTERNAL_SERVER_ERROR,
            ErrorCategory::CommunicationError => StatusCode::BAD_GATEWAY,
        };

        // Map the error to HTTP status
        let actual_status = error_to_status(&err);

        prop_assert_eq!(actual_status, expected_status,
            "For category {:?} with message '{}', expected status {:?} but got {:?}",
            category, message, expected_status, actual_status);

        // Verify the error can be serialized to JSON and deserialized back
        let json = serde_json::to_string(&err)
            .expect("AbstractionError should serialize to JSON");
        let deserialized: AbstractionError = serde_json::from_str(&json)
            .expect("JSON should deserialize back to AbstractionError");

        // Verify structural integrity after round-trip
        prop_assert_eq!(&deserialized.category, &category);
        prop_assert_eq!(&deserialized.message, &message);
        prop_assert_eq!(&deserialized.context.operation, &operation);
        prop_assert_eq!(&deserialized.context.resource, &resource);
        prop_assert_eq!(&deserialized.context.backend, &backend);
        prop_assert_eq!(&deserialized.backend_code, &backend_code);
    }
}
