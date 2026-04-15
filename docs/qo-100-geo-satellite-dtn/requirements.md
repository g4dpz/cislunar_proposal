# Requirements Document

## Introduction

This document specifies the requirements for Phase 1.5 of the cislunar amateur DTN project: QO-100 Geostationary Satellite DTN validation. Phase 1.5 extends the Phase 1 terrestrial DTN system to operate over the QO-100 (Es'hail-2) geostationary amateur radio satellite, validating the DTN protocol stack with real space delays and RF propagation through space.

QO-100 is a geostationary satellite at 25.9°E providing amateur radio transponder services with a 2.4 GHz uplink and 10.45 GHz downlink. The satellite introduces approximately 500ms round-trip time (250ms one-way light time) and validates DTN operation over a real space link before progressing to LEO CubeSat missions.

The system reuses the Phase 1 software stack: ION-DTN (BPv7, LTP, BPSec), the dtn-node Go orchestrator, and AX.25 framing. The primary changes are hardware-specific: 2.4 GHz uplink transmitter, 10.45 GHz downlink receiver (LNB + SDR), and QO-100 modem or SDR for digital mode operation. The geostationary orbit eliminates pass prediction complexity — the satellite is always visible from the ground station's location, providing an always-on contact window with minimal Doppler shift.

This spec validates DTN ping and store-and-forward operations over the QO-100 satellite link, demonstrating that the protocol stack handles real space delays and RF propagation characteristics before advancing to LEO orbital mechanics and CubeSat hardware.

## Glossary

- **QO-100**: Es'hail-2 geostationary amateur radio satellite at 25.9°E providing 2.4 GHz uplink and 10.45 GHz downlink transponder
- **GEO**: Geostationary Earth Orbit — satellite orbit at approximately 35,786 km altitude with zero inclination, appearing stationary from ground
- **Uplink**: Ground-to-satellite transmission path at 2.4 GHz (2400-2450 MHz amateur allocation)
- **Downlink**: Satellite-to-ground transmission path at 10.45 GHz (10.45-10.5 GHz amateur allocation)
- **LNB**: Low-Noise Block downconverter — converts 10.45 GHz downlink to intermediate frequency for SDR reception
- **RTT**: Round-Trip Time — time for a signal to travel from ground to satellite and back, approximately 500ms for QO-100
- **Light_Time**: One-way signal propagation delay from ground to satellite or satellite to ground, approximately 250ms for QO-100
- **Link_Budget**: Calculation of signal power from transmitter to receiver accounting for path loss, antenna gain, and system noise
- **Doppler_Shift**: Frequency shift due to relative motion between transmitter and receiver — minimal for GEO satellites
- **Transponder**: Satellite payload that receives uplink signals, frequency-translates them, and retransmits on downlink
- **Dish_Antenna**: Parabolic reflector antenna providing high gain for satellite communication, typically 60-90cm diameter for QO-100
- **BPA**: Bundle Protocol Agent — ION-DTN core engine (reused from Phase 1)
- **Bundle_Store**: Persistent bundle storage (reused from Phase 1)
- **Contact_Plan_Manager**: Manages communication windows (reused from Phase 1, simplified for always-on GEO contact)
- **CLA**: Convergence Layer Adapter — interfaces ION-DTN with radio hardware (reused from Phase 1)
- **Node_Controller**: Top-level orchestrator (reused from Phase 1)
- **BPSec**: Bundle Protocol Security providing HMAC-SHA-256 integrity (reused from Phase 1)

## Requirements

### Requirement 1: QO-100 Uplink Transmission

**User Story:** As an amateur radio operator, I want to transmit DTN bundles to the QO-100 satellite on the 2.4 GHz uplink, so that I can validate DTN operation over a real space link.

#### Acceptance Criteria

1. THE Uplink_Transmitter SHALL transmit AX.25 frames carrying LTP segments on the 2.4 GHz amateur radio band (2400-2450 MHz) to the QO-100 satellite
2. THE Uplink_Transmitter SHALL use a dish antenna with a minimum gain of 15 dBi to achieve the required uplink power budget for QO-100
3. THE Uplink_Transmitter SHALL operate at a data rate compatible with the QO-100 transponder bandwidth, not exceeding 2 MHz occupied bandwidth
4. THE Uplink_Transmitter SHALL include the amateur radio callsign in every transmitted AX.25 frame for regulatory compliance
5. THE Uplink_Transmitter SHALL coordinate frequency usage with other QO-100 users to avoid interference, selecting an uplink frequency within the 2.4 GHz amateur allocation that is not currently occupied

### Requirement 2: QO-100 Downlink Reception

**User Story:** As an amateur radio operator, I want to receive DTN bundles from the QO-100 satellite on the 10.45 GHz downlink, so that I can complete the space link validation.

#### Acceptance Criteria

1. THE Downlink_Receiver SHALL receive AX.25 frames carrying LTP segments on the 10.45 GHz amateur radio band (10.45-10.5 GHz) from the QO-100 satellite
2. THE Downlink_Receiver SHALL use an LNB to downconvert the 10.45 GHz signal to an intermediate frequency suitable for SDR or modem reception
3. THE Downlink_Receiver SHALL use a dish antenna with a minimum gain of 20 dBi to achieve the required downlink signal-to-noise ratio for QO-100
4. THE Downlink_Receiver SHALL demodulate and decode AX.25 frames from the downlink signal and deliver them to the CLA for LTP reassembly
5. THE Downlink_Receiver SHALL measure and report downlink signal quality metrics including RSSI, SNR, and bit error rate

### Requirement 3: Geostationary Contact Window

**User Story:** As a ground station operator, I want the system to treat the QO-100 satellite as an always-on contact, so that I can transmit and receive bundles without pass prediction or scheduling complexity.

#### Acceptance Criteria

1. THE Contact_Plan_Manager SHALL configure a single continuous Contact_Window for the QO-100 satellite with no end time, representing the always-visible geostationary link
2. THE Node_Controller SHALL treat the QO-100 contact as active at all times when the ground station hardware is operational
3. THE Contact_Plan_Manager SHALL NOT require orbital pass prediction or Doppler compensation for the QO-100 geostationary satellite
4. IF the ground station hardware (uplink transmitter or downlink receiver) fails, THEN THE Node_Controller SHALL mark the QO-100 contact as interrupted and retain all queued bundles for retry when hardware is restored

### Requirement 4: Space Link Round-Trip Time Handling

**User Story:** As a DTN node operator, I want the system to handle the approximately 500ms round-trip time to QO-100, so that LTP acknowledgments and bundle delivery work correctly over the space link.

#### Acceptance Criteria

1. THE LTP_Engine SHALL configure retransmission timeouts to account for the QO-100 round-trip time of approximately 500ms, setting the minimum timeout to at least 600ms
2. WHEN a bundle is transmitted to QO-100, THE Node_Controller SHALL expect LTP acknowledgments to arrive no sooner than 500ms after transmission
3. THE BPA SHALL correctly handle ping echo responses with round-trip times in the range of 500-600ms, accounting for space link propagation delay
4. THE Node_Controller SHALL measure and report the actual round-trip time for each bundle transmission over the QO-100 link

### Requirement 5: QO-100 Link Budget Validation

**User Story:** As a ground station operator, I want to validate that my uplink and downlink meet the QO-100 link budget requirements, so that I can achieve reliable communication through the satellite.

#### Acceptance Criteria

1. THE Uplink_Transmitter SHALL achieve a minimum effective isotropic radiated power (EIRP) of 50 dBm (100 watts EIRP) on the 2.4 GHz uplink to close the QO-100 uplink budget
2. THE Downlink_Receiver SHALL achieve a minimum G/T (antenna gain to system noise temperature ratio) of 10 dB/K on the 10.45 GHz downlink to close the QO-100 downlink budget
3. THE Node_Controller SHALL compute and report the uplink and downlink link margins based on measured transmit power, antenna gain, LNB noise figure, and received signal strength
4. IF the computed link margin falls below 3 dB on either uplink or downlink, THEN THE Node_Controller SHALL log a link-budget warning and recommend hardware adjustments

### Requirement 6: DTN Ping Over QO-100

**User Story:** As an amateur radio operator, I want to ping a remote ground station through the QO-100 satellite and receive an echo response, so that I can verify end-to-end DTN reachability over the space link.

#### Acceptance Criteria

1. WHEN a ping request bundle is transmitted to QO-100 and relayed to a remote ground station, THE remote ground station SHALL generate a ping response bundle and transmit it back through QO-100
2. WHEN the ping response is received at the originating ground station, THE Node_Controller SHALL compute and report the total round-trip time including both space link hops (ground → QO-100 → remote ground → QO-100 → originating ground)
3. THE Node_Controller SHALL validate that the measured round-trip time is consistent with the expected QO-100 propagation delay (approximately 1000ms for two space link hops plus terrestrial processing time)
4. THE BPA SHALL successfully correlate ping responses to their original requests using the bundle ID included in the response payload

### Requirement 7: Store-and-Forward Over QO-100

**User Story:** As an amateur radio operator, I want to send data bundles through the QO-100 satellite to a remote ground station, so that I can validate DTN store-and-forward operation over a real space link.

#### Acceptance Criteria

1. WHEN a data bundle is transmitted to QO-100 and relayed to a remote ground station, THE remote ground station SHALL receive the bundle intact and deliver it to the local application agent
2. THE BPA SHALL transmit queued bundles in priority order (critical, expedited, normal, bulk) during the QO-100 contact window
3. WHEN a transmitted bundle is acknowledged by the remote ground station via LTP, THE Bundle_Store SHALL delete the acknowledged bundle
4. FOR ALL valid data bundles transmitted through QO-100, the bundle received at the remote ground station SHALL be identical to the bundle sent by the originating ground station (end-to-end integrity property)
5. THE Node_Controller SHALL measure and report the end-to-end delivery latency for each bundle transmitted through QO-100

### Requirement 8: BPSec Integrity Over QO-100

**User Story:** As a network operator, I want BPSec integrity protection to work over the QO-100 satellite link, so that I can detect any tampering or corruption during space link transmission.

#### Acceptance Criteria

1. WHEN a bundle with a BPSec Block Integrity Block (BIB) is transmitted through QO-100, THE remote ground station SHALL verify the integrity block and accept the bundle if verification succeeds
2. IF a bundle's BPSec integrity verification fails at the remote ground station, THEN THE BPA SHALL discard the bundle and log the integrity failure
3. THE BPA SHALL NOT apply BPSec encryption (Block Confidentiality Blocks) in compliance with amateur radio regulations requiring transmissions to be unencrypted
4. FOR ALL bundles transmitted through QO-100 with BPSec integrity blocks, the HMAC-SHA-256 verification SHALL succeed at the remote ground station if the bundle was not corrupted or tampered with during transmission

### Requirement 9: Frequency Coordination and Interference Avoidance

**User Story:** As an amateur radio operator, I want to coordinate my QO-100 uplink frequency with other users, so that I do not cause interference on the shared transponder.

#### Acceptance Criteria

1. THE Uplink_Transmitter SHALL allow the operator to configure the uplink frequency within the 2.4 GHz amateur allocation (2400-2450 MHz)
2. THE Node_Controller SHALL provide a spectrum monitoring mode that displays current QO-100 downlink activity, allowing the operator to identify unused uplink frequencies
3. THE Uplink_Transmitter SHALL limit occupied bandwidth to no more than 2 MHz to avoid excessive transponder usage
4. THE Node_Controller SHALL log the selected uplink frequency and occupied bandwidth for each transmission session

### Requirement 10: Minimal Doppler Compensation

**User Story:** As a ground station operator, I want the system to handle the minimal Doppler shift from the QO-100 geostationary satellite, so that frequency tracking remains simple compared to LEO satellites.

#### Acceptance Criteria

1. THE Uplink_Transmitter SHALL NOT require dynamic Doppler compensation for the QO-100 geostationary satellite, as the satellite's relative velocity to the ground station is negligible
2. THE Downlink_Receiver SHALL tolerate a maximum Doppler shift of ±100 Hz on the 10.45 GHz downlink due to satellite stationkeeping maneuvers
3. IF the measured Doppler shift exceeds ±100 Hz, THEN THE Node_Controller SHALL log a Doppler anomaly warning indicating possible satellite maneuver or ground station pointing error

### Requirement 11: Hardware Configuration and Validation

**User Story:** As a ground station operator, I want to configure and validate my QO-100 hardware setup, so that I can ensure the system is ready for DTN operation over the satellite link.

#### Acceptance Criteria

1. THE Node_Controller SHALL load a configuration file specifying the uplink transmitter device, downlink receiver device, dish antenna parameters (gain, beamwidth), LNB local oscillator frequency, and QO-100 uplink/downlink frequencies
2. THE Node_Controller SHALL perform a hardware validation check at startup, verifying that the uplink transmitter and downlink receiver devices are accessible and operational
3. IF hardware validation fails, THEN THE Node_Controller SHALL log the specific hardware failure (transmitter not found, receiver not responding, LNB not powered) and refuse to start until the issue is resolved
4. THE Node_Controller SHALL provide a test mode that transmits a beacon on the QO-100 uplink and verifies reception on the downlink, confirming end-to-end hardware functionality

### Requirement 12: QO-100 Telemetry and Link Metrics

**User Story:** As a ground station operator, I want to monitor QO-100 link quality and performance, so that I can detect and respond to degraded space link conditions.

#### Acceptance Criteria

1. THE Node_Controller SHALL collect and report QO-100-specific telemetry including: uplink EIRP, downlink RSSI, downlink SNR, downlink bit error rate, measured round-trip time, and link margin
2. THE Node_Controller SHALL track cumulative QO-100 statistics including: total bundles transmitted through QO-100, total bundles received from QO-100, total bytes transmitted, total bytes received, and average delivery latency
3. THE Node_Controller SHALL expose QO-100 telemetry and statistics through the same local interface used for Phase 1 terrestrial telemetry (file, socket, or API)
4. WHEN a telemetry query is received, THE Node_Controller SHALL return the current QO-100 telemetry snapshot within 1 second

### Requirement 13: Reuse Phase 1 Software Stack

**User Story:** As a system architect, I want to reuse the Phase 1 software stack with minimal modifications, so that QO-100 validation builds on proven terrestrial DTN components.

#### Acceptance Criteria

1. THE QO-100 system SHALL reuse the Phase 1 BPA, Bundle_Store, Contact_Plan_Manager, CLA, and Node_Controller components without modification to their core logic
2. THE QO-100 system SHALL extend the CLA to support the QO-100 uplink transmitter and downlink receiver hardware interfaces in addition to the Phase 1 TNC4 interface
3. THE QO-100 system SHALL reuse the Phase 1 ION-DTN configuration format (ionrc, ltprc, bprc, bpsecrc) with QO-100-specific contact plan entries
4. THE QO-100 system SHALL reuse the Phase 1 dtn-node Go orchestrator CLI with a QO-100-specific configuration file

### Requirement 14: QO-100 Error Handling and Recovery

**User Story:** As a ground station operator, I want the system to handle QO-100 link failures gracefully and recover automatically, so that the DTN node remains operational despite space link disruptions.

#### Acceptance Criteria

1. IF the QO-100 uplink transmitter fails (device not responding, power amplifier fault, or antenna pointing error), THEN THE Node_Controller SHALL mark the QO-100 contact as interrupted, retain all queued bundles, and attempt to re-establish the uplink at a configurable retry interval
2. IF the QO-100 downlink receiver fails (LNB not powered, SDR device error, or signal loss), THEN THE Node_Controller SHALL mark the QO-100 contact as interrupted and attempt to re-establish the downlink at a configurable retry interval
3. IF the QO-100 transponder is temporarily unavailable (satellite eclipse, transponder maintenance, or uplink interference), THEN THE Node_Controller SHALL detect the loss of downlink signal within 10 seconds and suspend transmission until the transponder is restored
4. WHEN the QO-100 link is restored after an interruption, THE Node_Controller SHALL resume bundle transmission from the Bundle_Store in priority order without manual intervention

### Requirement 15: QO-100 Performance Targets

**User Story:** As a ground station operator, I want the QO-100 DTN system to operate within defined performance bounds, so that the system is responsive and predictable over the space link.

#### Acceptance Criteria

1. THE Node_Controller SHALL complete a full operation cycle (check QO-100 contact, transmit queued bundles, process received bundles, run cleanup) within 200 milliseconds, accounting for the increased round-trip time compared to Phase 1 terrestrial links
2. THE BPA SHALL successfully deliver a ping echo response over the QO-100 link with a total round-trip time of 500-600ms for a single ground-to-satellite-to-ground hop
3. THE BPA SHALL successfully deliver a data bundle over the QO-100 link with an end-to-end latency of 1000-1200ms for a ground-to-satellite-to-remote-ground path (two space link hops)
