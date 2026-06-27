# Requirements Document

## Introduction

This document specifies the requirements for Phase 1 of the RADIANT cislunar amateur DTN project: terrestrial DTN validation. Phase 1 deploys a DTN engine (ION-DTN or Hardy) via the DTN abstraction layer (`radiant-dtn-abstraction`) on Linux or macOS hosts connected via USB to Mobilinkd TNC4 terminal node controllers, which drive Yaesu FT-817 radios at 9600 baud through the 9600 baud data port using G3RUH-compatible GFSK modulation on VHF/UHF amateur bands.

The system is implemented in Rust, using the RADIANT crate ecosystem: `radiant-dtn-abstraction` for DTN engine orchestration (configuration generation, lifecycle, telemetry), `radiant-kiss` (a `no_std` KISS framing library designed for eventual STM32U585 flight hardware), and `radiant-cla` for convergence layer adapter logic.

The protocol stack is:
- Application (ping, store-and-forward)
- BPv7 (Bundle Protocol v7)
- LTP (Licklider Transmission Protocol)
- KISS (TNC Serial Framing) — LTP segments wrapped directly in KISS frames, NO AX.25
- USB Serial (TNC4)
- G3RUH GFSK (9600 baud)

The system supports two core operations: ping (DTN reachability test) and store-and-forward (point-to-point bundle delivery). There is no relay functionality — nodes do not forward bundles on behalf of other nodes. All bundle delivery is direct (source → destination).

Station identification is achieved through callsign-embedded DTN Endpoint IDs (`dtn://callsign-ssid/service`) present in every bundle, plus periodic beacon bundles transmitted every 10 minutes. This approach eliminates AX.25 entirely, reduces protocol overhead, and maintains full amateur radio regulatory compliance.

The DTN engine is managed through the DTN abstraction layer, which generates backend-specific configuration from a canonical YAML/JSON model, manages engine lifecycle, and provides unified telemetry. The orchestrator does not call ionadmin scripts directly — it configures the engine through the abstraction layer's trait-based adapter interface.

BPv7 bundle CRC integrity checks protect against accidental corruption. Per amateur radio regulations and project steering constraints, no cryptographic operations (encryption, HMAC, digital signatures, BPSec) are applied to transmitted signals. CRC validation is permitted as it falls under error detection/correction and does not obscure the meaning of communications.

This spec is scoped exclusively to terrestrial ground nodes. STM32U585 OBC, IQ baseband/SDR, Ettus B200mini, CGR contact prediction, orbital mechanics, space segment (CubeSat, cislunar), S-band/X-band communications, and flight-qualified hardware are out of scope.

## Glossary

- **BPA**: Bundle Protocol Agent — the DTN engine component that creates, receives, validates, stores, and delivers BPv7 bundles
- **Bundle**: A BPv7 protocol data unit carrying a payload between DTN endpoints
- **Bundle_Store**: Persistent storage subsystem on the local filesystem for bundles awaiting delivery
- **Contact_Plan_Manager**: Subsystem that maintains a manually configured schedule of communication windows between ground nodes
- **CLA**: Convergence Layer Adapter — interfaces the DTN engine's LTP stack with the Mobilinkd TNC4 over USB via KISS framing using `radiant-kiss`
- **Node_Controller**: Top-level Rust orchestrator binary that ties together DTN engine management (via `radiant-dtn-abstraction`), Bundle_Store, Contact_Plan_Manager, and CLA on each host node
- **DTN_Engine**: The BPv7 implementation (ION-DTN or Hardy) managed through the DTN abstraction layer
- **Abstraction_Layer**: The `radiant-dtn-abstraction` crate providing vendor-neutral configuration generation, lifecycle management, and monitoring for the DTN_Engine
- **LTP**: Licklider Transmission Protocol — provides reliable transfer with deferred acknowledgment, segments wrapped directly in KISS frames
- **KISS**: Keep It Simple Stupid — a minimal serial framing protocol for TNC hardware, implemented by the `radiant-kiss` no_std crate; carries LTP segments directly without AX.25
- **TNC4**: Mobilinkd TNC4 terminal node controller — USB-connected TNC interfacing with the FT-817 radio
- **FT-817**: Yaesu FT-817 portable transceiver with 9600 baud data port for G3RUH-compatible GFSK modulation
- **Ping**: DTN reachability test — send a bundle echo request and receive an echo response
- **Store_and_Forward**: Point-to-point bundle delivery where a source node sends a bundle directly to a destination node during a contact window
- **Contact_Window**: A scheduled time interval during which two ground nodes can communicate over their radio link
- **Endpoint_ID**: A DTN endpoint identifier using the "dtn" URI scheme embedding an amateur radio callsign (e.g., `dtn://g4dpz-1/mail`)
- **Callsign_EID**: An Endpoint_ID in the format `dtn://callsign-ssid/service` used for station identification compliance
- **Beacon_Bundle**: A periodic bundle transmitted every 10 minutes for amateur radio station identification, containing the station callsign and identification text
- **Canonical_Config**: The YAML/JSON configuration document consumed by the Abstraction_Layer to generate backend-specific DTN_Engine configuration

## Requirements

### Requirement 1: Bundle Creation and Validation

**User Story:** As an amateur radio operator, I want to create and validate DTN bundles, so that I can send well-formed messages through the terrestrial DTN network.

#### Acceptance Criteria

1. WHEN a user submits a message with a valid destination Callsign_EID and payload, THE BPA SHALL create a BPv7 bundle with the bundle version set to 7, source and destination Callsign_EIDs in `dtn://callsign-ssid/service` format, a CRC integrity check, a priority level (critical, expedited, normal, or bulk), and a positive lifetime value in seconds
2. WHEN a bundle is received from the CLA, THE BPA SHALL validate that the bundle version equals 7, the destination is a well-formed Callsign_EID, the lifetime is greater than zero, the creation timestamp does not exceed the current time, and the CRC is correct
3. IF a received bundle fails any validation check, THEN THE BPA SHALL discard the bundle and log the specific validation failure reason along with the source Callsign_EID
4. THE BPA SHALL support three bundle types: data bundles for store-and-forward payload delivery, ping request bundles for echo requests, and ping response bundles for echo responses
5. FOR ALL valid Bundle objects, serializing a Bundle to its BPv7 wire format and then parsing the wire format back SHALL produce a Bundle equivalent to the original (round-trip property)

### Requirement 2: Bundle Storage and Persistence

**User Story:** As a terrestrial DTN node operator, I want bundles to be stored persistently on the local filesystem and survive process restarts and power cycles, so that no messages are lost during disruptions.

#### Acceptance Criteria

1. WHEN a valid bundle is accepted, THE Bundle_Store SHALL persist the bundle to the local filesystem atomically, preventing corruption if the process is interrupted during the write
2. WHEN a bundle is stored and later retrieved by its bundle ID (source Callsign_EID, creation timestamp, sequence number), THE Bundle_Store SHALL return a bundle identical to the original (round-trip property)
3. THE Bundle_Store SHALL maintain a priority-ordered index so that bundles are retrieved in priority order: critical first, then expedited, then normal, then bulk
4. WHEN the Bundle_Store reaches the configured maximum storage capacity and a new bundle arrives, THE Bundle_Store SHALL evict expired bundles first, then the lowest-priority bundles with the earliest creation timestamps, to free sufficient space for the new bundle
5. WHEN evicting bundles, THE Bundle_Store SHALL preserve all critical-priority bundles until all expedited, normal, and bulk bundles have been evicted
6. THE Bundle_Store SHALL enforce that total stored bytes do not exceed the configured maximum storage capacity for the node
7. WHEN the Node_Controller process restarts, THE Bundle_Store SHALL reload its persisted state from the local filesystem and validate store integrity

### Requirement 3: Bundle Lifetime Enforcement

**User Story:** As a node operator, I want expired bundles to be automatically removed, so that stale data does not consume storage or bandwidth.

#### Acceptance Criteria

1. WHEN the Node_Controller runs a cleanup cycle, THE Bundle_Store SHALL delete all bundles whose creation timestamp plus lifetime is less than or equal to the current time
2. THE Bundle_Store SHALL contain zero expired bundles after a cleanup cycle completes

### Requirement 4: Ping Operation

**User Story:** As an amateur radio operator, I want to ping a remote terrestrial DTN node and receive an echo response, so that I can verify end-to-end DTN reachability over the radio link.

#### Acceptance Criteria

1. WHEN the BPA receives a ping request bundle addressed to a local endpoint, THE BPA SHALL generate exactly one ping response bundle with the destination set to the original sender's Callsign_EID
2. WHEN a ping response is generated, THE BPA SHALL queue the response in the Bundle_Store for delivery during the next available Contact_Window with the sender's node
3. WHEN a ping response is received at the originating node, THE Node_Controller SHALL compute and report the round-trip time from the original ping request creation timestamp to the response receipt time
4. THE BPA SHALL include the original ping request's bundle ID in the ping response payload so the originating node can correlate responses to requests

### Requirement 5: Store-and-Forward Operation

**User Story:** As an amateur radio operator, I want to send messages that are stored at my node and delivered directly to the destination node when a contact window opens, so that I can communicate across disrupted terrestrial links.

#### Acceptance Criteria

1. WHEN a data bundle is received whose destination matches a local Callsign_EID, THE BPA SHALL deliver the bundle payload to the local application agent
2. WHEN a data bundle is created or received whose destination is a remote Callsign_EID, THE BPA SHALL store the bundle in the Bundle_Store and queue it for direct delivery during the next Contact_Window with the destination node
3. THE BPA SHALL transmit queued bundles in priority order (critical first, then expedited, then normal, then bulk) during each Contact_Window
4. WHEN a transmitted bundle is acknowledged by the remote node via LTP, THE Bundle_Store SHALL delete the acknowledged bundle
5. IF a bundle transmission is not acknowledged within the LTP retransmission timeout, THEN THE Bundle_Store SHALL retain the bundle for retry during the next Contact_Window

### Requirement 6: No Relay Constraint

**User Story:** As a system architect, I want to enforce that terrestrial nodes only deliver bundles directly to their final destination, so that the system remains simple and predictable without multi-hop relay complexity.

#### Acceptance Criteria

1. THE BPA SHALL transmit a bundle only to the node matching the bundle's final destination Callsign_EID — the BPA SHALL NOT forward bundles on behalf of other nodes
2. WHEN the Node_Controller looks up a delivery route for a bundle, THE Contact_Plan_Manager SHALL return only direct Contact_Windows with the destination node, with no multi-hop paths

### Requirement 7: Contact Plan Management

**User Story:** As a ground station operator, I want to manage scheduled communication windows between terrestrial nodes, so that bundles are delivered during available contact opportunities.

#### Acceptance Criteria

1. THE Contact_Plan_Manager SHALL maintain a time-tagged schedule of Contact_Windows, each specifying a remote node Callsign_EID, start time, end time, data rate in bits per second, and link type (VHF or UHF)
2. WHEN queried for active contacts at a given time, THE Contact_Plan_Manager SHALL return all Contact_Windows whose start time is at or before the query time and whose end time is after the query time
3. WHEN queried for the next contact with a specific destination node, THE Contact_Plan_Manager SHALL return the earliest future Contact_Window matching that destination
4. THE Contact_Plan_Manager SHALL reject any contact plan update that would create overlapping Contact_Windows on the same link for a given node
5. THE Contact_Plan_Manager SHALL validate that all Contact_Windows fall within the plan's valid-from and valid-to time range
6. WHEN a contact plan is loaded or updated, THE Contact_Plan_Manager SHALL persist the plan to the local filesystem so it survives process restarts
7. THE Contact_Plan_Manager SHALL support loading contact plans from a YAML or JSON configuration file compatible with the Abstraction_Layer Canonical_Config format

### Requirement 8: Contact Window Execution

**User Story:** As a terrestrial DTN node, I want to transmit queued bundles during active contact windows, so that messages are delivered when communication links are available.

#### Acceptance Criteria

1. WHEN a Contact_Window becomes active (current time reaches the window start time), THE CLA SHALL establish the LTP link over KISS via the TNC4 and THE Node_Controller SHALL begin transmitting queued bundles destined for the contact's remote node
2. THE Node_Controller SHALL cease all transmission when the Contact_Window end time is reached
3. WHEN a Contact_Window completes, THE Node_Controller SHALL record link metrics (bytes transferred, duration, bundles sent, bundles received) and update contact statistics
4. IF the CLA fails to establish a link during a scheduled Contact_Window (TNC4 not responding, radio not keyed, or no KISS connection established), THEN THE Node_Controller SHALL mark the contact as missed, retain all queued bundles for the next window, and increment the contacts-missed counter

### Requirement 9: KISS Framing and LTP Convergence Layer

**User Story:** As an amateur radio operator, I want all DTN transmissions to use KISS-framed LTP segments over the TNC serial link, so that the protocol stack is minimal and efficient without unnecessary AX.25 overhead.

#### Acceptance Criteria

1. THE CLA SHALL wrap all outbound LTP segments in KISS frames using the `radiant-kiss` crate, with FEND (0xC0) delimiters, command byte 0x00 for data frames, and proper byte stuffing (0xC0 → 0xDB 0xDC, 0xDB → 0xDB 0xDD)
2. THE CLA SHALL extract LTP segments from received KISS frames by detecting FEND boundaries and reversing byte stuffing
3. THE CLA SHALL perform LTP segmentation for bundles that exceed the configured maximum KISS frame payload size, and reassemble received LTP segments into complete bundles
4. THE CLA SHALL interface with the Mobilinkd TNC4 via USB serial connection for all KISS frame operations
5. THE CLA SHALL drive the FT-817 radio at 9600 baud through its 9600 baud data port using G3RUH-compatible GFSK modulation
6. FOR ALL valid LTP segment byte sequences, wrapping a segment in a KISS frame and then extracting the segment from the KISS frame SHALL produce a byte sequence identical to the original (round-trip property)
7. THE CLA SHALL NOT include any AX.25 framing, headers, or addressing — LTP segments are carried directly in KISS frames

### Requirement 10: Station Identification via Callsign EIDs

**User Story:** As an amateur radio operator, I want every transmission to include my callsign for regulatory compliance, so that my station is properly identified without needing AX.25 addressing.

#### Acceptance Criteria

1. THE Node_Controller SHALL configure the DTN_Engine with a Callsign_EID in the format `dtn://callsign-ssid/service` as the node's primary Endpoint_ID, where the callsign is the operator's licensed amateur radio callsign in lowercase
2. THE BPA SHALL include the source Callsign_EID in every outbound bundle's primary block, ensuring the transmitting station's callsign is present in every over-the-air transmission
3. THE Node_Controller SHALL validate that the configured Callsign_EID contains a syntactically valid amateur radio callsign (one or two letter prefix, one or more digits, one to three letter suffix) with an SSID in the range 0–15
4. THE Node_Controller SHALL reject startup if no valid Callsign_EID is configured, logging an error indicating that station identification is required for regulatory compliance

### Requirement 11: Station Identification Beacon

**User Story:** As an amateur radio operator, I want my station to transmit periodic identification beacons, so that I comply with the regulatory requirement to identify at least every 10 minutes.

#### Acceptance Criteria

1. THE Node_Controller SHALL transmit a Beacon_Bundle every 10 minutes while the node is operational, regardless of whether a Contact_Window is currently active
2. THE Beacon_Bundle SHALL contain the source Callsign_EID in the bundle primary block and a human-readable identification string in the payload including the callsign and a description (e.g., "G4DPZ amateur radio DTN experimental station")
3. THE Beacon_Bundle SHALL use the destination EID `dtn://beacon` (a well-known broadcast-style endpoint) and a lifetime of 600 seconds (one beacon interval)
4. WHEN the Node_Controller starts, THE Node_Controller SHALL transmit an initial Beacon_Bundle within 30 seconds of startup
5. THE Node_Controller SHALL log each beacon transmission with a timestamp for regulatory compliance record-keeping

### Requirement 12: DTN Engine Configuration via Abstraction Layer

**User Story:** As a system integrator, I want the DTN engine to be configured and managed through the abstraction layer using canonical YAML/JSON configuration, so that I never write ionadmin scripts or backend-specific configs manually.

#### Acceptance Criteria

1. THE Node_Controller SHALL configure the DTN_Engine exclusively through the Abstraction_Layer interface, providing a Canonical_Config document specifying the local node identity (Callsign_EID and node number), neighbor definitions, KISS convergence layer parameters (TNC device path, baud rate, LTP engine IDs), and contact plan
2. THE Abstraction_Layer SHALL generate all backend-specific configuration (ionadmin scripts for ION-DTN, YAML for Hardy) from the Canonical_Config without manual intervention
3. THE Node_Controller SHALL use the Abstraction_Layer lifecycle operations (start, stop, restart, health check) to manage the DTN_Engine process
4. THE Node_Controller SHALL query DTN_Engine telemetry (bundle statistics, link state) through the Abstraction_Layer unified monitoring interface
5. WHEN the Canonical_Config is updated, THE Node_Controller SHALL apply changes through the Abstraction_Layer, using hot reconfiguration for supported operations (contact plan updates, link enable/disable) and engine restart for unsupported changes

### Requirement 13: No Cryptography (Amateur Radio Compliance)

**User Story:** As a network operator, I want to ensure the system complies with amateur radio regulations that prohibit encryption and codes intended to obscure the meaning of communications.

#### Acceptance Criteria

1. THE system SHALL NOT use any cryptographic operations (encryption, HMAC, digital signatures, BPSec Block Integrity Blocks, BPSec Block Confidentiality Blocks) on transmitted signals
2. THE system SHALL NOT apply any form of payload encryption, integrity blocks using cryptographic algorithms, or codes intended to obscure meaning
3. THE system SHALL rely on CRC validation (BPv7 bundle CRC and LTP checksum) for protection against accidental corruption — these are permitted as they are standard error detection mechanisms that do not obscure content
4. THE system SHALL ensure all transmitted data is inspectable by any third party with knowledge of the published protocol specification (BPv7 RFC 9171, LTP RFC 5326, KISS framing)

### Requirement 14: Priority-Based Message Handling

**User Story:** As a node operator, I want bundles to be handled according to their priority level, so that critical messages are delivered before less urgent ones.

#### Acceptance Criteria

1. THE BPA SHALL assign one of four priority levels to each bundle: critical (highest), expedited, normal, or bulk (lowest)
2. WHEN multiple bundles are queued for the same destination during a Contact_Window, THE Node_Controller SHALL transmit bundles in strict priority order — all critical bundles before any expedited, all expedited before any normal, all normal before any bulk
3. WHEN the Bundle_Store must evict bundles to free space, THE Bundle_Store SHALL evict bulk bundles first, then normal, then expedited — critical bundles SHALL be evicted only when no lower-priority bundles remain
4. THE BPA SHALL accept a default priority level from the Node_Controller configuration, applied to bundles that do not specify an explicit priority

### Requirement 15: Rate Limiting and Store Protection

**User Story:** As a node operator, I want to protect the bundle store from flooding, so that a misbehaving or malicious node cannot exhaust storage resources.

#### Acceptance Criteria

1. THE BPA SHALL enforce a configurable maximum bundle acceptance rate (bundles per second) per source Callsign_EID
2. IF the acceptance rate from a single source Callsign_EID exceeds the configured limit, THEN THE BPA SHALL reject additional bundles from that source and log the rate-limit event
3. THE BPA SHALL enforce a configurable maximum bundle size in bytes, rejecting any bundle whose total serialized size exceeds the limit

### Requirement 16: Node Health and Telemetry

**User Story:** As a node operator, I want to monitor node health and performance, so that I can detect and respond to anomalies in the terrestrial DTN network.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report telemetry including: uptime in seconds, storage utilization as a percentage of configured maximum, number of bundles currently stored, number of bundles delivered, number of bundles dropped (expired or evicted), and the timestamp of the last completed contact
2. THE Node_Controller SHALL track cumulative statistics including: total bundles received, total bundles sent, total bytes received, total bytes sent, average delivery latency in seconds, contacts completed, contacts missed, and beacons transmitted
3. THE Node_Controller SHALL expose telemetry and statistics through a local interface (file, socket, or API) accessible to the node operator
4. WHEN a telemetry query is received, THE Node_Controller SHALL return the current telemetry snapshot within 1 second
5. THE Node_Controller SHALL collect DTN_Engine-specific telemetry through the Abstraction_Layer unified monitoring interface and merge it with local node metrics

### Requirement 17: Error Handling and Recovery

**User Story:** As a node operator, I want the system to handle faults gracefully and recover automatically, so that the terrestrial DTN node remains operational despite disruptions.

#### Acceptance Criteria

1. IF the Bundle_Store reaches capacity and eviction cannot free sufficient space for an incoming bundle, THEN THE BPA SHALL reject the incoming bundle and return a storage-full error to the sender if the LTP session is still active
2. IF a CRC validation fails on a received bundle, THEN THE BPA SHALL discard the corrupted bundle and log the corruption event with the source Callsign_EID and link metrics
3. IF the Node_Controller process crashes and restarts, THEN THE Node_Controller SHALL reload the Bundle_Store and Contact_Plan_Manager state from the local filesystem and resume normal operation without manual intervention
4. IF the USB connection to the TNC4 is lost during operation, THEN THE CLA SHALL detect the disconnection within 5 seconds, mark the current contact as interrupted, retain all queued bundles, and attempt to re-establish the USB connection at a configurable retry interval
5. IF no direct Contact_Window exists for a bundle's destination, THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window with that destination is added to the contact plan or the bundle's lifetime expires
6. IF the DTN_Engine process exits unexpectedly, THEN THE Node_Controller SHALL detect the failure through the Abstraction_Layer health check, log the engine failure reason, and attempt automatic restart after a configurable backoff interval

### Requirement 18: Terrestrial Node Performance

**User Story:** As a node operator, I want the terrestrial DTN node to operate within defined performance bounds, so that the system is responsive and predictable.

#### Acceptance Criteria

1. THE Node_Controller SHALL complete a full operation cycle (check contacts, transmit queued bundles, process received bundles, run cleanup) within 100 milliseconds
2. THE Bundle_Store SHALL complete a single store or retrieve operation within 10 milliseconds
3. THE BPA SHALL complete bundle validation (version, Callsign_EID, lifetime, timestamp, and CRC checks) within 5 milliseconds per bundle
4. THE CLA SHALL complete KISS frame encoding or decoding (via `radiant-kiss`) within 1 millisecond per frame

### Requirement 19: Shared-Key Digest Extension (Future — Regulatory Review Required)

**User Story:** As a network operator, I want the option to include a shared-key HMAC digest as bundle metadata, so that receiving stations can verify bundle authenticity and reject spoofed transmissions.

#### Acceptance Criteria

1. THE BPA SHALL support an optional BPv7 extension block containing an HMAC-SHA256 digest computed over the bundle primary block fields (source EID, destination EID, creation timestamp, lifetime)
2. THE digest extension block SHALL be appended as metadata alongside the bundle — the payload content SHALL remain fully readable in plaintext and SHALL NOT be altered or obscured by the digest
3. THE digest SHALL be computed using a pre-shared key distributed out-of-band between cooperating stations
4. WHEN a bundle with a digest extension block is received, THE BPA SHALL verify the digest if the shared key is configured, and log a warning if verification fails indicating possible spoofing
5. IF a received bundle does not contain a digest extension block, THEN THE BPA SHALL accept the bundle normally (backward compatibility with stations not implementing this extension)
6. THIS requirement SHALL NOT be implemented until regulatory review confirms that a shared-key digest appended as metadata (with payload remaining in the clear) is compliant with amateur radio regulations (ITU Article 25, applicable national regulations)
7. THE algorithm, extension block format, and key derivation method SHALL be published as part of the RADIANT protocol specification, enabling any amateur operator to inspect and understand the mechanism
