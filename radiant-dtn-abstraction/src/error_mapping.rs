//! Error mapping utilities for backend adapters.
//!
//! Provides the `BackendErrorMapper` trait that adapters use to consistently
//! wrap raw backend-specific errors into `AbstractionError` with full context
//! preservation (backend_code, operation, resource, backend name).

use crate::error::{AbstractionError, ErrorCategory};

/// Trait for converting backend-specific errors into `AbstractionError`
/// with full context preservation.
///
/// Each backend adapter (ION-DTN, Hardy, etc.) implements this trait to provide
/// consistent error wrapping that preserves the raw error string as `backend_code`
/// and populates `ErrorContext` with operation and backend name.
pub trait BackendErrorMapper {
    /// The backend name this mapper represents (e.g., "ion-dtn", "hardy").
    fn backend_name(&self) -> &str;

    /// Map a raw lifecycle error (start, stop, restart, health, version failures)
    /// into an `AbstractionError` with `ErrorCategory::LifecycleError`.
    fn map_lifecycle_error(&self, raw_error: &str, operation: &str) -> AbstractionError {
        AbstractionError::new(
            ErrorCategory::LifecycleError,
            format!("{} lifecycle error: {}", self.backend_name(), raw_error),
            operation,
        )
        .with_backend(self.backend_name())
        .with_backend_code(raw_error)
    }

    /// Map a raw runtime error (hot reconfiguration failure, telemetry timeout)
    /// into an `AbstractionError` with `ErrorCategory::RuntimeError`.
    fn map_runtime_error(&self, raw_error: &str, operation: &str) -> AbstractionError {
        AbstractionError::new(
            ErrorCategory::RuntimeError,
            format!("{} runtime error: {}", self.backend_name(), raw_error),
            operation,
        )
        .with_backend(self.backend_name())
        .with_backend_code(raw_error)
    }

    /// Map a raw communication error (IPC failure, socket timeout, process unreachable)
    /// into an `AbstractionError` with `ErrorCategory::CommunicationError`.
    fn map_communication_error(&self, raw_error: &str, operation: &str) -> AbstractionError {
        AbstractionError::new(
            ErrorCategory::CommunicationError,
            format!("{} communication error: {}", self.backend_name(), raw_error),
            operation,
        )
        .with_backend(self.backend_name())
        .with_backend_code(raw_error)
    }
}

/// Error mapper for the ION-DTN backend adapter.
///
/// Maps raw ION errors (ionadmin failures, process exit codes, bplist parse errors)
/// into structured `AbstractionError` instances with `backend_code` preserving
/// the original ION error string and `ErrorContext.backend` set to "ion-dtn".
pub struct IonErrorMapper;

impl BackendErrorMapper for IonErrorMapper {
    fn backend_name(&self) -> &str {
        "ion-dtn"
    }
}

/// Error mapper for the Hardy backend adapter.
///
/// Maps raw Hardy errors (daemon startup failures, REST API errors, YAML parse errors)
/// into structured `AbstractionError` instances with `backend_code` preserving
/// the original Hardy error string and `ErrorContext.backend` set to "hardy".
pub struct HardyErrorMapper;

impl BackendErrorMapper for HardyErrorMapper {
    fn backend_name(&self) -> &str {
        "hardy"
    }
}

/// Enrich an existing `AbstractionError` with resource context.
///
/// This is a helper for the abstraction core to add resource information
/// to errors returned by backend adapters before passing them to callers.
///
/// # Example
/// ```
/// use radiant_dtn_abstraction::error::{AbstractionError, ErrorCategory};
/// use radiant_dtn_abstraction::error_mapping::enrich_error;
///
/// let err = AbstractionError::new(ErrorCategory::LifecycleError, "failed", "start");
/// let enriched = enrich_error(err, "link ltp-to-orbiter");
/// assert_eq!(enriched.context.resource, Some("link ltp-to-orbiter".to_string()));
/// ```
pub fn enrich_error(error: AbstractionError, resource: &str) -> AbstractionError {
    error.with_resource(resource)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ion_mapper_lifecycle_error() {
        let mapper = IonErrorMapper;
        let err = mapper.map_lifecycle_error("ionadmin exited with code 1", "start");

        assert_eq!(err.category, ErrorCategory::LifecycleError);
        assert_eq!(err.backend_code, Some("ionadmin exited with code 1".to_string()));
        assert_eq!(err.context.operation, "start");
        assert_eq!(err.context.backend, Some("ion-dtn".to_string()));
        assert!(err.message.contains("ion-dtn"));
        assert!(err.message.contains("ionadmin exited with code 1"));
    }

    #[test]
    fn test_ion_mapper_runtime_error() {
        let mapper = IonErrorMapper;
        let err = mapper.map_runtime_error("contact add rejected: overlap", "add_contact");

        assert_eq!(err.category, ErrorCategory::RuntimeError);
        assert_eq!(err.backend_code, Some("contact add rejected: overlap".to_string()));
        assert_eq!(err.context.operation, "add_contact");
        assert_eq!(err.context.backend, Some("ion-dtn".to_string()));
    }

    #[test]
    fn test_ion_mapper_communication_error() {
        let mapper = IonErrorMapper;
        let err = mapper.map_communication_error("connection refused", "collect_stats");

        assert_eq!(err.category, ErrorCategory::CommunicationError);
        assert_eq!(err.backend_code, Some("connection refused".to_string()));
        assert_eq!(err.context.operation, "collect_stats");
        assert_eq!(err.context.backend, Some("ion-dtn".to_string()));
    }

    #[test]
    fn test_hardy_mapper_lifecycle_error() {
        let mapper = HardyErrorMapper;
        let err = mapper.map_lifecycle_error("daemon crashed: segfault", "restart");

        assert_eq!(err.category, ErrorCategory::LifecycleError);
        assert_eq!(err.backend_code, Some("daemon crashed: segfault".to_string()));
        assert_eq!(err.context.operation, "restart");
        assert_eq!(err.context.backend, Some("hardy".to_string()));
        assert!(err.message.contains("hardy"));
    }

    #[test]
    fn test_hardy_mapper_runtime_error() {
        let mapper = HardyErrorMapper;
        let err = mapper.map_runtime_error("API returned 503", "add_neighbor");

        assert_eq!(err.category, ErrorCategory::RuntimeError);
        assert_eq!(err.backend_code, Some("API returned 503".to_string()));
        assert_eq!(err.context.operation, "add_neighbor");
        assert_eq!(err.context.backend, Some("hardy".to_string()));
    }

    #[test]
    fn test_hardy_mapper_communication_error() {
        let mapper = HardyErrorMapper;
        let err = mapper.map_communication_error("socket timeout after 5s", "link_states");

        assert_eq!(err.category, ErrorCategory::CommunicationError);
        assert_eq!(err.backend_code, Some("socket timeout after 5s".to_string()));
        assert_eq!(err.context.operation, "link_states");
        assert_eq!(err.context.backend, Some("hardy".to_string()));
    }

    #[test]
    fn test_enrich_error_adds_resource() {
        let mapper = IonErrorMapper;
        let err = mapper.map_lifecycle_error("failed to start", "start");
        let enriched = enrich_error(err, "link ltp-to-orbiter");

        assert_eq!(enriched.context.resource, Some("link ltp-to-orbiter".to_string()));
        // Other context preserved
        assert_eq!(enriched.context.operation, "start");
        assert_eq!(enriched.context.backend, Some("ion-dtn".to_string()));
        assert_eq!(enriched.backend_code, Some("failed to start".to_string()));
    }

    #[test]
    fn test_enrich_error_preserves_all_fields() {
        let err = AbstractionError::new(
            ErrorCategory::RuntimeError,
            "something broke",
            "add_contact",
        )
        .with_backend("hardy")
        .with_backend_code("ERR_503");

        let enriched = enrich_error(err, "contact node20→node30");

        assert_eq!(enriched.category, ErrorCategory::RuntimeError);
        assert_eq!(enriched.message, "something broke");
        assert_eq!(enriched.context.operation, "add_contact");
        assert_eq!(enriched.context.backend, Some("hardy".to_string()));
        assert_eq!(enriched.backend_code, Some("ERR_503".to_string()));
        assert_eq!(enriched.context.resource, Some("contact node20→node30".to_string()));
    }
}
