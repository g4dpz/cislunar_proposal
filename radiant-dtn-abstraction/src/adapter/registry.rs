//! Adapter registry for thread-safe registration and lookup of backend adapters.
//!
//! The `AdapterRegistry` manages a collection of `BackendAdapter` trait implementations,
//! allowing concurrent registration and lookup by name. It uses `Arc<RwLock<HashMap>>`
//! from tokio for safe concurrent access from multiple async tasks.

use std::collections::HashMap;
use std::sync::Arc;

use tokio::sync::RwLock;

use crate::error::{AbstractionError, ErrorCategory};

use super::traits::BackendAdapter;

/// Thread-safe registry for backend adapter implementations.
///
/// Provides concurrent registration and lookup of `BackendAdapter` instances
/// by name. Multiple readers can access the registry simultaneously, while
/// writes (registration) are exclusive.
///
/// # Example
///
/// ```ignore
/// let registry = AdapterRegistry::new();
/// registry.register("ion-dtn", Arc::new(IonAdapter::new())).await?;
/// let adapter = registry.get("ion-dtn").await?;
/// ```
pub struct AdapterRegistry {
    adapters: Arc<RwLock<HashMap<String, Arc<dyn BackendAdapter>>>>,
}

impl AdapterRegistry {
    /// Create a new empty adapter registry.
    pub fn new() -> Self {
        Self {
            adapters: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Register a backend adapter with the given name.
    ///
    /// # Errors
    ///
    /// Returns `AbstractionError` with category `ConfigurationError` if an adapter
    /// with the same name is already registered.
    pub async fn register(
        &self,
        name: &str,
        adapter: Arc<dyn BackendAdapter>,
    ) -> Result<(), AbstractionError> {
        let mut adapters = self.adapters.write().await;

        if adapters.contains_key(name) {
            return Err(AbstractionError::new(
                ErrorCategory::ConfigurationError,
                format!("Adapter already registered: {}", name),
                "register",
            )
            .with_resource(name.to_string()));
        }

        adapters.insert(name.to_string(), adapter);
        Ok(())
    }

    /// Look up a registered adapter by name.
    ///
    /// # Errors
    ///
    /// Returns `AbstractionError` with category `ConfigurationError` if no adapter
    /// is registered with the given name.
    pub async fn get(&self, name: &str) -> Result<Arc<dyn BackendAdapter>, AbstractionError> {
        let adapters = self.adapters.read().await;

        adapters.get(name).cloned().ok_or_else(|| {
            AbstractionError::new(
                ErrorCategory::ConfigurationError,
                format!("Adapter not found: {}", name),
                "get",
            )
            .with_resource(name.to_string())
        })
    }

    /// List all registered adapter names.
    ///
    /// Returns adapter names in no particular order.
    pub async fn list(&self) -> Vec<String> {
        let adapters = self.adapters.read().await;
        adapters.keys().cloned().collect()
    }
}

impl Default for AdapterRegistry {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::adapter::capability::{CapabilitySet, HotReconfigCapabilities, SecurityCapabilities};
    use crate::model::{Contact, Neighbor, NetworkConfiguration};
    use async_trait::async_trait;
    use std::path::Path;

    use super::super::traits::{
        BundleStatistics, ContactRef, GeneratedConfig, HealthStatus, LinkState, NodeRef,
    };

    /// A minimal mock adapter for testing the registry.
    struct MockAdapter {
        adapter_name: String,
    }

    impl MockAdapter {
        fn new(name: &str) -> Self {
            Self {
                adapter_name: name.to_string(),
            }
        }
    }

    #[async_trait]
    impl BackendAdapter for MockAdapter {
        fn name(&self) -> &str {
            &self.adapter_name
        }

        fn capabilities(&self) -> &CapabilitySet {
            // Return a static reference via a leaked box for testing simplicity
            Box::leak(Box::new(CapabilitySet {
                hot_reconfig: HotReconfigCapabilities {
                    add_contact: false,
                    remove_contact: false,
                    add_neighbor: false,
                    remove_neighbor: false,
                    enable_link: false,
                    disable_link: false,
                },
                convergence_layers: vec![],
                routing_strategies: vec![],
                security: SecurityCapabilities::none(),
            }))
        }

        async fn validate(&self, _config: &NetworkConfiguration) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn generate_config(
            &self,
            _config: &NetworkConfiguration,
            _output_dir: &Path,
        ) -> Result<GeneratedConfig, AbstractionError> {
            Ok(GeneratedConfig {
                files: HashMap::new(),
            })
        }

        async fn deploy(
            &self,
            _config: &NetworkConfiguration,
            _output_dir: &Path,
        ) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn start(&self, _config_dir: &Path) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn stop(&self) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn restart(&self, _config_dir: &Path) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn health(&self) -> Result<HealthStatus, AbstractionError> {
            Ok(HealthStatus {
                running: false,
                uptime_secs: None,
                message: None,
            })
        }

        async fn version(&self) -> Result<String, AbstractionError> {
            Ok("mock-1.0".to_string())
        }

        async fn add_contact(&self, _contact: &Contact) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn remove_contact(&self, _contact: &ContactRef) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn add_neighbor(&self, _neighbor: &Neighbor) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn remove_neighbor(&self, _node_ref: &NodeRef) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn enable_link(&self, _link_id: &str) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn disable_link(&self, _link_id: &str) -> Result<(), AbstractionError> {
            Ok(())
        }

        async fn collect_stats(&self) -> Result<BundleStatistics, AbstractionError> {
            Ok(BundleStatistics {
                bundles_sourced: 0,
                bundles_forwarded: 0,
                bundles_delivered: 0,
                bundles_expired: 0,
                bundles_queued: 0,
            })
        }

        async fn link_states(&self) -> Result<Vec<LinkState>, AbstractionError> {
            Ok(vec![])
        }
    }

    #[tokio::test]
    async fn test_register_and_get() {
        let registry = AdapterRegistry::new();
        let adapter = Arc::new(MockAdapter::new("test-backend"));

        registry.register("test-backend", adapter).await.unwrap();

        let retrieved = registry.get("test-backend").await.unwrap();
        assert_eq!(retrieved.name(), "test-backend");
    }

    #[tokio::test]
    async fn test_duplicate_registration_returns_error() {
        let registry = AdapterRegistry::new();
        let adapter1 = Arc::new(MockAdapter::new("ion-dtn"));
        let adapter2 = Arc::new(MockAdapter::new("ion-dtn"));

        registry.register("ion-dtn", adapter1).await.unwrap();
        let result = registry.register("ion-dtn", adapter2).await;

        assert!(result.is_err());
        let err = result.err().unwrap();
        assert_eq!(err.category, ErrorCategory::ConfigurationError);
        assert!(err.message.contains("already registered"));
    }

    #[tokio::test]
    async fn test_get_not_found_returns_error() {
        let registry = AdapterRegistry::new();

        let result = registry.get("nonexistent").await;

        assert!(result.is_err());
        let err = result.err().unwrap();
        assert_eq!(err.category, ErrorCategory::ConfigurationError);
        assert_eq!(err.message, "Adapter not found: nonexistent");
    }

    #[tokio::test]
    async fn test_list_empty_registry() {
        let registry = AdapterRegistry::new();
        let names = registry.list().await;
        assert!(names.is_empty());
    }

    #[tokio::test]
    async fn test_list_multiple_adapters() {
        let registry = AdapterRegistry::new();
        registry
            .register("ion-dtn", Arc::new(MockAdapter::new("ion-dtn")))
            .await
            .unwrap();
        registry
            .register("hardy", Arc::new(MockAdapter::new("hardy")))
            .await
            .unwrap();

        let mut names = registry.list().await;
        names.sort();
        assert_eq!(names, vec!["hardy", "ion-dtn"]);
    }
}
