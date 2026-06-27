//! Hardy DTN backend adapter implementation.
//!
//! Provides configuration generation, lifecycle management, hot reconfiguration,
//! and telemetry collection for the Hardy BPv7 implementation (Rust-native DTN engine).

pub mod config_gen;
pub mod hot_reconfig;
pub mod lifecycle;
pub mod telemetry;
