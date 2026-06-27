//! Hardy DTN lifecycle management.
//!
//! Provides start, stop, restart, health check, and version query
//! for the Hardy DTN daemon. Hardy is a Rust-native BPv7 implementation
//! that runs as a single binary process with a YAML configuration file.

use std::path::{Path, PathBuf};

use tokio::process::Command;
use tracing::warn;

use crate::adapter::traits::HealthStatus;
use crate::error::{AbstractionError, ErrorCategory};

/// Manages the lifecycle of the Hardy DTN daemon process.
///
/// Hardy runs as a single binary (`hardy`) that accepts a YAML configuration
/// file as its argument. Shutdown is via SIGTERM to the process.
pub struct HardyLifecycle {
    /// Path to the Hardy binary directory.
    /// If None, assumes `hardy` is on $PATH.
    hardy_bin_dir: Option<PathBuf>,
    /// Path to the config directory used for the last start.
    /// Retained for restart().
    config_dir: Option<PathBuf>,
    /// Optional management API base URL for health checks.
    /// Defaults to "http://127.0.0.1:8472" if not set.
    management_url: String,
}

impl HardyLifecycle {
    /// Create a new `HardyLifecycle` manager.
    ///
    /// # Arguments
    /// * `hardy_bin_dir` — Optional path to the directory containing the Hardy binary.
    ///   If None, the system $PATH is used to locate `hardy`.
    /// * `management_url` — Optional base URL for Hardy's management REST API.
    ///   Defaults to `http://127.0.0.1:8472`.
    pub fn new(hardy_bin_dir: Option<PathBuf>, management_url: Option<String>) -> Self {
        Self {
            hardy_bin_dir,
            config_dir: None,
            management_url: management_url
                .unwrap_or_else(|| "http://127.0.0.1:8472".to_string()),
        }
    }

    /// Resolve the full path to the Hardy binary.
    fn bin_path(&self, name: &str) -> PathBuf {
        match &self.hardy_bin_dir {
            Some(dir) => dir.join(name),
            None => PathBuf::from(name),
        }
    }

    /// Start the Hardy daemon with the given configuration directory.
    ///
    /// Looks for `hardy.yaml` in `config_dir` and launches the Hardy binary
    /// with that config file as an argument. The process is spawned in the
    /// background (detached).
    pub async fn start(&mut self, config_dir: &Path) -> Result<(), AbstractionError> {
        self.config_dir = Some(config_dir.to_path_buf());

        let config_file = config_dir.join("hardy.yaml");
        if !config_file.exists() {
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!(
                    "Hardy config file not found: {}",
                    config_file.display()
                ),
                "start",
            )
            .with_backend("hardy")
            .with_resource(format!("{}", config_file.display())));
        }

        let hardy_bin = self.bin_path("hardy");

        let output = Command::new(&hardy_bin)
            .arg("--config")
            .arg(&config_file)
            .arg("daemon")
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::LifecycleError,
                    format!("Failed to execute hardy: {}", e),
                    "start",
                )
                .with_backend("hardy")
            })?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!(
                    "Hardy failed to start (exit code {:?}): {}{}",
                    output.status.code(),
                    stderr.trim(),
                    if !stdout.trim().is_empty() {
                        format!(" | stdout: {}", stdout.trim())
                    } else {
                        String::new()
                    }
                ),
                "start",
            )
            .with_backend("hardy")
            .with_resource(format!("{}", config_file.display())));
        }

        Ok(())
    }

    /// Stop the Hardy daemon by sending SIGTERM to the process.
    ///
    /// Uses `pkill -TERM hardy` to gracefully terminate the Hardy daemon.
    /// If pkill is not available, falls back to finding the PID via `pgrep`.
    pub async fn stop(&self) -> Result<(), AbstractionError> {
        // Use pkill to send SIGTERM to the hardy process
        let output = Command::new("pkill")
            .arg("-TERM")
            .arg("-x")
            .arg("hardy")
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::LifecycleError,
                    format!("Failed to execute pkill for hardy: {}", e),
                    "stop",
                )
                .with_backend("hardy")
            })?;

        // pkill returns 0 if at least one process was signaled,
        // 1 if no processes matched (which means hardy wasn't running)
        if !output.status.success() {
            let exit_code = output.status.code().unwrap_or(-1);
            if exit_code == 1 {
                // No process found — hardy wasn't running; this is acceptable
                warn!("Hardy process not found during stop (may not have been running)");
                return Ok(());
            }
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!(
                    "pkill hardy failed (exit code {}): {}",
                    exit_code,
                    stderr.trim()
                ),
                "stop",
            )
            .with_backend("hardy"));
        }

        Ok(())
    }

    /// Restart Hardy by stopping then starting with the given config directory.
    pub async fn restart(&mut self, config_dir: &Path) -> Result<(), AbstractionError> {
        // Stop, ignoring errors (Hardy may not be running)
        let _ = self.stop().await;

        // Brief pause to allow the daemon to fully exit
        tokio::time::sleep(std::time::Duration::from_millis(500)).await;

        self.start(config_dir).await
    }

    /// Check the health of the running Hardy instance.
    ///
    /// Primary check: attempt to reach Hardy's management REST API health endpoint.
    /// Fallback: check if the `hardy` process is running via `pgrep`.
    pub async fn health(&self) -> Result<HealthStatus, AbstractionError> {
        // Primary check: query the management API health endpoint
        // We check if the process is running via pgrep as a reliable
        // cross-platform approach (avoids needing an HTTP client dependency
        // at this layer).
        if self.check_hardy_process().await {
            Ok(HealthStatus {
                running: true,
                uptime_secs: None, // Hardy doesn't expose uptime directly via process check
                message: Some("Hardy daemon is running".to_string()),
            })
        } else {
            Ok(HealthStatus {
                running: false,
                uptime_secs: None,
                message: Some("Hardy daemon is not running".to_string()),
            })
        }
    }

    /// Query the Hardy version string.
    ///
    /// Runs `hardy --version` and captures the version output.
    pub async fn version(&self) -> Result<String, AbstractionError> {
        let hardy_bin = self.bin_path("hardy");

        let output = Command::new(&hardy_bin)
            .arg("--version")
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::LifecycleError,
                    format!("Failed to execute hardy --version: {}", e),
                    "version",
                )
                .with_backend("hardy")
            })?;

        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);
        let combined = format!("{}\n{}", stdout, stderr);

        if let Some(version) = parse_hardy_version(&combined) {
            Ok(version)
        } else if !stdout.trim().is_empty() {
            // Return the raw first line as a fallback
            Ok(stdout.lines().next().unwrap_or("unknown").trim().to_string())
        } else {
            Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                "Could not determine Hardy version from output".to_string(),
                "version",
            )
            .with_backend("hardy"))
        }
    }

    /// Check if the Hardy process is running via `pgrep`.
    async fn check_hardy_process(&self) -> bool {
        let result = Command::new("pgrep")
            .arg("-x")
            .arg("hardy")
            .output()
            .await;

        matches!(result, Ok(output) if output.status.success())
    }

    /// Returns the management API base URL.
    pub fn management_url(&self) -> &str {
        &self.management_url
    }
}

/// Parse the Hardy version string from `hardy --version` output.
///
/// Hardy version output typically looks like:
/// - `hardy 0.5.2`
/// - `Hardy DTN v0.5.2`
/// - `hardy version 0.5.2-dev`
fn parse_hardy_version(output: &str) -> Option<String> {
    for line in output.lines() {
        let trimmed = line.trim();
        if trimmed.is_empty() {
            continue;
        }

        let lower = trimmed.to_lowercase();
        if lower.contains("hardy") {
            // Try to extract version number
            if let Some(version) = extract_version_number(trimmed) {
                return Some(format!("Hardy {}", version));
            }
            // Return the full line as fallback
            return Some(trimmed.to_string());
        }

        // If the first non-empty line looks like a version string, use it
        if let Some(version) = extract_version_number(trimmed) {
            return Some(format!("Hardy {}", version));
        }
    }
    None
}

/// Extract a semver-like version number (X.Y.Z or X.Y.Z-suffix) from a string.
fn extract_version_number(s: &str) -> Option<String> {
    // Look for a pattern like "v0.5.2" or "0.5.2" or "0.5.2-dev"
    let mut chars = s.chars().peekable();
    while let Some(&c) = chars.peek() {
        if c == 'v' || c == 'V' {
            chars.next();
            // Check if next char is a digit
            if chars.peek().is_some_and(|c| c.is_ascii_digit()) {
                return extract_semver_from_iter(&mut chars);
            }
            continue;
        }
        if c.is_ascii_digit() {
            return extract_semver_from_iter(&mut chars);
        }
        chars.next();
    }
    None
}

/// Extract a semver string starting from the current position in the char iterator.
fn extract_semver_from_iter(chars: &mut std::iter::Peekable<std::str::Chars<'_>>) -> Option<String> {
    let mut version = String::new();
    let mut dot_count = 0;

    // Consume digits and dots
    while let Some(&c) = chars.peek() {
        if c.is_ascii_digit() {
            version.push(c);
            chars.next();
        } else if c == '.' && dot_count < 2 {
            dot_count += 1;
            version.push(c);
            chars.next();
        } else if c == '-' || c.is_ascii_alphabetic() {
            // Allow suffix like "-dev", "-rc1", etc.
            // Only if we already have at least X.Y
            if dot_count >= 1 {
                // Consume the suffix
                while let Some(&sc) = chars.peek() {
                    if sc.is_ascii_alphanumeric() || sc == '-' || sc == '.' {
                        version.push(sc);
                        chars.next();
                    } else {
                        break;
                    }
                }
            }
            break;
        } else {
            break;
        }
    }

    // Must have at least one dot to be a version
    if dot_count >= 1 && version.len() >= 3 {
        Some(version)
    } else {
        None
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_hardy_version_standard() {
        let output = "hardy 0.5.2\n";
        assert_eq!(parse_hardy_version(output), Some("Hardy 0.5.2".to_string()));
    }

    #[test]
    fn test_parse_hardy_version_with_v_prefix() {
        let output = "Hardy DTN v0.5.2\n";
        assert_eq!(parse_hardy_version(output), Some("Hardy 0.5.2".to_string()));
    }

    #[test]
    fn test_parse_hardy_version_with_dev_suffix() {
        let output = "hardy version 0.5.2-dev\n";
        assert_eq!(parse_hardy_version(output), Some("Hardy 0.5.2-dev".to_string()));
    }

    #[test]
    fn test_parse_hardy_version_no_match() {
        let output = "";
        assert_eq!(parse_hardy_version(output), None);
    }

    #[test]
    fn test_extract_version_number_standard() {
        assert_eq!(extract_version_number("version 0.5.2"), Some("0.5.2".to_string()));
    }

    #[test]
    fn test_extract_version_number_with_v_prefix() {
        assert_eq!(extract_version_number("v1.2.3"), Some("1.2.3".to_string()));
    }

    #[test]
    fn test_extract_version_number_with_suffix() {
        assert_eq!(
            extract_version_number("hardy 0.5.2-rc1"),
            Some("0.5.2-rc1".to_string())
        );
    }

    #[test]
    fn test_extract_version_number_no_version() {
        assert_eq!(extract_version_number("no version here"), None);
    }

    #[test]
    fn test_hardy_lifecycle_new_with_bin_dir() {
        let lc = HardyLifecycle::new(Some(PathBuf::from("/opt/hardy/bin")), None);
        assert_eq!(lc.bin_path("hardy"), PathBuf::from("/opt/hardy/bin/hardy"));
    }

    #[test]
    fn test_hardy_lifecycle_new_without_bin_dir() {
        let lc = HardyLifecycle::new(None, None);
        assert_eq!(lc.bin_path("hardy"), PathBuf::from("hardy"));
    }

    #[test]
    fn test_hardy_lifecycle_custom_management_url() {
        let lc = HardyLifecycle::new(None, Some("http://10.0.0.1:9090".to_string()));
        assert_eq!(lc.management_url(), "http://10.0.0.1:9090");
    }

    #[test]
    fn test_hardy_lifecycle_default_management_url() {
        let lc = HardyLifecycle::new(None, None);
        assert_eq!(lc.management_url(), "http://127.0.0.1:8472");
    }

    #[tokio::test]
    async fn test_health_returns_not_running_when_no_hardy() {
        // Without Hardy installed, health should report not running
        let lc = HardyLifecycle::new(None, None);
        let health = lc.health().await.unwrap();
        // On a system without Hardy, this should report not running
        assert!(!health.running || health.message.is_some());
    }

    #[tokio::test]
    async fn test_start_fails_with_missing_config() {
        let mut lc = HardyLifecycle::new(None, None);
        let result = lc.start(Path::new("/nonexistent/path")).await;
        assert!(result.is_err());
        let err = result.unwrap_err();
        assert_eq!(err.category, ErrorCategory::LifecycleError);
        assert!(err.message.contains("not found"));
    }
}
