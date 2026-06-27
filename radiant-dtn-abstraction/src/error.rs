//! Common error types for the DTN Abstraction Layer.
//!
//! Defines `AbstractionError` with categories, context, and validation details
//! for consistent error reporting across all backends.

use std::fmt;

/// The primary error type for the DTN Abstraction Layer.
///
/// Contains a categorized error with contextual information about
/// which operation, resource, and backend produced the error.
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct AbstractionError {
    /// High-level classification of the error.
    pub category: ErrorCategory,
    /// Human-readable error message.
    pub message: String,
    /// Optional backend-specific error code or raw error string.
    pub backend_code: Option<String>,
    /// Contextual information about where the error occurred.
    pub context: ErrorContext,
}

/// Classification of errors produced by the abstraction layer.
#[derive(Debug, Clone, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum ErrorCategory {
    /// Invalid configuration structure, missing fields, referential integrity violations.
    ValidationError,
    /// Duplicate adapter registration, conflicting node numbers, capability mismatch.
    ConfigurationError,
    /// Engine failed to start, unexpected process exit, restart failure.
    LifecycleError,
    /// Hot reconfiguration command rejected, telemetry collection timeout.
    RuntimeError,
    /// Backend doesn't support the requested operation or convergence layer type.
    UnsupportedOperation,
    /// Cannot reach DTN engine process, IPC failure, socket timeout.
    CommunicationError,
}

/// Contextual information about where an error occurred.
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ErrorContext {
    /// The operation being performed when the error occurred (e.g., "generate_config", "start").
    pub operation: String,
    /// The resource affected, if applicable (e.g., "neighbor node 20", "link ltp-to-orbiter").
    pub resource: Option<String>,
    /// The backend adapter name, if applicable (e.g., "ion-dtn", "hardy").
    pub backend: Option<String>,
}

/// Structured detail for a single validation failure.
///
/// Used when the validation subsystem collects multiple errors (fail-slow validation).
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ValidationDetail {
    /// JSON-path-like location of the invalid value (e.g., "neighbors[0].links[1].remote_port").
    pub path: String,
    /// The validation rule that was violated (e.g., "referential_integrity", "temporal_order").
    pub rule: String,
    /// Human-readable description of the violation (e.g., "Neighbor references undefined node 99").
    pub message: String,
}

impl fmt::Display for ErrorCategory {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ErrorCategory::ValidationError => write!(f, "ValidationError"),
            ErrorCategory::ConfigurationError => write!(f, "ConfigurationError"),
            ErrorCategory::LifecycleError => write!(f, "LifecycleError"),
            ErrorCategory::RuntimeError => write!(f, "RuntimeError"),
            ErrorCategory::UnsupportedOperation => write!(f, "UnsupportedOperation"),
            ErrorCategory::CommunicationError => write!(f, "CommunicationError"),
        }
    }
}

impl fmt::Display for AbstractionError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "[{}] {}", self.category, self.message)?;

        // Append context information
        write!(f, " (operation: {}", self.context.operation)?;
        if let Some(ref resource) = self.context.resource {
            write!(f, ", resource: {}", resource)?;
        }
        if let Some(ref backend) = self.context.backend {
            write!(f, ", backend: {}", backend)?;
        }
        write!(f, ")")?;

        if let Some(ref code) = self.backend_code {
            write!(f, " [backend_code: {}]", code)?;
        }

        Ok(())
    }
}

impl std::error::Error for AbstractionError {}

impl AbstractionError {
    /// Create a new `AbstractionError` with the given category and message.
    pub fn new(category: ErrorCategory, message: impl Into<String>, operation: impl Into<String>) -> Self {
        Self {
            category,
            message: message.into(),
            backend_code: None,
            context: ErrorContext {
                operation: operation.into(),
                resource: None,
                backend: None,
            },
        }
    }

    /// Set the resource context on this error.
    pub fn with_resource(mut self, resource: impl Into<String>) -> Self {
        self.context.resource = Some(resource.into());
        self
    }

    /// Set the backend context on this error.
    pub fn with_backend(mut self, backend: impl Into<String>) -> Self {
        self.context.backend = Some(backend.into());
        self
    }

    /// Set the backend-specific error code.
    pub fn with_backend_code(mut self, code: impl Into<String>) -> Self {
        self.backend_code = Some(code.into());
        self
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_display_minimal() {
        let err = AbstractionError::new(
            ErrorCategory::ValidationError,
            "Missing required field",
            "validate",
        );
        let display = format!("{}", err);
        assert_eq!(display, "[ValidationError] Missing required field (operation: validate)");
    }

    #[test]
    fn test_display_full_context() {
        let err = AbstractionError::new(
            ErrorCategory::LifecycleError,
            "Engine failed to start",
            "start",
        )
        .with_resource("ion-dtn process")
        .with_backend("ion-dtn")
        .with_backend_code("exit_code:1");

        let display = format!("{}", err);
        assert_eq!(
            display,
            "[LifecycleError] Engine failed to start (operation: start, resource: ion-dtn process, backend: ion-dtn) [backend_code: exit_code:1]"
        );
    }

    #[test]
    fn test_error_trait() {
        let err = AbstractionError::new(
            ErrorCategory::RuntimeError,
            "Command rejected",
            "add_contact",
        );
        // Verify it implements std::error::Error
        let _: &dyn std::error::Error = &err;
    }

    #[test]
    fn test_error_category_display() {
        assert_eq!(format!("{}", ErrorCategory::ValidationError), "ValidationError");
        assert_eq!(format!("{}", ErrorCategory::ConfigurationError), "ConfigurationError");
        assert_eq!(format!("{}", ErrorCategory::LifecycleError), "LifecycleError");
        assert_eq!(format!("{}", ErrorCategory::RuntimeError), "RuntimeError");
        assert_eq!(format!("{}", ErrorCategory::UnsupportedOperation), "UnsupportedOperation");
        assert_eq!(format!("{}", ErrorCategory::CommunicationError), "CommunicationError");
    }

    #[test]
    fn test_validation_detail() {
        let detail = ValidationDetail {
            path: "neighbors[0].links[1].remote_port".to_string(),
            rule: "referential_integrity".to_string(),
            message: "Neighbor references undefined node 99".to_string(),
        };
        assert_eq!(detail.path, "neighbors[0].links[1].remote_port");
        assert_eq!(detail.rule, "referential_integrity");
        assert_eq!(detail.message, "Neighbor references undefined node 99");
    }

    #[test]
    fn test_serialization_roundtrip() {
        let err = AbstractionError::new(
            ErrorCategory::CommunicationError,
            "Socket timeout",
            "collect_stats",
        )
        .with_resource("link ltp-to-orbiter")
        .with_backend("ion-dtn")
        .with_backend_code("ETIMEDOUT");

        let json = serde_json::to_string(&err).unwrap();
        let deserialized: AbstractionError = serde_json::from_str(&json).unwrap();

        assert_eq!(deserialized.category, err.category);
        assert_eq!(deserialized.message, err.message);
        assert_eq!(deserialized.backend_code, err.backend_code);
        assert_eq!(deserialized.context.operation, err.context.operation);
        assert_eq!(deserialized.context.resource, err.context.resource);
        assert_eq!(deserialized.context.backend, err.context.backend);
    }
}
