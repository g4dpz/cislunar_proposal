# Implementation Plan: Terrestrial DTN Phase 1

## Overview

This implementation plan builds the `radiant-terrestrial` binary — a configured instance of the RADIANT DTN abstraction layer for terrestrial amateur radio DTN validation. The system composes 6 components: Node Controller, KISS CLA, Bundle Store, Contact Plan Manager, Beacon Timer, and Telemetry Collector. All code is Rust, using `tokio` async runtime, `radiant-dtn-abstraction`, `radiant-kiss`, and `radiant-cla` crates. CRC is the sole corruption detection mechanism — no cryptography of any kind.

## Tasks

- [ ] 1. Set up project structure, dependencies, and core types
  - [ ] 1.1 Create `radiant-terrestrial` crate with Cargo.toml
    - Initialize the crate with `[[bin]]` target for `radiant-terrestrial`
    - Add dependencies: `radiant-dtn-abstraction`, `radiant-kiss`, `radiant-cla`, `tokio` (full features), `serde`, `serde_yaml`, `serde_json`, `serialport`, `tracing`, `tracing-subscriber`, `thiserror`, `chrono`, `crc`, `proptest` (dev)
    - Create `src/main.rs` entry point and module declarations
    - _Requirements: 12.1, 13.1_

  - [ ] 1.2 Define core data types and enums
    - Implement `Priority` enum (Bulk, Normal, Expedited, Critical) with `Ord` derivation
    - Implement `BundleId` struct (source_eid, creation_timestamp, sequence_number)
    - Implement `BundleRecord` struct with all metadata fields
    - Implement `BundleType` enum (Data, PingRequest, PingResponse)
    - Implement `NodeError`, `StoreError`, `ClaError`, `PlanError` error types using `thiserror`
    - _Requirements: 1.1, 1.4, 14.1, 17.1, 17.2_

  - [ ] 1.3 Define configuration structs
    - Implement `TerrestrialNodeConfig` with `NetworkConfiguration` and `NodeOperationalConfig`
    - Implement `NodeOperationalConfig` with all fields (callsign_eid, store_path, max_storage_bytes, max_bundle_size, max_bundle_rate, default_priority, cycle_interval_ms, beacon_interval_secs, beacon_text, tnc_device, tnc_baud_rate, usb_retry_interval_secs, telemetry_path, engine_restart_backoff_secs)
    - Add `serde::Serialize` and `serde::Deserialize` derives
    - Implement config loading from YAML file
    - _Requirements: 12.1, 14.4_

  - [ ] 1.4 Implement Callsign EID validation
    - Write `validate_callsign_eid()` function enforcing: 1-2 letter prefix, 1+ digits, 1-3 letter suffix, SSID 0-15
    - Return structured `CallsignError` on failure with reason
    - Validate the `dtn://callsign-ssid` URI format
    - _Requirements: 10.3, 10.4_

  - [ ]* 1.5 Write property test for Callsign EID validation
    - **Property 17: Callsign EID Validation**
    - **Validates: Requirements 10.3**

- [ ] 2. Implement Bundle Store
  - [ ] 2.1 Implement filesystem-backed Bundle Store
    - Create `BundleStore` struct implementing `BundleStoreOps` trait
    - Implement atomic writes: write to temp file → fsync → rename
    - Implement `store()`, `retrieve()`, `delete()` operations
    - Implement `list_by_destination()` and `list_by_priority()` with priority ordering (critical first)
    - Implement `capacity()` reporting
    - _Requirements: 2.1, 2.2, 2.3, 2.6_

  - [ ]* 2.2 Write property test for Bundle Store round-trip
    - **Property 3: Bundle Store Round-Trip**
    - **Validates: Requirements 2.2**

  - [ ]* 2.3 Write property test for priority ordering invariant
    - **Property 4: Priority Ordering Invariant**
    - **Validates: Requirements 2.3, 5.3, 14.2**

  - [ ] 2.4 Implement eviction and expiry logic
    - Implement `evict_expired()` — delete all bundles where creation_timestamp + lifetime ≤ current_time
    - Implement `evict_for_space()` — evict lowest-priority oldest bundles first; critical only after all others gone
    - Implement capacity enforcement on store — reject if full and cannot evict
    - _Requirements: 2.4, 2.5, 3.1, 3.2, 14.3_

  - [ ]* 2.5 Write property test for store capacity invariant
    - **Property 5: Store Capacity Invariant**
    - **Validates: Requirements 2.6**

  - [ ]* 2.6 Write property test for eviction preserves critical bundles
    - **Property 6: Eviction Preserves Critical Bundles**
    - **Validates: Requirements 2.4, 2.5, 14.3**

  - [ ]* 2.7 Write property test for expiry cleanup completeness
    - **Property 8: Expiry Cleanup Completeness**
    - **Validates: Requirements 3.1, 3.2**

  - [ ] 2.8 Implement store reload from filesystem
    - Implement `reload()` — scan store directory, parse metadata files, rebuild in-memory index
    - Validate integrity of each entry on load, discard corrupted entries with log
    - _Requirements: 2.7, 17.3_

  - [ ]* 2.9 Write property test for store reload preserves bundles
    - **Property 7: Store Reload Preserves All Bundles**
    - **Validates: Requirements 2.7, 17.3**

- [ ] 3. Checkpoint - Bundle Store complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 4. Implement KISS CLA
  - [ ] 4.1 Implement KISS frame encoding and decoding
    - Create `KissCla` struct with `KissClaConfig`
    - Implement KISS frame encoding: FEND + CMD(0x00) + byte-stuffed data + FEND
    - Implement KISS frame decoding: detect FEND boundaries, reverse byte stuffing (0xDB 0xDC → 0xC0, 0xDB 0xDD → 0xDB)
    - Implement `LinkMetrics` tracking (bytes_sent, bytes_received, frames_sent, frames_received, framing_errors)
    - _Requirements: 9.1, 9.2, 9.7_

  - [ ]* 4.2 Write property test for KISS frame round-trip
    - **Property 12: KISS Frame Round-Trip**
    - **Validates: Requirements 9.1, 9.2, 9.6**

  - [ ] 4.3 Implement ConvergenceLayerAdapter trait for KissCla
    - Implement `send_segment()` — encode LTP segment in KISS frame, write to serial
    - Implement `recv_segment()` — read from serial, decode KISS frame, extract LTP segment
    - Implement `activate()` — open USB serial connection to TNC4
    - Implement `deactivate()` — close USB serial connection
    - Implement `is_active()` status check
    - _Requirements: 9.4, 9.5_

  - [ ] 4.4 Implement USB disconnection detection and reconnection
    - Detect USB disconnection within 5 seconds via serial read/write errors
    - Set `ClaStatus::Disconnected` on detection
    - Attempt reconnection at configurable retry interval
    - Track disconnection events in link metrics
    - _Requirements: 17.4_

  - [ ]* 4.5 Write property test for LTP segmentation correctness
    - **Property 13: LTP Segmentation Correctness**
    - **Validates: Requirements 9.3**

- [ ] 5. Implement Contact Plan Manager
  - [ ] 5.1 Implement ContactPlanManager struct and core operations
    - Create `ContactPlanManager` implementing `ContactPlanOps` trait
    - Implement `ContactWindow` struct with all fields (remote_node_eid, remote_node_number, start_time, end_time, rate_bps, link_type)
    - Implement `load()` and `load_from_file()` from canonical config format (YAML/JSON)
    - Implement `active_contacts(time)` — return all windows where start_time ≤ time < end_time
    - Implement `next_contact(dest_node, after)` — earliest future window for destination
    - Implement `update()` — add/update window, reject overlaps on same link
    - Implement `persist()` and `reload()` for filesystem persistence
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

  - [ ]* 5.2 Write property test for contact plan active query correctness
    - **Property 14: Contact Plan Active Query Correctness**
    - **Validates: Requirements 7.2**

  - [ ]* 5.3 Write property test for contact plan overlap rejection
    - **Property 15: Contact Plan Overlap Rejection**
    - **Validates: Requirements 7.4**

  - [ ]* 5.4 Write property test for contact plan serialization round-trip
    - **Property 16: Contact Plan Serialization Round-Trip**
    - **Validates: Requirements 7.6, 7.7**

  - [ ]* 5.5 Write property test for direct contact lookup (no relay)
    - **Property 21: Direct Contact Lookup (No Relay)**
    - **Validates: Requirements 6.2**

- [ ] 6. Checkpoint - CLA and Contact Plan complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 7. Implement Bundle Protocol Agent logic
  - [ ] 7.1 Implement bundle creation and CRC validation
    - Implement bundle creation with BPv7 version 7, source/destination Callsign_EIDs, CRC, priority, lifetime
    - Implement BPv7 bundle CRC computation using the `crc` crate
    - Implement bundle validation: version==7, valid destination Callsign_EID, lifetime>0, creation_timestamp ≤ current time, CRC correct
    - Log specific validation failure reason with source EID on rejection
    - _Requirements: 1.1, 1.2, 1.3, 13.3_

  - [ ]* 7.2 Write property test for bundle validation correctness
    - **Property 2: Bundle Validation Correctness**
    - **Validates: Requirements 1.2, 1.3**

  - [ ] 7.3 Implement bundle serialization/deserialization
    - Implement BPv7 wire format serialization (CBOR encoding per RFC 9171)
    - Implement BPv7 wire format deserialization with CRC verification
    - Support all three bundle types (Data, PingRequest, PingResponse)
    - _Requirements: 1.5_

  - [ ]* 7.4 Write property test for bundle serialization round-trip
    - **Property 1: Bundle Serialization Round-Trip**
    - **Validates: Requirements 1.5**

  - [ ] 7.5 Implement ping request/response handling
    - On receiving a ping request addressed to local endpoint, generate exactly one ping response
    - Set ping response destination to original sender's Callsign_EID
    - Include original request's BundleId in response payload
    - Queue response in Bundle Store for delivery
    - _Requirements: 4.1, 4.2, 4.4_

  - [ ]* 7.6 Write property test for ping response correctness
    - **Property 9: Ping Response Correctness**
    - **Validates: Requirements 4.1, 4.4**

  - [ ] 7.7 Implement routing logic (local delivery vs store-for-forwarding)
    - If destination matches local EID → deliver locally
    - If destination is remote → store in Bundle Store for direct delivery
    - Never forward to non-final-destination node (no relay)
    - _Requirements: 5.1, 5.2, 6.1_

  - [ ]* 7.8 Write property test for routing correctness
    - **Property 10: Routing Correctness (Local vs Remote)**
    - **Validates: Requirements 5.1, 5.2, 6.1**

  - [ ] 7.9 Implement ACK-driven store management
    - On LTP acknowledgment → delete bundle from store
    - On LTP timeout (no ACK) → retain bundle for retry in next contact window
    - _Requirements: 5.4, 5.5_

  - [ ]* 7.10 Write property test for ACK-driven store management
    - **Property 11: ACK-Driven Store Management**
    - **Validates: Requirements 5.4, 5.5**

  - [ ]* 7.11 Write property test for source Callsign EID presence
    - **Property 18: Source Callsign EID Presence**
    - **Validates: Requirements 10.1, 10.2**

- [ ] 8. Implement Rate Limiter and Bundle Size Enforcement
  - [ ] 8.1 Implement sliding-window rate limiter
    - Create `RateLimiter` struct with per-source-EID sliding window
    - Implement `check()` — accept if within configured max rate, reject with `RateLimitError` if exceeded
    - Log rate-limit events with source EID
    - _Requirements: 15.1, 15.2_

  - [ ]* 8.2 Write property test for rate limiter enforcement
    - **Property 19: Rate Limiter Enforcement**
    - **Validates: Requirements 15.1, 15.2**

  - [ ] 8.3 Implement maximum bundle size enforcement
    - Reject bundles exceeding configured max_bundle_size
    - Accept bundles at or below the limit (assuming other validation passes)
    - _Requirements: 15.3_

  - [ ]* 8.4 Write property test for maximum bundle size enforcement
    - **Property 20: Maximum Bundle Size Enforcement**
    - **Validates: Requirements 15.3**

- [ ] 9. Checkpoint - BPA and rate limiting complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 10. Implement Beacon Timer
  - [ ] 10.1 Implement BeaconTimer struct and logic
    - Create `BeaconTimer` with configurable interval (default 600s)
    - Implement `is_due(current_time)` — true if interval elapsed since last beacon
    - Implement `record_beacon(timestamp)` — update last beacon time, increment count
    - Initial beacon fires within 30 seconds of startup
    - _Requirements: 11.1, 11.4_

  - [ ] 10.2 Implement beacon bundle creation
    - Source EID: node's Callsign_EID
    - Destination EID: `dtn://beacon`
    - Lifetime: 600 seconds
    - Payload: human-readable identification text from config (e.g., "G4DPZ amateur radio DTN experimental station")
    - Include CRC for error detection
    - Log each beacon transmission with timestamp
    - _Requirements: 11.2, 11.3, 11.5_

  - [ ]* 10.3 Write property test for beacon timing regularity
    - **Property 22: Beacon Timing Regularity**
    - **Validates: Requirements 11.1**

- [ ] 11. Implement Telemetry Collector
  - [ ] 11.1 Implement TelemetryCollector struct
    - Create `TelemetryCollector` with `NodeHealth` and `NodeStatistics` snapshots
    - Implement `update_from_engine()` — merge DTN engine stats and link states
    - Implement `record_contact_completed()`, `record_contact_missed()`, `record_beacon()`
    - Implement `snapshot_health()` and `snapshot_stats()`
    - _Requirements: 16.1, 16.2, 16.5_

  - [ ] 11.2 Implement telemetry exposure via local interface
    - Write telemetry snapshots to configured path (file or unix socket)
    - Respond to telemetry queries within 1 second
    - _Requirements: 16.3, 16.4_

- [ ] 12. Implement Node Controller orchestration loop
  - [ ] 12.1 Implement Node Controller initialization
    - Load config from YAML
    - Validate Callsign_EID — reject startup if invalid
    - Initialize all components: Bundle Store (with reload), Contact Plan Manager, KISS CLA, Beacon Timer, Telemetry Collector
    - Configure DTN engine through abstraction layer
    - _Requirements: 10.4, 12.1, 12.2, 12.3, 17.3_

  - [ ] 12.2 Implement main async operation cycle
    - Check active contacts via Contact Plan Manager
    - Activate CLA if contact window active
    - Transmit queued bundles in priority order during active contact
    - Process received bundles: validate CRC, route (local delivery or store)
    - Handle ping requests → generate responses
    - Run bundle lifetime expiry cleanup
    - Check beacon timer → transmit if due
    - Collect telemetry
    - Deactivate CLA when contact window closes
    - Target cycle time: 100ms
    - _Requirements: 8.1, 8.2, 8.3, 14.2, 18.1_

  - [ ] 12.3 Implement contact window execution logic
    - On contact start: activate CLA, begin transmission
    - Transmit bundles in strict priority order (critical → expedited → normal → bulk)
    - On contact end: cease transmission, record link metrics, update telemetry
    - On CLA failure during contact: mark contact missed, retain bundles, increment counter
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [ ] 12.4 Implement error recovery and fault tolerance
    - USB disconnection: detect, mark contact interrupted, retry at interval
    - DTN engine crash: detect via abstraction layer health check, restart with backoff
    - Store corruption on reload: discard corrupted entries, log, continue
    - No direct contact window: retain bundle until window added or lifetime expires
    - _Requirements: 17.3, 17.4, 17.5, 17.6_

  - [ ] 12.5 Implement graceful shutdown
    - Flush pending store operations
    - Deactivate CLA
    - Stop DTN engine via abstraction layer
    - Persist contact plan state
    - _Requirements: 12.3_

- [ ] 13. Implement main binary entry point
  - [ ] 13.1 Wire main.rs with CLI argument parsing and startup
    - Parse config file path from CLI args
    - Load and validate configuration
    - Set up tracing/logging subscriber
    - Initialize NodeController
    - Set up shutdown signal handler (SIGTERM, SIGINT)
    - Call `node_controller.run(shutdown_rx).await`
    - _Requirements: 10.4, 12.1_

- [ ] 14. Final checkpoint - Full integration
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties (22 total)
- Unit tests validate specific examples and edge cases
- The `crc` crate handles BPv7 bundle CRC computation — no cryptographic crates needed
- No BPSec, HMAC, encryption, or digital signatures anywhere in the implementation
- All transmitted data remains fully inspectable per amateur radio regulations

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.3"] },
    { "id": 2, "tasks": ["1.4", "2.1", "4.1", "5.1"] },
    { "id": 3, "tasks": ["1.5", "2.2", "2.3", "2.4", "4.2", "4.3", "5.2", "5.3", "5.4", "5.5"] },
    { "id": 4, "tasks": ["2.5", "2.6", "2.7", "2.8", "4.4", "4.5"] },
    { "id": 5, "tasks": ["2.9", "7.1", "7.3"] },
    { "id": 6, "tasks": ["7.2", "7.4", "7.5", "7.7", "7.9", "8.1", "8.3"] },
    { "id": 7, "tasks": ["7.6", "7.8", "7.10", "7.11", "8.2", "8.4", "10.1"] },
    { "id": 8, "tasks": ["10.2", "10.3", "11.1"] },
    { "id": 9, "tasks": ["11.2", "12.1"] },
    { "id": 10, "tasks": ["12.2", "12.3", "12.4", "12.5"] },
    { "id": 11, "tasks": ["13.1"] }
  ]
}
```
