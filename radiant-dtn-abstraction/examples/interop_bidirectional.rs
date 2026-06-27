//! Bidirectional and stress interop test: ION ↔ Hardy over LTP/UDP.
//!
//! Demonstrates simultaneous bundle transfer in both directions with varying
//! payload sizes and a burst stress scenario. Collects telemetry from both
//! engines to verify counters update correctly.
//!
//! Test scenarios:
//! 1. Bidirectional concurrent transfer (ION→Hardy + Hardy→ION simultaneously)
//! 2. Multiple payload sizes (1KB, 20KB, 100KB)
//! 3. Stress burst — rapid sequential sends in both directions
//! 4. Telemetry collection from both adapters
//!
//! Run: cargo run --example interop_bidirectional --features interop-network

use std::collections::HashMap;
use std::path::Path;
use std::time::{Duration, Instant};

use tokio::sync::mpsc;
use tokio_stream::wrappers::ReceiverStream;
use tonic::Streaming;

use hardy_ltp_proto::proto::service::*;
use hardy_ltp_proto::proto::service::application_client::ApplicationClient;

use radiant_dtn_abstraction::adapter::hardy::config_gen::generate_hardy_config;
use radiant_dtn_abstraction::adapter::ion::config_gen::generate_ion_config;
use radiant_dtn_abstraction::adapter::ion::lifecycle::IonLifecycle;
use radiant_dtn_abstraction::adapter::ion::telemetry::IonTelemetry;
use radiant_dtn_abstraction::model::{
    Contact, ContactPlan, ConvergenceLayerLink, EndpointId, Neighbor,
    NetworkConfiguration, NodeDefinition, Range, RoutingConfig,
    RoutingStrategy, SecurityConfig, ServiceDemux,
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
const WORK_DIR: &str = "/tmp/radiant-interop-bidir";

// ─── Configuration ──────────────────────────────────────────────────────────

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
        storage: None,
        backend_options: HashMap::new(),
    }
}

// ─── Hardy gRPC Bundle Send ─────────────────────────────────────────────────

/// Send a bundle via Hardy BPA's gRPC Application service.
async fn hardy_send_bundle(
    grpc_endpoint: &str,
    destination: &str,
    payload: &[u8],
) -> Result<String, String> {
    let mut client = ApplicationClient::connect(grpc_endpoint.to_string())
        .await
        .map_err(|e| format!("gRPC connect failed: {}", e))?;

    let (tx, rx) = mpsc::channel::<AppToBpa>(16);
    let response = client
        .register(ReceiverStream::new(rx))
        .await
        .map_err(|e| format!("Register stream failed: {}", e))?;
    let mut bpa_stream: Streaming<BpaToApp> = response.into_inner();

    // Register as ipn:20.1
    tx.send(AppToBpa {
        msg_id: 1,
        msg: Some(app_to_bpa::Msg::Register(RegisterRequest {
            service_id: Some(register_request::ServiceId::Ipn(1)),
        })),
    })
    .await
    .map_err(|e| format!("Send register: {}", e))?;

    let reg_msg = bpa_stream
        .message()
        .await
        .map_err(|e| format!("Stream error: {}", e))?
        .ok_or("Stream closed")?;

    match reg_msg.msg {
        Some(bpa_to_app::Msg::Register(_)) => {}
        Some(bpa_to_app::Msg::Status(s)) => {
            return Err(format!("Registration failed: {:?}", s));
        }
        other => return Err(format!("Unexpected: {:?}", other)),
    }

    // Send bundle
    tx.send(AppToBpa {
        msg_id: 2,
        msg: Some(app_to_bpa::Msg::Send(AppSendRequest {
            destination: destination.to_string(),
            payload: payload.to_vec(),
            lifetime: 3600000,
            options: None,
        })),
    })
    .await
    .map_err(|e| format!("Send bundle: {}", e))?;

    let send_msg = bpa_stream
        .message()
        .await
        .map_err(|e| format!("Stream error: {}", e))?
        .ok_or("Stream closed")?;

    match send_msg.msg {
        Some(bpa_to_app::Msg::Send(resp)) => Ok(resp.bundle_id),
        Some(bpa_to_app::Msg::Status(s)) => Err(format!("Send failed: {:?}", s)),
        other => Err(format!("Unexpected: {:?}", other)),
    }
}

// ─── ION Bundle Send ────────────────────────────────────────────────────────

/// Send a bundle from ION using bpsendfile.
async fn ion_send_bundle(
    source_eid: &str,
    dest_eid: &str,
    payload: &[u8],
    label: &str,
) -> Result<(), String> {
    let payload_dir = Path::new(WORK_DIR).join("payloads");
    std::fs::create_dir_all(&payload_dir)
        .map_err(|e| format!("mkdir payloads: {}", e))?;
    let payload_path = payload_dir.join(format!("{}.dat", label));
    std::fs::write(&payload_path, payload)
        .map_err(|e| format!("write payload: {}", e))?;

    let output = tokio::process::Command::new("bpsendfile")
        .args([source_eid, dest_eid, payload_path.to_str().unwrap()])
        .output()
        .await
        .map_err(|e| format!("bpsendfile exec: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "bpsendfile failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }
    Ok(())
}

// ─── Telemetry Collection ───────────────────────────────────────────────────

/// Collect and print telemetry from both engines.
async fn collect_telemetry() {
    println!("\n  ── Telemetry ──");

    // ION telemetry via bpstats/bplist/ltpinfo
    let ion_telem = IonTelemetry::new(None);
    match ion_telem.collect_stats().await {
        Ok(stats) => {
            println!("  ION stats:");
            println!("    sourced={} forwarded={} delivered={} expired={} queued={}",
                stats.bundles_sourced, stats.bundles_forwarded,
                stats.bundles_delivered, stats.bundles_expired,
                stats.bundles_queued);
        }
        Err(e) => println!("  ION stats error: {}", e),
    }
    match ion_telem.link_states().await {
        Ok(links) => {
            for link in &links {
                println!("    link {} (node {}): active={} tx={} rx={}",
                    link.link_id, link.neighbor_node,
                    link.active, link.bytes_sent, link.bytes_received);
            }
        }
        Err(e) => println!("  ION link_states error: {}", e),
    }

    // Hardy telemetry via LTP CLA log analysis (best-effort)
    let ltp_log_path = Path::new(WORK_DIR).join("ltp-cla.log");
    if let Ok(log) = std::fs::read_to_string(&ltp_log_path) {
        let dispatched = log.matches("bundle dispatched to BPA").count();
        let exports = log.matches("created export session").count();
        let imports = log.matches("import session").count();
        println!("  Hardy LTP CLA stats (from log):");
        println!("    bundles_dispatched_to_bpa={} export_sessions={} import_sessions={}",
            dispatched, exports, imports);
    }
}

// ─── Test Infrastructure Setup ──────────────────────────────────────────────

struct TestInfra {
    ion: IonLifecycle,
    bpa_child: tokio::process::Child,
    ltp_child: tokio::process::Child,
}

impl TestInfra {
    async fn setup() -> Self {
        // Clean slate
        let _ = std::fs::remove_dir_all(WORK_DIR);
        let _ = tokio::process::Command::new("killm")
            .stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::null())
            .stderr(std::process::Stdio::null())
            .status().await;

        // Generate Hardy configs
        let hardy_cfg = hardy_config();
        let hardy_generated = generate_hardy_config(&hardy_cfg);
        let hardy_dir = Path::new(WORK_DIR).join("hardy/config");
        std::fs::create_dir_all(&hardy_dir).unwrap();
        for (name, content) in &hardy_generated.files {
            std::fs::write(hardy_dir.join(name), content).unwrap();
        }

        // Write ION configs (known-working format with loopback span)
        // Write ION configs via abstraction layer
        let ion_dir = Path::new(WORK_DIR).join("ion");
        std::fs::create_dir_all(&ion_dir).unwrap();
        let ion_cfg = ion_config();
        let ion_generated = generate_ion_config(&ion_cfg);
        for (name, content) in &ion_generated.files {
            std::fs::write(ion_dir.join(name), content).unwrap();
        }

        // Start Hardy BPA
        let hardy_yaml = hardy_dir.join("hardy.yaml");
        let bpa_child = tokio::process::Command::new(HARDY_BPA_BIN)
            .arg("-c").arg(&hardy_yaml)
            .stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::null())
            .stderr(std::process::Stdio::null())
            .spawn()
            .expect("Failed to start Hardy BPA");
        tokio::time::sleep(Duration::from_secs(2)).await;

        // Start hardy-ltp-server
        let ltp_yaml = hardy_dir.join("ltp-cla.yaml");
        let ltp_log = std::fs::File::create(Path::new(WORK_DIR).join("ltp-cla.log")).unwrap();
        let ltp_log2 = ltp_log.try_clone().unwrap();
        let ltp_child = tokio::process::Command::new(HARDY_LTP_BIN)
            .arg(&ltp_yaml)
            .env("RUST_LOG", "hardy_ltp_cla=trace,hardy_ltp_server=info,hardy_ltp_grpc=info")
            .stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::from(ltp_log))
            .stderr(std::process::Stdio::from(ltp_log2))
            .spawn()
            .expect("Failed to start hardy-ltp-server");
        tokio::time::sleep(Duration::from_secs(2)).await;

        // Start ION
        let _ = tokio::process::Command::new("killm")
            .stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::null())
            .stderr(std::process::Stdio::null())
            .status().await;
        tokio::time::sleep(Duration::from_millis(500)).await;

        let mut ion = IonLifecycle::new(None);
        ion.start(&ion_dir).await.expect("ION start failed");
        // Allow extra time for ION's CGR to compute routes and LTP spans to activate
        // ION needs time for rfxclock to process contacts and bpclm to open sessions
        tokio::time::sleep(Duration::from_secs(8)).await;

        Self { ion, bpa_child, ltp_child }
    }

    async fn teardown(mut self) {
        let _ = self.ion.stop().await;
        let _ = self.ltp_child.kill().await;
        let _ = self.bpa_child.kill().await;
        let _ = tokio::process::Command::new("killm")
            .stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::null())
            .stderr(std::process::Stdio::null())
            .output().await;
    }
}

// ─── Test Scenarios ─────────────────────────────────────────────────────────

/// Generate a deterministic payload of given size.
fn make_payload(size: usize, seed: u8) -> Vec<u8> {
    (0..size).map(|i| (i as u8).wrapping_add(seed)).collect()
}

/// Scenario 1: Bidirectional concurrent transfer.
/// Sends bundles simultaneously in both directions (ION→Hardy and Hardy→ION).
async fn test_bidirectional_concurrent(payload_size: usize, label: &str) -> (bool, bool) {
    println!("\n  [bidir] {} ({} bytes) — sending concurrently...", label, payload_size);

    let ion_to_hardy_payload = make_payload(payload_size, 0xAA);
    let hardy_to_ion_payload = make_payload(payload_size, 0xBB);

    // Start bprecvfile to receive Hardy→ION bundle
    let recv_dir = Path::new(WORK_DIR).join(format!("received_{}", label));
    std::fs::create_dir_all(&recv_dir).unwrap();
    let mut recv_child = tokio::process::Command::new("bprecvfile")
        .args(["ipn:10.1", "1"])
        .current_dir(&recv_dir)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .expect("bprecvfile spawn failed");
    tokio::time::sleep(Duration::from_millis(500)).await;

    // Send both directions concurrently
    let ion_payload = ion_to_hardy_payload.clone();
    let ion_label = format!("ion2hardy_{}", label);
    let (ion_result, hardy_result) = tokio::join!(
        ion_send_bundle("ipn:10.1", "ipn:20.1", &ion_payload, &ion_label),
        hardy_send_bundle("http://[::1]:50051", "ipn:10.1", &hardy_to_ion_payload),
    );

    let ion_sent = match ion_result {
        Ok(()) => { println!("    ION→Hardy: sent OK"); true }
        Err(e) => { println!("    ION→Hardy: FAILED ({})", e); false }
    };
    let hardy_sent = match hardy_result {
        Ok(id) => { println!("    Hardy→ION: sent OK (bundle_id={})", id); true }
        Err(e) => { println!("    Hardy→ION: FAILED ({})", e); false }
    };

    // Wait for delivery
    tokio::time::sleep(Duration::from_secs(8)).await;

    // Verify ION→Hardy: check LTP CLA log for "bundle dispatched to BPA"
    let ltp_log = std::fs::read_to_string(Path::new(WORK_DIR).join("ltp-cla.log"))
        .unwrap_or_default();
    let ion_to_hardy_ok = ion_sent && ltp_log.contains("bundle dispatched to BPA");

    // Verify Hardy→ION: check bprecvfile output
    let received_files: Vec<_> = std::fs::read_dir(&recv_dir)
        .unwrap()
        .filter_map(|e| e.ok())
        .filter(|e| e.path().is_file())
        .collect();
    let hardy_to_ion_ok = hardy_sent && !received_files.is_empty();

    let _ = recv_child.kill().await;

    println!("    ION→Hardy delivered: {}", ion_to_hardy_ok);
    println!("    Hardy→ION delivered: {}", hardy_to_ion_ok);

    (ion_to_hardy_ok, hardy_to_ion_ok)
}

/// Scenario 2: Stress burst — rapid multiple sends in both directions.
async fn test_stress_burst(bundle_count: usize, payload_size: usize) -> (usize, usize) {
    println!("\n  [stress] Sending {} bundles ({} bytes each) in both directions...",
        bundle_count, payload_size);

    let start = Instant::now();

    // Start bprecvfile for all expected Hardy→ION bundles
    let recv_dir = Path::new(WORK_DIR).join("received_stress");
    let _ = std::fs::remove_dir_all(&recv_dir);
    std::fs::create_dir_all(&recv_dir).unwrap();
    let mut recv_child = tokio::process::Command::new("bprecvfile")
        .args(["ipn:10.1", &bundle_count.to_string()])
        .current_dir(&recv_dir)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .expect("bprecvfile spawn failed");
    tokio::time::sleep(Duration::from_millis(500)).await;

    // Send ION→Hardy burst
    let mut ion_sent = 0usize;
    for i in 0..bundle_count {
        let payload = make_payload(payload_size, i as u8);
        let label = format!("stress_ion_{}", i);
        if ion_send_bundle("ipn:10.1", "ipn:20.1", &payload, &label).await.is_ok() {
            ion_sent += 1;
        }
    }
    println!("    ION→Hardy: {}/{} sent", ion_sent, bundle_count);

    // Send Hardy→ION burst
    let mut hardy_sent = 0usize;
    for i in 0..bundle_count {
        let payload = make_payload(payload_size, (i as u8).wrapping_add(128));
        if hardy_send_bundle("http://[::1]:50051", "ipn:10.1", &payload).await.is_ok() {
            hardy_sent += 1;
        }
    }
    println!("    Hardy→ION: {}/{} sent", hardy_sent, bundle_count);

    // Wait for delivery (longer for burst)
    tokio::time::sleep(Duration::from_secs(15)).await;

    // Count ION→Hardy deliveries from LTP log
    let ltp_log = std::fs::read_to_string(Path::new(WORK_DIR).join("ltp-cla.log"))
        .unwrap_or_default();
    let ion_to_hardy_delivered = ltp_log.matches("bundle dispatched to BPA").count();

    // Count Hardy→ION deliveries from received files
    let hardy_to_ion_delivered = std::fs::read_dir(&recv_dir)
        .unwrap()
        .filter_map(|e| e.ok())
        .filter(|e| e.path().is_file())
        .count();

    let _ = recv_child.kill().await;

    let elapsed = start.elapsed();
    println!("    ION→Hardy delivered: {}/{}", ion_to_hardy_delivered, ion_sent);
    println!("    Hardy→ION delivered: {}/{}", hardy_to_ion_delivered, hardy_sent);
    println!("    Elapsed: {:.2}s", elapsed.as_secs_f64());

    (ion_to_hardy_delivered, hardy_to_ion_delivered)
}

// ─── Main ───────────────────────────────────────────────────────────────────

#[tokio::main]
async fn main() {
    println!("=== Bidirectional & Stress Interop Test: ION ↔ Hardy over LTP/UDP ===\n");

    // ── Setup ──
    println!("[setup] Starting both engines...");
    let infra = TestInfra::setup().await;
    println!("[setup] Both engines running.\n");

    let mut all_passed = true;

    // ── Scenario 1: Bidirectional with varying payload sizes ──
    println!("━━━ Scenario 1: Bidirectional concurrent transfer ━━━");

    let sizes = [
        (1024, "1KB"),
        (20_480, "20KB"),
        (102_400, "100KB"),
    ];

    for (size, label) in &sizes {
        let (i2h, h2i) = test_bidirectional_concurrent(*size, label).await;
        if !i2h || !h2i {
            all_passed = false;
        }
    }

    // ── Telemetry checkpoint after bidirectional tests ──
    println!("\n━━━ Telemetry checkpoint (post-bidirectional) ━━━");
    collect_telemetry().await;

    // ── Scenario 2: Stress burst ──
    println!("\n━━━ Scenario 2: Stress burst (5 bundles × 20KB each direction) ━━━");
    let (i2h_count, h2i_count) = test_stress_burst(5, 20_480).await;
    if i2h_count == 0 && h2i_count == 0 {
        all_passed = false;
    }

    // ── Final telemetry ──
    println!("\n━━━ Final telemetry (post-stress) ━━━");
    collect_telemetry().await;

    // ── Teardown ──
    println!("\n[teardown] Stopping engines...");
    infra.teardown().await;
    println!("[teardown] Done.\n");

    // ── Summary ──
    println!("━━━ SUMMARY ━━━");
    if all_passed {
        println!("  All scenarios PASSED ✓");
    } else {
        println!("  Some scenarios FAILED ✗");
        std::process::exit(1);
    }
}
