# Requirements Document

## Introduction

This document specifies the requirements for a phased Delay/Disruption Tolerant Networking (DTN) system for amateur radio, progressing through four phases: terrestrial validation (RPi + Mobilinkd TNC4 + FT-817), CubeSat Engineering Model (STM32U585 + Ettus B200mini), LEO CubeSat flight (STM32U585 + flight IQ transceiver), and cislunar deep-space communication. The system uses ION-DTN (BPv7/LTP) over AX.25 with callsign-based addressing, supporting two core operations — ping and store-and-forward — with no relay functionality. Requirements are derived from the approved design document.

## Glossary

- **BPA**: Bundle Protocol Agent — the core DTN engine that creates, receives, stores, and delivers BPv7 bundles
- **Bundle**: A BPv7 protocol data unit carrying a payload between DTN endpoints
- **Bundle_Store**: Persistent storage subsystem for bundles awaiting delivery, backed by NVM on space nodes
- **Contact_Plan_Manager**: Subsystem that manages scheduled communication windows and uses CGR for contact prediction
- **CGR**: Contact Graph Routing — ION-DTN module used exclusively for contact prediction / pass scheduling (not multi-hop routing)
- **CLA**: Convergence Layer Adapter — abstracts the physical/link layer for bundle transmission across all radio links
- **Node_Controller**: Top-level orchestrator that ties together BPA, Bundle_Store, Contact_Plan_Manager, and CLA
- **ION-DTN**: NASA JPL's Interplanetary Overlay Network — the DTN implementation providing BPv7, LTP, and CGR
- **LTP**: Licklider Transmission Protocol — convergence layer running on top of AX.25 providing reliable transfer with deferred acknowledgment
- **AX.25**: Link-layer framing protocol providing callsign-based source/destination addressing for amateur radio compliance
- **STM32U585**: Ultra-low-power ARM Cortex-M33 MCU (160 MHz, 786 KB SRAM, 2 MB flash, hardware crypto, TrustZone) used as OBC for EM and flight nodes
- **B200mini**: Ettus Research USRP B200mini SDR — EM-only RF front-end (USB 3.0, 12-bit ADC/DAC, 70 MHz–6 GHz)
- **TNC4**: Mobilinkd TNC4 terminal node controller — USB-connected TNC for terrestrial nodes interfacing with FT-817
- **FT-817**: Yaesu FT-817 portable transceiver — terrestrial node radio with 9600 baud data port
- **NVM**: Non-volatile memory — external SPI/QSPI flash (64–256 MB) for persistent bundle storage on STM32U585 nodes
- **IQ_Transceiver**: Flight-qualified IQ baseband transceiver IC interfacing directly with STM32U585 for LEO and cislunar flight nodes
- **Ping**: DTN reachability test — send a bundle echo request and receive an echo response
- **Store_and_Forward**: A source node sends a bundle to a destination node, which stores it and delivers it when the destination becomes reachable during a contact window

## Requirements

### Requirement 1: Bundle Creation and Validation

**User Story:** As an amateur radio operator, I want to create and validate DTN bundles, so that I can send well-formed messages through the DTN network.

#### Acceptance Criteria

1. WHEN a user submits a message with a valid destination endpoint and payload, THE BPA SHALL create a BPv7 bundle with proper headers, endpoint addressing, CRC, priority, and lifetime fields
2. WHEN a bundle is received, THE BPA SHALL validate that the bundle version equals 7, the destination is a valid endpoint ID, the lifetime is greater than zero, the creation timestamp does not exceed the current time, and the CRC is correct
3. IF a received bundle fails any validation check, THEN THE BPA SHALL discard the bundle and log the validation failure
4. THE BPA SHALL support three bundle types: data bundles for store-and-forward, ping request bundles for echo requests, and ping response bundles for echo responses

### Requirement 2: Bundle Storage and Persistence

**User Story:** As a DTN node operator, I want bundles to be stored persistently and survive power cycles, so that no messages are lost during disruptions.

#### Acceptance Criteria

1. WHEN a valid bundle is accepted, THE Bundle_Store SHALL persist the bundle to non-volatile memory atomically, preventing corruption on power loss
2. WHEN a bundle is stored and later retrieved by its ID, THE Bundle_Store SHALL return a bundle identical to the original
3. THE Bundle_Store SHALL maintain a priority-ordered index so that bundles can be retrieved in priority order (critical > expedited > normal > bulk)
4. WHEN the Bundle_Store reaches capacity and a new bundle arrives, THE Bundle_Store SHALL evict expired bundles first, then lowest-priority bundles, to free sufficient space
5. WHEN evicting bundles, THE Bundle_Store SHALL preserve all critical-priority bundles until all lower-priority bundles have been evicted
6. THE Bundle_Store SHALL enforce that total stored bytes do not exceed the configured maximum storage capacity for the node
7. WHEN a power cycle occurs on an STM32U585-based node, THE Bundle_Store SHALL reload its state from external SPI/QSPI NVM and validate store integrity via CRC checks

### Requirement 3: Bundle Lifetime Enforcement

**User Story:** As a network operator, I want expired bundles to be automatically removed, so that stale data does not consume storage or bandwidth.

#### Acceptance Criteria

1. WHEN the Node_Controller runs a cleanup cycle, THE Bundle_Store SHALL delete all bundles whose creation timestamp plus lifetime is less than or equal to the current time
2. THE Bundle_Store SHALL contain zero expired bundles after a cleanup cycle completes

### Requirement 4: Ping Operation

**User Story:** As an amateur radio operator, I want to ping a DTN node and receive an echo response, so that I can verify end-to-end DTN reachability.

#### Acceptance Criteria

1. WHEN the BPA receives a ping request bundle, THE BPA SHALL generate exactly one ping response bundle addressed to the original sender's endpoint
2. WHEN a ping response is generated, THE BPA SHALL queue the response for delivery during the next available contact window with the sender
3. WHEN a ping response is received at the originating node, THE Node_Controller SHALL report the round-trip time to the operator

### Requirement 5: Store-and-Forward Operation

**User Story:** As an amateur radio operator, I want to send messages that are stored at intermediate nodes and delivered when the destination becomes reachable, so that I can communicate across disrupted links.

#### Acceptance Criteria

1. WHEN a data bundle is received whose destination is a local endpoint, THE BPA SHALL deliver the bundle to the local application agent
2. WHEN a data bundle is received whose destination is a remote endpoint, THE BPA SHALL store the bundle and queue it for direct delivery during the next contact window with the destination node
3. THE BPA SHALL transmit queued bundles in priority order (critical first, then expedited, normal, bulk) during each contact window
4. WHEN a transmitted bundle is acknowledged by the remote node, THE Bundle_Store SHALL delete the acknowledged bundle
5. IF a bundle transmission is not acknowledged, THEN THE Bundle_Store SHALL retain the bundle for retry during the next contact window

### Requirement 6: No Relay Constraint

**User Story:** As a system architect, I want to enforce that nodes only deliver bundles directly to their final destination, so that the system remains simple and predictable without multi-hop relay complexity.

#### Acceptance Criteria

1. THE BPA SHALL transmit a bundle only to the node matching the bundle's final destination endpoint — bundles are not forwarded on behalf of other nodes
2. WHEN the Node_Controller looks up a delivery route, THE Contact_Plan_Manager SHALL return only direct contact windows with the destination node, with no multi-hop paths

### Requirement 7: Contact Plan Management

**User Story:** As a ground station operator, I want the system to manage scheduled communication windows, so that bundles are delivered during available contact opportunities.

#### Acceptance Criteria

1. THE Contact_Plan_Manager SHALL maintain a time-tagged schedule of contact windows, each specifying a remote node, start time, end time, data rate, and link type
2. WHEN queried for active contacts at a given time, THE Contact_Plan_Manager SHALL return all contact windows whose start time is at or before the query time and whose end time is after the query time
3. WHEN queried for the next contact with a specific destination, THE Contact_Plan_Manager SHALL return the earliest future contact window matching that destination
4. THE Contact_Plan_Manager SHALL enforce that no overlapping contacts exist on the same link for a given node
5. THE Contact_Plan_Manager SHALL validate that all contact windows fall within the plan's valid-from and valid-to time range

### Requirement 8: CGR Contact Prediction

**User Story:** As a mission operator, I want the system to predict communication windows using orbital mechanics, so that contact plans are automatically generated from ephemeris data.

#### Acceptance Criteria

1. WHEN provided with orbital parameters and a list of ground station locations, THE Contact_Plan_Manager SHALL use ION-DTN's CGR engine to compute predicted contact windows over a specified time horizon
2. THE Contact_Plan_Manager SHALL return only predicted contacts where the maximum elevation angle meets or exceeds the ground station's minimum elevation threshold
3. THE Contact_Plan_Manager SHALL sort predicted contacts by start time in ascending order
4. THE Contact_Plan_Manager SHALL assign a confidence value to each predicted contact that decreases for windows further from the orbital parameter epoch
5. WHEN fresh orbital parameters (TLE/ephemeris data) are received, THE Contact_Plan_Manager SHALL re-compute predicted contact windows for the affected space node
6. THE Contact_Plan_Manager SHALL produce no overlapping predicted windows for the same ground station
7. THE Contact_Plan_Manager SHALL ensure all predicted contact windows fall within the requested time horizon boundaries

### Requirement 9: Contact Window Execution

**User Story:** As a DTN node, I want to transmit queued bundles during active contact windows, so that messages are delivered when communication links are available.

#### Acceptance Criteria

1. WHEN a contact window becomes active, THE CLA SHALL establish the link and THE Node_Controller SHALL begin transmitting queued bundles destined for the contact's remote node
2. THE Node_Controller SHALL cease all transmission when the contact window end time is reached
3. WHEN a contact window completes, THE Node_Controller SHALL update link metrics (RSSI, SNR, BER, bytes transferred) and contact statistics
4. IF the CLA fails to establish a link during a scheduled contact window, THEN THE Node_Controller SHALL mark the contact as missed, retain all queued bundles, and increment the contacts-missed counter


### Requirement 10: AX.25 and LTP Convergence Layer

**User Story:** As an amateur radio operator, I want all DTN links to use AX.25 framing with callsign addressing over LTP, so that every transmission complies with amateur radio regulations and provides reliable transfer.

#### Acceptance Criteria

1. THE CLA SHALL encapsulate all bundle transmissions in AX.25 frames carrying source and destination amateur radio callsigns, across all phases (terrestrial, EM, LEO, cislunar)
2. THE CLA SHALL run LTP sessions on top of AX.25 frames, providing reliable transfer with deferred acknowledgment for all bundle delivery
3. THE CLA SHALL perform LTP segmentation and reassembly for bundles that exceed a single AX.25 frame
4. THE CLA SHALL monitor link quality metrics including RSSI, SNR, and bit error rate during active contacts

### Requirement 11: Terrestrial Node Operation

**User Story:** As an amateur radio operator, I want to run a terrestrial DTN node using a Raspberry Pi, Mobilinkd TNC4, and FT-817, so that I can participate in DTN experiments with accessible hardware.

#### Acceptance Criteria

1. THE CLA SHALL interface with the Mobilinkd TNC4 via USB (not Bluetooth) for AX.25 packet operation on terrestrial nodes
2. THE CLA SHALL drive the FT-817 radio at 9600 baud through its 9600 baud data port using G3RUH-compatible GFSK modulation
3. WHEN operating as a terrestrial node, THE Node_Controller SHALL complete an operation cycle within 100 milliseconds
4. WHEN operating as a terrestrial node, THE Bundle_Store SHALL complete store operations within 10 milliseconds

### Requirement 12: Engineering Model Node Operation

**User Story:** As a mission developer, I want to validate the flight software stack on ground-based flight-representative hardware, so that I can confirm correct operation before flight commitment.

#### Acceptance Criteria

1. THE EM node SHALL run ION-DTN (BPv7/LTP over AX.25) on the STM32U585 OBC with identical software to the flight unit
2. THE CLA SHALL interface with the Ettus B200mini SDR as the RF front-end via a companion Raspberry Pi or PC running UHD, bridging IQ samples to the STM32U585 over SPI or UART/DMA
3. THE STM32U585 SHALL generate TX IQ samples and process RX IQ samples via its DMA engine for baseband DSP
4. WHEN operating as an EM node, THE Node_Controller SHALL complete an operation cycle within 1 second
5. THE EM node SHALL operate within 786 KB SRAM for concurrent bundle processing and IQ buffer management
6. THE EM node SHALL use external SPI/QSPI NVM (64–256 MB) for persistent bundle storage
7. WHEN the EM node enters idle state between simulated contact windows, THE STM32U585 SHALL enter Stop 2 ultra-low-power mode

### Requirement 13: LEO CubeSat Flight Node Operation

**User Story:** As a mission operator, I want the LEO CubeSat to autonomously store and deliver DTN bundles during orbital passes, so that ground stations can exchange messages via the satellite.

#### Acceptance Criteria

1. THE LEO flight node SHALL run ION-DTN (BPv7/LTP over AX.25) on the STM32U585 OBC with a flight-qualified IQ transceiver IC — no companion host or B200mini
2. THE CLA SHALL perform GMSK/BPSK modulation and demodulation at 9.6 kbps on UHF 437 MHz via IQ baseband on the STM32U585
3. WHEN operating as a LEO flight node, THE Node_Controller SHALL complete an operation cycle within 1 second
4. WHEN no contact window is active, THE STM32U585 SHALL enter Stop 2 ultra-low-power mode to comply with the 5–10 W average power budget
5. THE LEO flight node SHALL deliver bundles directly to destination ground stations during orbital passes — no relay to other intermediate nodes

### Requirement 14: Cislunar Node Operation

**User Story:** As a deep-space communications researcher, I want a cislunar DTN node that enables amateur participation in Earth–Moon delay-tolerant networking, so that the amateur community can experiment with deep-space communication.

#### Acceptance Criteria

1. THE cislunar node SHALL run ION-DTN (BPv7/LTP over AX.25) with BPSK modulation and strong FEC (LDPC or Turbo coding) at 500 bps on S-band 2.2 GHz
2. THE CLA SHALL account for 1–2 second one-way light-time delay in LTP session management for cislunar links
3. WHEN operating as a cislunar node, THE Node_Controller SHALL complete an operation cycle within 10 seconds
4. THE cislunar node SHALL support long-duration message storage for bundles awaiting delivery across extended contact gaps

### Requirement 15: Node Health and Telemetry

**User Story:** As a mission operator, I want to monitor node health and performance, so that I can detect and respond to anomalies.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report telemetry including uptime, storage utilization percentage, bundles stored, bundles forwarded, bundles dropped, and last contact time
2. WHILE operating as a space node, THE Node_Controller SHALL additionally report temperature and battery percentage
3. THE Node_Controller SHALL track cumulative statistics including total bundles received, total bundles sent, total bytes received, total bytes sent, average latency, contacts completed, and contacts missed

### Requirement 16: Security

**User Story:** As a network operator, I want bundle integrity protection and secure key storage, so that the DTN network is protected against spoofing and tampering.

#### Acceptance Criteria

1. THE BPA SHALL support BPSec (RFC 9172) integrity blocks for bundle origin authentication
2. WHILE operating on an STM32U585-based node, THE BPA SHALL use the hardware crypto accelerator (AES-256, SHA-256, PKA) for BPSec cryptographic operations
3. WHILE operating on an STM32U585-based node, THE Node_Controller SHALL store cryptographic keys and BPSec credentials in the TrustZone secure world, isolated from non-secure application code
4. THE BPA SHALL enforce rate limiting on bundle acceptance to prevent store flooding attacks
5. THE Node_Controller SHALL verify contact plan integrity using signed plans for space nodes

### Requirement 17: Error Handling and Recovery

**User Story:** As a node operator, I want the system to handle faults gracefully and recover automatically, so that the DTN network remains operational despite disruptions.

#### Acceptance Criteria

1. IF the Bundle_Store reaches capacity and eviction cannot free sufficient space, THEN THE BPA SHALL reject the incoming bundle and return an error to the sender if the link is still active
2. IF a CRC validation fails on a received bundle, THEN THE BPA SHALL discard the corrupted bundle and log the corruption event with link metrics
3. IF an unexpected power cycle occurs on an STM32U585-based node, THEN THE Node_Controller SHALL reload the bundle store from external NVM, validate integrity via CRC, and resume normal operation
4. IF store corruption is detected after a power cycle, THEN THE Bundle_Store SHALL rebuild from intact bundles only, discarding corrupted entries
5. IF no direct contact window exists for a bundle's destination, THEN THE Bundle_Store SHALL retain the bundle and re-evaluate when the contact plan is updated

### Requirement 18: Link Budget Feasibility

**User Story:** As a mission designer, I want to verify that the RF link closes with positive margin for each phase, so that communication is physically achievable.

#### Acceptance Criteria

1. THE link budget computation SHALL produce a positive margin for the LEO UHF configuration (2 W TX, omni antenna, 437 MHz, 500 km, 9.6 kbps, Yagi ground antenna)
2. THE link budget computation SHALL produce a positive margin for the cislunar S-band configuration (5 W TX, 10 dBi patch, 2.2 GHz, 384,000 km, 500 bps, 35 dBi ground dish, BPSK + LDPC)
3. THE link budget computation SHALL be monotonically decreasing with increasing distance for fixed transmit parameters
