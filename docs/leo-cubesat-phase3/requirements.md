# Requirements Document

## Introduction

This document specifies the requirements for Phase 3 of the cislunar amateur DTN project: LEO CubeSat Flight. Phase 3 takes the validated Phase 2 Engineering Model (EM) software and deploys it on the actual LEO CubeSat in orbit. The STM32U585 OBC is identical to Phase 2, running the same ION-DTN (BPv7/LTP over AX.25) firmware — C for the DTN/radio/DSP stack, with no companion host.

The key architectural change from Phase 2 is the elimination of the companion host and B200mini SDR. A flight-qualified IQ transceiver IC interfaces directly with the STM32U585 via DAC/ADC or SPI, replacing the B200mini + companion RPi/PC IQ bridge. The STM32U585 runs everything autonomously: ION-DTN BPA, LTP, AX.25 CLA, IQ baseband DSP, NVM bundle store, power management, TrustZone secure crypto, and CGR contact prediction. There is no ground operator controlling pass timing — the CubeSat predicts its own contact windows using ION-DTN's CGR module with SGP4/SDP4 orbit propagation from onboard TLE/ephemeris data.

The system operates at UHF 437 MHz at 9.6 kbps (GMSK/BPSK via IQ baseband). External SPI/QSPI NVM (64–256 MB) provides persistent bundle storage. Contact windows are 5–10 minutes per pass, 4–6 passes per day per ground station. Power budget is 5–10 W average active, with STM32U585 Stop 2 mode (~16 µA) between passes. BPSec provides integrity protection (HMAC-SHA-256 via hardware crypto, no encryption).

The system supports two core operations: ping (DTN reachability test) and store-and-forward (point-to-point bundle delivery). There is no relay functionality. All bundle delivery is direct (source → destination).

Phase 2 (EM) is complete and provides the validated STM32U585 firmware: ION-DTN BPA, LTP, AX.25 CLA plugin architecture, IQ baseband DSP, NVM bundle store with atomic writes, TrustZone secure key storage, hardware crypto BPSec, Stop 2 power management, static/pool memory allocation, ping and store-and-forward operations, and the no-relay constraint. Phase 3 adapts the CLA from the B200mini IQ bridge to the flight IQ transceiver IC direct interface, adds CGR-based contact prediction, and operates autonomously in orbit under real orbital dynamics (Doppler, varying elevation, radiation environment).

Out of scope: S-band / X-band (Phase 4), cislunar distances (Phase 4), LDPC/Turbo FEC (Phase 4), relay functionality.

## Glossary

- **STM32U585**: Ultra-low-power ARM Cortex-M33 MCU — 160 MHz, 2 MB flash, 786 KB SRAM, hardware crypto accelerator (AES-256, SHA-256, PKA), TrustZone security — the OBC for the LEO CubeSat (identical to Phase 2 EM)
- **Flight_Transceiver**: Flight-qualified IQ transceiver IC interfacing directly with the STM32U585 via DAC/ADC or SPI — replaces the Phase 2 B200mini + companion host IQ bridge
- **NVM**: External SPI/QSPI non-volatile memory (64–256 MB flash) connected to the STM32U585 for persistent bundle storage
- **BPA**: Bundle Protocol Agent — the core ION-DTN engine running on the STM32U585 that creates, receives, validates, stores, and delivers BPv7 bundles
- **Bundle_Store**: Persistent storage subsystem backed by external NVM for bundles awaiting delivery
- **CGR_Engine**: ION-DTN's Contact Graph Routing module running on the STM32U585, used exclusively for contact prediction (pass scheduling via SGP4/SDP4 orbit propagation) — not for multi-hop relay routing
- **Contact_Plan_Manager**: Subsystem that maintains CGR-predicted communication windows and manages contact scheduling autonomously onboard the CubeSat
- **CLA**: Convergence Layer Adapter — native ION-DTN CLA plugin running on the STM32U585 that provides AX.25 framing as the LTP link service layer, adapted from Phase 2 for the Flight_Transceiver direct interface (no companion host)
- **Node_Controller**: Top-level autonomous orchestrator running on the STM32U585 — manages the operation cycle (wake, transmit, receive, sleep) without external control
- **Firmware**: C code running on the STM32U585 (bare metal or lightweight RTOS) implementing ION-DTN BPv7/LTP, AX.25 CLA, IQ baseband DSP, NVM bundle store, CGR contact prediction, power management, and TrustZone secure crypto
- **ION-DTN**: NASA JPL's Interplanetary Overlay Network — the DTN implementation providing BPv7, LTP, CGR, and related protocols, cross-compiled for STM32U585
- **LTP**: Licklider Transmission Protocol — runs on top of AX.25 providing reliable transfer with deferred acknowledgment
- **AX.25**: Link-layer framing protocol providing callsign-based source/destination addressing for amateur radio compliance
- **BPSec**: Bundle Protocol Security (RFC 9172) — provides integrity blocks (HMAC-SHA-256) for bundle origin authentication
- **TrustZone**: ARM TrustZone hardware isolation on the STM32U585 — partitions the MCU into secure and non-secure worlds for key storage and crypto operations
- **Stop_2_Mode**: STM32U585 ultra-low-power sleep mode (~16 µA) with SRAM retention, used between orbital passes
- **DMA**: Direct Memory Access — STM32U585 peripheral for streaming IQ samples between memory and the Flight_Transceiver interface without CPU intervention
- **TLE**: Two-Line Element set — standard orbital parameter format used by SGP4/SDP4 propagators for orbit prediction
- **SGP4_SDP4**: Simplified General Perturbations / Simplified Deep-space Perturbations orbit propagation models used by the CGR_Engine to compute satellite position and predict ground station passes
- **Doppler_Compensation**: Frequency offset correction applied by the Firmware to account for relative velocity between the CubeSat and ground stations during orbital passes
- **Contact_Window**: A predicted or actual time interval during which the CubeSat has line-of-sight communication with a ground station
- **Ping**: DTN reachability test — send a bundle echo request and receive an echo response
- **Store_and_Forward**: Point-to-point bundle delivery where a source node sends a bundle directly to a destination node during a Contact_Window
- **Endpoint_ID**: A DTN endpoint identifier using the "dtn" or "ipn" URI scheme that uniquely addresses a node or application
- **Ground_Station_Catalog**: Onboard database of known ground station locations (latitude, longitude, altitude, minimum elevation angle) used by the CGR_Engine for pass prediction
- **Pool_Allocator**: Static/pool-based memory allocation system used by the Firmware — fixed-size block pools with no dynamic heap allocation

## Requirements

### Requirement 1: Flight Transceiver Direct IQ Interface

**User Story:** As a flight systems engineer, I want the STM32U585 to interface directly with the flight-qualified IQ transceiver IC via DAC/ADC or SPI, eliminating the companion host and B200mini, so that the CubeSat operates as a single autonomous unit in orbit.

#### Acceptance Criteria

1. THE Firmware SHALL generate TX IQ baseband samples on the STM32U585 using GMSK/BPSK modulation at 9.6 kbps and stream them directly to the Flight_Transceiver via DAC/ADC or SPI
2. THE Firmware SHALL receive RX IQ baseband samples directly from the Flight_Transceiver via ADC or SPI and demodulate them on the STM32U585 using GMSK/BPSK demodulation at 9.6 kbps
3. THE Firmware SHALL use the STM32U585 DMA engine for IQ sample streaming between memory and the Flight_Transceiver peripheral interface, avoiding CPU-bound sample transfers
4. THE CLA SHALL interface with the Flight_Transceiver IQ path (STM32U585 DMA → DAC/ADC or SPI → Flight_Transceiver) instead of the Phase 2 IQ_Bridge path (STM32U585 → SPI/UART → Companion_Host → USB 3.0 → B200mini)
5. THE Firmware SHALL configure the Flight_Transceiver for UHF 437 MHz center frequency with sufficient bandwidth to support 9.6 kbps GMSK/BPSK modulation
6. THE Firmware SHALL manage IQ sample buffers within the STM32U585 786 KB SRAM budget, sharing memory with the ION-DTN runtime, Bundle_Store index, CGR_Engine state, and TrustZone secure world
7. FOR ALL valid AX.25 frames, modulating a frame into IQ samples via the Flight_Transceiver path and then demodulating the IQ samples back SHALL produce a frame equivalent to the original (round-trip property for the flight baseband DSP path)

### Requirement 2: Autonomous CGR Contact Prediction

**User Story:** As a flight systems engineer, I want the CubeSat to autonomously predict its own ground station pass windows using onboard orbital parameters and SGP4/SDP4 propagation, so that the CubeSat schedules contact windows without ground operator intervention.

#### Acceptance Criteria

1. THE CGR_Engine SHALL compute predicted Contact_Windows between the CubeSat and ground stations listed in the Ground_Station_Catalog using SGP4/SDP4 orbit propagation from onboard TLE data
2. THE CGR_Engine SHALL predict pass windows with a time accuracy of 30 seconds or better for TLE data that is less than 7 days old
3. THE CGR_Engine SHALL compute for each predicted pass: start time, end time, maximum elevation angle, and estimated maximum Doppler shift at 437 MHz
4. THE CGR_Engine SHALL re-compute predicted Contact_Windows when fresh TLE or ephemeris data is received during a ground pass
5. THE CGR_Engine SHALL maintain a prediction horizon of at least 24 hours of future Contact_Windows
6. THE CGR_Engine SHALL filter predicted passes by a configurable minimum elevation angle (default 5 degrees) to exclude low-elevation passes with poor link quality
7. THE CGR_Engine SHALL execute contact prediction computations within the STM32U585 786 KB SRAM budget using the Pool_Allocator, sharing memory with the ION-DTN runtime and IQ buffers
8. THE CGR_Engine SHALL use CGR exclusively for contact prediction and pass scheduling — the CGR_Engine SHALL NOT compute multi-hop relay routes

### Requirement 3: Orbital Parameter Management

**User Story:** As a ground station operator, I want to upload fresh TLE/ephemeris data to the CubeSat during ground passes, so that contact predictions remain accurate as the orbit evolves.

#### Acceptance Criteria

1. WHEN fresh TLE data is received via a DTN bundle during a ground pass, THE Firmware SHALL validate the TLE format and epoch, update the onboard orbital parameters, and trigger the CGR_Engine to re-predict future Contact_Windows
2. THE Firmware SHALL persist the current TLE data to NVM so that orbital parameters survive power cycles and watchdog resets
3. IF received TLE data fails format validation or has an epoch older than the currently stored TLE, THEN THE Firmware SHALL reject the update and log the rejection reason
4. THE Firmware SHALL track the age of the current TLE data and include the TLE epoch in telemetry reports so ground operators can determine when a TLE update is needed
5. WHEN TLE data age exceeds a configurable threshold (default 14 days), THE Firmware SHALL flag a TLE-stale warning in telemetry and widen the Contact_Window margins by a configurable factor to compensate for reduced prediction accuracy

### Requirement 4: Ground Station Catalog

**User Story:** As a flight systems engineer, I want the CubeSat to maintain an onboard catalog of ground station locations, so that the CGR_Engine can predict passes over known stations.

#### Acceptance Criteria

1. THE Firmware SHALL store a Ground_Station_Catalog in NVM containing for each station: station identifier (callsign-based), geodetic latitude in degrees, geodetic longitude in degrees, altitude above WGS84 ellipsoid in meters, and minimum elevation angle in degrees
2. THE Firmware SHALL support a Ground_Station_Catalog of at least 32 ground stations
3. WHEN a catalog update bundle is received during a ground pass, THE Firmware SHALL validate the entry format and add or update the specified ground station entry in the catalog
4. THE Firmware SHALL persist the Ground_Station_Catalog to NVM so that it survives power cycles and watchdog resets
5. WHEN a ground station entry is added or updated, THE CGR_Engine SHALL re-predict Contact_Windows for the affected station

### Requirement 5: Doppler Compensation

**User Story:** As a communications engineer, I want the Firmware to compensate for Doppler frequency shift during orbital passes, so that the IQ baseband demodulator maintains lock at varying relative velocities.

#### Acceptance Criteria

1. THE Firmware SHALL compute the expected Doppler shift at 437 MHz based on the predicted pass geometry from the CGR_Engine (satellite position and velocity relative to the ground station)
2. THE Firmware SHALL apply Doppler_Compensation to the RX IQ baseband processing, adjusting the demodulator center frequency to track the predicted Doppler profile during each pass
3. THE Firmware SHALL apply Doppler_Compensation to the TX IQ baseband processing, pre-compensating the transmitted frequency so the ground station receives the signal at the nominal center frequency
4. THE Firmware SHALL support a Doppler range of at least plus or minus 10 kHz at 437 MHz (corresponding to LEO orbital velocities at 500 km altitude)
5. THE Firmware SHALL update the Doppler compensation at a rate sufficient to track the Doppler rate of change during a pass (at least once per second)

### Requirement 6: BPv7 Bundle Creation and Validation

**User Story:** As a ground station operator, I want the CubeSat to create and validate BPv7 bundles within its constrained SRAM, so that the flight DTN stack operates correctly in orbit.

#### Acceptance Criteria

1. WHEN a message with a valid destination Endpoint_ID and payload is submitted, THE BPA SHALL create a BPv7 bundle with the bundle version set to 7, proper source and destination Endpoint_IDs, a CRC integrity check, a priority level (critical, expedited, normal, or bulk), and a positive lifetime value in seconds
2. WHEN a bundle is received via the CLA, THE BPA SHALL validate that the bundle version equals 7, the destination is a well-formed Endpoint_ID, the lifetime is greater than zero, the creation timestamp does not exceed the current time, and the CRC is correct
3. IF a received bundle fails any validation check, THEN THE BPA SHALL discard the bundle and log the specific validation failure reason along with the source Endpoint_ID
4. THE BPA SHALL support three bundle types: data bundles for store-and-forward payload delivery, ping request bundles for echo requests, and ping response bundles for echo responses
5. FOR ALL valid Bundle objects, serializing a Bundle to its BPv7 wire format (CBOR) and then parsing the wire format back SHALL produce a Bundle equivalent to the original (round-trip property)
6. THE BPA SHALL complete bundle creation and validation using a working memory allocation that fits within the STM32U585 786 KB SRAM budget shared with IQ buffers, CGR_Engine state, and the ION-DTN runtime

### Requirement 7: NVM Bundle Storage and Persistence

**User Story:** As a flight systems engineer, I want bundles to be stored persistently on external SPI/QSPI NVM and survive power cycles, watchdog resets, and radiation-induced resets, so that no messages are lost during disruptions in orbit.

#### Acceptance Criteria

1. WHEN a valid bundle is accepted, THE Bundle_Store SHALL persist the bundle to external SPI/QSPI NVM atomically, preventing corruption if power is lost during the write
2. WHEN a bundle is stored and later retrieved by its bundle ID (source Endpoint_ID, creation timestamp, sequence number), THE Bundle_Store SHALL return a bundle identical to the original (round-trip property)
3. THE Bundle_Store SHALL maintain a priority-ordered index so that bundles are retrieved in priority order: critical first, then expedited, then normal, then bulk
4. WHEN the Bundle_Store reaches the configured maximum NVM capacity and a new bundle arrives, THE Bundle_Store SHALL evict expired bundles first, then the lowest-priority bundles with the earliest creation timestamps, to free sufficient space for the new bundle
5. WHEN evicting bundles, THE Bundle_Store SHALL preserve all critical-priority bundles until all expedited, normal, and bulk bundles have been evicted
6. THE Bundle_Store SHALL enforce that total stored bytes do not exceed the configured maximum NVM capacity (64–256 MB)
7. WHEN the STM32U585 restarts after a power cycle, watchdog reset, or radiation-induced reset, THE Bundle_Store SHALL reload its persisted state from external NVM and validate store integrity via CRC checks on each stored bundle
8. IF store corruption is detected during reload, THEN THE Bundle_Store SHALL rebuild from intact bundles only, discarding corrupted entries and logging each discarded bundle ID

### Requirement 8: Bundle Lifetime Enforcement

**User Story:** As a flight systems engineer, I want expired bundles to be automatically removed from NVM, so that stale data does not consume limited storage capacity in orbit.

#### Acceptance Criteria

1. WHEN the Firmware runs a cleanup cycle, THE Bundle_Store SHALL delete all bundles from NVM whose creation timestamp plus lifetime is less than or equal to the current time
2. THE Bundle_Store SHALL contain zero expired bundles after a cleanup cycle completes

### Requirement 9: Ping Operation in Orbit

**User Story:** As a ground station operator, I want to ping the LEO CubeSat during a pass and receive an echo response, so that I can verify end-to-end DTN reachability through the orbital link.

#### Acceptance Criteria

1. WHEN the BPA receives a ping request bundle addressed to a local endpoint, THE BPA SHALL generate exactly one ping response bundle with the destination set to the original sender's Endpoint_ID
2. WHEN a ping response is generated, THE BPA SHALL queue the response in the Bundle_Store for delivery during the current Contact_Window if the sender's ground station is still in view, or during the next predicted Contact_Window with the sender's station
3. WHEN a ping response is received at the originating ground station, the ground station SHALL compute and report the round-trip time from the original ping request creation timestamp to the response receipt time
4. THE BPA SHALL include the original ping request's bundle ID in the ping response payload so the originating station can correlate responses to requests

### Requirement 10: Store-and-Forward in Orbit

**User Story:** As a ground station operator, I want to send messages that are stored on the CubeSat NVM and delivered to the destination ground station when the CubeSat passes over that station, so that store-and-forward messaging works via satellite.

#### Acceptance Criteria

1. WHEN a data bundle is received whose destination matches a local Endpoint_ID on the CubeSat, THE BPA SHALL deliver the bundle payload to the local application agent
2. WHEN a data bundle is received whose destination is a remote Endpoint_ID (a ground station), THE BPA SHALL store the bundle in the NVM-backed Bundle_Store and queue it for direct delivery during the next predicted Contact_Window with the destination ground station
3. THE BPA SHALL transmit queued bundles in priority order (critical first, then expedited, then normal, then bulk) during each Contact_Window
4. WHEN a transmitted bundle is acknowledged by the remote ground station via LTP, THE Bundle_Store SHALL delete the acknowledged bundle from NVM
5. IF a bundle transmission is not acknowledged within the LTP retransmission timeout, THEN THE Bundle_Store SHALL retain the bundle for retry during the next Contact_Window with the destination station
6. IF no predicted Contact_Window exists for a bundle's destination ground station (station not in Ground_Station_Catalog or no future passes predicted), THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window becomes available or the bundle's lifetime expires

### Requirement 11: No Relay Constraint

**User Story:** As a system architect, I want to enforce that the CubeSat only delivers bundles directly to their final destination ground station, so that the system remains consistent with the no-relay architecture across all phases.

#### Acceptance Criteria

1. THE BPA SHALL transmit a bundle only to the node matching the bundle's final destination Endpoint_ID — the BPA SHALL NOT forward bundles on behalf of other nodes
2. WHEN the Node_Controller looks up a delivery route for a bundle, THE Contact_Plan_Manager SHALL return only direct Contact_Windows with the destination node, with no multi-hop paths

### Requirement 12: AX.25 and LTP Convergence Layer

**User Story:** As a regulatory compliance engineer, I want all DTN transmissions to use AX.25 framing with callsign addressing over LTP via the flight IQ baseband radio, so that every orbital transmission complies with amateur radio regulations.

#### Acceptance Criteria

1. THE CLA SHALL encapsulate all bundle transmissions in AX.25 frames carrying the source amateur radio callsign and the destination amateur radio callsign
2. THE CLA SHALL run LTP sessions on top of AX.25 frames, providing reliable transfer with deferred acknowledgment for all bundle delivery
3. THE CLA SHALL perform LTP segmentation for bundles that exceed a single AX.25 frame size, and reassemble received LTP segments into complete bundles
4. THE CLA SHALL interface with the Flight_Transceiver IQ path (STM32U585 DMA → DAC/ADC or SPI → Flight_Transceiver) as the physical transport
5. FOR ALL valid Bundle objects, encapsulating a bundle into AX.25/LTP frames, modulating to IQ via the Flight_Transceiver, demodulating from IQ, and reassembling the frames back into a bundle SHALL produce a bundle equivalent to the original (end-to-end round-trip property)

### Requirement 13: Autonomous Contact Window Execution

**User Story:** As a flight systems engineer, I want the CubeSat to autonomously wake from sleep, activate the radio, execute bundle transfers during predicted pass windows, and return to sleep, so that the CubeSat operates independently in orbit without ground operator intervention.

#### Acceptance Criteria

1. WHEN a CGR-predicted Contact_Window start time is reached, THE Firmware SHALL wake the STM32U585 from Stop_2_Mode via RTC alarm, initialize the Flight_Transceiver and IQ DSP, and begin processing queued bundles for the contact's ground station
2. THE Node_Controller SHALL transmit queued bundles destined for the contact's ground station in priority order during the Contact_Window
3. THE Node_Controller SHALL cease all transmission when the Contact_Window end time is reached
4. WHEN a Contact_Window completes, THE Node_Controller SHALL record link metrics (bytes transferred, duration, bundles sent, bundles received, signal quality, Doppler tracking accuracy) and update contact statistics
5. WHEN a Contact_Window completes and no further Contact_Windows are predicted within the next 60 seconds, THE Firmware SHALL deactivate the Flight_Transceiver, flush NVM, and transition the STM32U585 into Stop_2_Mode
6. IF the CLA fails to establish the IQ baseband link during a scheduled Contact_Window (Flight_Transceiver not responding, no AX.25 connection established, or signal quality below threshold), THEN THE Node_Controller SHALL mark the contact as missed, retain all queued bundles for the next window, and increment the contacts-missed counter

### Requirement 14: BPSec Integrity with Hardware Crypto

**User Story:** As a security engineer, I want bundle integrity protection using BPSec with the STM32U585 hardware crypto accelerator, so that bundle origin authentication is enforced in orbit while complying with amateur radio regulations.

#### Acceptance Criteria

1. THE BPA SHALL support BPSec (RFC 9172) Block Integrity Blocks (BIB) for bundle origin authentication using HMAC-SHA-256
2. THE BPA SHALL NOT apply BPSec Block Confidentiality Blocks (BCB) or any form of payload encryption, in compliance with amateur radio regulations requiring transmissions to be unencrypted
3. THE BPA SHALL use the STM32U585 hardware crypto accelerator (SHA-256, AES-256, PKA) for all BPSec HMAC-SHA-256 computations instead of software implementations
4. WHEN a bundle with a BIB is received, THE BPA SHALL verify the integrity block using the hardware crypto accelerator and discard the bundle if verification fails, logging the integrity failure with the source Endpoint_ID
5. THE Firmware SHALL store BPSec shared keys in the STM32U585 TrustZone secure world, isolated from non-secure application code

### Requirement 15: TrustZone Secure Key Storage

**User Story:** As a security engineer, I want cryptographic keys stored in the STM32U585 TrustZone secure world, so that keys are hardware-isolated from application firmware in the flight environment.

#### Acceptance Criteria

1. THE Firmware SHALL partition the STM32U585 into TrustZone secure and non-secure worlds, with BPSec keys and crypto operations executing in the secure world
2. THE Firmware SHALL expose a secure API from the TrustZone secure world that the non-secure BPA can call to request HMAC-SHA-256 signing and verification without exposing raw key material
3. IF non-secure code attempts to read TrustZone secure memory directly, THEN THE STM32U585 SHALL generate a hardware fault and the Firmware SHALL log the access violation
4. THE Firmware SHALL provision BPSec keys into the TrustZone secure world during initial firmware flashing or via a secure key update bundle received during a ground pass

### Requirement 16: Power Management and Stop 2 Mode

**User Story:** As a power systems engineer, I want the CubeSat to manage its power budget autonomously, transitioning between active and ultra-low-power states based on predicted contact windows, so that the CubeSat operates within its 5–10 W average power budget.

#### Acceptance Criteria

1. WHEN no Contact_Window is active and no bundle processing is pending, THE Firmware SHALL transition the STM32U585 into Stop_2_Mode
2. WHILE in Stop_2_Mode, THE STM32U585 SHALL consume no more than 20 µA (allowing margin above the nominal 16 µA specification)
3. WHEN a Contact_Window start time is reached, THE Firmware SHALL wake the STM32U585 from Stop_2_Mode via RTC alarm and resume normal operation within 10 milliseconds
4. THE Node_Controller SHALL log timestamped power state transitions (active → Stop_2_Mode, Stop_2_Mode → active) and the duration spent in each state for power budget analysis
5. THE Firmware SHALL set the RTC alarm for the next predicted Contact_Window start time before entering Stop_2_Mode, ensuring autonomous wake-up without external triggers

### Requirement 17: SRAM Memory Management

**User Story:** As an embedded systems engineer, I want the STM32U585 firmware to operate within the 786 KB SRAM constraint while concurrently running ION-DTN, IQ baseband DSP, CGR contact prediction, and bundle index management, so that the flight memory budget is validated.

#### Acceptance Criteria

1. THE Firmware SHALL operate within the STM32U585 786 KB SRAM for all concurrent operations: ION-DTN runtime, IQ sample buffers (TX and RX), AX.25/LTP frame buffers, bundle metadata index, CGR_Engine state and computation buffers, and TrustZone secure world allocations
2. THE Firmware SHALL use static or pool-based memory allocation (Pool_Allocator) for all runtime data structures, avoiding dynamic heap allocation that could cause fragmentation on the constrained MCU
3. THE Firmware SHALL report peak and current SRAM utilization as part of telemetry, broken down by subsystem (ION-DTN, IQ buffers, bundle index, CGR_Engine, TrustZone)
4. IF an operation would exceed the SRAM budget, THEN THE Firmware SHALL reject the operation and log the memory exhaustion event rather than corrupting adjacent memory regions

### Requirement 18: Priority-Based Message Handling

**User Story:** As a ground station operator, I want bundles to be handled according to their priority level on the CubeSat, so that critical messages are delivered before less urgent ones during limited pass windows.

#### Acceptance Criteria

1. THE BPA SHALL assign one of four priority levels to each bundle: critical (highest), expedited, normal, or bulk (lowest)
2. WHEN multiple bundles are queued for the same destination during a Contact_Window, THE Firmware SHALL transmit bundles in strict priority order — all critical bundles before any expedited, all expedited before any normal, all normal before any bulk
3. WHEN the Bundle_Store must evict bundles to free NVM space, THE Bundle_Store SHALL evict bulk bundles first, then normal, then expedited — critical bundles SHALL be evicted only when no lower-priority bundles remain
4. THE BPA SHALL accept a default priority level from the Firmware configuration, applied to bundles that do not specify an explicit priority

### Requirement 19: Rate Limiting and Store Protection

**User Story:** As a flight systems engineer, I want to protect the NVM bundle store from flooding by a misbehaving ground station, so that storage resources are not exhausted during limited pass windows.

#### Acceptance Criteria

1. THE BPA SHALL enforce a configurable maximum bundle acceptance rate (bundles per second) per source Endpoint_ID
2. IF the acceptance rate from a single source Endpoint_ID exceeds the configured limit, THEN THE BPA SHALL reject additional bundles from that source and log the rate-limit event
3. THE BPA SHALL enforce a configurable maximum bundle size in bytes, rejecting any bundle whose total serialized size exceeds the limit

### Requirement 20: Node Health and Telemetry

**User Story:** As a ground station operator, I want to receive telemetry from the CubeSat during passes, so that I can monitor the health and operational status of the flight node.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report telemetry including: uptime in seconds, NVM storage utilization as a percentage of configured maximum, number of bundles currently stored, number of bundles delivered, number of bundles dropped (expired or evicted), and the timestamp of the last completed contact
2. THE Node_Controller SHALL track cumulative statistics including: total bundles received, total bundles sent, total bytes received, total bytes sent, average delivery latency in seconds, contacts completed, and contacts missed
3. THE Firmware SHALL report STM32U585-specific telemetry including: current SRAM utilization (peak and current, by subsystem), power state (active or Stop_2_Mode), time spent in each power state, MCU temperature (internal sensor), IQ baseband signal quality metrics (SNR, bit error rate), current TLE epoch age, and CGR prediction horizon
4. THE Node_Controller SHALL package telemetry as a DTN bundle and transmit it to requesting ground stations during Contact_Windows
5. WHEN a telemetry request bundle is received, THE Node_Controller SHALL generate a telemetry response bundle within 1 second

### Requirement 21: Autonomous Error Handling and Fault Recovery

**User Story:** As a flight systems engineer, I want the CubeSat to handle faults autonomously including power loss, watchdog resets, radiation-induced upsets, transceiver failures, and memory corruption, so that the CubeSat recovers and resumes operation without ground intervention.

#### Acceptance Criteria

1. IF the Bundle_Store reaches NVM capacity and eviction cannot free sufficient space for an incoming bundle, THEN THE BPA SHALL reject the incoming bundle and return a storage-full error to the sender if the LTP session is still active
2. IF a CRC validation fails on a received bundle, THEN THE BPA SHALL discard the corrupted bundle and log the corruption event with the source Endpoint_ID and IQ link metrics
3. IF the STM32U585 experiences a power cycle, watchdog reset, or radiation-induced reset, THEN THE Firmware SHALL reload the Bundle_Store from external NVM, reload the TLE data and Ground_Station_Catalog from NVM, re-compute CGR contact predictions, and resume autonomous operation without ground intervention
4. IF the Flight_Transceiver becomes unresponsive during a Contact_Window, THEN THE Firmware SHALL attempt to reinitialize the Flight_Transceiver up to 3 times with a 1-second interval, and if all attempts fail, mark the contact as missed, retain all queued bundles, and enter Stop_2_Mode until the next predicted Contact_Window
5. IF no direct Contact_Window exists for a bundle's destination, THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window with that destination becomes available or the bundle's lifetime expires
6. THE Firmware SHALL use a hardware watchdog timer with a configurable timeout (default 30 seconds) to detect firmware hangs and trigger an automatic reset

### Requirement 22: Radiation Environment Considerations

**User Story:** As a flight systems engineer, I want the Firmware to detect and mitigate radiation-induced errors in SRAM and NVM, so that the CubeSat maintains data integrity in the LEO radiation environment.

#### Acceptance Criteria

1. THE Firmware SHALL protect critical data structures in SRAM (bundle metadata index, CGR state, contact plan, TLE data) using CRC or redundant copies so that single-bit upsets can be detected
2. WHEN a CRC mismatch or redundancy inconsistency is detected in a critical SRAM data structure, THE Firmware SHALL attempt to recover from the redundant copy or reload from NVM, and log the radiation event
3. THE Firmware SHALL validate NVM data integrity via CRC on every read operation, detecting corruption from radiation-induced bit flips in flash memory
4. THE Firmware SHALL include a radiation event counter in telemetry reports, tracking detected single-event upsets (SEUs) in SRAM and NVM

### Requirement 23: Onboard Time Management

**User Story:** As a flight systems engineer, I want the CubeSat to maintain accurate onboard time, so that CGR contact predictions, bundle timestamps, and contact window scheduling are correct.

#### Acceptance Criteria

1. THE Firmware SHALL maintain onboard time using the STM32U585 RTC, synchronized to UTC
2. WHEN a time synchronization bundle is received from a ground station during a pass, THE Firmware SHALL update the RTC to the received UTC time if the correction exceeds a configurable threshold (default 1 second)
3. THE Firmware SHALL include the current onboard time and the time since last synchronization in telemetry reports
4. IF the RTC has not been synchronized for more than a configurable period (default 7 days), THE Firmware SHALL flag a time-stale warning in telemetry

### Requirement 24: Autonomous Operation Cycle

**User Story:** As a flight systems engineer, I want the CubeSat to execute a fully autonomous operation cycle (predict passes, wake, communicate, sleep) without any ground operator control, so that the CubeSat operates independently between ground contacts.

#### Acceptance Criteria

1. THE Node_Controller SHALL execute the following autonomous cycle: compute next Contact_Window from CGR predictions, set RTC alarm for the Contact_Window start time, enter Stop_2_Mode, wake on RTC alarm, activate Flight_Transceiver and IQ DSP, execute bundle transfers during the Contact_Window, deactivate Flight_Transceiver, run bundle cleanup (expire, evict), update CGR predictions if needed, and repeat
2. THE Node_Controller SHALL complete a full operation cycle (check contacts, transmit queued bundles via IQ baseband, process received bundles, run cleanup) within 1 second on the STM32U585
3. THE Node_Controller SHALL operate indefinitely without ground intervention, using onboard TLE data and the Ground_Station_Catalog for all scheduling decisions
4. WHEN the Firmware boots after a reset, THE Node_Controller SHALL restore state from NVM (Bundle_Store, TLE, Ground_Station_Catalog, contact statistics) and resume the autonomous operation cycle within 5 seconds
