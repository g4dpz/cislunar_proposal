# Requirements Document

## Introduction

This document specifies the requirements for Phase 1 of the cislunar amateur DTN project: terrestrial DTN validation. Phase 1 deploys ION-DTN (NASA JPL's Interplanetary Overlay Network) on Linux or macOS hosts connected via USB to Mobilinkd TNC4 terminal node controllers, which drive Yaesu FT-817 radios at 9600 baud through the 9600 baud data port using G3RUH-compatible GFSK modulation on VHF/UHF amateur bands.

The system supports two core operations: ping (DTN reachability test) and store-and-forward (point-to-point bundle delivery). There is no relay functionality — nodes do not forward bundles on behalf of other nodes. All bundle delivery is direct (source → destination).

The protocol stack is BPv7 bundles over LTP sessions over AX.25 frames. AX.25 provides callsign-based source/destination addressing for amateur radio regulatory compliance. LTP provides reliable transfer with deferred acknowledgment. BPSec provides integrity protection (no encryption, per amateur radio regulations requiring transmissions to be unencrypted).

This spec is scoped exclusively to terrestrial ground nodes. STM32U585 OBC, IQ baseband/SDR, Ettus B200mini, CGR contact prediction, orbital mechanics, space segment (CubeSat, cislunar), S-band/X-band communications, and flight-qualified hardware are out of scope.

## Glossary

- **BPA**: Bundle Protocol Agent — the core ION-DTN engine that creates, receives, validates, stores, and delivers BPv7 bundles
- **Bundle**: A BPv7 protocol data unit carrying a payload between DTN endpoints
- **Bundle_Store**: Persistent storage subsystem on the local filesystem for bundles awaiting delivery
- **Contact_Plan_Manager**: Subsystem that maintains a manually configured schedule of communication windows between ground nodes
- **CLA**: Convergence Layer Adapter — interfaces ION-DTN's AX.25/LTP stack with the Mobilinkd TNC4 over USB
- **Node_Controller**: Top-level orchestrator that ties together BPA, Bundle_Store, Contact_Plan_Manager, and CLA on each host node
- **ION-DTN**: NASA JPL's Interplanetary Overlay Network — the DTN implementation providing BPv7, LTP, and related protocols
- **LTP**: Licklider Transmission Protocol — runs on top of AX.25 providing reliable transfer with deferred acknowledgment
- **AX.25**: Link-layer framing protocol providing callsign-based source/destination addressing for amateur radio compliance
- **TNC4**: Mobilinkd TNC4 terminal node controller — USB-connected TNC interfacing with the FT-817 radio
- **FT-817**: Yaesu FT-817 portable transceiver with 9600 baud data port for G3RUH-compatible GFSK modulation
- **BPSec**: Bundle Protocol Security (RFC 9172) — provides integrity blocks for bundle origin authentication
- **Ping**: DTN reachability test — send a bundle echo request and receive an echo response
- **Store_and_Forward**: Point-to-point bundle delivery where a source node sends a bundle directly to a destination node during a contact window
- **Contact_Window**: A scheduled time interval during which two ground nodes can communicate over their radio link
- **Endpoint_ID**: A DTN endpoint identifier using the "dtn" or "ipn" URI scheme that uniquely addresses a node or application

## Requirements

### Requirement 1: Bundle Creation and Validation

**User Story:** As an amateur radio operator, I want to create and validate DTN bundles, so that I can send well-formed messages through the terrestrial DTN network.

#### Acceptance Criteria

1. WHEN a user submits a message with a valid destination Endpoint_ID and payload, THE BPA SHALL create a BPv7 bundle with the bundle version set to 7, proper source and destination Endpoint_IDs, a CRC integrity check, a priority level (critical, expedited, normal, or bulk), and a positive lifetime value in seconds
2. WHEN a bundle is received from the CLA, THE BPA SHALL validate that the bundle version equals 7, the destination is a well-formed Endpoint_ID, the lifetime is greater than zero, the creation timestamp does not exceed the current time, and the CRC is correct
3. IF a received bundle fails any validation check, THEN THE BPA SHALL discard the bundle and log the specific validation failure reason along with the source Endpoint_ID
4. THE BPA SHALL support three bundle types: data bundles for store-and-forward payload delivery, ping request bundles for echo requests, and ping response bundles for echo responses
5. FOR ALL valid Bundle objects, serializing a Bundle to its BPv7 wire format and then parsing the wire format back SHALL produce a Bundle equivalent to the original (round-trip property)

### Requirement 2: Bundle Storage and Persistence

**User Story:** As a terrestrial DTN node operator, I want bundles to be stored persistently on the local filesystem and survive process restarts and power cycles, so that no messages are lost during disruptions.

#### Acceptance Criteria

1. WHEN a valid bundle is accepted, THE Bundle_Store SHALL persist the bundle to the local filesystem atomically, preventing corruption if the process is interrupted during the write
2. WHEN a bundle is stored and later retrieved by its bundle ID (source Endpoint_ID, creation timestamp, sequence number), THE Bundle_Store SHALL return a bundle identical to the original (round-trip property)
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

1. WHEN the BPA receives a ping request bundle addressed to a local endpoint, THE BPA SHALL generate exactly one ping response bundle with the destination set to the original sender's Endpoint_ID
2. WHEN a ping response is generated, THE BPA SHALL queue the response in the Bundle_Store for delivery during the next available Contact_Window with the sender's node
3. WHEN a ping response is received at the originating node, THE Node_Controller SHALL compute and report the round-trip time from the original ping request creation timestamp to the response receipt time
4. THE BPA SHALL include the original ping request's bundle ID in the ping response payload so the originating node can correlate responses to requests

### Requirement 5: Store-and-Forward Operation

**User Story:** As an amateur radio operator, I want to send messages that are stored at my node and delivered directly to the destination node when a contact window opens, so that I can communicate across disrupted terrestrial links.

#### Acceptance Criteria

1. WHEN a data bundle is received whose destination matches a local Endpoint_ID, THE BPA SHALL deliver the bundle payload to the local application agent
2. WHEN a data bundle is created or received whose destination is a remote Endpoint_ID, THE BPA SHALL store the bundle in the Bundle_Store and queue it for direct delivery during the next Contact_Window with the destination node
3. THE BPA SHALL transmit queued bundles in priority order (critical first, then expedited, then normal, then bulk) during each Contact_Window
4. WHEN a transmitted bundle is acknowledged by the remote node via LTP, THE Bundle_Store SHALL delete the acknowledged bundle
5. IF a bundle transmission is not acknowledged within the LTP retransmission timeout, THEN THE Bundle_Store SHALL retain the bundle for retry during the next Contact_Window

### Requirement 6: No Relay Constraint

**User Story:** As a system architect, I want to enforce that terrestrial nodes only deliver bundles directly to their final destination, so that the system remains simple and predictable without multi-hop relay complexity.

#### Acceptance Criteria

1. THE BPA SHALL transmit a bundle only to the node matching the bundle's final destination Endpoint_ID — the BPA SHALL NOT forward bundles on behalf of other nodes
2. WHEN the Node_Controller looks up a delivery route for a bundle, THE Contact_Plan_Manager SHALL return only direct Contact_Windows with the destination node, with no multi-hop paths


### Requirement 7: Contact Plan Management

**User Story:** As a ground station operator, I want to manage scheduled communication windows between terrestrial nodes, so that bundles are delivered during available contact opportunities.

#### Acceptance Criteria

1. THE Contact_Plan_Manager SHALL maintain a time-tagged schedule of Contact_Windows, each specifying a remote node ID, start time, end time, data rate in bits per second, and link type (VHF or UHF)
2. WHEN queried for active contacts at a given time, THE Contact_Plan_Manager SHALL return all Contact_Windows whose start time is at or before the query time and whose end time is after the query time
3. WHEN queried for the next contact with a specific destination node, THE Contact_Plan_Manager SHALL return the earliest future Contact_Window matching that destination
4. THE Contact_Plan_Manager SHALL reject any contact plan update that would create overlapping Contact_Windows on the same link for a given node
5. THE Contact_Plan_Manager SHALL validate that all Contact_Windows fall within the plan's valid-from and valid-to time range
6. WHEN a contact plan is loaded or updated, THE Contact_Plan_Manager SHALL persist the plan to the local filesystem so it survives process restarts
7. THE Contact_Plan_Manager SHALL support loading contact plans from a human-readable configuration file (ION-DTN contact plan format)

### Requirement 8: Contact Window Execution

**User Story:** As a terrestrial DTN node, I want to transmit queued bundles during active contact windows, so that messages are delivered when communication links are available.

#### Acceptance Criteria

1. WHEN a Contact_Window becomes active (current time reaches the window start time), THE CLA SHALL establish the AX.25/LTP link via the TNC4 and THE Node_Controller SHALL begin transmitting queued bundles destined for the contact's remote node
2. THE Node_Controller SHALL cease all transmission when the Contact_Window end time is reached
3. WHEN a Contact_Window completes, THE Node_Controller SHALL record link metrics (bytes transferred, duration, bundles sent, bundles received) and update contact statistics
4. IF the CLA fails to establish a link during a scheduled Contact_Window (TNC4 not responding, radio not keyed, or no AX.25 connection established), THEN THE Node_Controller SHALL mark the contact as missed, retain all queued bundles for the next window, and increment the contacts-missed counter

### Requirement 9: AX.25 and LTP Convergence Layer

**User Story:** As an amateur radio operator, I want all DTN transmissions to use AX.25 framing with callsign addressing over LTP, so that every transmission complies with amateur radio regulations and provides reliable transfer.

#### Acceptance Criteria

1. THE CLA SHALL encapsulate all bundle transmissions in AX.25 frames carrying the source amateur radio callsign and the destination amateur radio callsign
2. THE CLA SHALL run LTP sessions on top of AX.25 frames, providing reliable transfer with deferred acknowledgment for all bundle delivery
3. THE CLA SHALL perform LTP segmentation for bundles that exceed a single AX.25 frame size, and reassemble received LTP segments into complete bundles
4. THE CLA SHALL interface with the Mobilinkd TNC4 via USB serial connection (not Bluetooth) for all AX.25 packet operations
5. THE CLA SHALL drive the FT-817 radio at 9600 baud through its 9600 baud data port using G3RUH-compatible GFSK modulation
6. FOR ALL valid Bundle objects, encapsulating a bundle into AX.25/LTP frames and then reassembling the frames back into a bundle SHALL produce a bundle equivalent to the original (round-trip property)

### Requirement 10: BPSec Integrity Protection

**User Story:** As a network operator, I want bundle integrity protection using BPSec, so that the DTN network is protected against spoofing and tampering while complying with amateur radio regulations that prohibit encryption.

#### Acceptance Criteria

1. THE BPA SHALL support BPSec (RFC 9172) Block Integrity Blocks (BIB) for bundle origin authentication using HMAC-SHA-256
2. THE BPA SHALL NOT apply BPSec Block Confidentiality Blocks (BCB) or any form of payload encryption, in compliance with amateur radio regulations requiring transmissions to be unencrypted
3. WHEN a bundle with a BIB is received, THE BPA SHALL verify the integrity block and discard the bundle if verification fails, logging the integrity failure with the source Endpoint_ID
4. THE Node_Controller SHALL store BPSec shared keys in a configuration file on the local filesystem with file permissions restricted to the node operator's user account

### Requirement 11: Priority-Based Message Handling

**User Story:** As a node operator, I want bundles to be handled according to their priority level, so that critical messages are delivered before less urgent ones.

#### Acceptance Criteria

1. THE BPA SHALL assign one of four priority levels to each bundle: critical (highest), expedited, normal, or bulk (lowest)
2. WHEN multiple bundles are queued for the same destination during a Contact_Window, THE Node_Controller SHALL transmit bundles in strict priority order — all critical bundles before any expedited, all expedited before any normal, all normal before any bulk
3. WHEN the Bundle_Store must evict bundles to free space, THE Bundle_Store SHALL evict bulk bundles first, then normal, then expedited — critical bundles SHALL be evicted only when no lower-priority bundles remain
4. THE BPA SHALL accept a default priority level from the Node_Controller configuration, applied to bundles that do not specify an explicit priority

### Requirement 12: Rate Limiting and Store Protection

**User Story:** As a node operator, I want to protect the bundle store from flooding, so that a misbehaving or malicious node cannot exhaust storage resources.

#### Acceptance Criteria

1. THE BPA SHALL enforce a configurable maximum bundle acceptance rate (bundles per second) per source Endpoint_ID
2. IF the acceptance rate from a single source Endpoint_ID exceeds the configured limit, THEN THE BPA SHALL reject additional bundles from that source and log the rate-limit event
3. THE BPA SHALL enforce a configurable maximum bundle size in bytes, rejecting any bundle whose total serialized size exceeds the limit

### Requirement 13: Node Health and Telemetry

**User Story:** As a node operator, I want to monitor node health and performance, so that I can detect and respond to anomalies in the terrestrial DTN network.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report telemetry including: uptime in seconds, storage utilization as a percentage of configured maximum, number of bundles currently stored, number of bundles delivered, number of bundles dropped (expired or evicted), and the timestamp of the last completed contact
2. THE Node_Controller SHALL track cumulative statistics including: total bundles received, total bundles sent, total bytes received, total bytes sent, average delivery latency in seconds, contacts completed, and contacts missed
3. THE Node_Controller SHALL expose telemetry and statistics through a local interface (file, socket, or API) accessible to the node operator
4. WHEN a telemetry query is received, THE Node_Controller SHALL return the current telemetry snapshot within 1 second

### Requirement 14: Error Handling and Recovery

**User Story:** As a node operator, I want the system to handle faults gracefully and recover automatically, so that the terrestrial DTN node remains operational despite disruptions.

#### Acceptance Criteria

1. IF the Bundle_Store reaches capacity and eviction cannot free sufficient space for an incoming bundle, THEN THE BPA SHALL reject the incoming bundle and return a storage-full error to the sender if the LTP session is still active
2. IF a CRC validation fails on a received bundle, THEN THE BPA SHALL discard the corrupted bundle and log the corruption event with the source Endpoint_ID and link metrics
3. IF the Node_Controller process crashes and restarts, THEN THE Node_Controller SHALL reload the Bundle_Store and Contact_Plan_Manager state from the local filesystem and resume normal operation without manual intervention
4. IF the USB connection to the TNC4 is lost during operation, THEN THE CLA SHALL detect the disconnection within 5 seconds, mark the current contact as interrupted, retain all queued bundles, and attempt to re-establish the USB connection at a configurable retry interval
5. IF no direct Contact_Window exists for a bundle's destination, THEN THE Bundle_Store SHALL retain the bundle until a Contact_Window with that destination is added to the contact plan or the bundle's lifetime expires

### Requirement 15: Terrestrial Node Performance

**User Story:** As a node operator, I want the terrestrial DTN node to operate within defined performance bounds, so that the system is responsive and predictable.

#### Acceptance Criteria

1. THE Node_Controller SHALL complete a full operation cycle (check contacts, transmit queued bundles, process received bundles, run cleanup) within 100 milliseconds
2. THE Bundle_Store SHALL complete a single store or retrieve operation within 10 milliseconds
3. THE BPA SHALL complete bundle validation (version, Endpoint_ID, lifetime, timestamp, and CRC checks) within 5 milliseconds per bundle
