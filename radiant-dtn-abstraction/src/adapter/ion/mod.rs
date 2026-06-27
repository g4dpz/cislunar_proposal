//! ION-DTN backend adapter implementation.
//!
//! Provides configuration generation, lifecycle management, hot reconfiguration,
//! and telemetry collection for NASA JPL's ION-DTN (Interplanetary Overlay Network)
//! implementation.

pub mod config_gen;
pub mod hot_reconfig;
pub mod lifecycle;
pub mod telemetry;
