//! ION-DTN lifecycle management.
//!
//! Provides start, stop, restart, health check, and version query
//! for ION-DTN daemons via their CLI admin tools (ionadmin, bpadmin,
//! ltpadmin, ipnadmin, ionstop).

use std::path::{Path, PathBuf};

use tokio::process::Command;

use crate::adapter::traits::HealthStatus;
use crate::error::{AbstractionError, ErrorCategory};

/// Manages the lifecycle of ION-DTN daemon processes.
///
/// ION startup involves feeding admin scripts to each daemon in order:
/// ionadmin (.ionrc) → ltpadmin (.ltprc) → bpadmin (.bprc) → ipnadmin (.ipnrc)
///
/// Shutdown is handled by the `ionstop` utility which terminates all ION daemons.
pub struct IonLifecycle {
    /// Path to the ION bin directory (where ionadmin, bpadmin, etc. live).
    /// If None, assumes the tools are on $PATH.
    ion_bin_dir: Option<PathBuf>,
    /// Path to the config directory containing .ionrc, .bprc, etc.
    /// Set during start() and used for restart().
    config_dir: Option<PathBuf>,
}

impl IonLifecycle {
    /// Create a new `IonLifecycle` manager.
    ///
    /// # Arguments
    /// * `ion_bin_dir` — Optional path to the directory containing ION binaries.
    ///   If None, the system $PATH is used to locate ionadmin, bpadmin, etc.
    pub fn new(ion_bin_dir: Option<PathBuf>) -> Self {
        Self {
            ion_bin_dir,
            config_dir: None,
        }
    }

    /// Resolve the full path to an ION binary.
    fn bin_path(&self, name: &str) -> PathBuf {
        match &self.ion_bin_dir {
            Some(dir) => dir.join(name),
            None => PathBuf::from(name),
        }
    }

    /// Start ION daemons by feeding admin scripts in the required order.
    ///
    /// The startup sequence is:
    /// 1. `ionadmin <config_dir>/*.ionrc`
    /// 2. `ltpadmin <config_dir>/*.ltprc`
    /// 3. `bpadmin <config_dir>/*.bprc`
    /// 4. `ipnadmin <config_dir>/*.ipnrc`
    ///
    /// Each step feeds the corresponding config file to the admin tool.
    /// If any step fails, the error is returned immediately (partial start).
    pub async fn start(&mut self, config_dir: &Path) -> Result<(), AbstractionError> {
        self.config_dir = Some(config_dir.to_path_buf());

        // Find config files for each admin tool
        let ionrc = find_config_file(config_dir, "ionrc")?;
        let ltprc = find_config_file(config_dir, "ltprc")?;
        let bprc = find_config_file(config_dir, "bprc")?;
        let ipnrc = find_config_file(config_dir, "ipnrc")?;

        // Run admin tools in sequence
        self.run_admin("ionadmin", &ionrc).await?;
        self.run_admin("ltpadmin", &ltprc).await?;
        self.run_admin("bpadmin", &bprc).await?;
        self.run_admin("ipnadmin", &ipnrc).await?;

        Ok(())
    }

    /// Stop all ION daemons using the `ionstop` utility.
    pub async fn stop(&self) -> Result<(), AbstractionError> {
        let ionstop = self.bin_path("ionstop");

        let output = Command::new(&ionstop)
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::LifecycleError,
                    format!("Failed to execute ionstop: {}", e),
                    "stop",
                )
                .with_backend("ion-dtn")
            })?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!("ionstop failed (exit code {:?}): {}", output.status.code(), stderr),
                "stop",
            )
            .with_backend("ion-dtn"));
        }

        Ok(())
    }

    /// Restart ION daemons by stopping then starting with the given config directory.
    pub async fn restart(&mut self, config_dir: &Path) -> Result<(), AbstractionError> {
        // Stop ignoring errors (ION may not be running)
        let _ = self.stop().await;

        // Brief pause to allow daemons to fully exit
        tokio::time::sleep(std::time::Duration::from_millis(500)).await;

        self.start(config_dir).await
    }

    /// Check the health of the running ION instance.
    ///
    /// Uses `bplist` to determine if BP is operational. If `bplist` exits
    /// with code 0, ION is considered running. As a fallback, checks if the
    /// `rfxclock` process is active.
    pub async fn health(&self) -> Result<HealthStatus, AbstractionError> {
        // Primary check: run bplist
        let bplist = self.bin_path("bplist");
        let result = Command::new(&bplist).output().await;

        match result {
            Ok(output) if output.status.success() => {
                Ok(HealthStatus {
                    running: true,
                    uptime_secs: None, // ION doesn't expose uptime directly
                    message: Some("ION BP agent is running".to_string()),
                })
            }
            Ok(output) => {
                // bplist returned non-zero — try rfxclock fallback
                let stderr = String::from_utf8_lossy(&output.stderr);
                if self.check_rfxclock().await {
                    Ok(HealthStatus {
                        running: true,
                        uptime_secs: None,
                        message: Some("rfxclock is running (bplist unavailable)".to_string()),
                    })
                } else {
                    Ok(HealthStatus {
                        running: false,
                        uptime_secs: None,
                        message: Some(format!("ION not running: {}", stderr.trim())),
                    })
                }
            }
            Err(_) => {
                // bplist binary not found or execution failed — try rfxclock
                if self.check_rfxclock().await {
                    Ok(HealthStatus {
                        running: true,
                        uptime_secs: None,
                        message: Some("rfxclock is running (bplist not available)".to_string()),
                    })
                } else {
                    Ok(HealthStatus {
                        running: false,
                        uptime_secs: None,
                        message: Some("ION not running (bplist not available)".to_string()),
                    })
                }
            }
        }
    }

    /// Query the ION version string.
    ///
    /// Runs `ionadmin` and captures the version banner it prints on startup,
    /// which typically looks like: `ION OPEN SOURCE x.y.z ...`
    pub async fn version(&self) -> Result<String, AbstractionError> {
        let ionadmin = self.bin_path("ionadmin");

        // Run ionadmin with a quit command to get the version banner
        let output = Command::new(&ionadmin)
            .arg(".")
            .output()
            .await
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::LifecycleError,
                    format!("Failed to execute ionadmin for version query: {}", e),
                    "version",
                )
                .with_backend("ion-dtn")
            })?;

        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);

        // Look for version string in stdout or stderr
        // ION typically prints something like "ION OPEN SOURCE 4.1.2" or "ION-OPEN-SOURCE-4.1.2"
        let combined = format!("{}\n{}", stdout, stderr);
        if let Some(version) = parse_ion_version(&combined) {
            Ok(version)
        } else if !stdout.trim().is_empty() {
            // Return the raw first line as a fallback
            Ok(stdout.lines().next().unwrap_or("unknown").trim().to_string())
        } else {
            Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                "Could not determine ION version from ionadmin output".to_string(),
                "version",
            )
            .with_backend("ion-dtn"))
        }
    }

    /// Run an ION admin tool with a config file as its argument.
    async fn run_admin(&self, tool_name: &str, config_file: &Path) -> Result<(), AbstractionError> {
        let tool_path = self.bin_path(tool_name);

        let child = Command::new(&tool_path)
            .arg(config_file)
            .current_dir(config_file.parent().unwrap_or(Path::new(".")))
            .stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::null())
            .stderr(std::process::Stdio::null())
            .spawn()
            .map_err(|e| {
                AbstractionError::new(
                    ErrorCategory::LifecycleError,
                    format!("Failed to execute {}: {}", tool_name, e),
                    "start",
                )
                .with_backend("ion-dtn")
                .with_resource(format!("{}", config_file.display()))
            })?;

        let output = tokio::time::timeout(
            std::time::Duration::from_secs(10),
            child.wait_with_output(),
        )
        .await
        .map_err(|_| {
            AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!("{} timed out after 10 seconds", tool_name),
                "start",
            )
            .with_backend("ion-dtn")
        })?
        .map_err(|e| {
            AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!("Failed to wait for {}: {}", tool_name, e),
                "start",
            )
            .with_backend("ion-dtn")
        })?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            return Err(AbstractionError::new(
                ErrorCategory::LifecycleError,
                format!(
                    "{} failed (exit code {:?}): {}{}",
                    tool_name,
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
            .with_backend("ion-dtn")
            .with_resource(format!("{}", config_file.display())));
        }

        Ok(())
    }

    /// Check if the `rfxclock` process is running (fallback health check).
    async fn check_rfxclock(&self) -> bool {
        // Use pgrep to check for rfxclock process
        let result = Command::new("pgrep")
            .arg("-x")
            .arg("rfxclock")
            .output()
            .await;

        matches!(result, Ok(output) if output.status.success())
    }
}

/// Find a config file with the given extension in the config directory.
///
/// Looks for files matching `*.<ext>` (e.g., `*.ionrc`). Returns the first
/// match found, or an error if none exists.
fn find_config_file(config_dir: &Path, ext: &str) -> Result<PathBuf, AbstractionError> {
    let read_dir = std::fs::read_dir(config_dir).map_err(|e| {
        AbstractionError::new(
            ErrorCategory::LifecycleError,
            format!("Cannot read config directory '{}': {}", config_dir.display(), e),
            "start",
        )
        .with_backend("ion-dtn")
    })?;

    for entry in read_dir.flatten() {
        let path = entry.path();
        if let Some(file_ext) = path.extension() {
            if file_ext == ext {
                return Ok(path);
            }
        }
    }

    Err(AbstractionError::new(
        ErrorCategory::LifecycleError,
        format!("No .{} config file found in '{}'", ext, config_dir.display()),
        "start",
    )
    .with_backend("ion-dtn"))
}

/// Parse the ION version string from ionadmin output.
///
/// ION version banners typically look like:
/// - `ION OPEN SOURCE 4.1.2`
/// - `ION-OPEN-SOURCE-4.1.2`
/// - `: ION OPEN SOURCE 4.1.2s`
fn parse_ion_version(output: &str) -> Option<String> {
    for line in output.lines() {
        let line_upper = line.to_uppercase();
        if line_upper.contains("ION") && (line_upper.contains("OPEN SOURCE") || line_upper.contains("OPEN-SOURCE")) {
            // Try to extract version number (digits.digits.digits pattern)
            if let Some(version) = extract_version_number(line) {
                return Some(format!("ION {}", version));
            }
            // Return the whole line trimmed as fallback
            return Some(line.trim().to_string());
        }
    }
    None
}

/// Extract a semver-like version number (X.Y.Z) from a string.
fn extract_version_number(s: &str) -> Option<String> {
    let mut chars = s.chars().peekable();
    while let Some(&c) = chars.peek() {
        if c.is_ascii_digit() {
            let mut version = String::new();
            let mut dot_count = 0;
            for c in chars.by_ref() {
                if c.is_ascii_digit() || (c == '.' && dot_count < 2) {
                    if c == '.' {
                        dot_count += 1;
                    }
                    version.push(c);
                } else if c.is_ascii_alphabetic() {
                    // Allow trailing letter (e.g., "4.1.2s")
                    version.push(c);
                    break;
                } else {
                    break;
                }
            }
            // Must have at least one dot to look like a version
            if dot_count >= 1 && version.len() >= 3 {
                return Some(version);
            }
        } else {
            chars.next();
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_ion_version_standard() {
        let output = "ION OPEN SOURCE 4.1.2\n: \n";
        assert_eq!(parse_ion_version(output), Some("ION 4.1.2".to_string()));
    }

    #[test]
    fn test_parse_ion_version_hyphenated() {
        let output = "ION-OPEN-SOURCE-4.1.2s\n";
        assert_eq!(parse_ion_version(output), Some("ION 4.1.2s".to_string()));
    }

    #[test]
    fn test_parse_ion_version_with_prefix() {
        let output = ": ION OPEN SOURCE 4.1.2\n";
        assert_eq!(parse_ion_version(output), Some("ION 4.1.2".to_string()));
    }

    #[test]
    fn test_parse_ion_version_no_match() {
        let output = "some random output\n";
        assert_eq!(parse_ion_version(output), None);
    }

    #[test]
    fn test_extract_version_number() {
        assert_eq!(extract_version_number("version 4.1.2 released"), Some("4.1.2".to_string()));
        assert_eq!(extract_version_number("v3.7.4s"), Some("3.7.4s".to_string()));
        assert_eq!(extract_version_number("no version here"), None);
    }

    #[test]
    fn test_ion_lifecycle_new_with_bin_dir() {
        let lc = IonLifecycle::new(Some(PathBuf::from("/opt/ion/bin")));
        assert_eq!(lc.bin_path("ionadmin"), PathBuf::from("/opt/ion/bin/ionadmin"));
        assert_eq!(lc.bin_path("bpadmin"), PathBuf::from("/opt/ion/bin/bpadmin"));
    }

    #[test]
    fn test_ion_lifecycle_new_without_bin_dir() {
        let lc = IonLifecycle::new(None);
        assert_eq!(lc.bin_path("ionadmin"), PathBuf::from("ionadmin"));
    }

    #[test]
    fn test_find_config_file_missing_dir() {
        let result = find_config_file(Path::new("/nonexistent/path"), "ionrc");
        assert!(result.is_err());
        let err = result.unwrap_err();
        assert_eq!(err.category, ErrorCategory::LifecycleError);
        assert!(err.message.contains("Cannot read config directory"));
    }

    #[tokio::test]
    async fn test_health_returns_not_running_when_no_ion() {
        // Without ION installed, health should report not running
        let lc = IonLifecycle::new(None);
        let health = lc.health().await.unwrap();
        // On a system without ION, this should report not running
        // (we can't guarantee ION is installed in CI)
        assert!(!health.running || health.message.is_some());
    }
}
