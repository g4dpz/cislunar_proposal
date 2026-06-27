//! Hardy hot reconfiguration support.
//!
//! Implements runtime mutation of the contact plan and link state via
//! Hardy's management REST API. Hardy supports:
//! - Adding/removing contacts at runtime (POST/DELETE /api/contacts)
//! - Enabling/disabling links at runtime (via REST API)
//!
//! Hardy does NOT support:
//! - Adding/removing neighbors at runtime (requires restart)
//!
//! Operations that Hardy does not support return `UnsupportedOperation`.

use crate::adapter::capability::{
    CapabilitySet, HotReconfigCapabilities, SecurityCapabilities,
};
use crate::error::{AbstractionError, ErrorCategory};
use crate::model::contact_plan::Contact;
use crate::model::convergence::ConvergenceLayerType;
use crate::model::neighbor::Neighbor;
use crate::model::routing::RoutingStrategy;

use super::super::traits::{ContactRef, NodeRef};

/// Manages hot reconfiguration of a running Hardy instance.
///
/// Hardy exposes a management REST API for runtime contact plan mutations
/// and link state changes. Neighbor operations are not supported at runtime
/// and require a full daemon restart.
pub struct HardyHotReconfig {
    /// Base URL for Hardy's management REST API.
    /// Defaults to "http://127.0.0.1:8472".
    management_url: String,
}

impl HardyHotReconfig {
    /// Create a new `HardyHotReconfig` manager.
    ///
    /// # Arguments
    /// * `management_url` — Optional base URL for Hardy's management API.
    ///   Defaults to `http://127.0.0.1:8472`.
    pub fn new(management_url: Option<String>) -> Self {
        Self {
            management_url: management_url
                .unwrap_or_else(|| "http://127.0.0.1:8472".to_string()),
        }
    }

    /// Add a contact to the running Hardy contact plan via REST API.
    ///
    /// POST /api/contacts with JSON body containing the contact definition.
    pub async fn add_contact(&self, contact: &Contact) -> Result<(), AbstractionError> {
        let url = format!("{}/api/contacts", self.management_url);
        let body = serde_json::json!({
            "source_node": contact.source_node,
            "dest_node": contact.dest_node,
            "start_time": contact.start_time,
            "end_time": contact.end_time,
            "rate_bps": contact.rate_bps,
            "confidence": contact.confidence,
        });

        self.post_json(&url, &body, "add_contact").await
    }

    /// Remove a contact from the running Hardy contact plan via REST API.
    ///
    /// DELETE /api/contacts with JSON body identifying the contact.
    pub async fn remove_contact(&self, contact_ref: &ContactRef) -> Result<(), AbstractionError> {
        let url = format!("{}/api/contacts", self.management_url);
        let body = serde_json::json!({
            "source_node": contact_ref.source_node,
            "dest_node": contact_ref.dest_node,
            "start_time": contact_ref.start_time,
        });

        self.delete_json(&url, &body, "remove_contact").await
    }

    /// Add a neighbor — NOT SUPPORTED by Hardy at runtime.
    ///
    /// Returns `UnsupportedOperation`. Hardy requires a restart to add neighbors.
    pub async fn add_neighbor(&self, _neighbor: &Neighbor) -> Result<(), AbstractionError> {
        Err(AbstractionError::new(
            ErrorCategory::UnsupportedOperation,
            "Hardy does not support adding neighbors at runtime; restart required".to_string(),
            "add_neighbor",
        )
        .with_backend("hardy"))
    }

    /// Remove a neighbor — NOT SUPPORTED by Hardy at runtime.
    ///
    /// Returns `UnsupportedOperation`. Hardy requires a restart to remove neighbors.
    pub async fn remove_neighbor(&self, _node_ref: &NodeRef) -> Result<(), AbstractionError> {
        Err(AbstractionError::new(
            ErrorCategory::UnsupportedOperation,
            "Hardy does not support removing neighbors at runtime; restart required".to_string(),
            "remove_neighbor",
        )
        .with_backend("hardy"))
    }

    /// Enable a convergence layer link via Hardy's REST API.
    ///
    /// POST /api/links/{link_id}/enable
    pub async fn enable_link(&self, link_id: &str) -> Result<(), AbstractionError> {
        let url = format!("{}/api/links/{}/enable", self.management_url, link_id);
        let body = serde_json::json!({});
        self.post_json(&url, &body, "enable_link").await
    }

    /// Disable a convergence layer link via Hardy's REST API.
    ///
    /// POST /api/links/{link_id}/disable
    pub async fn disable_link(&self, link_id: &str) -> Result<(), AbstractionError> {
        let url = format!("{}/api/links/{}/disable", self.management_url, link_id);
        let body = serde_json::json!({});
        self.post_json(&url, &body, "disable_link").await
    }

    /// Send a POST request with a JSON body to Hardy's management API.
    ///
    /// Uses `tokio::process::Command` to invoke `curl` as a subprocess,
    /// avoiding the need for an HTTP client dependency at this layer.
    async fn post_json(
        &self,
        url: &str,
        body: &serde_json::Value,
        operation: &str,
    ) -> Result<(), AbstractionError> {
        let body_str = serde_json::to_string(body).map_err(|e| {
            AbstractionError::new(
                ErrorCategory::RuntimeError,
                format!("Failed to serialize request body: {}", e),
                operation,
            )
            .with_backend("hardy")
        })?;

        let output = tokio::process::Command::new("curl")
            .arg("-s")
            .arg("-X")
            .arg("POST")
            .arg("-H")
            .arg("Content-Type: application/json")
            .arg("-d")
            .arg(&body_str)
            .arg("-w")
            .arg("%{http_code}")
            .arg("-o")
            .arg("/dev/null")
            .arg(url)
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::CommunicationError,
                    format!("Failed to reach Hardy management API: {}", e),
                    operation,
                )
                .with_backend("hardy")
                .with_resource(url.to_string())
            })?;

        let status_str = String::from_utf8_lossy(&output.stdout);
        let http_code: u16 = status_str.trim().parse().unwrap_or(0);

        if !(200..300).contains(&http_code) {
            return Err(AbstractionError::new(
                ErrorCategory::RuntimeError,
                format!(
                    "Hardy management API returned HTTP {}: {}",
                    http_code, operation
                ),
                operation,
            )
            .with_backend("hardy")
            .with_backend_code(format!("http_{}", http_code))
            .with_resource(url.to_string()));
        }

        Ok(())
    }

    /// Send a DELETE request with a JSON body to Hardy's management API.
    async fn delete_json(
        &self,
        url: &str,
        body: &serde_json::Value,
        operation: &str,
    ) -> Result<(), AbstractionError> {
        let body_str = serde_json::to_string(body).map_err(|e| {
            AbstractionError::new(
                ErrorCategory::RuntimeError,
                format!("Failed to serialize request body: {}", e),
                operation,
            )
            .with_backend("hardy")
        })?;

        let output = tokio::process::Command::new("curl")
            .arg("-s")
            .arg("-X")
            .arg("DELETE")
            .arg("-H")
            .arg("Content-Type: application/json")
            .arg("-d")
            .arg(&body_str)
            .arg("-w")
            .arg("%{http_code}")
            .arg("-o")
            .arg("/dev/null")
            .arg(url)
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::CommunicationError,
                    format!("Failed to reach Hardy management API: {}", e),
                    operation,
                )
                .with_backend("hardy")
                .with_resource(url.to_string())
            })?;

        let status_str = String::from_utf8_lossy(&output.stdout);
        let http_code: u16 = status_str.trim().parse().unwrap_or(0);

        if !(200..300).contains(&http_code) {
            return Err(AbstractionError::new(
                ErrorCategory::RuntimeError,
                format!(
                    "Hardy management API returned HTTP {}: {}",
                    http_code, operation
                ),
                operation,
            )
            .with_backend("hardy")
            .with_backend_code(format!("http_{}", http_code))
            .with_resource(url.to_string()));
        }

        Ok(())
    }

    /// Returns the management API base URL.
    pub fn management_url(&self) -> &str {
        &self.management_url
    }
}

/// Returns the CapabilitySet for the Hardy backend.
///
/// Hardy supports:
/// - Hot reconfiguration: add_contact, remove_contact, enable_link, disable_link (via REST API)
/// - Convergence layers: LTP/UDP, TCP-CL, KISS, UDP
/// - Routing: CGR, Static, Default
/// - Security: No BPSec (amateur radio compliant)
///
/// Hardy does NOT support at runtime:
/// - add_neighbor, remove_neighbor (these require a daemon restart)
pub fn hardy_capabilities() -> CapabilitySet {
    CapabilitySet {
        hot_reconfig: HotReconfigCapabilities {
            add_contact: true,
            remove_contact: true,
            add_neighbor: false,
            remove_neighbor: false,
            enable_link: true,
            disable_link: true,
        },
        convergence_layers: vec![
            ConvergenceLayerType::LtpUdp,
            ConvergenceLayerType::TcpCl,
            ConvergenceLayerType::Kiss,
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
    fn test_hardy_capabilities_hot_reconfig() {
        let caps = hardy_capabilities();
        assert!(caps.hot_reconfig.add_contact);
        assert!(caps.hot_reconfig.remove_contact);
        assert!(!caps.hot_reconfig.add_neighbor);
        assert!(!caps.hot_reconfig.remove_neighbor);
        assert!(caps.hot_reconfig.enable_link);
        assert!(caps.hot_reconfig.disable_link);
    }

    #[test]
    fn test_hardy_capabilities_convergence_layers() {
        let caps = hardy_capabilities();
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::LtpUdp));
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::TcpCl));
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::Kiss));
        assert!(caps.convergence_layers.contains(&ConvergenceLayerType::Udp));
    }

    #[test]
    fn test_hardy_capabilities_routing() {
        let caps = hardy_capabilities();
        assert!(caps.routing_strategies.contains(&RoutingStrategy::Cgr));
        assert!(caps.routing_strategies.contains(&RoutingStrategy::Static));
        assert!(caps.routing_strategies.contains(&RoutingStrategy::Default));
    }

    #[test]
    fn test_hardy_capabilities_no_security() {
        let caps = hardy_capabilities();
        assert!(!caps.security.bpsec_bib);
        assert!(!caps.security.bpsec_bcb);
    }

    #[test]
    fn test_hardy_hot_reconfig_new_default_url() {
        let reconfig = HardyHotReconfig::new(None);
        assert_eq!(reconfig.management_url(), "http://127.0.0.1:8472");
    }

    #[test]
    fn test_hardy_hot_reconfig_new_custom_url() {
        let reconfig = HardyHotReconfig::new(Some("http://10.0.0.5:9090".to_string()));
        assert_eq!(reconfig.management_url(), "http://10.0.0.5:9090");
    }

    #[tokio::test]
    async fn test_add_neighbor_returns_unsupported() {
        let reconfig = HardyHotReconfig::new(None);
        let neighbor = Neighbor {
            node_number: 20,
            name: Some("Test".to_string()),
            links: vec![],
            rate_limit_bps: None,
        };
        let result = reconfig.add_neighbor(&neighbor).await;
        assert!(result.is_err());
        let err = result.unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("neighbors"));
    }

    #[tokio::test]
    async fn test_remove_neighbor_returns_unsupported() {
        let reconfig = HardyHotReconfig::new(None);
        let node_ref = NodeRef { node_number: 20 };
        let result = reconfig.remove_neighbor(&node_ref).await;
        assert!(result.is_err());
        let err = result.unwrap_err();
        assert_eq!(err.category, ErrorCategory::UnsupportedOperation);
        assert!(err.message.contains("neighbors"));
    }
}
