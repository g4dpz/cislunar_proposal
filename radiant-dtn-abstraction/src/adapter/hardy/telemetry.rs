//! Hardy DTN telemetry collection.
//!
//! Collects bundle statistics and per-neighbor link state information
//! from a running Hardy instance by querying its management REST API.
//!
//! Hardy exposes a JSON REST API for statistics:
//! - GET /api/stats — aggregate bundle counters
//! - GET /api/links — per-neighbor link state
//!
//! If the API is unreachable, this module degrades gracefully by returning
//! zero-valued statistics (matching the ION adapter's behavior).

use tokio::process::Command;
use tracing::warn;

use crate::adapter::traits::{BundleStatistics, LinkState};
use crate::error::AbstractionError;

/// Collects telemetry data from a running Hardy instance.
///
/// Queries Hardy's management REST API to gather bundle statistics and
/// link state information. If the API is unreachable, zero values are
/// returned (graceful degradation).
pub struct HardyTelemetry {
    /// Base URL for Hardy's management REST API.
    management_url: String,
}

impl HardyTelemetry {
    /// Create a new `HardyTelemetry` collector.
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

    /// Collect aggregate bundle statistics from the running Hardy instance.
    ///
    /// Queries GET /api/stats and parses the JSON response into BundleStatistics.
    /// If the API is unreachable or the response cannot be parsed, returns
    /// zero-valued statistics.
    pub async fn collect_stats(&self) -> Result<BundleStatistics, AbstractionError> {
        let url = format!("{}/api/stats", self.management_url);
        let json_output = self.http_get(&url).await;

        match json_output {
            Some(json_str) => Ok(parse_stats_json(&json_str)),
            None => {
                warn!(
                    url = %url,
                    "Could not reach Hardy stats API; returning zero counters"
                );
                Ok(BundleStatistics {
                    bundles_sourced: 0,
                    bundles_forwarded: 0,
                    bundles_delivered: 0,
                    bundles_expired: 0,
                    bundles_queued: 0,
                })
            }
        }
    }

    /// Query per-neighbor link states from the running Hardy instance.
    ///
    /// Queries GET /api/links and parses the JSON response into LinkState entries.
    /// If the API is unreachable or the response cannot be parsed, returns
    /// an empty vector.
    pub async fn link_states(&self) -> Result<Vec<LinkState>, AbstractionError> {
        let url = format!("{}/api/links", self.management_url);
        let json_output = self.http_get(&url).await;

        match json_output {
            Some(json_str) => Ok(parse_links_json(&json_str)),
            None => {
                warn!(
                    url = %url,
                    "Could not reach Hardy links API; returning empty link states"
                );
                Ok(Vec::new())
            }
        }
    }

    /// Execute an HTTP GET request using curl as a subprocess.
    ///
    /// Returns Some(response_body) on success (HTTP 2xx), or None on failure.
    /// This approach avoids requiring an HTTP client library dependency.
    async fn http_get(&self, url: &str) -> Option<String> {
        let output = Command::new("curl")
            .arg("-s")
            .arg("-f") // Fail silently on HTTP errors (non-2xx)
            .arg("--connect-timeout")
            .arg("5")
            .arg(url)
            .output()
            .await
            .ok()?;

        if !output.status.success() {
            return None;
        }

        let body = String::from_utf8_lossy(&output.stdout);
        if body.trim().is_empty() {
            return None;
        }

        Some(body.to_string())
    }

    /// Returns the management API base URL.
    pub fn management_url(&self) -> &str {
        &self.management_url
    }
}

// ─── JSON Parsers ───────────────────────────────────────────────────────────

/// Parse Hardy's /api/stats JSON response into BundleStatistics.
///
/// Expected JSON format:
/// ```json
/// {
///   "bundles_sourced": 42,
///   "bundles_forwarded": 38,
///   "bundles_delivered": 35,
///   "bundles_expired": 2,
///   "bundles_queued": 5
/// }
/// ```
///
/// Also handles nested formats like:
/// ```json
/// {
///   "stats": {
///     "sourced": 42,
///     "forwarded": 38,
///     ...
///   }
/// }
/// ```
///
/// Returns zero-valued statistics if parsing fails.
fn parse_stats_json(json_str: &str) -> BundleStatistics {
    let value: serde_json::Value = match serde_json::from_str(json_str) {
        Ok(v) => v,
        Err(e) => {
            warn!(
                error = %e,
                json_preview = &json_str[..json_str.len().min(200)],
                "Failed to parse Hardy stats JSON; returning zero counters"
            );
            return BundleStatistics {
                bundles_sourced: 0,
                bundles_forwarded: 0,
                bundles_delivered: 0,
                bundles_expired: 0,
                bundles_queued: 0,
            };
        }
    };

    // Try top-level fields first
    if let Some(stats) = extract_stats_from_object(&value) {
        return stats;
    }

    // Try nested "stats" object
    if let Some(stats_obj) = value.get("stats") {
        if let Some(stats) = extract_stats_from_object(stats_obj) {
            return stats;
        }
    }

    // Try nested "bundles" object
    if let Some(bundles_obj) = value.get("bundles") {
        if let Some(stats) = extract_stats_from_object(bundles_obj) {
            return stats;
        }
    }

    warn!(
        json_preview = &json_str[..json_str.len().min(200)],
        "Could not extract stats from Hardy JSON response; returning zero counters"
    );

    BundleStatistics {
        bundles_sourced: 0,
        bundles_forwarded: 0,
        bundles_delivered: 0,
        bundles_expired: 0,
        bundles_queued: 0,
    }
}

/// Extract BundleStatistics fields from a JSON object.
///
/// Recognizes various key naming conventions:
/// - bundles_sourced, sourced, src, originated
/// - bundles_forwarded, forwarded, fwd, relayed
/// - bundles_delivered, delivered, dlv, received
/// - bundles_expired, expired, exp, abandoned
/// - bundles_queued, queued, pending, stored
fn extract_stats_from_object(obj: &serde_json::Value) -> Option<BundleStatistics> {
    let map = obj.as_object()?;

    let mut stats = BundleStatistics {
        bundles_sourced: 0,
        bundles_forwarded: 0,
        bundles_delivered: 0,
        bundles_expired: 0,
        bundles_queued: 0,
    };
    let mut found_any = false;

    for (key, value) in map {
        let num = value.as_u64().unwrap_or(0);
        let key_lower = key.to_lowercase();
        let key_normalized = key_lower
            .strip_prefix("bundles_")
            .or_else(|| key_lower.strip_prefix("bundle_"))
            .unwrap_or(&key_lower);

        match key_normalized {
            "sourced" | "src" | "originated" => {
                stats.bundles_sourced = num;
                found_any = true;
            }
            "forwarded" | "fwd" | "relayed" => {
                stats.bundles_forwarded = num;
                found_any = true;
            }
            "delivered" | "dlv" | "received" => {
                stats.bundles_delivered = num;
                found_any = true;
            }
            "expired" | "exp" | "abandoned" | "ttl_expired" => {
                stats.bundles_expired = num;
                found_any = true;
            }
            "queued" | "pending" | "stored" | "in_queue" => {
                stats.bundles_queued = num;
                found_any = true;
            }
            _ => {}
        }
    }

    if found_any {
        Some(stats)
    } else {
        None
    }
}

/// Parse Hardy's /api/links JSON response into LinkState entries.
///
/// Expected JSON format:
/// ```json
/// [
///   {
///     "neighbor_node": 20,
///     "link_id": "ltp-to-orbiter",
///     "active": true,
///     "bytes_sent": 1024,
///     "bytes_received": 512
///   }
/// ]
/// ```
///
/// Also handles a wrapper object like:
/// ```json
/// { "links": [ ... ] }
/// ```
///
/// Returns an empty vector if parsing fails.
fn parse_links_json(json_str: &str) -> Vec<LinkState> {
    let value: serde_json::Value = match serde_json::from_str(json_str) {
        Ok(v) => v,
        Err(e) => {
            warn!(
                error = %e,
                json_preview = &json_str[..json_str.len().min(200)],
                "Failed to parse Hardy links JSON; returning empty link states"
            );
            return Vec::new();
        }
    };

    // Try as a top-level array
    if let Some(arr) = value.as_array() {
        return extract_link_states_from_array(arr);
    }

    // Try nested "links" array
    if let Some(links_val) = value.get("links") {
        if let Some(arr) = links_val.as_array() {
            return extract_link_states_from_array(arr);
        }
    }

    warn!(
        json_preview = &json_str[..json_str.len().min(200)],
        "Could not extract links from Hardy JSON response; returning empty"
    );

    Vec::new()
}

/// Extract LinkState entries from a JSON array.
fn extract_link_states_from_array(arr: &[serde_json::Value]) -> Vec<LinkState> {
    let mut links = Vec::new();

    for item in arr {
        if let Some(link) = extract_link_state_from_object(item) {
            links.push(link);
        }
    }

    links
}

/// Extract a single LinkState from a JSON object.
///
/// Recognizes fields:
/// - neighbor_node / peer_node / node_id / remote_node
/// - link_id / id / name
/// - active / up / running / state
/// - bytes_sent / sent_bytes / tx_bytes
/// - bytes_received / received_bytes / rx_bytes / bytes_rcvd
fn extract_link_state_from_object(obj: &serde_json::Value) -> Option<LinkState> {
    let map = obj.as_object()?;

    // Extract neighbor node ID
    let neighbor_node = map
        .get("neighbor_node")
        .or_else(|| map.get("peer_node"))
        .or_else(|| map.get("node_id"))
        .or_else(|| map.get("remote_node"))
        .and_then(|v| v.as_u64())?;

    // Extract link ID
    let link_id = map
        .get("link_id")
        .or_else(|| map.get("id"))
        .or_else(|| map.get("name"))
        .and_then(|v| v.as_str())
        .unwrap_or("unknown")
        .to_string();

    // Extract active status
    let active = map
        .get("active")
        .or_else(|| map.get("up"))
        .or_else(|| map.get("running"))
        .and_then(|v| v.as_bool())
        .or_else(|| {
            // Check for "state" string field
            map.get("state")
                .and_then(|v| v.as_str())
                .map(|s| {
                    let lower = s.to_lowercase();
                    lower == "active" || lower == "up" || lower == "running"
                })
        })
        .unwrap_or(false);

    // Extract bytes sent
    let bytes_sent = map
        .get("bytes_sent")
        .or_else(|| map.get("sent_bytes"))
        .or_else(|| map.get("tx_bytes"))
        .and_then(|v| v.as_u64())
        .unwrap_or(0);

    // Extract bytes received
    let bytes_received = map
        .get("bytes_received")
        .or_else(|| map.get("received_bytes"))
        .or_else(|| map.get("rx_bytes"))
        .or_else(|| map.get("bytes_rcvd"))
        .and_then(|v| v.as_u64())
        .unwrap_or(0);

    Some(LinkState {
        neighbor_node,
        link_id,
        active,
        bytes_sent,
        bytes_received,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    // ─── Stats JSON parsing tests ───────────────────────────────────────

    #[test]
    fn test_parse_stats_json_standard_format() {
        let json = r#"{
            "bundles_sourced": 42,
            "bundles_forwarded": 38,
            "bundles_delivered": 35,
            "bundles_expired": 2,
            "bundles_queued": 5
        }"#;
        let stats = parse_stats_json(json);
        assert_eq!(stats.bundles_sourced, 42);
        assert_eq!(stats.bundles_forwarded, 38);
        assert_eq!(stats.bundles_delivered, 35);
        assert_eq!(stats.bundles_expired, 2);
        assert_eq!(stats.bundles_queued, 5);
    }

    #[test]
    fn test_parse_stats_json_short_keys() {
        let json = r#"{
            "sourced": 10,
            "forwarded": 8,
            "delivered": 7,
            "expired": 1,
            "queued": 3
        }"#;
        let stats = parse_stats_json(json);
        assert_eq!(stats.bundles_sourced, 10);
        assert_eq!(stats.bundles_forwarded, 8);
        assert_eq!(stats.bundles_delivered, 7);
        assert_eq!(stats.bundles_expired, 1);
        assert_eq!(stats.bundles_queued, 3);
    }

    #[test]
    fn test_parse_stats_json_nested_stats_object() {
        let json = r#"{
            "stats": {
                "sourced": 100,
                "forwarded": 90,
                "delivered": 85,
                "expired": 5,
                "queued": 10
            }
        }"#;
        let stats = parse_stats_json(json);
        assert_eq!(stats.bundles_sourced, 100);
        assert_eq!(stats.bundles_forwarded, 90);
        assert_eq!(stats.bundles_delivered, 85);
        assert_eq!(stats.bundles_expired, 5);
        assert_eq!(stats.bundles_queued, 10);
    }

    #[test]
    fn test_parse_stats_json_invalid_json() {
        let json = "not valid json{{{";
        let stats = parse_stats_json(json);
        assert_eq!(stats.bundles_sourced, 0);
        assert_eq!(stats.bundles_forwarded, 0);
        assert_eq!(stats.bundles_delivered, 0);
        assert_eq!(stats.bundles_expired, 0);
        assert_eq!(stats.bundles_queued, 0);
    }

    #[test]
    fn test_parse_stats_json_empty_object() {
        let json = "{}";
        let stats = parse_stats_json(json);
        assert_eq!(stats.bundles_sourced, 0);
        assert_eq!(stats.bundles_forwarded, 0);
    }

    #[test]
    fn test_parse_stats_json_unrecognized_keys() {
        let json = r#"{"unknown_field": 42, "another": 99}"#;
        let stats = parse_stats_json(json);
        assert_eq!(stats.bundles_sourced, 0);
    }

    // ─── Links JSON parsing tests ───────────────────────────────────────

    #[test]
    fn test_parse_links_json_array_format() {
        let json = r#"[
            {
                "neighbor_node": 20,
                "link_id": "ltp-to-orbiter",
                "active": true,
                "bytes_sent": 1024,
                "bytes_received": 512
            },
            {
                "neighbor_node": 30,
                "link_id": "tcp-to-relay",
                "active": false,
                "bytes_sent": 0,
                "bytes_received": 0
            }
        ]"#;
        let links = parse_links_json(json);
        assert_eq!(links.len(), 2);

        assert_eq!(links[0].neighbor_node, 20);
        assert_eq!(links[0].link_id, "ltp-to-orbiter");
        assert!(links[0].active);
        assert_eq!(links[0].bytes_sent, 1024);
        assert_eq!(links[0].bytes_received, 512);

        assert_eq!(links[1].neighbor_node, 30);
        assert_eq!(links[1].link_id, "tcp-to-relay");
        assert!(!links[1].active);
        assert_eq!(links[1].bytes_sent, 0);
        assert_eq!(links[1].bytes_received, 0);
    }

    #[test]
    fn test_parse_links_json_wrapped_format() {
        let json = r#"{
            "links": [
                {
                    "neighbor_node": 20,
                    "link_id": "ltp-span-20",
                    "active": true,
                    "bytes_sent": 2048,
                    "bytes_received": 1024
                }
            ]
        }"#;
        let links = parse_links_json(json);
        assert_eq!(links.len(), 1);
        assert_eq!(links[0].neighbor_node, 20);
        assert_eq!(links[0].link_id, "ltp-span-20");
        assert!(links[0].active);
    }

    #[test]
    fn test_parse_links_json_alternative_keys() {
        let json = r#"[
            {
                "peer_node": 25,
                "id": "udp-link",
                "up": true,
                "tx_bytes": 4096,
                "rx_bytes": 2048
            }
        ]"#;
        let links = parse_links_json(json);
        assert_eq!(links.len(), 1);
        assert_eq!(links[0].neighbor_node, 25);
        assert_eq!(links[0].link_id, "udp-link");
        assert!(links[0].active);
        assert_eq!(links[0].bytes_sent, 4096);
        assert_eq!(links[0].bytes_received, 2048);
    }

    #[test]
    fn test_parse_links_json_state_string_field() {
        let json = r#"[
            {
                "node_id": 20,
                "name": "test-link",
                "state": "active",
                "bytes_sent": 100,
                "bytes_received": 50
            }
        ]"#;
        let links = parse_links_json(json);
        assert_eq!(links.len(), 1);
        assert!(links[0].active);
    }

    #[test]
    fn test_parse_links_json_inactive_state_string() {
        let json = r#"[
            {
                "node_id": 20,
                "name": "test-link",
                "state": "down",
                "bytes_sent": 100,
                "bytes_received": 50
            }
        ]"#;
        let links = parse_links_json(json);
        assert_eq!(links.len(), 1);
        assert!(!links[0].active);
    }

    #[test]
    fn test_parse_links_json_invalid_json() {
        let json = "not valid json";
        let links = parse_links_json(json);
        assert!(links.is_empty());
    }

    #[test]
    fn test_parse_links_json_empty_array() {
        let json = "[]";
        let links = parse_links_json(json);
        assert!(links.is_empty());
    }

    #[test]
    fn test_parse_links_json_missing_neighbor_node_skips_entry() {
        let json = r#"[
            {
                "link_id": "orphan-link",
                "active": true,
                "bytes_sent": 100,
                "bytes_received": 50
            }
        ]"#;
        let links = parse_links_json(json);
        // Entry without neighbor_node should be skipped
        assert!(links.is_empty());
    }

    // ─── HardyTelemetry struct tests ────────────────────────────────────

    #[test]
    fn test_hardy_telemetry_new_default_url() {
        let telem = HardyTelemetry::new(None);
        assert_eq!(telem.management_url(), "http://127.0.0.1:8472");
    }

    #[test]
    fn test_hardy_telemetry_new_custom_url() {
        let telem = HardyTelemetry::new(Some("http://10.0.0.1:9090".to_string()));
        assert_eq!(telem.management_url(), "http://10.0.0.1:9090");
    }

    #[tokio::test]
    async fn test_collect_stats_returns_zeros_when_api_unreachable() {
        // Use an unreachable URL to test graceful degradation
        let telem = HardyTelemetry::new(Some("http://127.0.0.1:1".to_string()));
        let stats = telem.collect_stats().await.unwrap();
        assert_eq!(stats.bundles_sourced, 0);
        assert_eq!(stats.bundles_forwarded, 0);
        assert_eq!(stats.bundles_delivered, 0);
        assert_eq!(stats.bundles_expired, 0);
        assert_eq!(stats.bundles_queued, 0);
    }

    #[tokio::test]
    async fn test_link_states_returns_empty_when_api_unreachable() {
        // Use an unreachable URL to test graceful degradation
        let telem = HardyTelemetry::new(Some("http://127.0.0.1:1".to_string()));
        let links = telem.link_states().await.unwrap();
        assert!(links.is_empty());
    }
}
