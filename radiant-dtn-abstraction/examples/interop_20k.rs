//! ION → Hardy 20KB LTP interop test using the radiant-dtn-abstraction layer.
//!
//! Demonstrates the abstraction layer generating configs and managing lifecycles
//! for both ION-DTN and Hardy, with the hardy-ltp-cla bridging LTP/UDP to gRPC.
//!
//! Run: cargo run --example interop_20k --features interop-network

use std::collections::HashMap;
use std::path::Path;
use std::time::Duration;

use radiant_dtn_abstraction::adapter::hardy::config_gen::generate_hardy_config;
use radiant_dtn_abstraction::adapter::ion::config_gen::generate_ion_config;
use radiant_dtn_abstraction::adapter::ion::lifecycle::IonLifecycle;
use radiant_dtn_abstraction::model::{
    Contact, ContactPlan, ConvergenceLayerLink, EndpointId, Neighbor, NetworkConfiguration,
    NodeDefinition, Range, RoutingConfig, RoutingStrategy, SecurityConfig, ServiceDemux,
};

const ION_NODE: u64 = 10;
const HARDY_NODE: u64 = 20;
const HARDY_LTP_BIN: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/../hardy-ltp-cla/target/debug/hardy-ltp-server"
);
const HARDY_BPA_BIN: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/../hardy/target/debug/hardy-bpa-server"
);
const WORK_DIR: &str = "/tmp/radiant-interop";

fn ion_config() -> NetworkConfiguration {
    NetworkConfiguration {
        version: "1.0".to_string(),
        backend: "ion-dtn".to_string(),
        local_node: NodeDefinition {
            node_number: ION_NODE,
            endpoint_id: Some(EndpointId::Ipn { node_number: ION_NODE, service_number: 0 }),
            callsign_eid: None,
            name: "ION Node".to_string(),
            services: vec![ServiceDemux { service_number: 1, description: Some("delivery".to_string()) }],
        },
        neighbors: vec![Neighbor {
            node_number: HARDY_NODE,
            name: Some("Hardy Node".to_string()),
            links: vec![ConvergenceLayerLink::LtpUdp {
                id: "ltp-to-hardy".to_string(),
                local_engine_id: ION_NODE,
                remote_engine_id: HARDY_NODE,
                remote_host: "127.0.0.1".to_string(),
                remote_port: 1113,
                local_port: 2113,
                mtu: Some(1400),
                segment_rate: None,
            }],
            rate_limit_bps: None,
        }],
        contact_plan: ContactPlan {
            contacts: vec![
                Contact { source_node: ION_NODE, dest_node: HARDY_NODE, start_time: 0, end_time: 86400, rate_bps: 10_000_000, confidence: 1.0 },
                Contact { source_node: HARDY_NODE, dest_node: ION_NODE, start_time: 0, end_time: 86400, rate_bps: 10_000_000, confidence: 1.0 },
            ],
            ranges: vec![
                Range { source_node: ION_NODE, dest_node: HARDY_NODE, owlt_secs: 0.001 },
                Range { source_node: HARDY_NODE, dest_node: ION_NODE, owlt_secs: 0.001 },
            ],
        },
        routing: RoutingConfig { strategy: RoutingStrategy::Cgr, static_routes: vec![] },
        security: Some(SecurityConfig { enabled: false }),
        storage: Some(radiant_dtn_abstraction::model::StorageConfig {
            path: format!("{}/ion", WORK_DIR),
            max_bytes: Some(104_857_600),
        }),
        backend_options: HashMap::new(),
    }
}

fn hardy_config() -> NetworkConfiguration {
    NetworkConfiguration {
        version: "1.0".to_string(),
        backend: "hardy".to_string(),
        local_node: NodeDefinition {
            node_number: HARDY_NODE,
            endpoint_id: Some(EndpointId::Ipn { node_number: HARDY_NODE, service_number: 0 }),
            callsign_eid: None,
            name: "Hardy Node".to_string(),
            services: vec![ServiceDemux { service_number: 1, description: Some("delivery".to_string()) }],
        },
        neighbors: vec![Neighbor {
            node_number: ION_NODE,
            name: Some("ION Node".to_string()),
            links: vec![ConvergenceLayerLink::LtpUdp {
                id: "ltp-to-ion".to_string(),
                local_engine_id: HARDY_NODE,
                remote_engine_id: ION_NODE,
                remote_host: "127.0.0.1".to_string(),
                remote_port: 2113,
                local_port: 1113,
                mtu: Some(1400),
                segment_rate: None,
            }],
            rate_limit_bps: None,
        }],
        contact_plan: ContactPlan {
            contacts: vec![
                Contact { source_node: ION_NODE, dest_node: HARDY_NODE, start_time: 0, end_time: 86400, rate_bps: 10_000_000, confidence: 1.0 },
                Contact { source_node: HARDY_NODE, dest_node: ION_NODE, start_time: 0, end_time: 86400, rate_bps: 10_000_000, confidence: 1.0 },
            ],
            ranges: vec![
                Range { source_node: ION_NODE, dest_node: HARDY_NODE, owlt_secs: 0.001 },
                Range { source_node: HARDY_NODE, dest_node: ION_NODE, owlt_secs: 0.001 },
            ],
        },
        routing: RoutingConfig { strategy: RoutingStrategy::Cgr, static_routes: vec![] },
        security: Some(SecurityConfig { enabled: false }),
        storage: None, // In-memory (no sqlite feature needed)
        backend_options: HashMap::new(),
    }
}

#[tokio::main]
async fn main() {
    println!("=== ION→Hardy 1MB Interop Test (abstraction layer) ===\n");

    // Clean slate
    let _ = std::fs::remove_dir_all(WORK_DIR);
    let _ = tokio::process::Command::new("killm").output().await;

    // Step 1: Generate all configs via abstraction layer
    println!("[1/6] Generating configs via abstraction layer...");

    let ion_cfg = ion_config();
    let hardy_cfg = hardy_config();

    // Generate configs via abstraction layer
    let ion_generated = generate_ion_config(&ion_cfg);
    let hardy_generated = generate_hardy_config(&hardy_cfg);

    // Write ION configs (now fully generated by abstraction layer including loopback entries)
    let ion_dir = Path::new(WORK_DIR).join("ion");
    std::fs::create_dir_all(&ion_dir).unwrap();
    for (name, content) in &ion_generated.files {
        std::fs::write(ion_dir.join(name), content).unwrap();
    }
    println!("  ION: {} files ({:?})", ion_generated.files.len(),
        ion_generated.files.keys().collect::<Vec<_>>());

    // Write Hardy configs
    let hardy_dir = Path::new(WORK_DIR).join("hardy/config");
    std::fs::create_dir_all(&hardy_dir).unwrap();
    for (name, content) in &hardy_generated.files {
        std::fs::write(hardy_dir.join(name), content).unwrap();
    }
    println!("  Hardy: {} files ({:?})", hardy_generated.files.len(),
        hardy_generated.files.keys().collect::<Vec<_>>());

    // Step 2: Start Hardy BPA
    println!("\n[2/6] Starting Hardy BPA...");
    let hardy_yaml = hardy_dir.join("hardy.yaml");
    let mut bpa_child = tokio::process::Command::new(HARDY_BPA_BIN)
        .arg("-c").arg(&hardy_yaml)
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .expect("Failed to start Hardy BPA");
    tokio::time::sleep(Duration::from_secs(2)).await;
    println!("  PID {}", bpa_child.id().unwrap());

    // Step 3: Start hardy-ltp-server
    println!("[3/6] Starting hardy-ltp-server...");
    let ltp_yaml = hardy_dir.join("ltp-cla.yaml");
    let ltp_log = std::fs::File::create(Path::new(WORK_DIR).join("ltp-cla.log")).unwrap();
    let ltp_log_clone = ltp_log.try_clone().unwrap();
    let mut ltp_child = tokio::process::Command::new(HARDY_LTP_BIN)
        .arg(&ltp_yaml)
        .env("RUST_LOG", "hardy_ltp_cla=trace,hardy_ltp_server=info,hardy_ltp_grpc=info")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::from(ltp_log))
        .stderr(std::process::Stdio::from(ltp_log_clone))
        .spawn()
        .expect("Failed to start hardy-ltp-server");
    tokio::time::sleep(Duration::from_secs(2)).await;
    println!("  PID {}", ltp_child.id().unwrap());

    // Step 4: Start ION via IonLifecycle
    println!("[4/6] Starting ION via IonLifecycle...");
    // Ensure clean ION state (no leftover shared memory)
    let _ = tokio::process::Command::new("killm")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .status().await;
    tokio::time::sleep(Duration::from_millis(500)).await;
    let mut ion = IonLifecycle::new(None);
    ion.start(&ion_dir).await.expect("ION start failed");
    tokio::time::sleep(Duration::from_secs(2)).await;
    let health = ion.health().await.expect("ION health failed");
    println!("  ION running: {}", health.running);

    // Step 5: Send 20KB bundle
    println!("[5/6] Sending 1MB bundle (ipn:10.1 → ipn:20.1)...");
    let payload_dir = Path::new(WORK_DIR).join("payloads");
    std::fs::create_dir_all(&payload_dir).unwrap();
    let payload_path = payload_dir.join("1mb.dat");
    let payload: Vec<u8> = (0..1048576).map(|i| (i % 256) as u8).collect();
    std::fs::write(&payload_path, &payload).unwrap();

    let send_result = tokio::process::Command::new("bpsendfile")
        .args(["ipn:10.1", "ipn:20.1", payload_path.to_str().unwrap()])
        .output()
        .await
        .expect("bpsendfile failed");
    assert!(send_result.status.success(), "bpsendfile failed: {:?}",
        String::from_utf8_lossy(&send_result.stderr));
    println!("  Sent!");

    // Step 6: Wait and verify
    println!("[6/6] Waiting for LTP delivery...");
    tokio::time::sleep(Duration::from_secs(10)).await;

    let log = std::fs::read_to_string(Path::new(WORK_DIR).join("ltp-cla.log"))
        .unwrap_or_default();
    let delivered = log.contains("bundle dispatched to BPA");

    if delivered {
        let bundle_line = log.lines()
            .find(|l| l.contains("bundle dispatched"))
            .unwrap_or("");
        println!("\n=== SUCCESS ===");
        println!("  {}", bundle_line.trim());
    } else {
        println!("\n=== FAILED ===");
        println!("  LTP CLA log:\n{}", log);
    }

    // Teardown
    println!("\nTearing down...");
    let _ = ion.stop().await;
    let _ = ltp_child.kill().await;
    let _ = bpa_child.kill().await;
    let _ = tokio::process::Command::new("killm")
        .stdin(std::process::Stdio::null())
        .output().await;
    println!("Done.");

    if !delivered {
        std::process::exit(1);
    }
}
