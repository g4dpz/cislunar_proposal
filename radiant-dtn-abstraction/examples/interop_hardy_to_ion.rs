//! Hardy → ION bundle delivery test over LTP/UDP.
//!
//! Demonstrates sending a bundle FROM Hardy TO ION using:
//! - Abstraction layer for config generation and ION lifecycle
//! - Hardy BPA's gRPC Application service for bundle injection
//! - hardy-ltp-server for LTP/UDP transport
//! - ION's bprecvfile for reception verification
//!
//! Run: cargo run --example interop_hardy_to_ion --features interop-network

use std::collections::HashMap;
use std::path::Path;
use std::time::Duration;

use tokio::sync::mpsc;
use tokio_stream::wrappers::ReceiverStream;
use tonic::Streaming;

use hardy_ltp_proto::proto::service::*;
use hardy_ltp_proto::proto::service::application_client::ApplicationClient;

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
        storage: None,
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
        storage: None,
        backend_options: HashMap::new(),
    }
}

/// Send a bundle via Hardy BPA's gRPC Application service.
async fn hardy_send_bundle(
    grpc_endpoint: &str,
    destination: &str,
    payload: &[u8],
) -> Result<String, String> {
    // Connect to the Application service
    let mut client = ApplicationClient::connect(grpc_endpoint.to_string())
        .await
        .map_err(|e| format!("Failed to connect to Hardy BPA Application service: {}", e))?;

    // Open bidirectional stream
    let (tx, rx) = mpsc::channel::<AppToBpa>(16);
    let response = client
        .register(ReceiverStream::new(rx))
        .await
        .map_err(|e| format!("Failed to open Application stream: {}", e))?;
    let mut bpa_stream: Streaming<BpaToApp> = response.into_inner();

    // Send RegisterRequest (register as ipn:20.1 sender endpoint)
    tx.send(AppToBpa {
        msg_id: 1,
        msg: Some(app_to_bpa::Msg::Register(RegisterRequest {
            service_id: Some(register_request::ServiceId::Ipn(1)),
        })),
    })
    .await
    .map_err(|e| format!("Failed to send register: {}", e))?;

    // Wait for RegisterResponse
    let reg_msg = bpa_stream
        .message()
        .await
        .map_err(|e| format!("Stream error waiting for register response: {}", e))?
        .ok_or("Stream closed before register response")?;

    match reg_msg.msg {
        Some(bpa_to_app::Msg::Register(resp)) => {
            println!("    Registered as: {}", resp.endpoint_id);
        }
        Some(bpa_to_app::Msg::Status(status)) => {
            return Err(format!("Registration failed: {:?}", status));
        }
        other => {
            return Err(format!("Unexpected register response: {:?}", other));
        }
    }

    // Send AppSendRequest
    tx.send(AppToBpa {
        msg_id: 2,
        msg: Some(app_to_bpa::Msg::Send(AppSendRequest {
            destination: destination.to_string(),
            payload: payload.to_vec(),
            lifetime: 3600000, // 1 hour in ms
            options: None,
        })),
    })
    .await
    .map_err(|e| format!("Failed to send bundle: {}", e))?;

    // Wait for SendResponse
    let send_msg = bpa_stream
        .message()
        .await
        .map_err(|e| format!("Stream error waiting for send response: {}", e))?
        .ok_or("Stream closed before send response")?;

    match send_msg.msg {
        Some(bpa_to_app::Msg::Send(resp)) => {
            Ok(resp.bundle_id)
        }
        Some(bpa_to_app::Msg::Status(status)) => {
            Err(format!("Send failed: {:?}", status))
        }
        other => {
            Err(format!("Unexpected send response: {:?}", other))
        }
    }
}

#[tokio::main]
async fn main() {
    println!("=== Hardy→ION 20KB Interop Test (abstraction layer) ===\n");

    // Clean slate
    let _ = std::fs::remove_dir_all(WORK_DIR);
    let _ = tokio::process::Command::new("killm")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .status().await;

    // Step 1: Generate configs
    println!("[1/7] Generating configs via abstraction layer...");
    let hardy_cfg = hardy_config();
    let hardy_generated = generate_hardy_config(&hardy_cfg);

    // Write Hardy configs
    let hardy_dir = Path::new(WORK_DIR).join("hardy/config");
    std::fs::create_dir_all(&hardy_dir).unwrap();
    for (name, content) in &hardy_generated.files {
        std::fs::write(hardy_dir.join(name), content).unwrap();
    }
    println!("  Hardy: {:?}", hardy_generated.files.keys().collect::<Vec<_>>());

    // Write ION configs via abstraction layer
    let ion_dir = Path::new(WORK_DIR).join("ion");
    std::fs::create_dir_all(&ion_dir).unwrap();
    let ion_cfg = ion_config();
    let ion_generated = generate_ion_config(&ion_cfg);
    for (name, content) in &ion_generated.files {
        std::fs::write(ion_dir.join(name), content).unwrap();
    }
    println!("  ION: {} files", ion_generated.files.len());

    // Step 2: Start Hardy BPA
    println!("\n[2/7] Starting Hardy BPA...");
    let hardy_yaml = hardy_dir.join("hardy.yaml");
    let mut bpa_child = tokio::process::Command::new(HARDY_BPA_BIN)
        .arg("-c").arg(&hardy_yaml)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .expect("Failed to start Hardy BPA");
    tokio::time::sleep(Duration::from_secs(2)).await;
    println!("  PID {}", bpa_child.id().unwrap());

    // Step 3: Start hardy-ltp-server
    println!("[3/7] Starting hardy-ltp-server...");
    let ltp_yaml = hardy_dir.join("ltp-cla.yaml");
    let ltp_log = std::fs::File::create(Path::new(WORK_DIR).join("ltp-cla.log")).unwrap();
    let ltp_log2 = ltp_log.try_clone().unwrap();
    let mut ltp_child = tokio::process::Command::new(HARDY_LTP_BIN)
        .arg(&ltp_yaml)
        .env("RUST_LOG", "hardy_ltp_cla=trace,hardy_ltp_server=info,hardy_ltp_grpc=info")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::from(ltp_log))
        .stderr(std::process::Stdio::from(ltp_log2))
        .spawn()
        .expect("Failed to start hardy-ltp-server");
    tokio::time::sleep(Duration::from_secs(2)).await;
    println!("  PID {}", ltp_child.id().unwrap());

    // Step 4: Start ION
    println!("[4/7] Starting ION via IonLifecycle...");
    let _ = tokio::process::Command::new("killm")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .status().await;
    tokio::time::sleep(Duration::from_millis(500)).await;
    let mut ion = IonLifecycle::new(None);
    ion.start(&ion_dir).await.expect("ION start failed");
    tokio::time::sleep(Duration::from_secs(2)).await;
    println!("  ION running");

    // Step 5: Start bprecvfile on ION side to receive the bundle
    println!("[5/7] Starting bprecvfile on ION (ipn:10.1)...");
    let recv_dir = Path::new(WORK_DIR).join("received");
    std::fs::create_dir_all(&recv_dir).unwrap();
    let mut recv_child = tokio::process::Command::new("bprecvfile")
        .args(["ipn:10.1", "1"])
        .current_dir(&recv_dir)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .expect("Failed to start bprecvfile");
    tokio::time::sleep(Duration::from_secs(1)).await;
    println!("  bprecvfile waiting for bundle");

    // Step 6: Send bundle from Hardy to ION
    println!("[6/7] Sending 20KB bundle from Hardy (ipn:20.1 → ipn:10.1)...");
    let payload: Vec<u8> = (0..20480).map(|i| (i % 256) as u8).collect();

    match hardy_send_bundle("http://[::1]:50051", "ipn:10.1", &payload).await {
        Ok(bundle_id) => {
            println!("  Bundle accepted: {}", bundle_id);
        }
        Err(e) => {
            println!("  ERROR: {}", e);
            // Teardown and exit
            let _ = recv_child.kill().await;
            let _ = ion.stop().await;
            let _ = ltp_child.kill().await;
            let _ = bpa_child.kill().await;
            let _ = tokio::process::Command::new("killm")
                .stdin(std::process::Stdio::null())
                .output().await;
            std::process::exit(1);
        }
    }

    // Step 7: Wait for delivery and verify
    println!("[7/7] Waiting for ION to receive bundle...");
    tokio::time::sleep(Duration::from_secs(10)).await;

    // Check if bprecvfile created a file
    let received_files: Vec<_> = std::fs::read_dir(&recv_dir)
        .unwrap()
        .filter_map(|e| e.ok())
        .filter(|e| e.path().is_file())
        .collect();

    // Also check LTP CLA log for export activity
    let ltp_log_content = std::fs::read_to_string(Path::new(WORK_DIR).join("ltp-cla.log"))
        .unwrap_or_default();
    let has_export = ltp_log_content.contains("created export session");

    if !received_files.is_empty() {
        let received_path = received_files[0].path();
        let received_data = std::fs::read(&received_path).unwrap();
        let payload_match = received_data == payload;
        println!("\n=== SUCCESS ===");
        println!("  Received file: {} ({} bytes)", received_path.display(), received_data.len());
        println!("  Payload matches: {}", payload_match);
        if has_export {
            println!("  LTP export session created (Hardy→ION path confirmed)");
        }
    } else {
        println!("\n=== PARTIAL ===");
        if has_export {
            println!("  LTP CLA created export session (bundle left Hardy)");
            println!("  But bprecvfile didn't capture it (may need more time or ION config)");
        } else {
            println!("  No export session in LTP CLA log");
            println!("  Bundle may not have been routed to the LTP CLA");
        }
        println!("\n  LTP CLA log (last 10 lines):");
        for line in ltp_log_content.lines().rev().take(10).collect::<Vec<_>>().into_iter().rev() {
            println!("    {}", line);
        }
    }

    // Teardown
    println!("\nTearing down...");
    let _ = recv_child.kill().await;
    let _ = ion.stop().await;
    let _ = ltp_child.kill().await;
    let _ = bpa_child.kill().await;
    let _ = tokio::process::Command::new("killm")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .output().await;
    println!("Done.");

    if received_files.is_empty() {
        std::process::exit(1);
    }
}
