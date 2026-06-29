//! ION-DTN backend adapter implementation.
//!
//! Provides configuration generation, lifecycle management, hot reconfiguration,
//! and telemetry collection for NASA JPL's ION-DTN (Interplanetary Overlay Network)
//! implementation.

pub mod config_gen;
pub mod hot_reconfig;
pub mod lifecycle;
pub mod telemetry;

use std::path::{Path, PathBuf};

use async_trait::async_trait;
use tokio::sync::Mutex;

use crate::adapter::capability::CapabilitySet;
use crate::adapter::traits::{
    BackendAdapter, BundleStatistics, ContactRef, GeneratedConfig, HealthStatus, LinkState, NodeRef,
};
use crate::error::AbstractionError;
use crate::model::{Contact, Neighbor, NetworkConfiguration};

/// Concrete ION-DTN backend adapter implementing [`BackendAdapter`].
///
/// Composes the ION sub-components (lifecycle, telemetry, hot-reconfig)
/// into a single unified adapter interface. Uses interior mutability for
/// the lifecycle component since `IonLifecycle::start` requires `&mut self`.
pub struct IonAdapter {
    lifecycle: Mutex<lifecycle::IonLifecycle>,
    telemetry: telemetry::IonTelemetry,
    hot_reconfig: hot_reconfig::IonHotReconfig,
    capabilities: CapabilitySet,
}

impl IonAdapter {
    /// Create a new `IonAdapter`.
    ///
    /// # Arguments
    /// * `ion_bin_dir` — Optional path to the directory containing ION binaries.
    ///   If None, the system $PATH is used to locate ION tools.
    pub fn new(ion_bin_dir: Option<PathBuf>) -> Self {
        Self {
            lifecycle: Mutex::new(lifecycle::IonLifecycle::new(ion_bin_dir.clone())),
            telemetry: telemetry::IonTelemetry::new(ion_bin_dir.clone()),
            hot_reconfig: hot_reconfig::IonHotReconfig::new(ion_bin_dir),
            capabilities: hot_reconfig::ion_capabilities(),
        }
    }
}

#[async_trait]
impl BackendAdapter for IonAdapter {
    fn name(&self) -> &str {
        "ion-dtn"
    }

    fn capabilities(&self) -> &CapabilitySet {
        &self.capabilities
    }

    async fn validate(&self, _config: &NetworkConfiguration) -> Result<(), AbstractionError> {
        // ION accepts all canonical configurations currently representable
        // in the model. Future: validate LTP-specific constraints.
        Ok(())
    }

    async fn generate_config(
        &self,
        config: &NetworkConfiguration,
        _output_dir: &Path,
    ) -> Result<GeneratedConfig, AbstractionError> {
        Ok(config_gen::generate_ion_config(config))
    }

    async fn deploy(
        &self,
        config: &NetworkConfiguration,
        output_dir: &Path,
    ) -> Result<(), AbstractionError> {
        let generated = config_gen::generate_ion_config(config);

        std::fs::create_dir_all(output_dir).map_err(|e| {
            AbstractionError::new(
                crate::error::ErrorCategory::LifecycleError,
                format!("Failed to create output directory '{}': {}", output_dir.display(), e),
                "deploy",
            )
            .with_backend("ion-dtn")
        })?;

        for (filename, content) in &generated.files {
            let file_path = output_dir.join(filename);
            std::fs::write(&file_path, content).map_err(|e| {
                AbstractionError::new(
                    crate::error::ErrorCategory::LifecycleError,
                    format!("Failed to write '{}': {}", file_path.display(), e),
                    "deploy",
                )
                .with_backend("ion-dtn")
            })?;
        }

        Ok(())
    }

    // ─── Lifecycle operations ───────────────────────────────────────────

    async fn start(&self, config_dir: &Path) -> Result<(), AbstractionError> {
        let mut lc = self.lifecycle.lock().await;
        lc.start(config_dir).await
    }

    async fn stop(&self) -> Result<(), AbstractionError> {
        let lc = self.lifecycle.lock().await;
        lc.stop().await
    }

    async fn restart(&self, config_dir: &Path) -> Result<(), AbstractionError> {
        let mut lc = self.lifecycle.lock().await;
        lc.restart(config_dir).await
    }

    async fn health(&self) -> Result<HealthStatus, AbstractionError> {
        let lc = self.lifecycle.lock().await;
        lc.health().await
    }

    async fn version(&self) -> Result<String, AbstractionError> {
        let lc = self.lifecycle.lock().await;
        lc.version().await
    }

    // ─── Hot reconfiguration operations ─────────────────────────────────

    async fn add_contact(&self, contact: &Contact) -> Result<(), AbstractionError> {
        self.hot_reconfig.add_contact(contact).await
    }

    async fn remove_contact(&self, contact: &ContactRef) -> Result<(), AbstractionError> {
        self.hot_reconfig.remove_contact(contact).await
    }

    async fn add_neighbor(&self, neighbor: &Neighbor) -> Result<(), AbstractionError> {
        self.hot_reconfig.add_neighbor(neighbor).await
    }

    async fn remove_neighbor(&self, node_ref: &NodeRef) -> Result<(), AbstractionError> {
        self.hot_reconfig.remove_neighbor(node_ref).await
    }

    async fn enable_link(&self, link_id: &str) -> Result<(), AbstractionError> {
        self.hot_reconfig.enable_link(link_id).await
    }

    async fn disable_link(&self, link_id: &str) -> Result<(), AbstractionError> {
        self.hot_reconfig.disable_link(link_id).await
    }

    // ─── Telemetry ──────────────────────────────────────────────────────

    async fn collect_stats(&self) -> Result<BundleStatistics, AbstractionError> {
        self.telemetry.collect_stats().await
    }

    async fn link_states(&self) -> Result<Vec<LinkState>, AbstractionError> {
        self.telemetry.link_states().await
    }
}
