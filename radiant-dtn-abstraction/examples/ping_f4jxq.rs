//! Ping F4JXQ's Hardy node over TCPCLv4 (internet).
//!
//! Starts a local Hardy BPA + TCPCLv4 CLA, connects to F4JXQ's live Hardy
//! node at 44.27.131.233:4556, sends a bundle to ipn:1.7 (echo service),
//! and waits for the echo response.
//!
//! This demonstrates live inter-operator BPv7 DTN over the internet
//! between two amateur radio stations (G4DPZ ↔ F4JXQ).
//!
//! Prerequisites:
//! - hardy-bpa-server built at ../hardy/target/debug/hardy-bpa-server
//! - hardy-tcpclv4-server built at ../hardy/target/debug/hardy-tcpclv4-server
//! - Network connectivity to 44.27.131.233:4556
//!
//! Run: cargo run --example ping_f4jxq --features interop-network

use std::path::Path;
use std::time::Duration;

use tokio::sync::mpsc;
use tokio_stream::wrappers::ReceiverStream;
use tonic::Streaming;

use hardy_ltp_proto::proto::service::*;
use hardy_ltp_proto::proto::service::application_client::ApplicationClient;

const HARDY_BPA_BIN: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/../hardy/target/debug/hardy-bpa-server"
);
const HARDY_TCPCL_BIN: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/../hardy/target/debug/hardy-tcpclv4-server"
);
const WORK_DIR: &str = "/tmp/radiant-ping-f4jxq";

// F4JXQ's node
const F4JXQ_HOST: &str = "44.27.131.233";
const F4JXQ_PORT: u16 = 4556;
const F4JXQ_ECHO_EID: &str = "ipn:1.7";

#[tokio::main]
async fn main() {
    println!("=== RADIANT DTN Ping: G4DPZ -> F4JXQ ===");
    println!("  Target: {} ({}:{})", F4JXQ_ECHO_EID, F4JXQ_HOST, F4JXQ_PORT);
    println!();

    // Clean up
    let _ = std::fs::remove_dir_all(WORK_DIR);
    std::fs::create_dir_all(WORK_DIR).unwrap();

    // Write configs
    println!("[1/4] Writing configs...");
    let bpa_config_path = Path::new(WORK_DIR).join("hardy.yaml");
    std::fs::write(&bpa_config_path, "\
admin-endpoints:\n  - \"dtn://g4dpz/\"\n  - \"ipn:10.0\"\n\n\
node-ids:\n  - \"dtn://g4dpz/\"\n  - \"ipn:10.0\"\n\n\
grpc:\n  address: \"[::1]:50051\"\n  services: [\"application\", \"cla\", \"service\"]\n\n\
built-in-services:\n  echo: [7, \"echo\"]\n").unwrap();

    let tcpcl_config_path = Path::new(WORK_DIR).join("tcpclv4.yaml");
    std::fs::write(&tcpcl_config_path, format!("\
bpa-address: \"http://[::1]:50051\"\n\
cla-name: \"tcpclv4-to-f4jxq\"\n\
address: \"0.0.0.0:14556\"\n\
segment-mru: 16384\n\
transfer-mru: 562949953421312\n\
peers:\n  - \"{}:{}\"\n", F4JXQ_HOST, F4JXQ_PORT)).unwrap();

    // Start Hardy BPA
    println!("[2/4] Starting Hardy BPA...");
    let bpa_log = std::fs::File::create(Path::new(WORK_DIR).join("bpa.log")).unwrap();
    let bpa_log2 = bpa_log.try_clone().unwrap();
    let mut bpa_child = tokio::process::Command::new(HARDY_BPA_BIN)
        .arg("-c").arg(&bpa_config_path)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::from(bpa_log))
        .stderr(std::process::Stdio::from(bpa_log2))
        .spawn()
        .expect("Failed to start Hardy BPA");
    println!("  PID: {}", bpa_child.id().unwrap());
    tokio::time::sleep(Duration::from_secs(5)).await;

    // Verify BPA started
    let bpa_log_content = std::fs::read_to_string(Path::new(WORK_DIR).join("bpa.log"))
        .unwrap_or_default();
    if !bpa_log_content.contains("Started successfully") {
        println!("  ERROR: BPA failed to start:");
        for line in bpa_log_content.lines() { println!("    {}", line); }
        let _ = bpa_child.kill().await;
        return;
    }
    println!("  BPA started OK");

    // Start TCPCLv4 CLA
    println!("[3/4] Starting TCPCLv4 CLA...");
    let tcpcl_log = std::fs::File::create(Path::new(WORK_DIR).join("tcpclv4.log")).unwrap();
    let tcpcl_log2 = tcpcl_log.try_clone().unwrap();
    let mut tcpcl_child = tokio::process::Command::new(HARDY_TCPCL_BIN)
        .arg("-c").arg(&tcpcl_config_path)
        .env("RUST_LOG", "info")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::from(tcpcl_log))
        .stderr(std::process::Stdio::from(tcpcl_log2))
        .spawn()
        .expect("Failed to start TCPCLv4 CLA");
    println!("  PID: {}", tcpcl_child.id().unwrap());
    tokio::time::sleep(Duration::from_secs(8)).await;

    // Check connection
    let tcpcl_log_content = std::fs::read_to_string(Path::new(WORK_DIR).join("tcpclv4.log"))
        .unwrap_or_default();
    if tcpcl_log_content.contains("Connected to configured peer") {
        println!("  Connected to F4JXQ!");
    } else {
        println!("  WARNING: No peer connection confirmed");
        for line in tcpcl_log_content.lines().take(10) { println!("    {}", line); }
    }

    // Send ping
    println!("[4/4] Sending echo ping...");
    let ping_payload: Vec<u8> = b"RADIANT G4DPZ ping - 73 de Dave".to_vec();
    println!("  Payload: {} bytes", ping_payload.len());

    let result = send_ping("http://[::1]:50051", F4JXQ_ECHO_EID, &ping_payload).await;
    match result {
        Ok(Some(response)) => {
            let matches = response == ping_payload;
            println!();
            println!("=== ECHO RESPONSE RECEIVED ===");
            println!("  From: {}", F4JXQ_ECHO_EID);
            println!("  Sent: {} bytes", ping_payload.len());
            println!("  Received: {} bytes", response.len());
            println!("  Data integrity: {}", if matches { "MATCH ✓" } else { "MISMATCH ✗" });
        }
        Ok(None) => {
            println!();
            println!("=== NO RESPONSE (timeout) ===");
            println!("  Bundle sent OK but no echo within 15s.");
        }
        Err(e) => {
            println!();
            println!("=== ERROR: {} ===", e);
        }
    }

    // Teardown
    println!();
    println!("Tearing down...");
    let _ = tcpcl_child.kill().await;
    let _ = bpa_child.kill().await;
    println!("Done. 73!");
}

/// Send a bundle to a destination and wait for a response.
///
/// Registers as ipn:10.1 (service 1 = delivery) to receive echo responses.
/// Note: Do NOT register as service 7 — the BPA's built-in echo handler
/// already owns that service ID.
async fn send_ping(
    grpc_endpoint: &str,
    destination: &str,
    payload: &[u8],
) -> Result<Option<Vec<u8>>, String> {
    let mut client = ApplicationClient::connect(grpc_endpoint.to_string())
        .await
        .map_err(|e| format!("gRPC connect failed: {}", e))?;

    let (tx, rx) = mpsc::channel::<AppToBpa>(16);
    let response = client
        .register(ReceiverStream::new(rx))
        .await
        .map_err(|e| format!("Stream open failed: {}", e))?;
    let mut bpa_stream: Streaming<BpaToApp> = response.into_inner();

    // Register as service 1 (bundle delivery) to receive responses
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
        .ok_or("BPA closed stream (registration rejected)")?;

    match reg_msg.msg {
        Some(bpa_to_app::Msg::Register(resp)) => {
            println!("  Registered as: {}", resp.endpoint_id);
        }
        Some(bpa_to_app::Msg::Status(s)) => {
            return Err(format!("Registration failed: {:?}", s));
        }
        other => return Err(format!("Unexpected: {:?}", other)),
    }

    // Send the ping bundle
    tx.send(AppToBpa {
        msg_id: 2,
        msg: Some(app_to_bpa::Msg::Send(AppSendRequest {
            destination: destination.to_string(),
            payload: payload.to_vec(),
            lifetime: 300000, // 5 minutes
            options: None,
        })),
    })
    .await
    .map_err(|e| format!("Send bundle: {}", e))?;

    let send_msg = bpa_stream
        .message()
        .await
        .map_err(|e| format!("Stream error: {}", e))?
        .ok_or("Stream closed after send")?;

    match send_msg.msg {
        Some(bpa_to_app::Msg::Send(resp)) => {
            println!("  Bundle sent: {}", resp.bundle_id);
        }
        Some(bpa_to_app::Msg::Status(s)) => {
            return Err(format!("Send failed: {:?}", s));
        }
        _ => {}
    }

    // Wait for echo response
    println!("  Waiting for echo response (15s)...");
    let timeout_result = tokio::time::timeout(Duration::from_secs(15), async {
        loop {
            match bpa_stream.message().await {
                Ok(Some(msg)) => {
                    if let Some(bpa_to_app::Msg::Receive(recv)) = msg.msg {
                        return Ok(Some(recv.payload));
                    }
                }
                Ok(None) => return Ok(None),
                Err(e) => return Err(format!("Stream error: {}", e)),
            }
        }
    })
    .await;

    match timeout_result {
        Ok(result) => result,
        Err(_) => Ok(None), // Timeout
    }
}
