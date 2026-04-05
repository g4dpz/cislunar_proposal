# Requirements Document

## Introduction

This document specifies the requirements for Phase 4 of the cislunar amateur DTN project: Cislunar Mission. Phase 4 extends DTN operations from LEO (Phase 3) to cislunar distances (~384,000 km Earth–Moon), enabling amateur participation in deep-space delay-tolerant networking.

The system uses the same ION-DTN (BPv7/LTP over AX.25) protocol stack validated in Phases 1–3. The key changes from Phase 3 are: S-band 2.2 GHz replaces UHF 437 MHz, 500 bps replaces 9.6 kbps, BPSK + LDPC/Turbo FEC replaces GMSK/BPSK, 1–2 second one-way light-time delay (vs. milliseconds in LEO), hours-long contact arcs replace 5–10 minute LEO passes, Tier 3/4 ground stations (3–5m dishes) replace Tier 1/2 stations, external NVM is expanded to 256 MB–1 GB for long-duration storage, power budget increases to 10–20 W, and enhanced radiation tolerance is required for the cislunar environment beyond the Van Allen belts.

The OBC baseline is the STM32U585 (same as Phase 3), with the option to upgrade to a more capable processor if mission requirements demand it. The design is processor-flexible — all interfaces are defined to work on the STM32U585 baseline while accommodating a higher-capability OBC.

The system supports two core operations: ping (DTN reachability test) and store-and-forward (point-to-point bundle delivery). There is no relay functionality. All bundle delivery is direct (source → destination). LTP's deferred acknowledgment model is critical at cislunar distances where round-trip times are 2–4 seconds.

CGR contact prediction is adapted for cislunar orbital mechanics — using numerical orbit propagation or pre-computed ephemeris tables instead of SGP4/SDP4 (which is designed for near-Earth orbits). Contact windows are hours-long arcs with slower Doppler dynamics but longer duration than LEO passes.

Phase 3 (LEO CubeSat Flight) is complete and provides the validated firmware: ION-DTN BPA, LTP, AX.25 CLA plugin architecture, IQ baseband DSP, NVM bundle store with atomic writes, TrustZone secure key storage, hardware crypto BPSec, pool allocator, CGR framework, Doppler compensation, radiation monitor, autonomous operation cycle, and all supporting subsystems.

Out of scope: relay functionality, X-band (future enhancement), optical communication, Mars-relay simulations.

## Glossary

- **OBC**: Onboard Computer — STM32U585 as baseline (160 MHz Cortex-M33, 2 MB flash, 786 KB SRAM, hardware crypto, TrustZone) or a more capable processor if mission requirements demand it
- **Flight_Transceiver**: Flight-qualified S-band IQ transceiver IC interfacing with the OBC via DAC/ADC or SPI — operates at 2.2 GHz with BPSK modulation
- **NVM**: External non-volatile memory (256 MB–1 GB) connected to the OBC via SPI/QSPI for persistent bundle storage — expanded from Phase 3's 64–256 MB to support long-duration cislunar storage
- **BPA**: Bundle Protocol Agent — the core ION-DTN engine running on the OBC that creates, receives, validates, stores, and delivers BPv7 bundles
- **Bundle_Store**: Persistent storage subsystem backed by external NVM for bundles awaiting delivery
- **CGR_Engine**: ION-DTN's Contact Graph Routing module running on the OBC, adapted for cislunar orbital mechanics — used exclusively for contact prediction (pass scheduling) using numerical propagation or pre-computed ephemeris, not for multi-hop relay routing
- **Contact_Plan_Manager**: Subsystem that maintains CGR-predicted communication windows and manages contact scheduling autonomously onboard the cislunar payload
- **CLA**: Convergence Layer Adapter — native ION-DTN CLA plugin running on the OBC that provides AX.25 framing as the LTP link service layer, adapted for S-band 2.2 GHz at 500 bps with BPSK + LDPC/Turbo FEC
- **Node_Controller**: Top-level autonomous orchestrator running on the OBC — manages the operation cycle (wake, transmit, receive, sleep) without external control
- **Firmware**: C code running on the OBC implementing ION-DTN BPv7/LTP, AX.25 CLA, IQ baseband DSP, NVM bundle store, CGR contact prediction, power management, and TrustZone secure crypto
- **ION-DTN**: NASA JPL's Interplanetary Overlay Network — the DTN implementation providing BPv7, LTP, CGR, and related protocols
- **LTP**: Licklider Transmission Protocol — runs on top of AX.25 providing reliable transfer with deferred acknowledgment; deferred ACK is critical at cislunar distances (2–4 second RTT)
- **AX.25**: Link-layer framing protocol providing callsign-based source/destination addressing for amateur radio compliance
- **BPSec**: Bundle Protocol Security (RFC 9172) — provides integrity blocks (HMAC-SHA-256) for bundle origin authentication
- **TrustZone**: ARM TrustZone hardware isolation on the OBC — partitions the MCU into secure and non-secure worlds for key storage and crypto operations
- **DMA**: Direct Memory Access — OBC peripheral for streaming IQ samples between memory and the Flight_Transceiver interface without CPU intervention
- **LDPC**: Low-Density Parity-Check code — strong forward error correction used for the cislunar S-band link at 500 bps
- **Turbo_Code**: Turbo error correction code — alternative strong FEC option for the cislunar S-band link
- **BPSK**: Binary Phase Shift Keying — modulation scheme for the cislunar S-band link
- **Doppler_Compensation**: Frequency offset correction applied by the Firmware to account for relative velocity between the cislunar payload and ground stations
- **Contact_Window**: A predicted or actual time interval during which the cislunar payload has line-of-sight communication with a ground station — hours-long arcs at cislunar distances
- **Ping**: DTN reachability test — send a bundle echo request and receive an echo response; RTT is 2–4 seconds at cislunar distances
- **Store_and_Forward**: Point-to-point bundle delivery where a source node sends a bundle directly to a destination node during a Contact_Window
- **Endpoint_ID**: A DTN endpoint identifier using the "dtn" or "ipn" URI scheme that uniquely addresses a node or application
- **Ground_Station_Catalog**: Onboard database of known Tier 3/4 ground station locations used by the CGR_Engine for pass prediction
- **Pool_Allocator**: Static/pool-based memory allocation system used by the Firmware — fixed-size block pools with no dynamic heap allocation
- **Ephemeris**: Pre-computed position/velocity table for the cislunar payload's orbit, used by the CGR_Engine for contact prediction instead of SGP4/SDP4
- **Light_Time_Delay**: One-way signal propagation delay between Earth and the cislunar payload — 1–2 seconds at Earth–Moon distance (~384,000 km)
- **Link_Margin**: The excess signal strength above the minimum required for reliable demodulation — approximately 7 dB for the cislunar S-band link at 500 bps
- **Tier_3_Station**: Deep-space amateur or institutional ground station with a 3–5m dish, S-band capability, and low-noise front end
- **Tier_4_Station**: University or partner ground station with a large dish providing backbone support for cislunar operations
- **Stop_2_Mode**: Ultra-low-power sleep mode with SRAM retention, used between contact arcs
- **Radiation_Monitor**: Subsystem that protects critical SRAM data structures against radiation-induced single-event upsets (SEUs) — enhanced for the cislunar radiation environment beyond the Van Allen belts

## Requirements

### Requirement 1: S-Band IQ Transceiver Interface

**User Story:** As a flight systems engineer, I want the OBC to interface directly with the flight-qualified S-band IQ transceiver IC at 2.2 GHz, so that the cislunar payload transmits and receives at S-band frequencies suitable for Earth–Moon distances.

#### Acceptance Criteria

1. THE Firmware SHALL generate TX IQ baseband samples on the OBC using BPSK modulation at 500 bps and stream them directly to the Flight_Transceiver via DAC/ADC or SPI
2. THE Firmware SHALL receive RX IQ baseband samples directly from the Flight_Transceiver via ADC or SPI and demodulate them on the OBC using BPSK demodulation at 500 bps
3. THE Firmware SHALL use the OBC DMA engine for IQ sample streaming between memory and the Flight_Transceiver peripheral interface, avoiding CPU-bound sample transfers
4. THE CLA SHALL interface with the Flight_Transceiver IQ path (OBC DMA → DAC/ADC or SPI → Flight_Transceiver) for S-band 2.2 GHz operation
5. THE Firmware SHALL configure the Flight_Transceiver for S-band 2.2 GHz center frequency with sufficient bandwidth to support 500 bps BPSK modulation
6. THE Firmware SHALL manage IQ sample buffers within the OBC SRAM budget, sharing memory with the ION-DTN runtime, Bundle_Store index, CGR_Engine state, and TrustZone secure world
7. FOR ALL valid AX.25 frames, modulating a frame into IQ samples via the S-band Flight_Transceiver path and then demodulating the IQ samples back SHALL produce a frame equivalent to the original (round-trip property for the S-band baseband DSP path)

### Requirement 2: LDPC/Turbo Forward Error Correction

**User Story:** As a communications engineer, I want the Firmware to apply strong LDPC or Turbo FEC encoding and decoding on the S-band link, so that the cislunar payload achieves reliable communication at 500 bps with only 7 dB link margin.

#### Acceptance Criteria

1. THE Firmware SHALL apply LDPC or Turbo FEC encoding to all transmitted data before BPSK modulation on the S-band link
2. THE Firmware SHALL apply LDPC or Turbo FEC decoding to all received data after BPSK demodulation on the S-band link
3. THE Firmware SHALL support a coding rate that achieves a bit error rate of 1e-5 or better at an Eb/N0 of 2 dB or less (consistent with the 7 dB link margin budget at 500 bps)
4. THE Firmware SHALL perform FEC encoding and decoding within the OBC SRAM budget, sharing memory with the ION-DTN runtime and IQ buffers
5. FOR ALL valid data blocks, encoding a block with LDPC or Turbo FEC and then decoding the encoded block (without channel errors) SHALL produce a block identical to the original (round-trip property for the FEC codec)

### Requirement 3: LTP Deferred Acknowledgment for Cislunar Delay

**User Story:** As a protocol engineer, I want LTP to handle 1–2 second one-way light-time delays gracefully with deferred acknowledgment, so that the DTN convergence layer operates correctly at cislunar distances.

#### Acceptance Criteria

1. THE LTP engine SHALL support configurable retransmission timers that account for the 2–4 second round-trip time at Earth–Moon distance (1–2 second one-way light-time delay plus processing time)
2. THE LTP engine SHALL use deferred acknowledgment, allowing the sender to continue transmitting new segments during the acknowledgment delay without blocking
3. THE LTP engine SHALL maintain session state for the duration of the round-trip delay plus a configurable margin (default 10 seconds total session timeout for cislunar)
4. THE LTP engine SHALL support concurrent LTP sessions to maximize throughput during hours-long Contact_Windows at 500 bps
5. IF an LTP acknowledgment is not received within the configured retransmission timer, THEN THE LTP engine SHALL retransmit the unacknowledged segments up to a configurable maximum retry count (default 5 retries)

### Requirement 4: Cislunar CGR Contact Prediction

**User Story:** As a flight systems engineer, I want the CGR_Engine to predict contact windows for cislunar orbits using numerical propagation or pre-computed ephemeris tables, so that the payload autonomously schedules hours-long communication arcs with Tier 3/4 ground stations.

#### Acceptance Criteria

1. THE CGR_Engine SHALL compute predicted Contact_Windows between the cislunar payload and ground stations listed in the Ground_Station_Catalog using numerical orbit propagation or pre-computed ephemeris tables suitable for cislunar orbital mechanics
2. THE CGR_Engine SHALL predict contact arcs with durations ranging from minutes to hours, reflecting cislunar orbital dynamics (not the 5–10 minute LEO passes of Phase 3)
3. THE CGR_Engine SHALL compute for each predicted contact: start time, end time, maximum elevation angle, estimated maximum Doppler shift at 2.2 GHz, and estimated one-way light-time delay
4. THE CGR_Engine SHALL re-compute predicted Contact_Windows when fresh ephemeris data is received during a ground pass
5. THE CGR_Engine SHALL maintain a prediction horizon of at least 48 hours of future Contact_Windows (extended from Phase 3's 24 hours due to longer orbital periods)
6. THE CGR_Engine SHALL filter predicted contacts by a configurable minimum elevation angle (default 5 degrees) to exclude low-elevation contacts with poor link quality
7. THE CGR_Engine SHALL execute contact prediction computations within the OBC SRAM budget using the Pool_Allocator
8. THE CGR_Engine SHALL use CGR exclusively for contact prediction and pass scheduling — the CGR_Engine SHALL NOT compute multi-hop relay routes

### Requirement 5: Cislunar Ephemeris Management

**User Story:** As a ground station operator, I want to upload fresh ephemeris data to the cislunar payload during contact arcs, so that contact predictions remain accurate as the orbit evolves.

#### Acceptance Criteria

1. WHEN fresh ephemeris data is received via a DTN bundle during a contact arc, THE Firmware SHALL validate the ephemeris format and epoch, update the onboard orbital parameters, and trigger the CGR_Engine to re-predict future Contact_Windows
2. THE Firmware SHALL persist the current ephemeris data to NVM so that orbital parameters survive power cycles and watchdog resets
3. IF received ephemeris data fails format validation or has an epoch older than the currently stored ephemeris, THEN THE Firmware SHALL reject the update and log the rejection reason
4. THE Firmware SHALL track the age of the current ephemeris data and include the ephemeris epoch in telemetry reports so ground operators can determine when an ephemeris update is needed
5. WHEN ephemeris data age exceeds a configurable threshold (default 7 days), THE Firmware SHALL flag an ephemeris-stale warning in telemetry and widen the Contact_Window margins by a configurable factor to compensate for reduced prediction accuracy

### Requirement 6: Cislunar Doppler Compensation

**User Story:** As a communications engineer, I want the Firmware to compensate for Doppler frequency shift at S-band 2.2 GHz during cislunar contact arcs, so that the IQ baseband demodulator maintains lock over hours-long passes with slowly varying Doppler.

#### Acceptance Criteria

1. THE Firmware SHALL compute the expected Doppler shift at 2.2 GHz based on the predicted contact geometry from the CGR_Engine (payload position and velocity relative to the ground station)
2. THE Firmware SHALL apply Doppler_Compensation to the RX IQ baseband processing, adjusting the demodulator center frequency to track the predicted Doppler profile during each contact arc
3. THE Firmware SHALL apply Doppler_Compensation to the TX IQ baseband processing, pre-compensating the transmitted frequency so the ground station receives the signal at the nominal center frequency
4. THE Firmware SHALL support a Doppler range of at least plus or minus 5 kHz at 2.2 GHz (corresponding to cislunar orbital velocities)
5. THE Firmware SHALL update the Doppler compensation at a rate sufficient to track the Doppler rate of change during a cislunar contact arc (at least once per 10 seconds, reflecting the slower Doppler dynamics compared to LEO)

### Requirement 7: BPv7 Bundle Creation and Validation

**User Story:** As a ground station operator, I want the cislunar payload to create and validate BPv7 bundles within its constrained resources, so that the flight DTN stack operates correctly at cislunar distances.

#### Acceptance Criteria

1. WHEN a message with a valid destination Endpoint_ID and payload is submitted, THE BPA SHALL create a BPv7 bundle with the bundle version set to 7, proper source and destination Endpoint_IDs, a CRC integrity check, a priority level (critical, expedited, normal, or bulk), and a positive lifetime value in seconds
2. WHEN a bundle is received via the CLA, THE BPA SHALL validate that the bundle version equals 7, the destination is a well-formed Endpoint_ID, the lifetime is greater than zero, the creation timestamp does not exceed the current time, and the CRC is correct
3. IF a received bundle fails any validation check, THEN THE BPA SHALL discard the bundle and log the specific validation failure reason along with the source Endpoint_ID
4. THE BPA SHALL support three bundle types: data bundles for store-and-forward payload delivery, ping request bundles for echo requests, and ping response bundles for echo responses
5. FOR ALL valid Bundle objects, serializing a Bundle to its BPv7 wire format (CBOR) and then parsing the wire format back SHALL produce a Bundle equivalent to the original (round-trip property)
6. THE BPA SHALL complete bundle creation and validation using a working memory allocation that fits within the OBC SRAM budget shared with IQ buffers, FEC codec state, CGR_Engine state, and the ION-DTN runtime

### Requirement 8: NVM Bundle Storage and Persistence (256 MB–1 GB)

**User Story:** As a flight systems engineer, I want bundles to be stored persistently on expanded external NVM (256 MB–1 GB) and survive power cycles, watchdog resets, and radiation-induced resets, so that no messages are lost during the long-duration cislunar mission.

#### Acceptance Criteria

1. WHEN a valid bundle is accepted, THE Bundle_Store SHALL persist the bundle to external SPI/QSPI NVM atomically, preventing corruption if power is lost during the write
2. WHEN a bundle is stored and later retrieved by its bundle ID (source Endpoint_ID, creation timestamp, sequence number), THE Bundle_Store SHALL return a bundle identical to the original (round-trip property)
3. THE Bundle_Store SHALL maintain a priority-ordered index so that bundles are retrieved in priority order: critical first, then expedited, then normal, then bulk
4. WHEN the Bundle_Store reaches the configured maximum NVM capacity and a new bundle arrives, THE Bundle_Store SHALL evict expired bundles first, then the lowest-priority bundles with the earliest creation timestamps, to free sufficient space for the new bundle
5. WHEN evicting bundles, THE Bundle_Store SHALL preserve all critical-priority bundles until all expedited, normal, and bulk bundles have been evicted
6. THE Bundle_Store SHALL enforce that total stored bytes do not exceed the configured maximum NVM capacity (256 MB–1 GB)
7. WHEN the OBC restarts after a power cycle, watchdog reset, or radiation-induced reset, THE Bundle_Store SHALL reload its persisted state from external NVM and validate store integrity via CRC checks on each stored bundle
8. IF store corruption is detected during reload, THEN THE Bundle_Store SHALL rebuild from intact bundles only, discarding corrupted entries and logging each discarded bundle ID

### Requirement 9: Bundle Lifetime Enforcement

**User Story:** As a flight systems engineer, I want expired bundles to be automatically removed from NVM, so that stale data does not consume limited storage capacity during the long-duration cislunar mission.

#### Acceptance Criteria

1. WHEN the Firmware runs a cleanup cycle, THE Bundle_Store SHALL delete all bundles from NVM whose creation timestamp plus lifetime is less than or equal to the current time
2. THE Bundle_Store SHALL contain zero expired bundles after a cleanup cycle completes

### Requirement 10: Cislunar Ping Operation

**User Story:** As a ground station operator, I want to ping the cislunar payload during a contact arc and receive an echo response, so that I can verify end-to-end DTN reachability across Earth–Moon distance with measured round-trip time.

#### Acceptance Criteria

1. WHEN the BPA receives a ping request bundle addressed to a local endpoint, THE BPA SHALL generate exactly one ping response bundle with the destination set to the original sender's Endpoint_ID
2. WHEN a ping response is generated, THE BPA SHALL queue the response in the Bundle_Store for delivery during the current Contact_Window if the sender's ground station is still in view, or during the next predicted Contact_Window with the sender's station
3. WHEN a ping response is received at the originating ground station, the ground station SHALL compute and report the round-trip time from the original ping request creation timestamp to the response receipt time (expected 2–4 seconds at cislunar distance)
4. THE BPA SHALL include the original ping request's bundle ID in the ping response payload so the originating station can correlate responses to requests

### Requirement 11: Cislunar Store-and-Forward

**User Story:** As a ground station operator, I want to send messages that are stored on the cislunar payload NVM and delivered to the destination ground station when the payload has a contact arc with that station, so that store-and-forward messaging works across Earth–Moon distance.

#### Acceptance Criteria

1. WHEN a data bundle is received whose destination matches a local Endpoint_ID on the cislunar payload, THE BPA SHALL deliver the bundle payload to the local application agent
2. WHEN a data bundle is received whose destination is a remote Endpoint_ID (a ground station), THE BPA SHALL store the bundle in the NVM-backed Bundle_Store and queue it for direct delivery during the next predicted Contact_Window with the destination ground station
3. THE BPA SHALL transmit queued bundles in priority order (critical first, then expedited, then normal, then bulk) during each Contact_Window
4. WHEN a transmitted bundle is acknowledged by the remote ground station via LTP (accounting for the 2–4 second round-trip acknowledgment delay), THE Bundle_Store SHALL delete the acknowledged bundle from NVM
5. IF a bundle transmission is not acknowledged within the LTP retransmission timeout (configured for cislunar delay), THEN THE Bundle_Store SHALL retain the bundle for retry during the current Contact_Window (if time remains) or the next Contact_Window with the destination station
6. IF no predicted Contact_Window exists for a bundle's destination ground station (station not in Ground_Station_Catalog or no future contacts predicted), THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window becomes available or the bundle's lifetime expires

### Requirement 12: No Relay Constraint

**User Story:** As a system architect, I want to enforce that the cislunar payload only delivers bundles directly to their final destination ground station, so that the system remains consistent with the no-relay architecture across all phases.

#### Acceptance Criteria

1. THE BPA SHALL transmit a bundle only to the node matching the bundle's final destination Endpoint_ID — the BPA SHALL NOT forward bundles on behalf of other nodes
2. WHEN the Node_Controller looks up a delivery route for a bundle, THE Contact_Plan_Manager SHALL return only direct Contact_Windows with the destination node, with no multi-hop paths

### Requirement 13: AX.25 and LTP Convergence Layer (S-Band)

**User Story:** As a regulatory compliance engineer, I want all DTN transmissions to use AX.25 framing with callsign addressing over LTP via the S-band IQ baseband radio, so that every cislunar transmission complies with amateur radio regulations.

#### Acceptance Criteria

1. THE CLA SHALL encapsulate all bundle transmissions in AX.25 frames carrying the source amateur radio callsign and the destination amateur radio callsign
2. THE CLA SHALL run LTP sessions on top of AX.25 frames, providing reliable transfer with deferred acknowledgment for all bundle delivery
3. THE CLA SHALL perform LTP segmentation for bundles that exceed a single AX.25 frame size, and reassemble received LTP segments into complete bundles
4. THE CLA SHALL interface with the Flight_Transceiver IQ path (OBC DMA → DAC/ADC or SPI → Flight_Transceiver) as the physical transport at S-band 2.2 GHz
5. FOR ALL valid Bundle objects, encapsulating a bundle into AX.25/LTP frames, encoding with LDPC/Turbo FEC, modulating to IQ via the S-band Flight_Transceiver, demodulating from IQ, decoding FEC, and reassembling the frames back into a bundle SHALL produce a bundle equivalent to the original (end-to-end round-trip property)

### Requirement 14: Autonomous Contact Arc Execution

**User Story:** As a flight systems engineer, I want the cislunar payload to autonomously wake from sleep, activate the S-band radio, execute bundle transfers during predicted contact arcs (potentially hours long), and return to sleep, so that the payload operates independently without ground operator intervention.

#### Acceptance Criteria

1. WHEN a CGR-predicted Contact_Window start time is reached, THE Firmware SHALL wake the OBC from Stop_2_Mode via RTC alarm, initialize the Flight_Transceiver and IQ DSP with LDPC/Turbo FEC, and begin processing queued bundles for the contact's ground station
2. THE Node_Controller SHALL transmit queued bundles destined for the contact's ground station in priority order during the Contact_Window
3. THE Node_Controller SHALL cease all transmission when the Contact_Window end time is reached
4. WHEN a Contact_Window completes, THE Node_Controller SHALL record link metrics (bytes transferred, duration, bundles sent, bundles received, signal quality, Doppler tracking accuracy, FEC decode statistics) and update contact statistics
5. WHEN a Contact_Window completes and no further Contact_Windows are predicted within the next 5 minutes, THE Firmware SHALL deactivate the Flight_Transceiver, flush NVM, and transition the OBC into Stop_2_Mode
6. IF the CLA fails to establish the S-band IQ baseband link during a scheduled Contact_Window (Flight_Transceiver not responding, no AX.25 connection established, or signal quality below threshold), THEN THE Node_Controller SHALL mark the contact as missed, retain all queued bundles for the next window, and increment the contacts-missed counter

### Requirement 15: BPSec Integrity with Hardware Crypto

**User Story:** As a security engineer, I want bundle integrity protection using BPSec with the OBC hardware crypto accelerator, so that bundle origin authentication is enforced at cislunar distances while complying with amateur radio regulations.

#### Acceptance Criteria

1. THE BPA SHALL support BPSec (RFC 9172) Block Integrity Blocks (BIB) for bundle origin authentication using HMAC-SHA-256
2. THE BPA SHALL NOT apply BPSec Block Confidentiality Blocks (BCB) or any form of payload encryption, in compliance with amateur radio regulations requiring transmissions to be unencrypted
3. THE BPA SHALL use the OBC hardware crypto accelerator for all BPSec HMAC-SHA-256 computations instead of software implementations
4. WHEN a bundle with a BIB is received, THE BPA SHALL verify the integrity block using the hardware crypto accelerator and discard the bundle if verification fails, logging the integrity failure with the source Endpoint_ID
5. THE Firmware SHALL store BPSec shared keys in the OBC TrustZone secure world, isolated from non-secure application code

### Requirement 16: TrustZone Secure Key Storage

**User Story:** As a security engineer, I want cryptographic keys stored in the OBC TrustZone secure world, so that keys are hardware-isolated from application firmware in the cislunar flight environment.

#### Acceptance Criteria

1. THE Firmware SHALL partition the OBC into TrustZone secure and non-secure worlds, with BPSec keys and crypto operations executing in the secure world
2. THE Firmware SHALL expose a secure API from the TrustZone secure world that the non-secure BPA can call to request HMAC-SHA-256 signing and verification without exposing raw key material
3. IF non-secure code attempts to read TrustZone secure memory directly, THEN THE OBC SHALL generate a hardware fault and the Firmware SHALL log the access violation
4. THE Firmware SHALL provision BPSec keys into the TrustZone secure world during initial firmware flashing or via a secure key update bundle received during a contact arc

### Requirement 17: Power Management (10–20 W Budget)

**User Story:** As a power systems engineer, I want the cislunar payload to manage its power budget autonomously within 10–20 W, transitioning between active and ultra-low-power states based on predicted contact arcs, so that the payload operates within its power budget during the long-duration mission.

#### Acceptance Criteria

1. WHEN no Contact_Window is active and no bundle processing is pending, THE Firmware SHALL transition the OBC into Stop_2_Mode
2. WHILE in Stop_2_Mode, THE OBC SHALL consume no more than 20 µA (allowing margin above the nominal specification)
3. WHEN a Contact_Window start time is reached, THE Firmware SHALL wake the OBC from Stop_2_Mode via RTC alarm and resume normal operation within 10 milliseconds
4. THE Node_Controller SHALL log timestamped power state transitions (active → Stop_2_Mode, Stop_2_Mode → active) and the duration spent in each state for power budget analysis
5. THE Firmware SHALL set the RTC alarm for the next predicted Contact_Window start time before entering Stop_2_Mode, ensuring autonomous wake-up without external triggers
6. WHILE a Contact_Window is active, THE Firmware SHALL operate the S-band Flight_Transceiver, LDPC/Turbo FEC codec, and IQ DSP within the 10–20 W power envelope

### Requirement 18: SRAM Memory Management

**User Story:** As an embedded systems engineer, I want the OBC firmware to operate within the SRAM constraint while concurrently running ION-DTN, IQ baseband DSP, LDPC/Turbo FEC codec, CGR contact prediction, and bundle index management, so that the flight memory budget is validated.

#### Acceptance Criteria

1. THE Firmware SHALL operate within the OBC SRAM for all concurrent operations: ION-DTN runtime, IQ sample buffers (TX and RX), LDPC/Turbo FEC codec state, AX.25/LTP frame buffers, bundle metadata index, CGR_Engine state and computation buffers, and TrustZone secure world allocations
2. THE Firmware SHALL use static or pool-based memory allocation (Pool_Allocator) for all runtime data structures, avoiding dynamic heap allocation that could cause fragmentation on the constrained MCU
3. THE Firmware SHALL report peak and current SRAM utilization as part of telemetry, broken down by subsystem (ION-DTN, IQ buffers, FEC codec, bundle index, CGR_Engine, TrustZone)
4. IF an operation would exceed the SRAM budget, THEN THE Firmware SHALL reject the operation and log the memory exhaustion event rather than corrupting adjacent memory regions

### Requirement 19: Priority-Based Message Handling

**User Story:** As a ground station operator, I want bundles to be handled according to their priority level on the cislunar payload, so that critical messages are delivered before less urgent ones during contact arcs.

#### Acceptance Criteria

1. THE BPA SHALL assign one of four priority levels to each bundle: critical (highest), expedited, normal, or bulk (lowest)
2. WHEN multiple bundles are queued for the same destination during a Contact_Window, THE Firmware SHALL transmit bundles in strict priority order — all critical bundles before any expedited, all expedited before any normal, all normal before any bulk
3. WHEN the Bundle_Store must evict bundles to free NVM space, THE Bundle_Store SHALL evict bulk bundles first, then normal, then expedited — critical bundles SHALL be evicted only when no lower-priority bundles remain
4. THE BPA SHALL accept a default priority level from the Firmware configuration, applied to bundles that do not specify an explicit priority

### Requirement 20: Rate Limiting and Store Protection

**User Story:** As a flight systems engineer, I want to protect the NVM bundle store from flooding by a misbehaving ground station, so that storage resources are not exhausted during contact arcs.

#### Acceptance Criteria

1. THE BPA SHALL enforce a configurable maximum bundle acceptance rate (bundles per second) per source Endpoint_ID
2. IF the acceptance rate from a single source Endpoint_ID exceeds the configured limit, THEN THE BPA SHALL reject additional bundles from that source and log the rate-limit event
3. THE BPA SHALL enforce a configurable maximum bundle size in bytes, rejecting any bundle whose total serialized size exceeds the limit

### Requirement 21: Node Health and Telemetry

**User Story:** As a ground station operator, I want to receive telemetry from the cislunar payload during contact arcs, so that I can monitor the health and operational status of the deep-space node.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report telemetry including: uptime in seconds, NVM storage utilization as a percentage of configured maximum, number of bundles currently stored, number of bundles delivered, number of bundles dropped (expired or evicted), and the timestamp of the last completed contact
2. THE Node_Controller SHALL track cumulative statistics including: total bundles received, total bundles sent, total bytes received, total bytes sent, average delivery latency in seconds, contacts completed, and contacts missed
3. THE Firmware SHALL report OBC-specific telemetry including: current SRAM utilization (peak and current, by subsystem), power state (active or Stop_2_Mode), time spent in each power state, MCU temperature (internal sensor), IQ baseband signal quality metrics (SNR, bit error rate), FEC decode statistics (corrected errors, uncorrectable frames), current ephemeris epoch age, CGR prediction horizon, and radiation event counters
4. THE Node_Controller SHALL package telemetry as a DTN bundle and transmit it to requesting ground stations during Contact_Windows
5. WHEN a telemetry request bundle is received, THE Node_Controller SHALL generate a telemetry response bundle within 1 second

### Requirement 22: Autonomous Error Handling and Fault Recovery

**User Story:** As a flight systems engineer, I want the cislunar payload to handle faults autonomously including power loss, watchdog resets, radiation-induced upsets, transceiver failures, and memory corruption, so that the payload recovers and resumes operation without ground intervention.

#### Acceptance Criteria

1. IF the Bundle_Store reaches NVM capacity and eviction cannot free sufficient space for an incoming bundle, THEN THE BPA SHALL reject the incoming bundle and return a storage-full error to the sender if the LTP session is still active
2. IF a CRC validation fails on a received bundle, THEN THE BPA SHALL discard the corrupted bundle and log the corruption event with the source Endpoint_ID and IQ link metrics
3. IF the OBC experiences a power cycle, watchdog reset, or radiation-induced reset, THEN THE Firmware SHALL reload the Bundle_Store from external NVM, reload the ephemeris data and Ground_Station_Catalog from NVM, re-compute CGR contact predictions, and resume autonomous operation without ground intervention
4. IF the Flight_Transceiver becomes unresponsive during a Contact_Window, THEN THE Firmware SHALL attempt to reinitialize the Flight_Transceiver up to 3 times with a 1-second interval, and if all attempts fail, mark the contact as missed, retain all queued bundles, and enter Stop_2_Mode until the next predicted Contact_Window
5. IF no direct Contact_Window exists for a bundle's destination, THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window with that destination becomes available or the bundle's lifetime expires
6. THE Firmware SHALL use a hardware watchdog timer with a configurable timeout (default 30 seconds) to detect firmware hangs and trigger an automatic reset

### Requirement 23: Enhanced Radiation Tolerance (Cislunar Environment)

**User Story:** As a flight systems engineer, I want the Firmware to detect and mitigate radiation-induced errors with enhanced tolerance for the cislunar radiation environment beyond the Van Allen belts, so that the payload maintains data integrity during the deep-space mission.

#### Acceptance Criteria

1. THE Firmware SHALL protect critical data structures in SRAM (bundle metadata index, CGR state, contact plan, ephemeris data, FEC codec state) using CRC and redundant copies so that single-bit upsets can be detected
2. WHEN a CRC mismatch or redundancy inconsistency is detected in a critical SRAM data structure, THE Firmware SHALL attempt to recover from the redundant copy or reload from NVM, and log the radiation event
3. THE Firmware SHALL validate NVM data integrity via CRC on every read operation, detecting corruption from radiation-induced bit flips in flash memory
4. THE Firmware SHALL include a radiation event counter in telemetry reports, tracking detected single-event upsets (SEUs) in SRAM and NVM
5. THE Firmware SHALL perform SRAM integrity validation at a higher frequency than Phase 3 (at least once per minute during active operation and once per wake cycle during sleep), reflecting the elevated radiation flux in the cislunar environment
6. THE Firmware SHALL implement triple modular redundancy (TMR) for the most critical control variables (contact plan active flag, current contact index, power state) to tolerate single-event upsets without requiring recovery from redundant copies

### Requirement 24: Onboard Time Management

**User Story:** As a flight systems engineer, I want the cislunar payload to maintain accurate onboard time, so that CGR contact predictions, bundle timestamps, LTP session timers, and contact arc scheduling are correct at cislunar distances.

#### Acceptance Criteria

1. THE Firmware SHALL maintain onboard time using the OBC RTC, synchronized to UTC
2. WHEN a time synchronization bundle is received from a ground station during a contact arc, THE Firmware SHALL update the RTC to the received UTC time if the correction exceeds a configurable threshold (default 1 second)
3. THE Firmware SHALL include the current onboard time and the time since last synchronization in telemetry reports
4. IF the RTC has not been synchronized for more than a configurable period (default 7 days), THE Firmware SHALL flag a time-stale warning in telemetry

### Requirement 25: Autonomous Operation Cycle

**User Story:** As a flight systems engineer, I want the cislunar payload to execute a fully autonomous operation cycle (predict contact arcs, wake, communicate, sleep) without any ground operator control, so that the payload operates independently between contact arcs.

#### Acceptance Criteria

1. THE Node_Controller SHALL execute the following autonomous cycle: compute next Contact_Window from CGR predictions, set RTC alarm for the Contact_Window start time, enter Stop_2_Mode, wake on RTC alarm, activate Flight_Transceiver and IQ DSP with LDPC/Turbo FEC, execute bundle transfers during the Contact_Window, deactivate Flight_Transceiver, run bundle cleanup (expire, evict), update CGR predictions if needed, and repeat
2. THE Node_Controller SHALL complete a full operation cycle iteration (check contacts, transmit queued bundles via IQ baseband, process received bundles, run cleanup) within 2 seconds on the OBC (relaxed from Phase 3's 1 second to accommodate FEC processing overhead)
3. THE Node_Controller SHALL operate indefinitely without ground intervention, using onboard ephemeris data and the Ground_Station_Catalog for all scheduling decisions
4. WHEN the Firmware boots after a reset, THE Node_Controller SHALL restore state from NVM (Bundle_Store, ephemeris, Ground_Station_Catalog, contact statistics) and resume the autonomous operation cycle within 5 seconds

### Requirement 26: Ground Station Catalog (Tier 3/4)

**User Story:** As a flight systems engineer, I want the cislunar payload to maintain an onboard catalog of Tier 3/4 ground station locations, so that the CGR_Engine can predict contact arcs with stations capable of S-band cislunar communication.

#### Acceptance Criteria

1. THE Firmware SHALL store a Ground_Station_Catalog in NVM containing for each station: station identifier (callsign-based), geodetic latitude in degrees, geodetic longitude in degrees, altitude above WGS84 ellipsoid in meters, minimum elevation angle in degrees, and antenna gain in dBi
2. THE Firmware SHALL support a Ground_Station_Catalog of at least 32 ground stations
3. WHEN a catalog update bundle is received during a contact arc, THE Firmware SHALL validate the entry format and add or update the specified ground station entry in the catalog
4. THE Firmware SHALL persist the Ground_Station_Catalog to NVM so that it survives power cycles and watchdog resets
5. WHEN a ground station entry is added or updated, THE CGR_Engine SHALL re-predict Contact_Windows for the affected station
