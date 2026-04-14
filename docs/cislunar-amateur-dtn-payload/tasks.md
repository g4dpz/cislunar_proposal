# Implementation Plan: Cislunar Amateur DTN Payload

## Overview

This implementation plan covers the complete four-phase DTN system for amateur radio: terrestrial validation (RPi + Mobilinkd TNC4 + FT-817), CubeSat Engineering Model (STM32U585 + Ettus B200mini), LEO CubeSat flight (STM32U585 + flight IQ transceiver), and cislunar deep-space communication. The system uses ION-DTN (BPv7/LTP) over AX.25 with callsign-based addressing, supporting ping and store-and-forward operations with no relay functionality.

Implementation is in Go, leveraging ION-DTN for core DTN functionality (BPv7, LTP, bundle storage, priority handling, lifetime enforcement, BPSec). Our code provides: AX.25 frame validation, ION-DTN configuration management, node orchestration, telemetry collection, contact plan management (CGR-based pass prediction), and integration testing.

## Tasks

- [x] 1. Core DTN Infrastructure (Shared Across All Phases)
  - [x] 1.1 Implement Bundle Protocol Agent (BPA) Go wrapper
    - Create `pkg/bpa/` package wrapping ION-DTN bundle operations
    - Implement bundle creation, validation, and type handling (data, ping request, ping response)
    - Implement ping echo request/response handling
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2_

  - [x] 1.2 Implement Bundle Store Go wrapper
    - Create `pkg/store/` package wrapping ION-DTN bundle storage
    - Implement priority-ordered retrieval and capacity management
    - Implement eviction policy (expired first, then lowest priority)
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_

  - [x] 1.3 Implement Contact Plan Manager with CGR integration
    - Create `pkg/contact/` package for contact plan management
    - Implement contact window scheduling and active contact queries
    - Integrate ION-DTN's CGR for orbital pass prediction (LEO/cislunar)
    - Implement direct contact lookup (no multi-hop routing)
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

  - [x] 1.4 Implement Convergence Layer Adapter (CLA) abstraction
    - Create `pkg/cla/` package with CLA interface for all link types
    - Implement AX.25/LTP framing abstraction
    - Support multiple CLA types: VHF/UHF TNC, UHF IQ (B200mini), UHF/S-band/X-band IQ (flight)
    - _Requirements: 10.1, 10.2, 10.3, 10.4_

  - [x] 1.5 Implement Node Controller orchestrator
    - Create `pkg/node/` package for top-level node orchestration
    - Implement autonomous operation cycle (store-check-deliver loop)
    - Implement health monitoring and telemetry collection
    - Support all node types: terrestrial, EM, LEO, cislunar
    - _Requirements: 15.1, 15.2, 15.3_

  - [x]* 1.6 Write property test for Bundle Store/Retrieve Round-Trip
    - **Property 1: Bundle Store/Retrieve Round-Trip**
    - **Validates: Requirement 2.2**
    - Use Go property testing library (e.g., gopter or rapid)
    - Generate arbitrary valid BPv7 bundles, store and retrieve, verify identity

  - [x]* 1.7 Write property test for Bundle Validation Correctness
    - **Property 2: Bundle Validation Correctness**
    - **Validates: Requirements 1.1, 1.2, 1.3**
    - Generate bundles with various valid/invalid fields, verify accept/reject behavior

  - [x]* 1.8 Write property test for Priority Ordering Invariant
    - **Property 3: Priority Ordering Invariant**
    - **Validates: Requirements 2.3, 5.3**
    - Generate arbitrary bundle sets with mixed priorities, verify ordering

  - [x]* 1.9 Write property test for Eviction Policy Ordering
    - **Property 4: Eviction Policy Ordering**
    - **Validates: Requirements 2.4, 2.5**
    - Generate store-at-capacity scenarios, verify eviction order

  - [x]* 1.10 Write property test for Store Capacity Bound
    - **Property 5: Store Capacity Bound**
    - **Validates: Requirement 2.6**
    - Generate arbitrary store/delete sequences, verify capacity never exceeded

- [x] 2. Checkpoint - Core DTN infrastructure complete

- [x] 3. Phase 1: Terrestrial DTN Validation (RPi + TNC4 + FT-817)
  - [x] 3.1 Implement terrestrial CLA for Mobilinkd TNC4
    - Create `pkg/cla/tnc4/` package for TNC4 USB serial interface
    - Implement KISS framing over USB serial
    - Support 9600 baud G3RUH-compatible GFSK on VHF/UHF
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [x] 3.2 Create ION-DTN configuration generator for terrestrial nodes
    - Generate ionrc, ltprc, bprc, ipnrc config files
    - Configure KISS CLA with TNC4 device paths
    - Support two-node terrestrial setup
    - _Requirements: 7.1, 11.1, 11.2_

  - [x] 3.3 Implement terrestrial node CLI
    - Create `cmd/terrestrial-node/` CLI for starting terrestrial DTN nodes
    - Parse config file (node ID, callsign, TNC device, contact plan)
    - Start ION-DTN, monitor health, expose telemetry
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [x]* 3.4 Write integration test for terrestrial ping
    - Test DTN ping between two terrestrial nodes over TNC4 + FT-817
    - Verify echo request/response and RTT measurement
    - _Requirements: 4.1, 4.2, 4.3_

  - [x]* 3.5 Write integration test for terrestrial store-and-forward
    - Test bundle delivery between two terrestrial nodes
    - Verify priority-based delivery and ACK handling
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [x]* 3.6 Write property test for Bundle Lifetime Enforcement
    - **Property 6: Bundle Lifetime Enforcement**
    - **Validates: Requirements 3.1, 3.2**
    - Generate bundles with various lifetimes, verify cleanup removes expired bundles

  - [x]* 3.7 Write property test for Ping Echo Correctness
    - **Property 7: Ping Echo Correctness**
    - **Validates: Requirements 4.1, 4.2**
    - Generate arbitrary ping requests, verify exactly one echo response generated

- [x] 4. Checkpoint - Phase 1 terrestrial validation complete

- [x] 5. Phase 2: CubeSat Engineering Model (STM32U585 + B200mini)
  - [x] 5.1 Implement IQ baseband interface for STM32U585
    - Create `pkg/iq/` package for IQ sample generation and processing
    - Implement GFSK/GMSK/BPSK modulation and demodulation
    - Support DMA-driven IQ streaming
    - _Requirements: 12.2, 12.3_

  - [x] 5.2 Implement B200mini SDR bridge interface
    - Create `pkg/sdr/b200mini/` package for UHD bridge communication
    - Implement SPI/UART bridge to STM32U585 for IQ samples
    - Support USB 3.0 IQ streaming via companion RPi/PC running UHD
    - _Requirements: 12.2, 12.3_

  - [x] 5.3 Implement EM CLA for UHF IQ (B200mini)
    - Create `pkg/cla/uhf_iq_b200/` package for EM UHF link
    - Integrate IQ baseband with B200mini SDR bridge
    - Support 9.6 kbps UHF 437 MHz operation
    - _Requirements: 12.1, 12.2, 12.3_

  - [x] 5.4 Implement external NVM interface for STM32U585
    - Create `pkg/nvm/` package for SPI/QSPI flash interface
    - Implement persistent bundle storage (64-256 MB)
    - Support atomic store/delete with CRC validation
    - _Requirements: 12.6, 2.7_

  - [x] 5.5 Implement STM32U585 power management
    - Create `pkg/power/` package for ultra-low-power modes
    - Implement Stop 2 mode entry/exit (~16 µA idle)
    - Support wake-on-contact for scheduled passes
    - _Requirements: 12.7_

  - [x] 5.6 Create ION-DTN configuration generator for EM node
    - Generate config files for STM32U585 + B200mini setup
    - Configure contact windows for simulated orbital passes (8 min, 9.6 kbps)
    - _Requirements: 12.1, 12.4_

  - [x] 5.7 Implement EM node CLI
    - Create `cmd/em-node/` CLI for starting EM DTN node
    - Support simulated pass testing with realistic timing
    - Monitor power/memory telemetry for flight budget validation
    - _Requirements: 12.1, 12.4, 12.5, 12.6, 12.7_

  - [x]* 5.8 Write integration test for EM simulated pass
    - Test bundle upload/download during simulated 8-minute pass
    - Verify store-and-forward through EM node
    - Validate power budget and SRAM usage (786 KB)
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7_

  - [x]* 5.9 Write property test for Local vs Remote Delivery Routing
    - **Property 8: Local vs Remote Delivery Routing**
    - **Validates: Requirements 5.1, 5.2**
    - Generate bundles with local/remote destinations, verify routing behavior

  - [x]* 5.10 Write property test for ACK Deletes, No-ACK Retains
    - **Property 9: ACK Deletes, No-ACK Retains**
    - **Validates: Requirements 5.4, 5.5**
    - Simulate ACK/no-ACK scenarios, verify bundle retention behavior

- [x] 6. Checkpoint - Phase 2 EM validation complete

- [x] 7. Phase 3: LEO CubeSat Flight (STM32U585 + Flight IQ Transceiver)
  - [x] 7.1 Implement flight IQ transceiver interface
    - Create `pkg/radio/iq_transceiver/` package for flight IQ radio
    - Implement DAC/ADC or SPI interface to STM32U585
    - Support GMSK/BPSK at 9.6 kbps UHF 437 MHz
    - _Requirements: 13.1, 13.2_

  - [x] 7.2 Implement LEO CLA for UHF IQ (flight)
    - Create `pkg/cla/uhf_iq/` package for LEO UHF link
    - Integrate IQ baseband with flight transceiver (no B200mini)
    - Support autonomous operation during orbital passes
    - _Requirements: 13.1, 13.2, 13.3_

  - [x] 7.3 Implement CGR-based pass prediction for LEO
    - Extend `pkg/contact/` with LEO orbital parameter support
    - Implement SGP4/SDP4 orbit propagation via ION-DTN CGR
    - Generate contact windows for ground station passes (5-10 min, 4-6/day)
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

  - [x] 7.4 Create ION-DTN configuration generator for LEO node
    - Generate config files for STM32U585 + flight IQ transceiver
    - Configure CGR-predicted contact windows
    - Support TLE/ephemeris updates for re-prediction
    - _Requirements: 13.1, 13.3, 8.5_

  - [x] 7.5 Implement LEO node CLI
    - Create `cmd/leo-node/` CLI for LEO CubeSat operation
    - Support autonomous store-and-forward during passes
    - Monitor telemetry (temperature, battery, storage)
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5_

  - [x]* 7.6 Write integration test for LEO pass simulation
    - Test ground-to-LEO-to-ground store-and-forward
    - Verify direct delivery (no relay)
    - Validate 5-10 W average power budget
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5_

  - [x]* 7.7 Write property test for No Relay — Direct Delivery Only
    - **Property 10: No Relay — Direct Delivery Only**
    - **Validates: Requirements 6.1, 6.2, 13.5**
    - Generate arbitrary bundle routing scenarios, verify no relay behavior

  - [x]* 7.8 Write property test for Active Contacts Query Correctness
    - **Property 11: Active Contacts Query Correctness**
    - **Validates: Requirement 7.2**
    - Generate arbitrary contact plans and query times, verify active contact results

  - [x]* 7.9 Write property test for Next Contact Lookup Correctness
    - **Property 12: Next Contact Lookup Correctness**
    - **Validates: Requirement 7.3**
    - Generate arbitrary contact plans and destinations, verify next contact lookup

- [x] 8. Checkpoint - Phase 3 LEO flight validation complete

- [x] 9. Phase 4: Cislunar Deep-Space Communication
  - [x] 9.1 Implement S-band/X-band IQ transceiver interface
    - Create `pkg/radio/sband_transceiver/` package for cislunar radio
    - Support BPSK + LDPC/Turbo coding at 500 bps S-band 2.2 GHz
    - Account for 1-2 second one-way light-time delay
    - _Requirements: 14.1, 14.2_

  - [x] 9.2 Implement cislunar CLA for S-band IQ
    - Create `pkg/cla/sband_iq/` package for cislunar link
    - Integrate IQ baseband with S-band transceiver
    - Support long-delay LTP session management
    - _Requirements: 14.1, 14.2, 14.3_

  - [x] 9.3 Implement CGR-based pass prediction for cislunar
    - Extend `pkg/contact/` with cislunar orbital parameter support
    - Implement lunar orbit propagation via ION-DTN CGR
    - Generate contact windows with 1-2 second delay and confidence degradation
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

  - [x] 9.4 Create ION-DTN configuration generator for cislunar node
    - Generate config files for cislunar payload
    - Configure CGR-predicted contact windows with light-time delay
    - Support long-duration message storage
    - _Requirements: 14.1, 14.3, 14.4_

  - [x] 9.5 Implement cislunar node CLI
    - Create `cmd/cislunar-node/` CLI for cislunar payload operation
    - Support autonomous store-and-forward across extended contact gaps
    - Monitor telemetry for deep-space operation
    - _Requirements: 14.1, 14.2, 14.3, 14.4_

  - [x]* 9.6 Write integration test for cislunar store-and-forward
    - Test Earth-to-cislunar-to-Earth bundle delivery
    - Verify 1-2 second delay handling
    - Validate 500 bps S-band link budget (5-7 dB margin)
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 18.2_

  - [x]* 9.7 Write property test for Contact Plan Validity Invariants
    - **Property 13: Contact Plan Validity Invariants**
    - **Validates: Requirements 7.4, 7.5**
    - Generate arbitrary contact plans, verify validity constraints

  - [x]* 9.8 Write property test for CGR Prediction Validity
    - **Property 14: CGR Prediction Validity**
    - **Validates: Requirements 8.1, 8.6, 8.7**
    - Generate arbitrary orbital parameters and time horizons, verify prediction validity

  - [x]* 9.9 Write property test for CGR Elevation Threshold
    - **Property 15: CGR Elevation Threshold**
    - **Validates: Requirement 8.2**
    - Generate arbitrary ground stations and predictions, verify elevation thresholds

  - [x]* 9.10 Write property test for CGR Sorted Output
    - **Property 16: CGR Sorted Output**
    - **Validates: Requirement 8.3**
    - Generate arbitrary predictions, verify sorted by start time

  - [x]* 9.11 Write property test for CGR Confidence Monotonicity
    - **Property 17: CGR Confidence Monotonicity**
    - **Validates: Requirement 8.4**
    - Generate arbitrary predictions, verify confidence decreases with time from epoch

- [x] 10. Checkpoint - Phase 4 cislunar validation complete

- [x] 11. Security and Error Handling
  - [x] 11.1 Implement BPSec integration
    - Create `pkg/security/` package for BPSec operations
    - Integrate ION-DTN BPSec for integrity blocks (RFC 9172)
    - Use STM32U585 hardware crypto accelerator (AES-256, SHA-256)
    - Store keys in TrustZone secure world
    - _Requirements: 16.1, 16.2, 16.3_

  - [x] 11.2 Implement rate limiting
    - Add rate limiting to BPA bundle acceptance
    - Prevent store flooding attacks
    - _Requirements: 16.4_

  - [x] 11.3 Implement contact plan integrity verification
    - Add signed contact plan support for space nodes
    - Verify plan signatures before loading
    - _Requirements: 16.5_

  - [x] 11.4 Implement error recovery for store full
    - Handle store-at-capacity scenarios with eviction
    - Log eviction events for telemetry
    - _Requirements: 17.1_

  - [x] 11.5 Implement error recovery for bundle corruption
    - Handle CRC validation failures
    - Log corruption events with link metrics
    - _Requirements: 17.2_

  - [x] 11.6 Implement error recovery for power loss
    - Reload bundle store from NVM after power cycle
    - Validate store integrity via CRC
    - Rebuild from intact bundles if corruption detected
    - _Requirements: 17.3, 17.4_

  - [x] 11.7 Implement error recovery for missed contacts
    - Retain bundles when contact window missed
    - Increment contacts-missed counter
    - _Requirements: 9.4_

  - [x] 11.8 Implement error recovery for no contact available
    - Retain bundles when no direct contact exists
    - Re-evaluate when contact plan updated
    - _Requirements: 17.5_

  - [x]* 11.9 Write property test for No Transmission After Window End
    - **Property 18: No Transmission After Window End**
    - **Validates: Requirement 9.2**
    - Generate arbitrary contact windows and transmission attempts, verify no transmission after end

  - [x]* 11.10 Write property test for Missed Contact Retains Bundles
    - **Property 19: Missed Contact Retains Bundles**
    - **Validates: Requirement 9.4**
    - Simulate missed contacts, verify bundle retention and counter increment

  - [x]* 11.11 Write property test for Rate Limiting
    - **Property 24: Rate Limiting**
    - **Validates: Requirement 16.4**
    - Generate rapid bundle submission sequences, verify rate limiting behavior

  - [x]* 11.12 Write property test for Bundles Retained When No Contact Available
    - **Property 25: Bundles Retained When No Contact Available**
    - **Validates: Requirements 17.5, 5.5**
    - Generate scenarios with no available contacts, verify bundle retention

- [x] 12. Checkpoint - Security and error handling complete

- [x] 13. Link Budget Validation and Protocol Testing
  - [x] 13.1 Implement link budget computation
    - Create `pkg/linkbudget/` package for RF link analysis
    - Implement free-space path loss (FSPL) calculation
    - Compute link margin for LEO UHF and cislunar S-band
    - _Requirements: 18.1, 18.2, 18.3_

  - [x] 13.2 Implement AX.25 callsign framing validation
    - Verify all CLA implementations produce valid AX.25 frames
    - Verify source/destination callsigns in every frame
    - _Requirements: 10.1_

  - [x] 13.3 Implement LTP segmentation/reassembly testing
    - Test large bundle segmentation and reassembly
    - Verify round-trip identity for bundles exceeding single frame
    - _Requirements: 10.3_

  - [x]* 13.4 Write property test for AX.25 Callsign Framing
    - **Property 20: AX.25 Callsign Framing**
    - **Validates: Requirement 10.1**
    - Generate arbitrary bundles, verify AX.25 framing with callsigns

  - [x]* 13.5 Write property test for LTP Segmentation/Reassembly Round-Trip
    - **Property 21: LTP Segmentation/Reassembly Round-Trip**
    - **Validates: Requirement 10.3**
    - Generate arbitrary large bundles, verify segmentation/reassembly identity

  - [x]* 13.6 Write property test for Modulation/Demodulation Round-Trip
    - **Property 22: Modulation/Demodulation Round-Trip**
    - **Validates: Requirement 13.2**
    - Generate arbitrary payloads, verify modulation/demodulation identity

  - [x]* 13.7 Write property test for Statistics Consistency
    - **Property 23: Statistics Consistency**
    - **Validates: Requirement 15.3**
    - Generate arbitrary operation sequences, verify statistics monotonicity and consistency

  - [x]* 13.8 Write property test for Link Margin Monotonically Decreasing with Distance
    - **Property 26: Link Margin Monotonically Decreasing with Distance**
    - **Validates: Requirement 18.3**
    - Generate arbitrary distance pairs, verify link margin decreases with distance

- [ ] 14. Checkpoint - Link budget and protocol testing complete

- [ ] 15. End-to-End Integration and Documentation
  - [ ] 15.1 Run full four-phase integration test
    - Test terrestrial → EM → LEO → cislunar progression
    - Verify ping and store-and-forward across all phases
    - Validate telemetry and health monitoring
    - _Requirements: all_

  - [ ] 15.2 Run extended duration test
    - Run all nodes for 24+ hours with periodic bundle exchanges
    - Verify no memory leaks, no crashes, accurate telemetry
    - _Requirements: 15.1, 15.2, 15.3_

  - [ ]* 15.3 Create operational documentation
    - Write README with setup instructions for all phases
    - Document configuration files and device mappings
    - Include troubleshooting guide
    - _Requirements: all_

  - [ ]* 15.4 Create developer documentation
    - Document Go package architecture
    - Document ION-DTN integration points
    - Document CGR contact prediction workflow
    - _Requirements: all_

- [ ] 16. Final checkpoint - Cislunar Amateur DTN Payload complete

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at phase boundaries
- Property tests validate universal correctness properties from the design document
- Unit tests and integration tests validate specific examples and end-to-end flows
- ION-DTN provides core DTN functionality; our Go code provides orchestration, telemetry, and phase-specific interfaces
- The four phases build progressively: terrestrial → EM → LEO → cislunar
- CGR is used exclusively for contact prediction / pass scheduling, NOT for multi-hop relay routing
- All bundle delivery is direct (source → destination); no relay functionality
