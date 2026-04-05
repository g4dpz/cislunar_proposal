# Requirements Document

## Introduction

This document specifies the requirements for Phase 2 of the cislunar amateur DTN project: the CubeSat Engineering Model (EM). Phase 2 validates the flight software stack on ground-based, flight-representative hardware before committing to orbital deployment. The EM uses an STM32U585 ultra-low-power ARM Cortex-M33 OBC (160 MHz, 2 MB flash, 786 KB SRAM, hardware crypto AES/SHA/PKA, TrustZone) running ION-DTN (BPv7/LTP over AX.25) on bare metal or a lightweight RTOS, with C firmware for the DTN/radio stack and Go orchestration on a companion host.

The RF front-end is an Ettus Research USRP B200mini SDR (USB 3.0, 12-bit ADC/DAC, 70 MHz–6 GHz, full-duplex IQ), connected to a companion Raspberry Pi or PC running the UHD driver. The companion host bridges IQ samples to/from the STM32U585 via SPI/UART/DMA. The STM32U585 generates TX IQ samples and processes RX IQ samples directly via its DMA engine — the same baseband DSP code that will fly. The B200mini is EM-only; the flight unit replaces it with a dedicated IQ transceiver IC. External SPI/QSPI NVM (64–256 MB) provides persistent bundle storage.

The system supports two core operations: ping (DTN reachability test) and store-and-forward (point-to-point bundle delivery). There is no relay functionality. All bundle delivery is direct (source → destination). The protocol stack is BPv7 bundles over LTP sessions over AX.25 frames with callsign-based addressing for amateur radio regulatory compliance. BPSec provides integrity protection using HMAC-SHA-256 via the STM32U585 hardware crypto accelerator (no encryption, per amateur radio regulations).

Phase 2 operates at UHF 437 MHz at 9.6 kbps, matching the flight configuration. Simulated orbital pass testing validates store-and-forward under realistic contact windows (5–10 min, 4–6 passes/day). Power budget profiling validates STM32U585 Stop 2 ultra-low-power mode (~16 µA idle) and active power consumption (5–10 W average).

Phase 1 (terrestrial validation with RPi + TNC4 + FT-817) is complete and provides the ION-DTN protocol stack, CLA plugin architecture, ping/store-and-forward operations, and no-relay constraint that Phase 2 inherits and adapts for the STM32U585 IQ baseband architecture.

Out of scope: flight-qualified IQ transceiver IC (Phase 3), orbital deployment (Phase 3), CGR contact prediction / orbital mechanics (Phase 3), S-band / X-band / cislunar communications (Phase 4), relay functionality.

### Development Hardware Note

Phase 2 EM development and testing uses an **ST NUCLEO-F753ZI** development board (STM32F753, ARM Cortex-M7, 216 MHz, 512 KB SRAM, 1 MB flash) as the initial test platform. The NUCLEO-F753ZI provides more headroom than the flight-target STM32U585 (Cortex-M33, 160 MHz, 786 KB SRAM), allowing firmware bring-up and debugging before constraining to the U585's tighter resource budget. The firmware is written to be portable between the F7 and U585 — peripheral abstraction via STM32 HAL ensures the same application code runs on both. TrustZone and hardware crypto features are validated on the U585 target once the core firmware is stable on the NUCLEO-F753ZI. SRAM budget validation (786 KB constraint) is enforced via pool allocator configuration even when running on the larger-SRAM F753.

## Glossary

- **STM32U585**: Ultra-low-power ARM Cortex-M33 MCU — 160 MHz, 2 MB flash, 786 KB SRAM, hardware crypto accelerator (AES-256, SHA-256, PKA), TrustZone security — the flight-target OBC for EM and flight nodes
- **NUCLEO_F753ZI**: ST NUCLEO-F753ZI development board (STM32F753, ARM Cortex-M7, 216 MHz, 512 KB SRAM, 1 MB flash) — the initial development and testing platform for Phase 2 EM firmware before transitioning to the STM32U585 flight target
- **B200mini**: Ettus Research USRP B200mini SDR — USB 3.0, 12-bit ADC/DAC, 70 MHz–6 GHz, full-duplex IQ streaming — EM-only RF front-end
- **Companion_Host**: Raspberry Pi or PC running the UHD driver, acting as USB host for the B200mini and bridging IQ samples to/from the STM32U585 via SPI/UART/DMA
- **UHD**: USRP Hardware Driver — Ettus Research library for controlling the B200mini SDR from the Companion_Host
- **IQ_Bridge**: The SPI/UART/DMA interface between the Companion_Host and the STM32U585 that carries baseband IQ samples in both TX and RX directions
- **NVM**: External SPI/QSPI non-volatile memory (64–256 MB flash) connected to the STM32U585 for persistent bundle storage
- **BPA**: Bundle Protocol Agent — the core ION-DTN engine running on the STM32U585 that creates, receives, validates, stores, and delivers BPv7 bundles
- **Bundle_Store**: Persistent storage subsystem backed by external NVM for bundles awaiting delivery
- **Contact_Plan_Manager**: Subsystem that maintains manually configured communication windows for simulated orbital passes (no CGR in Phase 2)
- **CLA**: Convergence Layer Adapter — native ION-DTN CLA plugin running on the STM32U585 that provides AX.25 framing as the LTP link service layer, adapted from Phase 1 for IQ baseband instead of TNC4
- **Node_Controller**: Top-level orchestrator — Go process on the Companion_Host managing the STM32U585 firmware lifecycle, contact scheduling, and telemetry collection
- **Firmware**: C code running on the STM32U585 (bare metal or lightweight RTOS) implementing ION-DTN BPv7/LTP, AX.25 CLA, IQ baseband DSP, NVM bundle store, and power management
- **ION-DTN**: NASA JPL's Interplanetary Overlay Network — the DTN implementation providing BPv7, LTP, and related protocols, cross-compiled for STM32U585
- **LTP**: Licklider Transmission Protocol — runs on top of AX.25 providing reliable transfer with deferred acknowledgment
- **AX.25**: Link-layer framing protocol providing callsign-based source/destination addressing for amateur radio compliance
- **BPSec**: Bundle Protocol Security (RFC 9172) — provides integrity blocks (HMAC-SHA-256) for bundle origin authentication
- **TrustZone**: ARM TrustZone hardware isolation on the STM32U585 — partitions the MCU into secure and non-secure worlds for key storage and crypto operations
- **Stop_2_Mode**: STM32U585 ultra-low-power sleep mode (~16 µA) with SRAM retention, used between simulated contact windows
- **DMA**: Direct Memory Access — STM32U585 peripheral for streaming IQ samples between memory and the IQ_Bridge interface without CPU intervention
- **Ping**: DTN reachability test — send a bundle echo request and receive an echo response
- **Store_and_Forward**: Point-to-point bundle delivery where a source node sends a bundle directly to a destination node during a contact window
- **Contact_Window**: A scheduled time interval during which the EM node can communicate with a ground node over the UHF radio link
- **Endpoint_ID**: A DTN endpoint identifier using the "dtn" or "ipn" URI scheme that uniquely addresses a node or application

## Requirements

### Requirement 1: STM32U585 ION-DTN Bundle Creation and Validation

**User Story:** As an EM test operator, I want the STM32U585 firmware to create and validate BPv7 bundles within its constrained SRAM, so that the flight DTN stack is validated on representative hardware.

#### Acceptance Criteria

1. WHEN a message with a valid destination Endpoint_ID and payload is submitted, THE BPA SHALL create a BPv7 bundle with the bundle version set to 7, proper source and destination Endpoint_IDs, a CRC integrity check, a priority level (critical, expedited, normal, or bulk), and a positive lifetime value in seconds
2. WHEN a bundle is received via the CLA, THE BPA SHALL validate that the bundle version equals 7, the destination is a well-formed Endpoint_ID, the lifetime is greater than zero, the creation timestamp does not exceed the current time, and the CRC is correct
3. IF a received bundle fails any validation check, THEN THE BPA SHALL discard the bundle and log the specific validation failure reason along with the source Endpoint_ID
4. THE BPA SHALL support three bundle types: data bundles for store-and-forward payload delivery, ping request bundles for echo requests, and ping response bundles for echo responses
5. FOR ALL valid Bundle objects, serializing a Bundle to its BPv7 wire format (CBOR) and then parsing the wire format back SHALL produce a Bundle equivalent to the original (round-trip property)
6. THE BPA SHALL complete bundle creation and validation using a working memory allocation that fits within the STM32U585 786 KB SRAM budget shared with IQ buffers and the ION-DTN runtime

### Requirement 2: NVM Bundle Storage and Persistence

**User Story:** As an EM test operator, I want bundles to be stored persistently on external SPI/QSPI NVM and survive power cycles and watchdog resets, so that no messages are lost during disruptions on the constrained STM32U585 platform.

#### Acceptance Criteria

1. WHEN a valid bundle is accepted, THE Bundle_Store SHALL persist the bundle to external SPI/QSPI NVM atomically, preventing corruption if power is lost during the write
2. WHEN a bundle is stored and later retrieved by its bundle ID (source Endpoint_ID, creation timestamp, sequence number), THE Bundle_Store SHALL return a bundle identical to the original (round-trip property)
3. THE Bundle_Store SHALL maintain a priority-ordered index so that bundles are retrieved in priority order: critical first, then expedited, then normal, then bulk
4. WHEN the Bundle_Store reaches the configured maximum NVM capacity and a new bundle arrives, THE Bundle_Store SHALL evict expired bundles first, then the lowest-priority bundles with the earliest creation timestamps, to free sufficient space for the new bundle
5. WHEN evicting bundles, THE Bundle_Store SHALL preserve all critical-priority bundles until all expedited, normal, and bulk bundles have been evicted
6. THE Bundle_Store SHALL enforce that total stored bytes do not exceed the configured maximum NVM capacity (64–256 MB)
7. WHEN the STM32U585 restarts after a power cycle or watchdog reset, THE Bundle_Store SHALL reload its persisted state from external NVM and validate store integrity via CRC checks on each stored bundle
8. IF store corruption is detected during reload, THEN THE Bundle_Store SHALL rebuild from intact bundles only, discarding corrupted entries and logging each discarded bundle ID

### Requirement 3: Bundle Lifetime Enforcement

**User Story:** As an EM test operator, I want expired bundles to be automatically removed from NVM, so that stale data does not consume limited storage capacity.

#### Acceptance Criteria

1. WHEN the Firmware runs a cleanup cycle, THE Bundle_Store SHALL delete all bundles from NVM whose creation timestamp plus lifetime is less than or equal to the current time
2. THE Bundle_Store SHALL contain zero expired bundles after a cleanup cycle completes

### Requirement 4: Ping Operation on EM Hardware

**User Story:** As an EM test operator, I want to ping the STM32U585 EM node over the B200mini RF link and receive an echo response, so that I can verify end-to-end DTN reachability through the IQ baseband radio path.

#### Acceptance Criteria

1. WHEN the BPA receives a ping request bundle addressed to a local endpoint, THE BPA SHALL generate exactly one ping response bundle with the destination set to the original sender's Endpoint_ID
2. WHEN a ping response is generated, THE BPA SHALL queue the response in the Bundle_Store for delivery during the next available Contact_Window with the sender's node
3. WHEN a ping response is received at the originating ground node, THE Node_Controller SHALL compute and report the round-trip time from the original ping request creation timestamp to the response receipt time
4. THE BPA SHALL include the original ping request's bundle ID in the ping response payload so the originating node can correlate responses to requests

### Requirement 5: Store-and-Forward on EM Hardware

**User Story:** As an EM test operator, I want to send messages that are stored on the STM32U585 NVM and delivered to the destination ground node when a simulated contact window opens, so that I can validate store-and-forward under flight-representative constraints.

#### Acceptance Criteria

1. WHEN a data bundle is received whose destination matches a local Endpoint_ID on the STM32U585, THE BPA SHALL deliver the bundle payload to the local application agent
2. WHEN a data bundle is created or received whose destination is a remote Endpoint_ID, THE BPA SHALL store the bundle in the NVM-backed Bundle_Store and queue it for direct delivery during the next Contact_Window with the destination node
3. THE BPA SHALL transmit queued bundles in priority order (critical first, then expedited, then normal, then bulk) during each Contact_Window
4. WHEN a transmitted bundle is acknowledged by the remote node via LTP, THE Bundle_Store SHALL delete the acknowledged bundle from NVM
5. IF a bundle transmission is not acknowledged within the LTP retransmission timeout, THEN THE Bundle_Store SHALL retain the bundle for retry during the next Contact_Window

### Requirement 6: No Relay Constraint

**User Story:** As a system architect, I want to enforce that the EM node only delivers bundles directly to their final destination, so that the system remains consistent with the Phase 1 no-relay architecture.

#### Acceptance Criteria

1. THE BPA SHALL transmit a bundle only to the node matching the bundle's final destination Endpoint_ID — the BPA SHALL NOT forward bundles on behalf of other nodes
2. WHEN the Node_Controller looks up a delivery route for a bundle, THE Contact_Plan_Manager SHALL return only direct Contact_Windows with the destination node, with no multi-hop paths

### Requirement 7: IQ Baseband Radio Interface

**User Story:** As an EM test operator, I want the STM32U585 to generate and process IQ baseband samples via DMA, bridged through the Companion_Host to the B200mini SDR, so that the flight baseband DSP code is validated with a real RF front-end.

#### Acceptance Criteria

1. THE Firmware SHALL generate TX IQ baseband samples on the STM32U585 using GFSK/G3RUH modulation at 9.6 kbps and stream them to the Companion_Host via the IQ_Bridge (SPI or UART/DMA)
2. THE Firmware SHALL receive RX IQ baseband samples from the Companion_Host via the IQ_Bridge and demodulate them on the STM32U585 using GFSK/G3RUH demodulation at 9.6 kbps
3. THE Firmware SHALL use the STM32U585 DMA engine for IQ sample streaming, avoiding CPU-bound sample transfers between memory and the IQ_Bridge peripheral
4. THE Companion_Host SHALL run the UHD driver to control the B200mini, converting between IQ sample streams (USB 3.0 to/from B200mini) and the IQ_Bridge interface (SPI/UART/DMA to/from STM32U585)
5. THE IQ_Bridge SHALL operate at UHF 437 MHz center frequency with sufficient sample rate to support 9.6 kbps GFSK/G3RUH modulation
6. THE Firmware SHALL manage IQ sample buffers within the STM32U585 786 KB SRAM budget, sharing memory with the ION-DTN runtime and Bundle_Store index
7. FOR ALL valid AX.25 frames, modulating a frame into IQ samples and then demodulating the IQ samples back SHALL produce a frame equivalent to the original (round-trip property for the baseband DSP path)

### Requirement 8: AX.25 and LTP Convergence Layer on STM32U585

**User Story:** As an EM test operator, I want all DTN transmissions to use AX.25 framing with callsign addressing over LTP via the IQ baseband radio, so that every transmission complies with amateur radio regulations and the CLA architecture from Phase 1 is validated on flight hardware.

#### Acceptance Criteria

1. THE CLA SHALL encapsulate all bundle transmissions in AX.25 frames carrying the source amateur radio callsign and the destination amateur radio callsign
2. THE CLA SHALL run LTP sessions on top of AX.25 frames, providing reliable transfer with deferred acknowledgment for all bundle delivery
3. THE CLA SHALL perform LTP segmentation for bundles that exceed a single AX.25 frame size, and reassemble received LTP segments into complete bundles
4. THE CLA SHALL interface with the IQ baseband radio path (STM32U585 DMA → IQ_Bridge → Companion_Host → B200mini) instead of the Phase 1 TNC4 USB serial path
5. FOR ALL valid Bundle objects, encapsulating a bundle into AX.25/LTP frames, modulating to IQ, demodulating from IQ, and reassembling the frames back into a bundle SHALL produce a bundle equivalent to the original (end-to-end round-trip property)

### Requirement 9: Contact Plan Management for Simulated Passes

**User Story:** As an EM test operator, I want to configure simulated orbital pass windows (5–10 min duration, 4–6 passes/day) so that I can validate store-and-forward behavior under realistic contact schedules.

#### Acceptance Criteria

1. THE Contact_Plan_Manager SHALL maintain a time-tagged schedule of Contact_Windows, each specifying a remote node ID, start time, end time, data rate (9.6 kbps), and link type (UHF IQ via B200mini)
2. WHEN queried for active contacts at a given time, THE Contact_Plan_Manager SHALL return all Contact_Windows whose start time is at or before the query time and whose end time is after the query time
3. WHEN queried for the next contact with a specific destination node, THE Contact_Plan_Manager SHALL return the earliest future Contact_Window matching that destination
4. THE Contact_Plan_Manager SHALL reject any contact plan update that would create overlapping Contact_Windows on the same link
5. THE Contact_Plan_Manager SHALL validate that all Contact_Windows fall within the plan's valid-from and valid-to time range
6. THE Contact_Plan_Manager SHALL support configuring simulated orbital pass schedules with window durations of 5–10 minutes and inter-pass gaps of 60–90 minutes (4–6 passes per day)
7. WHEN a contact plan is loaded or updated, THE Contact_Plan_Manager SHALL persist the plan so it survives Companion_Host and STM32U585 restarts

### Requirement 10: Contact Window Execution

**User Story:** As an EM node, I want to transmit queued bundles during active simulated pass windows over the IQ baseband radio, so that store-and-forward delivery is validated under realistic timing constraints.

#### Acceptance Criteria

1. WHEN a Contact_Window becomes active (current time reaches the window start time), THE CLA SHALL activate the IQ baseband radio link and THE Node_Controller SHALL begin transmitting queued bundles destined for the contact's remote node
2. THE Node_Controller SHALL cease all transmission when the Contact_Window end time is reached
3. WHEN a Contact_Window completes, THE Node_Controller SHALL record link metrics (bytes transferred, duration, bundles sent, bundles received, IQ signal quality) and update contact statistics
4. IF the CLA fails to establish the IQ baseband link during a scheduled Contact_Window (B200mini not responding, IQ_Bridge failure, or no AX.25 connection established), THEN THE Node_Controller SHALL mark the contact as missed, retain all queued bundles for the next window, and increment the contacts-missed counter

### Requirement 11: BPSec Integrity with Hardware Crypto

**User Story:** As a network operator, I want bundle integrity protection using BPSec with the STM32U585 hardware crypto accelerator, so that the EM validates the same security path that will fly while complying with amateur radio regulations.

#### Acceptance Criteria

1. THE BPA SHALL support BPSec (RFC 9172) Block Integrity Blocks (BIB) for bundle origin authentication using HMAC-SHA-256
2. THE BPA SHALL NOT apply BPSec Block Confidentiality Blocks (BCB) or any form of payload encryption, in compliance with amateur radio regulations requiring transmissions to be unencrypted
3. THE BPA SHALL use the STM32U585 hardware crypto accelerator (SHA-256, AES-256, PKA) for all BPSec HMAC-SHA-256 computations instead of software implementations
4. WHEN a bundle with a BIB is received, THE BPA SHALL verify the integrity block using the hardware crypto accelerator and discard the bundle if verification fails, logging the integrity failure with the source Endpoint_ID
5. THE Firmware SHALL store BPSec shared keys in the STM32U585 TrustZone secure world, isolated from non-secure application code

### Requirement 12: TrustZone Secure Key Storage

**User Story:** As a security engineer, I want cryptographic keys stored in the STM32U585 TrustZone secure world, so that the EM validates the flight security architecture where keys are hardware-isolated from application firmware.

#### Acceptance Criteria

1. THE Firmware SHALL partition the STM32U585 into TrustZone secure and non-secure worlds, with BPSec keys and crypto operations executing in the secure world
2. THE Firmware SHALL expose a secure API from the TrustZone secure world that the non-secure BPA can call to request HMAC-SHA-256 signing and verification without exposing raw key material
3. IF non-secure code attempts to read TrustZone secure memory directly, THEN THE STM32U585 SHALL generate a hardware fault and the Firmware SHALL log the access violation
4. THE Firmware SHALL provision BPSec keys into the TrustZone secure world during initial firmware flashing or via a secure key injection protocol over the debug interface

### Requirement 13: Power Management and Stop 2 Mode

**User Story:** As an EM test operator, I want to profile the STM32U585 power consumption across active and idle states, so that I can validate the power budget for the flight mission (5–10 W average active, ~16 µA Stop 2 idle).

#### Acceptance Criteria

1. WHEN no Contact_Window is active and no bundle processing is pending, THE Firmware SHALL transition the STM32U585 into Stop 2 ultra-low-power mode
2. WHILE in Stop 2 mode, THE STM32U585 SHALL consume no more than 20 µA (allowing margin above the nominal 16 µA specification)
3. WHEN a Contact_Window start time is reached, THE Firmware SHALL wake the STM32U585 from Stop 2 mode via RTC alarm or external interrupt and resume normal operation within 10 milliseconds
4. THE Node_Controller SHALL log timestamped power state transitions (active → Stop 2, Stop 2 → active) and the duration spent in each state for power budget analysis
5. THE Node_Controller SHALL compute and report average power consumption over a configurable measurement window for comparison against the 5–10 W active power budget

### Requirement 14: SRAM Memory Management

**User Story:** As an embedded systems engineer, I want the STM32U585 firmware to operate within the 786 KB SRAM constraint while concurrently running ION-DTN, IQ baseband DSP, and bundle index management, so that the flight memory budget is validated.

#### Acceptance Criteria

1. THE Firmware SHALL operate within the STM32U585 786 KB SRAM for all concurrent operations: ION-DTN runtime, IQ sample buffers (TX and RX), AX.25/LTP frame buffers, bundle metadata index, and TrustZone secure world allocations
2. THE Firmware SHALL use static or pool-based memory allocation for all runtime data structures, avoiding dynamic heap allocation that could cause fragmentation on the constrained MCU
3. THE Firmware SHALL report peak and current SRAM utilization as part of telemetry, broken down by subsystem (ION-DTN, IQ buffers, bundle index, TrustZone)
4. IF an operation would exceed the SRAM budget, THEN THE Firmware SHALL reject the operation and log the memory exhaustion event rather than corrupting adjacent memory regions

### Requirement 15: Split Architecture — Go Orchestration and C Firmware

**User Story:** As a system integrator, I want the Go orchestration on the Companion_Host and the C firmware on the STM32U585 to communicate reliably over a defined interface, so that the split architecture is validated for the flight configuration.

#### Acceptance Criteria

1. THE Node_Controller (Go on Companion_Host) SHALL communicate with the Firmware (C on STM32U585) over a serial command interface (UART) for control messages including contact activation, contact deactivation, telemetry requests, and firmware status queries
2. THE Companion_Host SHALL bridge IQ samples between the B200mini (USB 3.0 / UHD) and the STM32U585 (SPI/UART/DMA) without modifying the sample data
3. THE Node_Controller SHALL detect loss of communication with the STM32U585 within 5 seconds and attempt reconnection at a configurable retry interval
4. THE Firmware SHALL respond to telemetry requests from the Node_Controller within 500 milliseconds
5. THE Node_Controller SHALL manage the simulated pass schedule and instruct the Firmware to activate or deactivate the IQ radio link at the appropriate Contact_Window boundaries

### Requirement 16: Priority-Based Message Handling

**User Story:** As an EM test operator, I want bundles to be handled according to their priority level on the constrained STM32U585 platform, so that critical messages are delivered before less urgent ones.

#### Acceptance Criteria

1. THE BPA SHALL assign one of four priority levels to each bundle: critical (highest), expedited, normal, or bulk (lowest)
2. WHEN multiple bundles are queued for the same destination during a Contact_Window, THE Firmware SHALL transmit bundles in strict priority order — all critical bundles before any expedited, all expedited before any normal, all normal before any bulk
3. WHEN the Bundle_Store must evict bundles to free NVM space, THE Bundle_Store SHALL evict bulk bundles first, then normal, then expedited — critical bundles SHALL be evicted only when no lower-priority bundles remain
4. THE BPA SHALL accept a default priority level from the Node_Controller configuration, applied to bundles that do not specify an explicit priority

### Requirement 17: Rate Limiting and Store Protection

**User Story:** As an EM test operator, I want to protect the NVM bundle store from flooding, so that a misbehaving ground node cannot exhaust the constrained storage resources.

#### Acceptance Criteria

1. THE BPA SHALL enforce a configurable maximum bundle acceptance rate (bundles per second) per source Endpoint_ID
2. IF the acceptance rate from a single source Endpoint_ID exceeds the configured limit, THEN THE BPA SHALL reject additional bundles from that source and log the rate-limit event
3. THE BPA SHALL enforce a configurable maximum bundle size in bytes, rejecting any bundle whose total serialized size exceeds the limit

### Requirement 18: Node Health and Telemetry

**User Story:** As an EM test operator, I want to monitor the STM32U585 node health, power state, memory utilization, and RF performance, so that I can characterize the EM for flight readiness.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report telemetry including: uptime in seconds, NVM storage utilization as a percentage of configured maximum, number of bundles currently stored, number of bundles delivered, number of bundles dropped (expired or evicted), and the timestamp of the last completed contact
2. THE Node_Controller SHALL track cumulative statistics including: total bundles received, total bundles sent, total bytes received, total bytes sent, average delivery latency in seconds, contacts completed, and contacts missed
3. THE Firmware SHALL report STM32U585-specific telemetry including: current SRAM utilization (peak and current, by subsystem), power state (active or Stop 2), time spent in each power state, MCU temperature (internal sensor), and IQ baseband signal quality metrics (SNR, bit error rate)
4. THE Node_Controller SHALL expose telemetry through a local interface on the Companion_Host accessible to the test operator
5. WHEN a telemetry query is received, THE Node_Controller SHALL return the current telemetry snapshot within 1 second

### Requirement 19: Error Handling and Fault Recovery

**User Story:** As an EM test operator, I want the system to handle faults gracefully including power loss, watchdog resets, IQ_Bridge failures, and memory corruption, so that the EM validates flight-grade fault tolerance.

#### Acceptance Criteria

1. IF the Bundle_Store reaches NVM capacity and eviction cannot free sufficient space for an incoming bundle, THEN THE BPA SHALL reject the incoming bundle and return a storage-full error to the sender if the LTP session is still active
2. IF a CRC validation fails on a received bundle, THEN THE BPA SHALL discard the corrupted bundle and log the corruption event with the source Endpoint_ID and IQ link metrics
3. IF the STM32U585 experiences a power cycle or watchdog reset, THEN THE Firmware SHALL reload the Bundle_Store from external NVM, validate integrity via CRC, and resume normal operation without manual intervention
4. IF the IQ_Bridge connection between the Companion_Host and the STM32U585 is lost during operation, THEN THE Node_Controller SHALL detect the disconnection within 5 seconds, mark the current contact as interrupted, retain all queued bundles, and attempt to re-establish the connection at a configurable retry interval
5. IF the B200mini becomes unresponsive or the UHD driver reports an error, THEN THE Companion_Host SHALL log the failure, notify the Node_Controller, and attempt to reinitialize the B200mini at a configurable retry interval
6. IF no direct Contact_Window exists for a bundle's destination, THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window with that destination is added to the contact plan or the bundle's lifetime expires

### Requirement 20: Simulated Orbital Pass Testing

**User Story:** As an EM test operator, I want to run end-to-end simulated orbital pass tests with realistic timing (5–10 min windows, 90 min gaps, 4–6 passes/day), so that I can validate the complete store-and-forward cycle under flight-representative conditions.

#### Acceptance Criteria

1. THE Node_Controller SHALL support automated test sequences that configure a series of simulated pass windows with configurable duration (5–10 minutes), inter-pass gap (60–90 minutes), and number of passes per day (4–6)
2. WHEN a simulated pass window opens, THE Firmware SHALL wake from Stop 2 mode, activate the IQ baseband radio, and begin processing queued bundles for the contact's remote node
3. WHEN a simulated pass window closes, THE Firmware SHALL deactivate the IQ baseband radio and transition back to Stop 2 mode
4. THE Node_Controller SHALL record per-pass metrics including: bundles uploaded to the EM, bundles downloaded from the EM, total bytes transferred, pass duration, wake-up latency, and power consumption during the pass
5. THE Node_Controller SHALL generate a test report summarizing all passes in a test sequence, including aggregate throughput, delivery success rate, and power budget compliance

### Requirement 21: EM Operation Cycle Performance

**User Story:** As an EM test operator, I want the STM32U585 firmware to complete its operation cycle within defined time bounds, so that the system is responsive during the limited simulated pass windows.

#### Acceptance Criteria

1. THE Firmware SHALL complete a full operation cycle (check contacts, transmit queued bundles via IQ baseband, process received bundles, run cleanup) within 1 second on the STM32U585
2. THE Bundle_Store SHALL complete a single store or retrieve operation on external NVM within 50 milliseconds
3. THE BPA SHALL complete bundle validation (version, Endpoint_ID, lifetime, timestamp, and CRC checks) within 10 milliseconds per bundle on the STM32U585
