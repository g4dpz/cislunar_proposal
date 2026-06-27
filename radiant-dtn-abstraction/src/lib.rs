//! # radiant-dtn-abstraction
//!
//! Vendor-neutral DTN configuration generation, lifecycle management, and monitoring
//! for BPv7 Delay/Disruption Tolerant Networking implementations within the RADIANT project.
//!
//! This crate provides:
//! - A canonical configuration model (YAML/JSON serializable)
//! - Trait-based backend adapters (ION-DTN, Hardy)
//! - Engine lifecycle state machine
//! - Event bus for operational notifications
//! - Configuration validation
//! - HTTP/JSON management API
//!
//! ## Amateur Radio Compliance
//!
//! This crate operates within the amateur radio service. No encryption or BPSec is
//! applied over amateur radio links. Callsign-embedded EIDs (dtn://callsign-ssid/service)
//! are first-class citizens in the configuration model.

#![allow(clippy::result_large_err)]

pub mod adapter;
pub mod api;
pub mod error;
pub mod error_mapping;
pub mod events;
pub mod lifecycle;
pub mod model;
pub mod validation;
