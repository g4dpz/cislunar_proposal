//! ION-DTN hot reconfiguration support.
//!
//! Implements runtime mutation of the contact plan, neighbor table, and
//! convergence layer link state by piping commands to ION admin tools
//! (`ionadmin`, `bpadmin`, `ltpadmin`, `ipnadmin`) via stdin.

use std::path::PathBuf;
use std::process::Stdio;

use tokio::io::AsyncWriteExt;
use tokio::process::Command;

use crate::adapter::capability::{
    CapabilitySet, HotReconfigCapabilities, SecurityCapabilities,
};
use crate::error::{AbstractionError, ErrorCategory};
use crate::model::contact_plan::Contact;
use crate::model::convergence::ConvergenceLayerType;
use crate::model::neighbor::{ConvergenceLayerLink, Neighbor};
use crate::model::routing::RoutingStrategy;

use super::super::traits::{ContactRef, NodeRef};

/// Manages hot reconfiguration of a running ION-DTN instance.
///
/// Commands are piped to the appropriate admin tool's stdin. ION admin tools
/// accept interactive commands when run without a script file argument.
pub struct IonHotReconfig {
    /// Path to the ION bin directory (where ionadmin, bpadmin, etc. live).
    /// If None, assumes the tools are on $PATH.
    ion_bin_dir: Option<PathBuf>,
}

impl IonHotReconfig {
    /// Create a new `IonHotReconfig` manager.
    ///
    /// # Arguments
    /// * `ion_bin_dir` — Optional path to the directory containing ION binaries.
    ///   If None, the system $PATH is used to locate ionadmin, bpadmin, etc.
    pub fn new(ion_bin_dir: Option<PathBuf>) -> Self {
        Self { ion_bin_dir }
    }

    /// Resolve the full path to an ION binary.
    fn bin_path(&self, name: &str) -> PathBuf {
        match &self.ion_bin_dir {
            Some(dir) => dir.join(name),
            None => PathBuf::from(name),
        }
    }

    /// Add a contact to the running ION contact plan via `ionadmin`.
    ///
    /// Pipes: `a contact +<start_time> +<end_time> <source> <dest> <rate>`
    pub async fn add_contact(&self, contact: &Contact) -> Result<(), AbstractionError> {
        let cmd = format!(
            "a contact +{} +{} {} {} {}\nq\n",
            contact.start_time, contact.end_time, contact.source_node, contact.dest_node, contact.rate_bps
        );
        self.pipe_to_admin("ionadmin", &cmd, "add_contact").await
    }

    /// Remove a contact from the running ION contact plan via `ionadmin`.
    ///
    /// Pipes: `d contact +<start_time> <source> <dest>`
    pub async fn remove_contact(&self, contact_ref: &ContactRef) -> Result<(), AbstractionError> {
        let cmd = format!(
            "d contact +{} {} {}\nq\n",
            contact_ref.start_time, contact_ref.source_node, contact_ref.dest_node
        );
        self.pipe_to_admin("ionadmin", &cmd, "remove_contact").await
    }

    /// Add a neighbor to the running ION instance.
    ///
    /// This involves multiple admin tools:
    /// - `bpadmin`: add outduct for each LTP link
    /// - `ipnadmin`: add plan entry routing to the neighbor
    pub async fn add_neighbor(&self, neighbor: &Neighbor) -> Result<(), AbstractionError> {
        // For each LTP link, add an outduct via bpadmin
        for link in &neighbor.links {
            if let Some(engine_id) = Self::extract_engine_id(link) {
                let bpadmin_cmd = format!("a outduct ltp {} ltpclo\nq\n", engine_id);
                self.pipe_to_admin("bpadmin", &bpadmin_cmd, "add_neighbor")
                    .await?;
            }
        }

        // Add a plan entry via ipnadmin for the neighbor node
        // Use the first LTP link's engine_id for the plan
        if let Some(engine_id) = neighbor.links.iter().find_map(Self::extract_engine_id) {
            let ipnadmin_cmd = format!(
                "a plan {} ltp/{}\nq\n",
                neighbor.node_number, engine_id
            );
            self.pipe_to_admin("ipnadmin", &ipnadmin_cmd, "add_neighbor")
                .await?;
        }

        Ok(())
    }

    /// Remove a neighbor from the running ION instance.
    ///
    /// This involves multiple admin tools:
    /// - `bpadmin`: delete outduct for the neighbor's LTP engine_id
    /// - `ipnadmin`: delete plan entry for the neighbor node
    pub async fn remove_neighbor(&self, node_ref: &NodeRef) -> Result<(), AbstractionError> {
        // Delete the plan entry via ipnadmin
        let ipnadmin_cmd = format!("d plan {}\nq\n", node_ref.node_number);
        self.pipe_to_admin("ipnadmin", &ipnadmin_cmd, "remove_neighbor")
            .await?;

        // Delete outduct via bpadmin — use node_number as engine_id convention
        let bpadmin_cmd = format!("d outduct ltp {}\nq\n", node_ref.node_number);
        self.pipe_to_admin("bpadmin", &bpadmin_cmd, "remove_neighbor")
            .await?;

        Ok(())
    }

    /// Enable a convergence layer link via `bpadmin`.
    ///
    /// Pipes: `r start ltp/<engine_id>` to resume a stopped duct.
    /// The `link_id` is expected to be the LTP engine_id as a string.
    pub async fn enable_link(&self, link_id: &str) -> Result<(), AbstractionError> {
        let cmd = format!("r start ltp/{}\nq\n", link_id);
        self.pipe_to_admin("bpadmin", &cmd, "enable_link").await
    }

    /// Disable a convergence layer link via `bpadmin`.
    ///
    /// Pipes: `r stop ltp/<engine_id>` to stop a running duct.
    /// The `link_id` is expected to be the LTP engine_id as a string.
    pub async fn disable_link(&self, link_id: &str) -> Result<(), AbstractionError> {
        let cmd = format!("r stop ltp/{}\nq\n", link_id);
        self.pipe_to_admin("bpadmin", &cmd, "disable_link").await
    }

    /// Pipe a command string to an ION admin tool's stdin.
    ///
    /// Spawns the tool with stdin piped, writes the command, closes stdin,
    /// and waits for the process to exit. Returns an error if the tool
    /// exits with a non-zero status.
    async fn pipe_to_admin(
        &self,
        tool_name: &str,
        command: &str,
        operation: &str,
    ) -> Result<(), AbstractionError> {
        let tool_path = self.bin_path(tool_name);

        let mut child = Command::new(&tool_path)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::RuntimeError,
                    format!("Failed to spawn {}: {}", tool_name, e),
                    operation,
                )
                .with_backend("ion-dtn")
            })?;

        // Write command to stdin
        if let Some(mut stdin) = child.stdin.take() {
            stdin.write_all(command.as_bytes()).await.map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::RuntimeError,
                    format!("Failed to write to {} stdin: {}", tool_name, e),
                    operation,
                )
                .with_backend("ion-dtn")
            })?;
            // stdin is dropped here, closing the pipe
        }

        let output = child.wait_with_output().await.map_err(|e| {
            AbstractionError::new(
                ErrorCategory::RuntimeError,
                format!("Failed to wait for {}: {}", tool_name, e),
                operation,
            )
            .with_backend("ion-dtn")
        })?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            return Err(AbstractionError::new(
                ErrorCategory::RuntimeError,
                format!(
                    "{} command failed (exit code {:?}): {}{}",
                    tool_name,
                    output.status.code(),
                    stderr.trim(),
                    if !stdout.trim().is_empty() {
                        format!(" | stdout: {}", stdout.trim())
                    } else {
                        String::new()
                    }
                ),
                operation,
            )
            .with_backend("ion-dtn")
            .with_backend_code(format!(
                "exit_code:{}",
                output.status.code().unwrap_or(-1)
            )));
        }

        Ok(())
    }

    /// Extract the remote LTP engine_id from a convergence layer link.
    fn extract_engine_id(link: &ConvergenceLayerLink) -> Option<u64> {
        match link {
            ConvergenceLayerLink::LtpUdp {
                remote_engine_id, ..
            } => Some(*remote_engine_id),
            ConvergenceLayerLink::Kiss {
                remote_engine_id, ..
            } => Some(*remote_engine_id),
            _ => None,
        }
    }
}

/// Returns the CapabilitySet for the ION-DTN backend.
///
/// ION supports all hot reconfiguration operations, all convergence layer
/// types relevant to amateur radio, CGR/Static/Default routing, and no
/// BPSec (per amateur radio regulations prohibiting encryption).
pub fn ion_capabilities() -> CapabilitySet {
    CapabilitySet {
        hot_reconfig: HotReconfigCapabilities::all(),
        convergence_layers: vec![
            ConvergenceLayerType::LtpUdp,
            ConvergenceLayerType::Kiss,
            ConvergenceLayerType::TcpCl,
            ConvergenceLayerType::Udp,
        ],
        routing_strategies: vec![
            RoutingStrategy::Cgr,
            RoutingStrategy::Static,
            RoutingStrategy::Default,
        ],
        security: SecurityCapabilities::none(), // No BPSec for amateur radio
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ion_capabilities_hot_reconfig_all() {
        let caps = ion_capabilities();
        assert!(caps.hot_reconfig.add_contact);
        assert!(caps.hot_reconfig.remove_contact);
        assert!(caps.hot_reconfig.add_neighbor);
        assert!(caps.hot_reconfig.remove_neighbor);
        assert!(caps.hot_reconfig.enable_link);
        assert!(caps.hot_reconfig.disable_link);
    }

    #[test]
    fn test_ion_capabilities_convergence_layers() {
        let caps = ion_capabilities();
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::LtpUdp));
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::Kiss));
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::TcpCl));
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::Udp));
    }

    #[test]
    fn test_ion_capabilities_routing() {
        let caps = ion_capabilities();
        assert!(caps.routing_strategies.contains(&RoutingStrategy::Cgr));
        assert!(caps.routing_strategies.contains(&RoutingStrategy::Static));
        assert!(caps.routing_strategies.contains(&RoutingStrategy::Default));
    }

    #[test]
    fn test_ion_capabilities_no_security() {
        let caps = ion_capabilities();
        assert!(!caps.security.bpsec_bib);
        assert!(!caps.security.bpsec_bcb);
    }

    #[test]
    fn test_extract_engine_id_ltp_udp() {
        let link = ConvergenceLayerLink::LtpUdp {
            id: "ltp-to-orbiter".to_string(),
            local_engine_id: 10,
            remote_engine_id: 20,
            remote_host: "192.168.1.20".to_string(),
            remote_port: 1113,
            local_port: 2113,
            mtu: None,
            segment_rate: None,
        };
        assert_eq!(IonHotReconfig::extract_engine_id(&link), Some(20));
    }

    #[test]
    fn test_extract_engine_id_kiss() {
        let link = ConvergenceLayerLink::Kiss {
            id: "kiss-tnc".to_string(),
            tnc_device: "/dev/ttyUSB0".to_string(),
            baud_rate: 9600,
            local_engine_id: 10,
            remote_engine_id: 20,
            frame_size: None,
        };
        assert_eq!(IonHotReconfig::extract_engine_id(&link), Some(20));
    }

    #[test]
    fn test_extract_engine_id_tcpcl_returns_none() {
        let link = ConvergenceLayerLink::TcpCl {
            id: "tcp-link".to_string(),
            remote_host: "192.168.1.30".to_string(),
            remote_port: 4556,
            local_port: None,
            keepalive_interval_secs: None,
        };
        assert_eq!(IonHotReconfig::extract_engine_id(&link), None);
    }

    #[test]
    fn test_extract_engine_id_udp_returns_none() {
        let link = ConvergenceLayerLink::Udp {
            id: "udp-link".to_string(),
            remote_host: "192.168.1.30".to_string(),
            remote_port: 5000,
            local_port: None,
        };
        assert_eq!(IonHotReconfig::extract_engine_id(&link), None);
    }

    #[test]
    fn test_ion_hot_reconfig_new_with_bin_dir() {
        let reconfig = IonHotReconfig::new(Some(PathBuf::from("/opt/ion/bin")));
        assert_eq!(reconfig.bin_path("ionadmin"), PathBuf::from("/opt/ion/bin/ionadmin"));
        assert_eq!(reconfig.bin_path("bpadmin"), PathBuf::from("/opt/ion/bin/bpadmin"));
    }

    #[test]
    fn test_ion_hot_reconfig_new_without_bin_dir() {
        let reconfig = IonHotReconfig::new(None);
        assert_eq!(reconfig.bin_path("ionadmin"), PathBuf::from("ionadmin"));
    }
}
