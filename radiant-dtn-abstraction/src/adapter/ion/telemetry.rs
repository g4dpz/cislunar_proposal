//! ION-DTN telemetry collection.
//!
//! Collects bundle statistics and per-neighbor link state information
//! from a running ION-DTN instance by parsing CLI tool output.
//!
//! ION exposes telemetry via several CLI tools:
//! - `bpstats` / `bpstats2` — bundle protocol statistics counters
//! - `bplist` — lists bundles currently in the database (queue depth)
//! - `ltpinfo` — LTP session and span information
//!
//! Since ION's exact CLI output format varies by version, this module
//! implements best-effort parsers that return zero values on parse failure
//! (graceful degradation) and log warnings for unexpected formats.

use std::path::PathBuf;

use tokio::process::Command;
use tracing::warn;

use crate::adapter::traits::{BundleStatistics, LinkState};
use crate::error::AbstractionError;

/// Collects telemetry data from a running ION-DTN instance.
///
/// Uses ION CLI tools (`bpstats`, `bplist`, `ltpinfo`) to gather
/// bundle statistics and link state information. Parsing is best-effort:
/// if a tool's output cannot be parsed, zero values are returned rather
/// than failing the entire collection.
pub struct IonTelemetry {
    /// Path to the ION bin directory (where bpstats, bplist, ltpinfo live).
    /// If None, assumes the tools are on $PATH.
    ion_bin_dir: Option<PathBuf>,
}

impl IonTelemetry {
    /// Create a new `IonTelemetry` collector.
    ///
    /// # Arguments
    /// * `ion_bin_dir` — Optional path to the directory containing ION binaries.
    ///   If None, the system $PATH is used to locate bpstats, bplist, etc.
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

    /// Collect aggregate bundle statistics from the running ION instance.
    ///
    /// Runs `bpstats` to gather sourced/forwarded/delivered/expired counters,
    /// and `bplist` to determine the number of bundles currently queued.
    ///
    /// If `bpstats` output cannot be parsed, counters default to zero.
    /// If `bplist` output cannot be parsed, bundles_queued defaults to zero.
    pub async fn collect_stats(&self) -> Result<BundleStatistics, AbstractionError> {
        // Collect counters from bpstats
        let stats = self.run_bpstats().await;

        // Collect queue depth from bplist
        let queued = self.run_bplist_count().await;

        Ok(BundleStatistics {
            bundles_sourced: stats.bundles_sourced,
            bundles_forwarded: stats.bundles_forwarded,
            bundles_delivered: stats.bundles_delivered,
            bundles_expired: stats.bundles_expired,
            bundles_queued: queued,
        })
    }

    /// Query per-neighbor link states from the running ION instance.
    ///
    /// Runs `ltpinfo` or parses `ltpadmin` output to determine which LTP
    /// spans are active and their byte counters. Returns an empty vec if
    /// no link state information can be retrieved.
    pub async fn link_states(&self) -> Result<Vec<LinkState>, AbstractionError> {
        self.run_ltpinfo().await
    }

    /// Run `bpstats` and parse the output into partial BundleStatistics.
    ///
    /// `bpstats` output format varies but typically contains lines like:
    /// ```text
    /// sourced: 42
    /// forwarded: 38
    /// delivered: 35
    /// expired: 2
    /// ```
    /// Or key=value format:
    /// ```text
    /// bundles_sourced=42 bundles_forwarded=38 bundles_delivered=35 bundles_expired=2
    /// ```
    async fn run_bpstats(&self) -> BundleStatistics {
        let bpstats = self.bin_path("bpstats");

        let output = match Command::new(&bpstats).output().await {
            Ok(output) => output,
            Err(e) => {
                warn!(
                    tool = "bpstats",
                    error = %e,
                    "Failed to execute bpstats; returning zero counters"
                );
                return BundleStatistics::default();
            }
        };

        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);
        let combined = format!("{}\n{}", stdout, stderr);

        parse_bpstats_output(&combined)
    }

    /// Run `bplist` and count the number of bundles in the output.
    ///
    /// `bplist` lists one bundle per line (excluding header/footer).
    /// The line count (minus headers) gives us the queue depth.
    async fn run_bplist_count(&self) -> u64 {
        let bplist = self.bin_path("bplist");

        let output = match Command::new(&bplist).output().await {
            Ok(output) => output,
            Err(e) => {
                warn!(
                    tool = "bplist",
                    error = %e,
                    "Failed to execute bplist; returning zero for bundles_queued"
                );
                return 0;
            }
        };

        let stdout = String::from_utf8_lossy(&output.stdout);
        parse_bplist_count(&stdout)
    }

    /// Run `ltpinfo` or `ltpadmin` to gather per-span link state.
    ///
    /// Attempts to parse LTP span information including:
    /// - Remote engine ID (used as neighbor_node)
    /// - Active/inactive state
    /// - Bytes sent/received
    async fn run_ltpinfo(&self) -> Result<Vec<LinkState>, AbstractionError> {
        // Try ltpinfo first
        let ltpinfo = self.bin_path("ltpinfo");
        let output = match Command::new(&ltpinfo).output().await {
            Ok(output) if output.status.success() => output,
            _ => {
                // Fallback: try piping "l span" to ltpadmin
                return self.run_ltpadmin_list_spans().await;
            }
        };

        let stdout = String::from_utf8_lossy(&output.stdout);
        Ok(parse_ltpinfo_output(&stdout))
    }

    /// Fallback: pipe `l span` to `ltpadmin` to list LTP spans.
    async fn run_ltpadmin_list_spans(&self) -> Result<Vec<LinkState>, AbstractionError> {
        let ltpadmin = self.bin_path("ltpadmin");

        let output = match Command::new(&ltpadmin)
            .arg("l span")
            .output()
            .await
        {
            Ok(output) => output,
            Err(e) => {
                warn!(
                    tool = "ltpadmin",
                    error = %e,
                    "Failed to execute ltpadmin for span listing; returning empty link states"
                );
                return Ok(Vec::new());
            }
        };

        let stdout = String::from_utf8_lossy(&output.stdout);
        Ok(parse_ltpinfo_output(&stdout))
    }
}

// ─── Parsers ────────────────────────────────────────────────────────────────

/// Parse `bpstats` output into BundleStatistics.
///
/// Supports multiple output formats:
/// 1. Key-colon-value lines: `sourced: 42`
/// 2. Key=value pairs: `bundles_sourced=42`
/// 3. Whitespace-separated tabular: `sourced  42  forwarded  38 ...`
///
/// Unknown formats return zero values with a logged warning.
fn parse_bpstats_output(output: &str) -> BundleStatistics {
    let mut stats = BundleStatistics::default();
    let mut parsed_any = false;

    for line in output.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        // Try key: value format (e.g., "sourced: 42" or "bundles sourced: 42")
        if let Some((key, value)) = parse_colon_separated(line) {
            if let Ok(val) = value.parse::<u64>() {
                if apply_stat_value(&mut stats, &key, val) {
                    parsed_any = true;
                }
            }
            continue;
        }

        // Try key=value format (e.g., "bundles_sourced=42")
        for token in line.split_whitespace() {
            if let Some((key, value)) = parse_equals_separated(token) {
                if let Ok(val) = value.parse::<u64>() {
                    if apply_stat_value(&mut stats, &key, val) {
                        parsed_any = true;
                    }
                }
            }
        }
    }

    if !parsed_any && !output.trim().is_empty() {
        warn!(
            output_preview = &output[..output.len().min(200)],
            "Could not parse bpstats output; returning zero counters"
        );
    }

    stats
}

/// Parse a `key: value` line, returning (normalized_key, value_str).
fn parse_colon_separated(line: &str) -> Option<(String, &str)> {
    let mut parts = line.splitn(2, ':');
    let key = parts.next()?.trim();
    let value = parts.next()?.trim();

    if key.is_empty() || value.is_empty() {
        return None;
    }

    let normalized = normalize_stat_key(key);
    Some((normalized, value))
}

/// Parse a `key=value` token, returning (normalized_key, value_str).
fn parse_equals_separated(token: &str) -> Option<(String, &str)> {
    let mut parts = token.splitn(2, '=');
    let key = parts.next()?.trim();
    let value = parts.next()?.trim();

    if key.is_empty() || value.is_empty() {
        return None;
    }

    let normalized = normalize_stat_key(key);
    Some((normalized, value))
}

/// Normalize a statistics key to a canonical form for matching.
///
/// Strips common prefixes ("bundles_", "bundles ", "bundle_") and
/// converts to lowercase with underscores.
fn normalize_stat_key(key: &str) -> String {
    let lower = key.to_lowercase();
    let stripped = lower
        .strip_prefix("bundles_")
        .or_else(|| lower.strip_prefix("bundles "))
        .or_else(|| lower.strip_prefix("bundle_"))
        .unwrap_or(&lower);

    stripped.replace([' ', '-'], "_")
}

/// Apply a parsed counter value to the appropriate field in BundleStatistics.
/// Returns true if the key was recognized.
fn apply_stat_value(stats: &mut BundleStatistics, key: &str, value: u64) -> bool {
    match key {
        "sourced" | "src" | "originated" => {
            stats.bundles_sourced = value;
            true
        }
        "forwarded" | "fwd" | "relayed" => {
            stats.bundles_forwarded = value;
            true
        }
        "delivered" | "dlv" | "received" => {
            stats.bundles_delivered = value;
            true
        }
        "expired" | "exp" | "ttl_expired" | "abandoned" => {
            stats.bundles_expired = value;
            true
        }
        "queued" | "pending" | "in_queue" | "stored" => {
            stats.bundles_queued = value;
            true
        }
        _ => false,
    }
}

/// Parse `bplist` output to count the number of bundles queued.
///
/// `bplist` typically outputs one bundle per line. Lines that are
/// empty, contain only headers/separators, or start with common
/// header prefixes are excluded from the count.
fn parse_bplist_count(output: &str) -> u64 {
    let mut count: u64 = 0;

    for line in output.lines() {
        let trimmed = line.trim();

        // Skip empty lines
        if trimmed.is_empty() {
            continue;
        }

        // Skip common header/separator patterns
        if is_bplist_header_line(trimmed) {
            continue;
        }

        // Each remaining line represents a bundle
        count += 1;
    }

    count
}

/// Determine if a line from bplist output is a header or separator.
fn is_bplist_header_line(line: &str) -> bool {
    let lower = line.to_lowercase();

    // Common header patterns
    if lower.starts_with("bundle") && lower.contains("list") {
        return true;
    }
    if lower.starts_with("---") || lower.starts_with("===") {
        return true;
    }
    if lower.starts_with('#') {
        return true;
    }
    if lower.contains("no bundles") || lower.contains("empty") {
        return true;
    }

    false
}

/// Parse `ltpinfo` or `ltpadmin` output for per-span link state.
///
/// ION LTP span information typically includes:
/// - Engine/span number (used as neighbor_node)
/// - Segment activity (active sessions indicate link up)
/// - Byte counters
///
/// Example formats:
/// ```text
/// span 20: active, bytes_sent=1024, bytes_rcvd=512
/// ```
/// or:
/// ```text
/// Engine 20  segments_sent: 42  segments_rcvd: 38  bytes_sent: 1048576  bytes_rcvd: 524288
/// ```
fn parse_ltpinfo_output(output: &str) -> Vec<LinkState> {
    let mut links = Vec::new();

    for line in output.lines() {
        let trimmed = line.trim();
        if trimmed.is_empty() {
            continue;
        }

        if let Some(link) = parse_ltp_span_line(trimmed) {
            links.push(link);
        }
    }

    if links.is_empty() && !output.trim().is_empty() {
        warn!(
            output_preview = &output[..output.len().min(200)],
            "Could not parse ltpinfo output; returning empty link states"
        );
    }

    links
}

/// Attempt to parse a single line of LTP span information into a LinkState.
///
/// Recognizes patterns like:
/// - `span 20: active, bytes_sent=1024, bytes_rcvd=512`
/// - `Engine 20  ...`
/// - `20  active  1024  512`
fn parse_ltp_span_line(line: &str) -> Option<LinkState> {
    let lower = line.to_lowercase();

    // Try "span N" or "engine N" prefix pattern
    let engine_id = extract_span_engine_id(&lower)?;

    // Determine if the link is active
    let active = lower.contains("active")
        || lower.contains("running")
        || lower.contains("up");
    let inactive_explicit = lower.contains("inactive")
        || lower.contains("stopped")
        || lower.contains("down");
    let is_active = active && !inactive_explicit;

    // Extract byte counters
    let bytes_sent = extract_counter(&lower, &["bytes_sent", "bytes sent", "bytessent", "sent"])
        .unwrap_or(0);
    let bytes_received =
        extract_counter(&lower, &["bytes_rcvd", "bytes_received", "bytes received", "bytesrcvd", "rcvd"])
            .unwrap_or(0);

    Some(LinkState {
        neighbor_node: engine_id,
        link_id: format!("ltp-span-{}", engine_id),
        active: is_active,
        bytes_sent,
        bytes_received,
    })
}

/// Extract the engine/span ID number from a line.
///
/// Looks for patterns like "span 20", "engine 20", or a leading number.
fn extract_span_engine_id(line: &str) -> Option<u64> {
    // Try "span N" or "engine N" pattern
    for prefix in &["span", "engine", "peer"] {
        if let Some(rest) = line.strip_prefix(prefix) {
            // Rest might be " 20: active" or " 20 running" etc.
            let rest = rest.trim();
            if let Some(num_str) = rest.split(|c: char| !c.is_ascii_digit()).next() {
                if !num_str.is_empty() {
                    if let Ok(id) = num_str.parse::<u64>() {
                        return Some(id);
                    }
                }
            }
        }
    }

    // Try leading number (tabular format: "20  active  1024  512")
    if let Some(first_token) = line.split_whitespace().next() {
        let cleaned = first_token.trim_end_matches(|c: char| !c.is_ascii_digit());
        let digits: String = cleaned.chars().take_while(|c| c.is_ascii_digit()).collect();
        if !digits.is_empty() {
            if let Ok(id) = digits.parse::<u64>() {
                // Only accept if it looks like a reasonable engine ID (not a byte counter)
                if id <= 100_000 {
                    return Some(id);
                }
            }
        }
    }

    None
}

/// Extract a numeric counter value following a given key label.
///
/// Handles formats like:
/// - `key=value`
/// - `key: value`
/// - `key value`
fn extract_counter(line: &str, keys: &[&str]) -> Option<u64> {
    for key in keys {
        // Try key=value
        if let Some(pos) = line.find(&format!("{}=", key)) {
            let after = &line[pos + key.len() + 1..];
            if let Some(val) = parse_leading_number(after) {
                return Some(val);
            }
        }

        // Try key: value
        if let Some(pos) = line.find(&format!("{}:", key)) {
            let after = &line[pos + key.len() + 1..];
            let after = after.trim();
            if let Some(val) = parse_leading_number(after) {
                return Some(val);
            }
        }

        // Try "key value" (whitespace separated, key as a whole word)
        if let Some(pos) = line.find(key) {
            let after = &line[pos + key.len()..];
            let after = after.trim().trim_start_matches(['=', ':', ' ']);
            if let Some(val) = parse_leading_number(after) {
                return Some(val);
            }
        }
    }

    None
}

/// Parse a leading decimal number from a string slice.
fn parse_leading_number(s: &str) -> Option<u64> {
    let s = s.trim();
    let num_str: String = s.chars().take_while(|c| c.is_ascii_digit()).collect();
    if num_str.is_empty() {
        return None;
    }
    num_str.parse::<u64>().ok()
}

#[cfg(test)]
mod tests {
    use super::*;

    // ─── bpstats parser tests ───────────────────────────────────────────

    #[test]
    fn test_parse_bpstats_colon_format() {
        let output = "\
sourced: 42
forwarded: 38
delivered: 35
expired: 2
";
        let stats = parse_bpstats_output(output);
        assert_eq!(stats.bundles_sourced, 42);
        assert_eq!(stats.bundles_forwarded, 38);
        assert_eq!(stats.bundles_delivered, 35);
        assert_eq!(stats.bundles_expired, 2);
    }

    #[test]
    fn test_parse_bpstats_equals_format() {
        let output = "bundles_sourced=100 bundles_forwarded=90 bundles_delivered=85 bundles_expired=5";
        let stats = parse_bpstats_output(output);
        assert_eq!(stats.bundles_sourced, 100);
        assert_eq!(stats.bundles_forwarded, 90);
        assert_eq!(stats.bundles_delivered, 85);
        assert_eq!(stats.bundles_expired, 5);
    }

    #[test]
    fn test_parse_bpstats_prefixed_colon_format() {
        let output = "\
bundles sourced: 10
bundles forwarded: 8
bundles delivered: 7
bundles expired: 1
";
        let stats = parse_bpstats_output(output);
        assert_eq!(stats.bundles_sourced, 10);
        assert_eq!(stats.bundles_forwarded, 8);
        assert_eq!(stats.bundles_delivered, 7);
        assert_eq!(stats.bundles_expired, 1);
    }

    #[test]
    fn test_parse_bpstats_empty_output() {
        let stats = parse_bpstats_output("");
        assert_eq!(stats.bundles_sourced, 0);
        assert_eq!(stats.bundles_forwarded, 0);
        assert_eq!(stats.bundles_delivered, 0);
        assert_eq!(stats.bundles_expired, 0);
    }

    #[test]
    fn test_parse_bpstats_garbage_returns_zeros() {
        let output = "this is not parseable telemetry data\nrandom noise\n";
        let stats = parse_bpstats_output(output);
        assert_eq!(stats.bundles_sourced, 0);
        assert_eq!(stats.bundles_forwarded, 0);
        assert_eq!(stats.bundles_delivered, 0);
        assert_eq!(stats.bundles_expired, 0);
    }

    // ─── bplist parser tests ────────────────────────────────────────────

    #[test]
    fn test_parse_bplist_count_with_bundles() {
        let output = "\
Bundle list:
---
ipn:10.1 -> ipn:20.1 (1700000042)
ipn:10.1 -> ipn:20.2 (1700000043)
ipn:30.1 -> ipn:10.1 (1700000044)
";
        assert_eq!(parse_bplist_count(output), 3);
    }

    #[test]
    fn test_parse_bplist_count_empty() {
        let output = "";
        assert_eq!(parse_bplist_count(output), 0);
    }

    #[test]
    fn test_parse_bplist_count_no_bundles_message() {
        let output = "No bundles in database.\n";
        assert_eq!(parse_bplist_count(output), 0);
    }

    #[test]
    fn test_parse_bplist_count_with_header_and_separator() {
        let output = "\
# Bundle database contents
---
ipn:10.1 -> ipn:20.1
ipn:10.1 -> ipn:20.2
";
        assert_eq!(parse_bplist_count(output), 2);
    }

    // ─── ltpinfo parser tests ───────────────────────────────────────────

    #[test]
    fn test_parse_ltpinfo_span_format() {
        let output = "\
span 20: active, bytes_sent=1024, bytes_rcvd=512
span 30: inactive, bytes_sent=0, bytes_rcvd=0
";
        let links = parse_ltpinfo_output(output);
        assert_eq!(links.len(), 2);

        assert_eq!(links[0].neighbor_node, 20);
        assert_eq!(links[0].link_id, "ltp-span-20");
        assert!(links[0].active);
        assert_eq!(links[0].bytes_sent, 1024);
        assert_eq!(links[0].bytes_received, 512);

        assert_eq!(links[1].neighbor_node, 30);
        assert_eq!(links[1].link_id, "ltp-span-30");
        assert!(!links[1].active);
        assert_eq!(links[1].bytes_sent, 0);
        assert_eq!(links[1].bytes_received, 0);
    }

    #[test]
    fn test_parse_ltpinfo_engine_format() {
        let output = "engine 20 active bytes_sent: 2048 bytes_received: 1024\n";
        let links = parse_ltpinfo_output(output);
        assert_eq!(links.len(), 1);
        assert_eq!(links[0].neighbor_node, 20);
        assert!(links[0].active);
        assert_eq!(links[0].bytes_sent, 2048);
        assert_eq!(links[0].bytes_received, 1024);
    }

    #[test]
    fn test_parse_ltpinfo_empty() {
        let links = parse_ltpinfo_output("");
        assert!(links.is_empty());
    }

    #[test]
    fn test_parse_ltpinfo_garbage_returns_empty() {
        let output = "some unrecognized format\nnot ltp data\n";
        let links = parse_ltpinfo_output(output);
        assert!(links.is_empty());
    }

    // ─── Helper function tests ──────────────────────────────────────────

    #[test]
    fn test_normalize_stat_key() {
        assert_eq!(normalize_stat_key("bundles_sourced"), "sourced");
        assert_eq!(normalize_stat_key("bundles forwarded"), "forwarded");
        assert_eq!(normalize_stat_key("bundle_delivered"), "delivered");
        assert_eq!(normalize_stat_key("expired"), "expired");
        assert_eq!(normalize_stat_key("Bundles_Sourced"), "sourced");
    }

    #[test]
    fn test_extract_span_engine_id() {
        assert_eq!(extract_span_engine_id("span 20: active"), Some(20));
        assert_eq!(extract_span_engine_id("engine 30 running"), Some(30));
        assert_eq!(extract_span_engine_id("peer 15, bytes=100"), Some(15));
        assert_eq!(extract_span_engine_id("20 active 1024 512"), Some(20));
        assert_eq!(extract_span_engine_id("not a span line"), None);
    }

    #[test]
    fn test_extract_counter() {
        assert_eq!(
            extract_counter("bytes_sent=1024, bytes_rcvd=512", &["bytes_sent"]),
            Some(1024)
        );
        assert_eq!(
            extract_counter("bytes_rcvd=512", &["bytes_rcvd"]),
            Some(512)
        );
        assert_eq!(
            extract_counter("bytes sent: 2048", &["bytes sent"]),
            Some(2048)
        );
        assert_eq!(
            extract_counter("no match here", &["bytes_sent"]),
            None
        );
    }

    #[test]
    fn test_parse_leading_number() {
        assert_eq!(parse_leading_number("1024,"), Some(1024));
        assert_eq!(parse_leading_number("  512 bytes"), Some(512));
        assert_eq!(parse_leading_number("abc"), None);
        assert_eq!(parse_leading_number(""), None);
    }

    #[test]
    fn test_apply_stat_value_recognized_keys() {
        let mut stats = BundleStatistics::default();
        assert!(apply_stat_value(&mut stats, "sourced", 10));
        assert!(apply_stat_value(&mut stats, "fwd", 8));
        assert!(apply_stat_value(&mut stats, "delivered", 7));
        assert!(apply_stat_value(&mut stats, "exp", 1));
        assert!(apply_stat_value(&mut stats, "queued", 3));

        assert_eq!(stats.bundles_sourced, 10);
        assert_eq!(stats.bundles_forwarded, 8);
        assert_eq!(stats.bundles_delivered, 7);
        assert_eq!(stats.bundles_expired, 1);
        assert_eq!(stats.bundles_queued, 3);
    }

    #[test]
    fn test_apply_stat_value_unrecognized_key() {
        let mut stats = BundleStatistics::default();
        assert!(!apply_stat_value(&mut stats, "unknown_counter", 99));
        // Stats should remain at zero
        assert_eq!(stats.bundles_sourced, 0);
    }

    #[test]
    fn test_is_bplist_header_line() {
        assert!(is_bplist_header_line("Bundle list:"));
        assert!(is_bplist_header_line("---"));
        assert!(is_bplist_header_line("==="));
        assert!(is_bplist_header_line("# header comment"));
        assert!(is_bplist_header_line("No bundles in database."));
        assert!(!is_bplist_header_line("ipn:10.1 -> ipn:20.1"));
    }

    // ─── IonTelemetry struct tests ──────────────────────────────────────

    #[test]
    fn test_ion_telemetry_new_with_bin_dir() {
        let telem = IonTelemetry::new(Some(PathBuf::from("/opt/ion/bin")));
        assert_eq!(telem.bin_path("bpstats"), PathBuf::from("/opt/ion/bin/bpstats"));
        assert_eq!(telem.bin_path("bplist"), PathBuf::from("/opt/ion/bin/bplist"));
        assert_eq!(telem.bin_path("ltpinfo"), PathBuf::from("/opt/ion/bin/ltpinfo"));
    }

    #[test]
    fn test_ion_telemetry_new_without_bin_dir() {
        let telem = IonTelemetry::new(None);
        assert_eq!(telem.bin_path("bpstats"), PathBuf::from("bpstats"));
        assert_eq!(telem.bin_path("ltpinfo"), PathBuf::from("ltpinfo"));
    }
}
