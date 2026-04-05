# Implementation Plan: CubeSat Engineering Model (Phase 2)

## Overview

This plan implements the Phase 2 CubeSat EM system in two parallel tracks: C firmware on the STM32U585 (ION-DTN BPv7/LTP, AX.25 CLA, IQ baseband DSP, NVM bundle store, TrustZone crypto, power management, static pool allocator) and Go orchestration on the Companion Host (Node Controller, Contact Plan Manager, IQ Bridge, telemetry, test orchestration). Tasks are ordered so each builds on the previous, with property-based tests (theft for C, rapid for Go) placed close to the code they validate.

## Tasks

- [ ] 1. Half-Duplex AX.25 Transfer Validation via IQ Baseband (Pre-ION-DTN)
  - [ ] 1.1 Implement basic AX.25 frame construction and parsing (C — STM32U585)
    - Implement AX.25 UI frame builder with source/destination callsigns and SSID
    - Implement AX.25 frame parser to extract callsigns and information field
    - Validate callsign encoding/decoding (bit-shifted format per AX.25 spec)
    - Pool-allocated buffers from POOL_FRAME_BUFFER
    - _Requirements: 8.1_

  - [ ] 1.2 Implement IQ baseband modulation/demodulation for AX.25 frames (C — STM32U585)
    - Implement GFSK/G3RUH modulation: AX.25 frame bytes → IQ baseband samples at 9.6 kbps
    - Implement GFSK/G3RUH demodulation: IQ baseband samples → AX.25 frame bytes
    - Configure DMA double-buffered streaming for IQ samples
    - IQ sample buffers from POOL_IQ_BUFFER within 786 KB SRAM budget
    - _Requirements: 7.1, 7.2, 7.3, 7.6_

  - [ ] 1.3 Implement IQ Bridge and B200mini integration (Go — Companion Host)
    - Use UHD library to control B200mini: configure UHF 437 MHz, sample rate, TX/RX gain
    - Bridge IQ samples between B200mini (USB 3.0) and STM32U585 (SPI/UART DMA)
    - Transparent bridging — no sample modification
    - _Requirements: 7.4, 7.5_

  - [ ] 1.4 Implement half-duplex AX.25 send/receive test harness
    - Build a test that sends an AX.25 UI frame from the STM32U585 EM through the IQ baseband path (STM32U585 → IQ Bridge → B200mini → over-the-air UHF 437 MHz)
    - Build a test that receives an AX.25 UI frame at the EM from a ground station (over-the-air → B200mini → IQ Bridge → STM32U585)
    - Demonstrate half-duplex operation: EM transmits, ground station receives, then roles swap
    - Verify frames are received intact with correct source/destination callsigns
    - _Requirements: 7.7, 8.1, 8.4, 8.5_

  - [ ] 1.5 Write AX.25 frame round-trip test (EM ↔ ground station)
    - Send a frame from ground station → B200mini → STM32U585 (receive and verify)
    - Send a frame from STM32U585 → B200mini → ground station (receive and verify)
    - Confirm half-duplex AX.25 link over IQ baseband is operational before proceeding to ION-DTN integration
    - _Requirements: 7.7, 8.5_

- [ ] 2. Checkpoint — Half-duplex AX.25 over IQ baseband validated
  - Ensure AX.25 frames can be sent and received between the STM32U585 EM and a ground station over the B200mini IQ baseband path at 9.6 kbps UHF before proceeding to ION-DTN integration.

- [ ] 3. Static Memory Pool Allocator (C — STM32U585)
  - [ ] 1.1 Implement pool allocator core (`pool_init`, `pool_alloc`, `pool_free`, `pool_stats`, `pool_total_used_bytes`, `pool_peak_used_bytes`)
    - Define `pool_id_t` enum (POOL_BUNDLE_PAYLOAD, POOL_IQ_BUFFER, POOL_FRAME_BUFFER, POOL_INDEX_ENTRY, POOL_GENERAL)
    - Implement free-list per pool with O(1) alloc/free from statically allocated SRAM regions
    - Track peak and current usage per pool
    - Return NULL on exhaustion — no undefined behavior
    - _Requirements: 14.2, 14.3, 14.4_

  - [ ]* 1.2 Write unit tests for pool allocator
    - Allocate until exhaustion, verify NULL returned
    - Free and re-allocate, verify blocks reused
    - Peak tracking across alloc/free sequences
    - Multi-pool isolation (allocating from one pool does not affect another)
    - _Requirements: 14.2_

- [ ] 4. TrustZone Secure Crypto Service (C — STM32U585)
  - [ ] 2.1 Implement TrustZone secure/non-secure partition and secure API
    - Configure SAU/IDAU for secure/non-secure memory regions
    - Implement `secure_hmac_sign`, `secure_hmac_verify`, `secure_provision_key`, `secure_get_key_count` as Non-Secure Callable (NSC) functions
    - Use STM32U585 HASH peripheral (hardware crypto accelerator) for HMAC-SHA-256
    - Store keys in secure flash, never exposed to non-secure world
    - Log SecureFault on unauthorized access attempts
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 11.3, 11.5_

  - [ ]* 2.2 Write unit tests for TrustZone secure API
    - HMAC sign/verify with known NIST test vectors
    - Key provisioning and retrieval by key_id
    - Rejection of invalid key_id
    - _Requirements: 12.2_


- [ ] 5. Bundle Protocol Agent — Core (C — STM32U585)
  - [ ] 3.1 Implement BPA data types and bundle creation (`bpa_create_bundle`, `bpa_create_ping`)
    - Define `endpoint_id_t`, `bundle_id_t`, `priority_t`, `bundle_type_t`, `bundle_t`, `bpa_error_t`
    - Implement bundle creation for data, ping request, and ping response types
    - All buffers allocated from pool allocator (POOL_BUNDLE_PAYLOAD) — no malloc
    - Set version=7, CRC-32, source/destination EIDs, priority, lifetime
    - Support default priority from configuration
    - _Requirements: 1.1, 1.4, 1.6, 16.1, 16.4_

  - [ ] 3.2 Implement bundle validation (`bpa_validate_bundle`)
    - Validate version==7, well-formed destination EID, lifetime>0, creation_timestamp<=current_time, CRC correct
    - Return specific `bpa_error_t` codes for each failure
    - Discard invalid bundles and log failure reason with source EID
    - _Requirements: 1.2, 1.3_

  - [ ] 3.3 Implement BPv7 CBOR serialization/deserialization (`bpa_serialize`, `bpa_deserialize`)
    - Serialize bundle to BPv7 CBOR wire format into caller-provided buffer
    - Deserialize CBOR wire format into `bundle_t`, allocating payload from pool
    - Include primary block, payload block, and optional BPSec BIB
    - _Requirements: 1.5_

  - [ ]* 3.4 Write property test: Bundle Creation Correctness (theft)
    - **Property 1: Bundle Creation Correctness**
    - Generate random valid endpoints, payloads (within pool block size), priorities, lifetimes
    - Create bundle via `bpa_create_bundle`. Verify version==7, EIDs set, CRC valid, priority matches, lifetime matches
    - **Validates: Requirements 1.1**

  - [ ]* 3.5 Write property test: Bundle Validation Correctness (theft)
    - **Property 2: Bundle Validation Correctness**
    - Generate random bundles with random field mutations (bad version, empty EID, zero lifetime, future timestamp, bad CRC)
    - Verify validator accepts iff all fields valid, rejects with correct error code otherwise
    - **Validates: Requirements 1.2, 1.3, 19.2**

  - [ ]* 3.6 Write property test: Bundle Serialization Round-Trip (theft)
    - **Property 3: Bundle Serialization Round-Trip**
    - Generate random valid bundles. Serialize to CBOR. Deserialize. Assert equality
    - **Validates: Requirements 1.5**

  - [ ] 3.7 Implement ping echo handling (`bpa_generate_ping_response`)
    - Generate exactly one ping response per ping request
    - Set destination to original sender's EID
    - Include original request's bundle ID in response payload
    - Queue response in Bundle Store for delivery
    - _Requirements: 4.1, 4.2, 4.4_

  - [ ]* 3.8 Write property test: Ping Echo Correctness (theft)
    - **Property 9: Ping Echo Correctness**
    - Generate random ping requests with random source EIDs. Process via `bpa_generate_ping_response`
    - Verify exactly one response with correct destination and request bundle ID in payload
    - **Validates: Requirements 4.1, 4.2, 4.4**

  - [ ] 3.9 Implement BPSec integrity (`bpa_apply_integrity`, `bpa_verify_integrity`)
    - Apply BPSec BIB (HMAC-SHA-256) via TrustZone secure API call (`secure_hmac_sign`)
    - Verify BIB via TrustZone secure API call (`secure_hmac_verify`)
    - No BCB/encryption — amateur radio compliance
    - Discard bundles that fail integrity verification, log with source EID
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [ ]* 3.10 Write property test: BPSec Integrity Round-Trip (theft)
    - **Property 20: BPSec Integrity Round-Trip**
    - Generate random bundles and keys. Apply integrity via TrustZone mock. Verify passes. Mutate bundle. Verify fails
    - **Validates: Requirements 11.1, 11.4**

  - [ ]* 3.11 Write property test: No Encryption Constraint (theft)
    - **Property 21: No Encryption Constraint**
    - Generate random bundles. Process through BPA. Verify no BCB blocks present in output
    - **Validates: Requirements 11.2**

  - [ ] 3.12 Implement rate limiting and bundle size enforcement
    - Implement sliding-window rate limiter per source EID (`rate_limiter_entry_t`, `rate_limiter_config_t`)
    - Enforce configurable max bundles/sec per source EID
    - Enforce configurable max bundle size in bytes
    - Reject with `BPA_ERR_RATE_LIMITED` or `BPA_ERR_OVERSIZED`, log events
    - _Requirements: 17.1, 17.2, 17.3_

  - [ ]* 3.13 Write property test: Rate Limiting (theft)
    - **Property 22: Rate Limiting**
    - Generate random submission sequences at various rates from random source EIDs
    - Verify correct acceptance/rejection based on configured rate limit
    - **Validates: Requirements 17.1, 17.2**

  - [ ]* 3.14 Write property test: Bundle Size Limit (theft)
    - **Property 23: Bundle Size Limit**
    - Generate random bundles of varying sizes. Verify oversized rejected, within-limit accepted
    - **Validates: Requirements 17.3**

  - [ ] 3.15 Implement local vs remote delivery routing
    - If destination matches local EID, deliver to local application agent
    - If destination is remote, store in Bundle Store and queue for direct delivery
    - Enforce no-relay constraint: transmit only to final destination node
    - _Requirements: 5.1, 5.2, 6.1, 6.2_

  - [ ]* 3.16 Write property test: Local vs Remote Delivery Routing (theft)
    - **Property 10: Local vs Remote Delivery Routing**
    - Generate random bundles with destinations matching and not matching local EIDs
    - Verify correct routing (local delivery vs store-and-forward)
    - **Validates: Requirements 5.1, 5.2**

  - [ ]* 3.17 Write property test: No Relay — Direct Delivery Only (theft)
    - **Property 12: No Relay — Direct Delivery Only**
    - Generate random bundles and contacts. Verify bundles only transmitted to contacts matching their destination
    - **Validates: Requirements 6.1, 6.2**

  - [ ] 3.18 Implement `bpa_release_bundle` for pool memory cleanup
    - Free pool-allocated payload and raw_cbor buffers
    - _Requirements: 14.2_

- [ ] 6. Checkpoint — Pool allocator, TrustZone, and BPA core
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 7. NVM Bundle Store (C — STM32U585)
  - [ ] 5.1 Implement NVM storage layer (`store_init`, `store_put`, `store_get`, `store_delete`, `store_flush`)
    - Define `nvm_header_t`, `nvm_bundle_entry_t` NVM layout structures
    - Implement atomic writes: write to temp sector → CRC-32 validate → commit
    - Maintain in-SRAM priority-ordered metadata index (~64 KB) from POOL_INDEX_ENTRY
    - `store_put`: persist bundle to NVM atomically, update SRAM index
    - `store_get`: retrieve by bundle_id_t, allocate payload from pool
    - `store_delete`: remove from NVM and SRAM index
    - _Requirements: 2.1, 2.2, 2.3, 2.6_

  - [ ]* 5.2 Write property test: Bundle Store/Retrieve Round-Trip (theft)
    - **Property 4: Bundle Store/Retrieve Round-Trip**
    - Generate random valid bundles. Store to mock NVM. Retrieve by ID. Assert equality
    - **Validates: Requirements 2.2**

  - [ ] 5.3 Implement priority-ordered listing and destination queries (`store_list_by_priority`, `store_list_by_destination`)
    - List bundles in priority order: critical > expedited > normal > bulk
    - Filter by destination EID for contact-window transmission
    - _Requirements: 2.3, 5.3, 16.2_

  - [ ]* 5.4 Write property test: Priority Ordering Invariant (theft)
    - **Property 5: Priority Ordering Invariant**
    - Generate random bundle sets with random priorities. Store. List by priority. Verify non-increasing priority sequence
    - **Validates: Requirements 2.3, 5.3, 16.2**

  - [ ] 5.5 Implement capacity enforcement and eviction (`store_capacity`, `store_evict_expired`, `store_evict_lowest`)
    - Enforce total stored bytes ≤ configured max NVM capacity
    - Eviction order: expired first, then bulk → normal → expedited; critical last
    - Within same priority, evict earliest creation timestamp first
    - _Requirements: 2.4, 2.5, 2.6, 16.3_

  - [ ]* 5.6 Write property test: Eviction Policy Ordering (theft)
    - **Property 6: Eviction Policy Ordering**
    - Generate random stores at capacity with mixed priorities and lifetimes. Trigger eviction. Verify expired first, then ascending priority, oldest first within same priority, critical last
    - **Validates: Requirements 2.4, 2.5, 16.3**

  - [ ]* 5.7 Write property test: Store Capacity Bound (theft)
    - **Property 7: Store Capacity Bound**
    - Generate random store/delete operation sequences. Verify total bytes never exceeds max after each operation
    - **Validates: Requirements 2.6**

  - [ ] 5.8 Implement lifetime cleanup (`store_evict_expired`)
    - Delete all bundles whose creation_timestamp + lifetime ≤ current_time
    - Verify zero expired bundles remain after cleanup
    - _Requirements: 3.1, 3.2_

  - [ ]* 5.9 Write property test: Bundle Lifetime Enforcement (theft)
    - **Property 8: Bundle Lifetime Enforcement**
    - Generate random bundles with random lifetimes. Advance time. Run cleanup. Verify zero expired bundles remain
    - **Validates: Requirements 3.1, 3.2**

  - [ ] 5.10 Implement NVM reload and corruption recovery (`store_reload`)
    - Reload persisted state from NVM after power cycle / watchdog reset
    - Validate CRC-32 on each stored bundle
    - Discard corrupted entries, log each discarded bundle ID
    - Rebuild SRAM priority index from intact bundles
    - _Requirements: 2.7, 2.8, 19.3_

  - [ ]* 5.11 Write unit tests for NVM reload and corruption recovery
    - Simulate corrupted NVM entries, verify discarded and logged
    - Verify intact bundles recovered and index rebuilt
    - _Requirements: 2.7, 2.8_

  - [ ] 5.12 Implement ACK-based deletion and no-ACK retention
    - On LTP ACK: delete acknowledged bundle from NVM
    - On LTP timeout: retain bundle for retry during next contact window
    - _Requirements: 5.4, 5.5_

  - [ ]* 5.13 Write property test: ACK Deletes, No-ACK Retains (theft)
    - **Property 11: ACK Deletes, No-ACK Retains**
    - Generate random transmission scenarios with random ACK outcomes. Verify ACKed bundles deleted, unACKed retained
    - **Validates: Requirements 5.4, 5.5**

  - [ ]* 5.14 Write property test: Bundle Retention When No Contact Available (theft)
    - **Property 25: Bundle Retention When No Contact Available**
    - Generate bundles with no matching contacts. Verify retention until contact added or lifetime expires
    - **Validates: Requirements 19.6**

- [ ] 8. Checkpoint — NVM Bundle Store complete
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 9. IQ Baseband DSP (C — STM32U585)
  - [ ] 7.1 Implement DSP engine core (`dsp_init`, `dsp_modulate_frame`, `dsp_demodulate`)
    - Define `iq_sample_t`, `dsp_config_t` data types
    - Implement GFSK/G3RUH modulation: AX.25 frame bytes → IQ baseband samples at 9.6 kbps
    - Implement GFSK/G3RUH demodulation: IQ baseband samples → AX.25 frame bytes
    - Carrier/clock recovery and bit synchronization in demodulator
    - IQ sample buffers allocated from POOL_IQ_BUFFER within 786 KB SRAM budget
    - _Requirements: 7.1, 7.2, 7.5, 7.6_

  - [ ] 7.2 Implement DMA double-buffered streaming (`dsp_start_streaming`, `dsp_stop_streaming`, DMA ISR callbacks)
    - Configure STM32U585 DMA channels for TX and RX IQ sample streaming
    - Implement ping-pong double-buffering via `dsp_tx_half_complete_callback`, `dsp_tx_complete_callback`, `dsp_rx_half_complete_callback`, `dsp_rx_complete_callback`
    - ISR-driven — no CPU polling or CPU-bound sample transfers
    - `dsp_get_memory_usage` for telemetry reporting
    - _Requirements: 7.3, 7.6_

  - [ ]* 7.3 Write unit tests for IQ DSP modulation/demodulation
    - Modulate known AX.25 frames, verify IQ output
    - Demodulate known IQ samples, verify frame output
    - DMA buffer management and memory usage tracking
    - _Requirements: 7.1, 7.2_

- [ ] 10. AX.25 CLA Plugin (C — STM32U585)
  - [ ] 8.1 Implement CLA plugin lifecycle and ION-DTN registration (`cla_init`, `cla_activate_link`, `cla_deactivate_link`, `cla_shutdown`)
    - Define `cla_status_t`, `link_metrics_t`, `callsign_t`, `cla_config_t`
    - Register as native ION-DTN CLA plugin implementing LTP link service adapter
    - Configure DMA channels for IQ streaming on link activation
    - Stop DMA and flush buffers on link deactivation
    - _Requirements: 8.4_

  - [ ] 8.2 Implement CLA send/receive paths (`ax25iq_send_segment`, `ax25iq_recv_process`)
    - `ax25iq_send_segment`: wrap LTP segment in AX.25 frame with source/destination callsigns → modulate GFSK/G3RUH → stream IQ via DMA
    - `ax25iq_recv_process`: demodulate IQ from DMA → extract AX.25 frames → deliver LTP segments to ION's LTP engine
    - AX.25 framing with amateur radio callsigns in every frame
    - LTP segmentation/reassembly handled by ION-DTN's LTP engine natively
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [ ] 8.3 Implement link metrics collection (`cla_status`, `cla_get_metrics`)
    - Track RSSI, SNR, BER, bytes transferred, frames sent/received
    - _Requirements: 10.3, 18.3_

  - [ ]* 8.4 Write property test: AX.25 Callsign Framing (theft)
    - **Property 14: AX.25 Callsign Framing**
    - Generate random bundles. Transmit through CLA. Verify output AX.25 frames carry valid source/dest callsigns
    - **Validates: Requirements 8.1**

  - [ ]* 8.5 Write property test: End-to-End Radio Path Round-Trip (theft)
    - **Property 13: End-to-End Radio Path Round-Trip**
    - Generate random valid bundles. Push through full stack: BPv7 → LTP → AX.25 → IQ mod → IQ demod → AX.25 → LTP → BPv7. Assert bundle equality
    - **Validates: Requirements 7.7, 8.5**

- [ ] 11. Power Manager (C — STM32U585)
  - [ ] 9.1 Implement power state management (`power_init`, `power_enter_stop2`, `power_should_sleep`, `power_get_metrics`, `power_log_transition`)
    - Configure RTC and backup domain
    - Enter Stop 2: disable peripherals (DMA, SPI, UART data), retain SRAM, set RTC alarm for wake time
    - Wake from Stop 2 via RTC alarm or external interrupt — resume within 10 ms
    - Track cumulative active/Stop 2 time and transition count
    - Measure wake-up latency (Stop 2 → active)
    - Coordinate with CLA (stop DMA before sleep) and store (flush NVM before sleep)
    - _Requirements: 13.1, 13.2, 13.3, 13.4_

  - [ ]* 9.2 Write unit tests for power manager
    - State transition logging correctness
    - RTC alarm configuration for wake times
    - `power_should_sleep` logic (no active contact, no pending work)
    - _Requirements: 13.1, 13.3_

- [ ] 12. UART Command Interface (C — STM32U585)
  - [ ] 10.1 Implement UART command handler (`uart_cmd_init`, `uart_cmd_process`, `uart_cmd_send_telemetry`, `uart_cmd_send_status`)
    - Define `uart_cmd_type_t`, `uart_cmd_frame_t`, `uart_resp_frame_t` and command payloads (`cmd_contact_activate_t`, `telemetry_response_t`)
    - Parse incoming command frames with CRC-16 validation
    - Dispatch commands to BPA, CLA, power manager, store subsystems
    - Send telemetry and status responses back to Node Controller
    - Respond to telemetry requests within 500 ms
    - Non-blocking: integrates into firmware main loop without stalling ION-DTN
    - _Requirements: 15.1, 15.4, 18.3_

  - [ ]* 10.2 Write unit tests for UART command handler
    - Command frame parsing with valid/invalid CRC
    - Response frame construction
    - Command dispatch to correct subsystems
    - _Requirements: 15.1_

- [ ] 13. Firmware Main Loop and Operation Cycle (C — STM32U585)
  - [ ] 11.1 Implement firmware main loop integrating all C subsystems
    - Initialize: pool allocator → TrustZone → NVM store reload → BPA → CLA → DSP → power manager → UART handler
    - Main loop: process UART commands → check contacts → transmit queued bundles via IQ baseband → process received bundles → run cleanup → check sleep condition
    - Complete full operation cycle within 1 second
    - Handle error scenarios: NVM full (evict + reject), CRC failure (discard + log), pool exhaustion (reject + log)
    - Report SRAM utilization by subsystem as telemetry
    - _Requirements: 21.1, 21.2, 21.3, 14.1, 14.3, 19.1, 19.2, 19.3_

  - [ ]* 11.2 Write unit tests for firmware operation cycle
    - Verify cycle completes within 1 second (mock peripherals)
    - Verify NVM store/retrieve within 50 ms target
    - Verify BPA validation within 10 ms per bundle
    - _Requirements: 21.1, 21.2, 21.3_

- [ ] 14. Checkpoint — All C firmware components complete
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 15. Contact Plan Manager (Go — Companion Host)
  - [ ] 13.1 Implement Contact Plan Manager (`LoadPlan`, `LoadFromFile`, `GetActiveContacts`, `GetNextContact`, `FindDirectContact`, `UpdatePlan`, `Persist`, `Reload`)
    - Define `LinkType`, `ContactWindow`, `ContactPlan` data types
    - Validate: all contacts within valid-from/valid-to, no overlapping contacts on same link, DataRate > 0, StartTime < EndTime
    - Active contacts query: return windows where startTime ≤ t < endTime
    - Next contact lookup: earliest future window matching destination
    - Direct contact only — no multi-hop paths (no-relay constraint)
    - Persist plan to Companion Host filesystem (JSON), reload on restart
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.7, 6.2_

  - [ ] 13.2 Implement simulated pass generation (`GenerateSimulatedPasses`)
    - Generate series of pass windows with configurable duration (5–10 min), inter-pass gap (60–90 min), pass count (4–6)
    - _Requirements: 9.6_

  - [ ]* 13.3 Write property test: Active Contacts Query Correctness (rapid)
    - **Property 15: Active Contacts Query Correctness**
    - Generate random contact plans and query times. Verify returned set is exactly the active contacts (startTime ≤ t < endTime)
    - **Validates: Requirements 9.2**

  - [ ]* 13.4 Write property test: Next Contact Lookup Correctness (rapid)
    - **Property 16: Next Contact Lookup Correctness**
    - Generate random plans, destinations, times. Verify result is earliest future matching contact
    - **Validates: Requirements 9.3**

  - [ ]* 13.5 Write property test: Contact Plan Validity Invariants (rapid)
    - **Property 17: Contact Plan Validity Invariants**
    - Generate random contact plans. Verify all contacts within valid range and no overlaps on same link
    - **Validates: Requirements 9.4, 9.5**

- [ ] 16. Node Controller — Core (Go — Companion Host)
  - [ ] 14.1 Implement UART communication layer for Go ↔ C firmware
    - Define `NodeConfig`, UART command/response frame encoding/decoding matching C firmware protocol
    - Implement serial open, command send, response receive with CRC-16 validation
    - Detect loss of UART communication within 5 seconds, attempt reconnection at configurable retry interval
    - _Requirements: 15.1, 15.3_

  - [ ] 14.2 Implement Node Controller lifecycle (`Initialize`, `Run`, `RunCycle`, `Shutdown`)
    - Open UART, connect B200mini via UHD, load contact plan from file
    - Main loop: check contact schedule → send CONTACT_ACTIVATE/DEACTIVATE at window boundaries → collect telemetry → handle errors
    - `RunCycle` for single-cycle testing
    - Graceful shutdown: deactivate contacts, flush state
    - _Requirements: 15.1, 15.5, 10.1, 10.2_

  - [ ] 14.3 Implement contact window execution logic
    - Send CONTACT_ACTIVATE to firmware when window start time reached
    - Cease transmission command (CONTACT_DEACTIVATE) at window end time
    - Record link metrics per contact (bytes transferred, duration, bundles sent/received, IQ signal quality)
    - Handle missed contacts: IQ link failure → mark missed, retain bundles, increment contacts-missed counter
    - _Requirements: 10.1, 10.2, 10.3, 10.4_

  - [ ]* 14.4 Write property test: No Transmission After Window End (rapid)
    - **Property 18: No Transmission After Window End**
    - Generate random contact windows and time sequences. Verify no transmission command sent after end time
    - **Validates: Requirements 10.2**

  - [ ]* 14.5 Write property test: Missed Contact Retains Bundles (rapid)
    - **Property 19: Missed Contact Retains Bundles**
    - Generate random failed contacts. Verify bundles retained and missed counter incremented by exactly one
    - **Validates: Requirements 10.4**

- [ ] 17. Node Controller — Telemetry and Health (Go — Companion Host)
  - [ ] 15.1 Implement telemetry collection and health reporting (`Health`, `Statistics`, `RequestFirmwareTelemetry`)
    - Define `FirmwareTelemetry`, `NodeHealth`, `NodeStatistics` data types
    - Query firmware via UART TELEMETRY_REQUEST, parse `telemetry_response_t`
    - Track cumulative statistics: total bundles received/sent, bytes received/sent, average latency, contacts completed/missed
    - Report NVM utilization, bundles stored/delivered/dropped, uptime, last contact time
    - Expose telemetry through local interface on Companion Host
    - Return telemetry snapshot within 1 second
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5_

  - [ ] 15.2 Implement power state logging
    - Log timestamped power state transitions (active → Stop 2, Stop 2 → active) and duration in each state
    - Compute and report average power consumption over configurable measurement window
    - _Requirements: 13.4, 13.5_

  - [ ]* 15.3 Write property test: Statistics Monotonicity and Consistency (rapid)
    - **Property 24: Statistics Monotonicity and Consistency**
    - Generate random operation sequences. Verify cumulative stats (bundles received/sent, bytes received/sent, contacts completed/missed) are monotonically non-decreasing
    - **Validates: Requirements 18.2**

- [ ] 18. Checkpoint — Contact Plan Manager and Node Controller core complete
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 19. IQ Bridge Service (Go — Companion Host)
  - [ ] 17.1 Implement IQ Bridge between B200mini and STM32U585
    - Use UHD library to control B200mini: configure UHF 437 MHz center frequency, sample rate for 9.6 kbps GFSK/G3RUH, TX/RX gain
    - Bridge IQ samples between B200mini (USB 3.0) and STM32U585 (SPI/UART DMA) — transparent, no modification
    - Handle B200mini initialization, error detection, and reinitialization at configurable retry interval
    - _Requirements: 7.4, 7.5, 15.2, 19.5_

  - [ ]* 17.2 Write unit tests for IQ Bridge
    - Verify UHD configuration parameters (frequency, sample rate, gain)
    - Verify transparent sample bridging (no modification)
    - Verify B200mini error handling and reinitialization
    - _Requirements: 7.4, 15.2, 19.5_

- [ ] 20. Node Controller — Error Handling and Fault Recovery (Go — Companion Host)
  - [ ] 18.1 Implement error handling for IQ Bridge disconnection, B200mini failure, and UART loss
    - IQ Bridge disconnection: detect within 5 seconds, mark contact interrupted, retain bundles, attempt reconnection
    - B200mini failure: log UHD error, notify Node Controller, attempt reinitialization
    - UART loss: detect within 5 seconds, attempt reconnection at configurable retry interval
    - _Requirements: 19.4, 19.5, 15.3_

  - [ ]* 18.2 Write unit tests for error handling
    - Simulate IQ Bridge disconnection, verify detection and recovery
    - Simulate B200mini failure, verify logging and reinitialization attempt
    - Simulate UART loss, verify detection and reconnection
    - _Requirements: 19.4, 19.5_

- [ ] 21. Node Controller — Simulated Pass Test Orchestration (Go — Companion Host)
  - [ ] 19.1 Implement automated simulated pass test sequences (`RunSimulatedPassTest`)
    - Define `PassMetrics`, `TestReport` data types
    - Configure series of simulated pass windows (5–10 min duration, 60–90 min gaps, 4–6 passes/day)
    - Per pass: wake firmware from Stop 2 → activate IQ radio → process bundles → deactivate → return to Stop 2
    - Record per-pass metrics: bundles uploaded/downloaded, bytes transferred, duration, wake-up latency, power consumption, link SNR/BER
    - Generate test report: aggregate throughput, delivery success rate, power budget compliance
    - _Requirements: 20.1, 20.2, 20.3, 20.4, 20.5_

  - [ ]* 19.2 Write unit tests for simulated pass test orchestration
    - Verify pass schedule generation with correct timing
    - Verify per-pass metric collection
    - Verify test report aggregation
    - _Requirements: 20.1, 20.4, 20.5_

- [ ] 22. Checkpoint — All Go components complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 23. Integration Wiring — Go ↔ C Full System
  - [ ] 21.1 Wire UART command interface end-to-end (Go Node Controller ↔ C UART handler)
    - Verify all command types work: CONTACT_ACTIVATE, CONTACT_DEACTIVATE, TELEMETRY_REQUEST, STATUS_QUERY, POWER_SLEEP, POWER_WAKE, SET_DEFAULT_PRIORITY, SET_RATE_LIMIT, SET_MAX_BUNDLE_SIZE
    - Verify CRC-16 integrity on both sides
    - Verify telemetry response within 500 ms
    - _Requirements: 15.1, 15.4_

  - [ ] 21.2 Wire IQ Bridge end-to-end (B200mini ↔ Companion Host ↔ STM32U585 DMA)
    - Verify IQ sample flow: B200mini → UHD → IQ Bridge → SPI/UART DMA → STM32U585 DSP (and reverse)
    - Verify transparent bridging — no sample modification
    - _Requirements: 7.4, 15.2_

  - [ ] 21.3 Wire contact window execution end-to-end
    - Node Controller schedules contact → sends CONTACT_ACTIVATE → firmware wakes, activates CLA/DSP → bundles transmitted in priority order → CONTACT_DEACTIVATE → firmware enters Stop 2
    - Verify priority-ordered transmission during contact window
    - Verify no transmission after window end
    - _Requirements: 10.1, 10.2, 5.3, 16.2_

  - [ ]* 21.4 Write integration tests for end-to-end store-and-forward
    - Ground station → B200mini → IQ Bridge → STM32U585 (store) → (next pass) → STM32U585 (retrieve) → IQ Bridge → B200mini → destination
    - Verify bundle delivered intact through full RF path
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 7.7, 8.5_

  - [ ]* 21.5 Write integration tests for end-to-end ping
    - Ground station pings EM node through full RF path
    - Verify echo response received with correct RTT
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ]* 21.6 Write integration tests for power cycle recovery
    - Populate NVM store, simulate power cycle, verify store reloaded and operation resumes
    - _Requirements: 2.7, 2.8, 19.3_

  - [ ]* 21.7 Write integration tests for fault scenarios
    - IQ Bridge disconnection during active contact: verify detection within 5 seconds, bundles retained
    - B200mini failure: verify Node Controller handles gracefully
    - SRAM budget: run all subsystems concurrently during simulated pass, verify total SRAM ≤ 786 KB via pool stats
    - _Requirements: 19.4, 19.5, 14.1_

- [ ] 24. Final checkpoint — Full system integration complete
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- C firmware tests use [theft](https://github.com/silentbicycle/theft) for property-based testing
- Go Companion Host tests use [rapid](https://github.com/flyingmutant/rapid) for property-based testing
- Each property test references a specific correctness property from the design document
- ION-DTN is cross-compiled for Cortex-M33 via cgo — the CLA is a native ION-DTN plugin, not a wrapper
- All C firmware memory allocation uses the static pool allocator (task 1) — no malloc/free anywhere
- Checkpoints at tasks 4, 6, 12, 16, 20, and 22 ensure incremental validation
