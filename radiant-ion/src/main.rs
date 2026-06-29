//! radiant-ion: ION-DTN configuration management and lifecycle control.
//!
//! A standalone binary that wraps the ION-DTN adapter from `radiant-dtn-abstraction`
//! into a usable CLI/daemon. Reads canonical YAML configuration, generates ION admin
//! scripts, manages ION's lifecycle, collects telemetry, and exposes an HTTP/JSON API
//! for remote management.
//!
//! ## Amateur Radio Compliance
//!
//! No encryption is configured by default. All bundle payloads carried over amateur
//! radio links remain unencrypted per ITU Radio Regulations (Article 25) and national
//! amateur radio regulations.

use std::net::SocketAddr;
use std::path::{Path, PathBuf};
use std::sync::Arc;

use clap::{Parser, Subcommand};
use tokio::net::TcpListener;
use tracing::{error, info};

use radiant_dtn_abstraction::adapter::ion::config_gen::generate_ion_config;
use radiant_dtn_abstraction::adapter::ion::lifecycle::IonLifecycle;
use radiant_dtn_abstraction::adapter::ion::telemetry::IonTelemetry;
use radiant_dtn_abstraction::adapter::registry::AdapterRegistry;
use radiant_dtn_abstraction::api::routes::build_router;
use radiant_dtn_abstraction::api::AppState;
use radiant_dtn_abstraction::events::bus::EventBus;
use radiant_dtn_abstraction::model::NetworkConfiguration;

#[derive(Parser)]
#[command(name = "radiant-ion", about = "ION-DTN configuration and lifecycle manager")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Generate ION admin scripts from a canonical YAML config
    Generate {
        /// Path to the canonical YAML configuration file
        config: String,
        /// Output directory for generated files (default: current dir)
        #[arg(short, long, default_value = ".")]
        output_dir: String,
    },
    /// Generate configs and start ION
    Start {
        /// Path to the canonical YAML configuration file
        config: String,
    },
    /// Stop ION (ionstop + killm)
    Stop,
    /// Show ION health and bundle statistics
    Status,
    /// Start ION and run the HTTP/JSON management API
    Serve {
        /// Path to the canonical YAML configuration file
        config: String,
        /// HTTP API port (default: 3000)
        #[arg(short, long, default_value = "3000")]
        port: u16,
    },
}

#[tokio::main]
async fn main() {
    // Initialize tracing with env filter (controlled via RUST_LOG)
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info")),
        )
        .init();

    let cli = Cli::parse();

    let result = match cli.command {
        Commands::Generate { config, output_dir } => cmd_generate(&config, &output_dir).await,
        Commands::Start { config } => cmd_start(&config).await,
        Commands::Stop => cmd_stop().await,
        Commands::Status => cmd_status().await,
        Commands::Serve { config, port } => cmd_serve(&config, port).await,
    };

    if let Err(e) = result {
        error!("{}", e);
        std::process::exit(1);
    }
}

/// Load and deserialize the canonical YAML configuration file.
fn load_config(path: &str) -> Result<NetworkConfiguration, String> {
    let file = std::fs::File::open(path)
        .map_err(|e| format!("Failed to open config file '{}': {}", path, e))?;
    let config: NetworkConfiguration = serde_yaml::from_reader(file)
        .map_err(|e| format!("Failed to parse YAML config '{}': {}", path, e))?;
    Ok(config)
}

/// Generate ION admin scripts from YAML config and write to output directory.
async fn cmd_generate(config_path: &str, output_dir: &str) -> Result<(), String> {
    let config = load_config(config_path)?;

    info!(
        "Generating ION config for node {} ({})",
        config.local_node.node_number, config.local_node.name
    );

    let generated = generate_ion_config(&config);

    // Ensure output directory exists
    std::fs::create_dir_all(output_dir)
        .map_err(|e| format!("Failed to create output directory '{}': {}", output_dir, e))?;

    let output_path = Path::new(output_dir);
    for (filename, content) in &generated.files {
        let file_path = output_path.join(filename);
        std::fs::write(&file_path, content)
            .map_err(|e| format!("Failed to write '{}': {}", file_path.display(), e))?;
        info!("  Generated: {}", file_path.display());
    }

    println!(
        "Generated {} ION config files in '{}':",
        generated.files.len(),
        output_dir
    );
    for filename in generated.files.keys() {
        println!("  {}", filename);
    }

    Ok(())
}

/// Generate configs to a temp directory and start ION.
async fn cmd_start(config_path: &str) -> Result<(), String> {
    let config = load_config(config_path)?;

    info!(
        "Starting ION for node {} ({})",
        config.local_node.node_number, config.local_node.name
    );

    // Generate configs to a temp directory
    let config_dir = generate_to_temp(&config)?;

    // Start ION
    let mut lifecycle = IonLifecycle::new(None);
    lifecycle
        .start(&config_dir)
        .await
        .map_err(|e| format!("Failed to start ION: {}", e))?;

    // Check health
    let health = lifecycle
        .health()
        .await
        .map_err(|e| format!("Failed to check ION health: {}", e))?;

    println!("ION started successfully.");
    println!(
        "  Running: {}",
        if health.running { "yes" } else { "no" }
    );
    if let Some(msg) = &health.message {
        println!("  Status: {}", msg);
    }

    Ok(())
}

/// Stop ION daemons.
async fn cmd_stop() -> Result<(), String> {
    info!("Stopping ION...");

    let lifecycle = IonLifecycle::new(None);
    lifecycle
        .stop()
        .await
        .map_err(|e| format!("Failed to stop ION: {}", e))?;

    println!("ION stopped.");
    Ok(())
}

/// Show ION health and bundle statistics.
async fn cmd_status() -> Result<(), String> {
    let lifecycle = IonLifecycle::new(None);
    let telemetry = IonTelemetry::new(None);

    // Health check
    let health = lifecycle
        .health()
        .await
        .map_err(|e| format!("Failed to check ION health: {}", e))?;

    println!("=== ION Health ===");
    println!(
        "  Running: {}",
        if health.running { "yes" } else { "no" }
    );
    if let Some(msg) = &health.message {
        println!("  Message: {}", msg);
    }
    if let Some(uptime) = health.uptime_secs {
        println!("  Uptime:  {}s", uptime);
    }

    // Bundle statistics
    let stats = telemetry
        .collect_stats()
        .await
        .map_err(|e| format!("Failed to collect stats: {}", e))?;

    println!("\n=== Bundle Statistics ===");
    println!("  Sourced:    {}", stats.bundles_sourced);
    println!("  Forwarded:  {}", stats.bundles_forwarded);
    println!("  Delivered:  {}", stats.bundles_delivered);
    println!("  Expired:    {}", stats.bundles_expired);
    println!("  Queued:     {}", stats.bundles_queued);

    // Link states
    let links = telemetry
        .link_states()
        .await
        .map_err(|e| format!("Failed to get link states: {}", e))?;

    println!("\n=== Link States ===");
    if links.is_empty() {
        println!("  No active links reported.");
    } else {
        for link in &links {
            println!(
                "  {} (node {}): {} | TX: {} bytes | RX: {} bytes",
                link.link_id,
                link.neighbor_node,
                if link.active { "ACTIVE" } else { "DOWN" },
                link.bytes_sent,
                link.bytes_received
            );
        }
    }

    Ok(())
}

/// Start ION and run the HTTP/JSON management API daemon.
async fn cmd_serve(config_path: &str, port: u16) -> Result<(), String> {
    let config = load_config(config_path)?;

    info!(
        "Starting ION for node {} ({}) with HTTP API on port {}",
        config.local_node.node_number, config.local_node.name, port
    );

    // Generate configs to a temp directory and start ION
    let config_dir = generate_to_temp(&config)?;

    let mut lifecycle = IonLifecycle::new(None);
    lifecycle
        .start(&config_dir)
        .await
        .map_err(|e| format!("Failed to start ION: {}", e))?;

    info!("ION started. Launching HTTP API...");

    // Build application state for the API
    let registry = Arc::new(AdapterRegistry::new());
    let event_bus = Arc::new(EventBus::new(256));
    let state = AppState::new(registry, event_bus);

    // Store the loaded config into state
    {
        let mut cfg_lock = state.config.write().await;
        *cfg_lock = Some(config);
    }

    // Build the router
    let router = build_router(state);

    // Bind and serve
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    let listener = TcpListener::bind(addr)
        .await
        .map_err(|e| format!("Failed to bind to {}: {}", addr, e))?;

    info!("HTTP API listening on http://{}", addr);
    println!("radiant-ion serving on http://{}", addr);
    println!("Press Ctrl+C to stop.");

    // Serve with graceful shutdown on SIGINT/SIGTERM
    axum::serve(listener, router)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .map_err(|e| format!("Server error: {}", e))?;

    // Graceful shutdown: stop ION
    info!("Shutting down ION...");
    let lifecycle = IonLifecycle::new(None);
    if let Err(e) = lifecycle.stop().await {
        error!("Warning: failed to stop ION during shutdown: {}", e);
    }

    println!("radiant-ion stopped.");
    Ok(())
}

/// Generate ION config files to a temporary directory.
fn generate_to_temp(config: &NetworkConfiguration) -> Result<PathBuf, String> {
    let temp_dir = std::env::temp_dir().join(format!(
        "radiant-ion-node{}",
        config.local_node.node_number
    ));
    std::fs::create_dir_all(&temp_dir)
        .map_err(|e| format!("Failed to create temp config dir: {}", e))?;

    let generated = generate_ion_config(config);
    for (filename, content) in &generated.files {
        let file_path = temp_dir.join(filename);
        std::fs::write(&file_path, content)
            .map_err(|e| format!("Failed to write temp config '{}': {}", file_path.display(), e))?;
    }

    info!("Generated ION configs in {}", temp_dir.display());
    Ok(temp_dir)
}

/// Wait for SIGINT (Ctrl+C) or SIGTERM for graceful shutdown.
async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c()
            .await
            .expect("Failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
            .expect("Failed to install SIGTERM handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => info!("Received SIGINT, shutting down..."),
        _ = terminate => info!("Received SIGTERM, shutting down..."),
    }
}
