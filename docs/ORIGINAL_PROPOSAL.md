# Cislunar Amateur Delay/Disruption Tolerant Networking (DTN) Payload
## Phased Demonstration Proposal (Terrestrial → Engineering Model → LEO CubeSat → Cislunar)

---

## 1. Executive Summary

This proposal outlines a four-phase approach to developing and deploying an amateur radio payload capable of demonstrating Delay/Disruption Tolerant Networking (DTN) beyond Earth orbit. The mission progresses through:

1. Terrestrial DTN validation using amateur radio infrastructure (RPi + Mobilinkd TNC4 + FT-817)
2. CubeSat Engineering Model (EM) ground testing on flight-representative hardware (STM32U585 + Ettus B200mini SDR)
3. Low Earth Orbit (LEO) CubeSat DTN payload demonstration (STM32U585 + flight IQ transceiver)
4. Cislunar DTN node enabling deep-space amateur communications

The system uses NASA JPL's ION-DTN implementation (BPv7/LTP) running over AX.25 link-layer framing across all phases. All packets carry amateur radio callsigns for source and destination, satisfying regulatory requirements. The system supports two core operations: **ping** (DTN reachability testing) and **store-and-forward** (point-to-point bundle delivery). There is **no relay functionality** — nodes do not forward bundles on behalf of other nodes.

The objective is to establish a technically feasible, community-driven pathway toward the first amateur-operated interplanetary-style communication system.

---

## 2. Mission Objectives

### Primary Objectives
- Demonstrate ION-DTN protocols (BPv7/LTP over AX.25) in constrained and disrupted communication environments
- Enable store-and-forward messaging and DTN ping for amateur radio operators
- Validate DTN performance across increasing latency regimes (ms → seconds → minutes)

### Secondary Objectives
- Provide an open experimental platform for education and research
- Develop reusable open-source DTN flight and ground software based on ION-DTN
- Encourage global participation in deep-space communication experiments

---

## 3. System Overview

The proposed system consists of:

- Space segment (CubeSat and later cislunar payload, both using STM32U585 OBC with IQ baseband radio)
- Ground segment (amateur ground stations, tiered by capability)
- Engineering Model segment (ground-based flatsat for flight validation)
- DTN protocol stack: ION-DTN (BPv7 bundles over LTP sessions over AX.25 frames)

Key features:
- Store-and-forward message delivery (no relay)
- DTN ping for reachability testing
- AX.25 callsign-based addressing on all transmissions (regulatory compliance)
- LTP reliable transfer with deferred acknowledgment
- CGR-based contact prediction for pass scheduling (not multi-hop routing)
- Priority-based message handling and eviction

### Protocol Stack (All Phases)

| Layer | Protocol | Purpose |
|-------|----------|---------|
| Application | ION-DTN BPv7 | Bundle creation, storage, delivery, ping |
| Transport | LTP (RFC 5326) | Reliable transfer with deferred ACK |
| Link | AX.25 | Callsign framing, amateur radio compliance |
| Physical | GFSK/GMSK/BPSK | Modulation (varies by phase) |

---

## 4. Phase 1: Terrestrial DTN Demonstration

### 4.1 Objectives
- Validate ION-DTN (BPv7/LTP) over AX.25 amateur radio links
- Demonstrate ping and store-and-forward operations
- Build community engagement and participation

### 4.2 Architecture
- Fixed and mobile amateur radio stations
- UHF/VHF packet radio links at 9600 baud
- Raspberry Pi hosts connected via USB to Mobilinkd TNC4 TNCs

### 4.3 Hardware
- **Host computer**: Raspberry Pi
- **TNC**: Mobilinkd TNC4 (USB connection to RPi)
- **Radio**: Yaesu FT-817 (9600 baud data port, G3RUH-compatible GFSK)
- **Data rate**: 9600 baud

### 4.4 Implementation
- Use existing amateur bands for packet communication
- Deploy ION-DTN on Raspberry Pi hosts
- AX.25/LTP convergence layer stack with callsign addressing
- No relay — direct store-and-forward only

### 4.5 Experiments
- DTN ping reachability testing between ground nodes
- Store-and-forward messaging between ground stations
- Scheduled connectivity windows with artificial delay injection

### 4.6 Deliverables
- ION-DTN ground station software package
- Performance data under varying conditions
- Community documentation and tutorials

---

## 5. Phase 2: CubeSat Engineering Model (EM)

### 5.1 Objectives
- Validate flight software stack on flight-representative hardware
- Confirm store-and-forward and ping operations under flight constraints
- Profile power budget and validate thermal/vacuum readiness
- Bridge terrestrial testing and flight commitment

### 5.2 Hardware
- **OBC**: STM32U585 ultra-low-power ARM Cortex-M33 (160 MHz, 2 MB flash, 786 KB SRAM, hardware crypto, TrustZone)
- **RF front-end**: Ettus Research USRP B200mini SDR (USB 3.0, 12-bit ADC/DAC, 70 MHz–6 GHz, full-duplex IQ)
- **IQ bridge**: Companion Raspberry Pi or PC running UHD driver, bridging IQ samples to STM32U585 via SPI/UART/DMA
- **Storage**: External SPI/QSPI NVM (64–256 MB) for persistent bundle store
- **Data rate**: 9.6 kbps UHF (matching flight configuration)

### 5.3 Architecture
The EM uses the identical STM32U585 board and ION-DTN software that will fly. The STM32U585 generates and processes IQ baseband samples directly via its DMA engine. Since the STM32U585 lacks USB 3.0 host capability, the B200mini connects to a companion RPi/PC which bridges IQ samples to the STM32U585. The B200mini is EM-only — the flight unit replaces it with a dedicated flight-qualified IQ transceiver IC.

### 5.4 Validation Activities
- End-to-end DTN store-and-forward through EM with real over-the-air RF
- DTN ping echo request/response on flight hardware
- Simulated orbital pass testing (5–10 min windows, 4–6 passes/day)
- Power budget profiling (STM32U585 Stop 2 mode ~16 µA idle)
- Store-and-forward under constrained resources (786 KB SRAM, 64–256 MB NVM)
- IQ baseband radio integration validation
- B200mini-to-flight-transceiver transition characterization
- Fault injection (power loss, watchdog resets, memory corruption)

### 5.5 Deliverables
- Validated flight software stack on STM32U585
- Power budget and thermal characterization data
- IQ interface specification for flight transceiver transition

---

## 6. Phase 3: LEO CubeSat DTN Payload

### 6.1 Objectives
- Demonstrate ground-to-space DTN ping and store-and-forward
- Validate ION-DTN operations under orbital dynamics
- No relay — direct delivery to destination ground stations only

### 6.2 Payload Description
- **OBC**: STM32U585 (identical to EM)
- **RF**: Flight-qualified IQ transceiver IC interfacing directly with STM32U585 via DAC/ADC or SPI
- **Mass**: ~1–3 kg
- **Power**: 5–10 W average (STM32U585 Stop 2 mode between passes)
- **Storage**: External SPI/QSPI NVM (64–256 MB)

### 6.3 Communications
- UHF 437 MHz for global accessibility
- Data rate: 9.6 kbps (GMSK/BPSK via IQ baseband)
- AX.25/LTP/BPv7 protocol stack (ION-DTN)
- All frames carry source/destination callsigns

### 6.4 DTN Functionality
- Store-and-forward message delivery (no relay)
- DTN ping for reachability testing
- Scheduled contact windows with ground stations (CGR-predicted)
- Priority-based message handling and eviction

### 6.5 Contact Prediction
- ION-DTN's CGR module predicts pass windows using orbital parameters (TLE/ephemeris)
- SGP4/SDP4 orbit propagation for line-of-sight computation
- CGR used exclusively for contact prediction / pass scheduling — not multi-hop routing
- Contact predictions updated when fresh orbital data is received during ground passes

### 6.6 Ground Segment
- Distributed amateur ground stations (Tier 1–4)
- Open-source ION-DTN client software
- Automated scheduling via CGR-predicted contact windows

### 6.7 Experiments
- End-to-end DTN ping via satellite
- Store-and-forward messaging via satellite (direct delivery)
- Latency-tolerant file transfer

### 6.8 Deliverables
- Flight-proven DTN payload design
- Operational procedures for amateur users
- Public dataset of DTN performance in orbit

---

## 7. Phase 4: Cislunar DTN Mission

### 7.1 Objectives
- Extend DTN operations beyond Earth orbit
- Demonstrate long-delay networking (1–2 seconds one-way)
- Enable amateur participation in deep-space communication

### 7.2 Mission Concept
- Hosted payload on a cislunar spacecraft or lunar CubeSat
- Highly elliptical Earth orbit or lunar transfer trajectory

### 7.3 Payload Description
- **OBC**: STM32U585 or more capable processor (flexible)
- **RF**: IQ baseband S-band transceiver
- **Mass**: <5 kg
- **Power**: 10–20 W
- **Storage**: External NVM (256 MB–1 GB)

### 7.4 Communications
- Primary: S-band 2.2 GHz at 500 bps (BPSK + LDPC/Turbo FEC)
- Backup: UHF beacon for tracking
- Directional antenna required
- AX.25/LTP/BPv7 protocol stack (ION-DTN)

### 7.5 Ground Segment
- Tier 3/4 stations: 3–5 m dishes or phased arrays (S-band/X-band)
- Optional collaboration with university ground stations

### 7.6 DTN Functionality
- Store-and-forward with long-duration message storage (no relay)
- DTN ping for reachability testing
- CGR-based contact prediction for scheduling
- Priority-based message handling

### 7.7 Experiments
- Earth–cislunar DTN ping and store-and-forward messaging
- Delay-tolerant file transfer
- Network resilience under extreme latency

### 7.8 Deliverables
- First amateur cislunar DTN node
- Deep-space communication datasets
- Open framework for future interplanetary amateur missions

---

## 8. Technical Challenges

- Link budget constraints at cislunar distances (7 dB margin at 500 bps S-band)
- Limited power and antenna gain
- STM32U585 SRAM constraints (786 KB for concurrent bundle processing and IQ buffers)
- Radiation effects on electronics
- Frequency coordination and regulatory compliance
- IQ transceiver IC selection for flight (must match B200mini interface characteristics)

Mitigation strategies include conservative data rates, robust error correction (LDPC/Turbo), phased validation through the EM phase, and STM32U585 TrustZone/hardware crypto for security.

---

## 9. Community and Open Access

The project will be fully open and community-driven:

- Open-source ION-DTN ground station software
- Open hardware designs for terrestrial nodes (RPi + TNC4 + FT-817)
- Public documentation and data access
- Inclusive participation for licensed amateur operators worldwide

---

## 10. Partnerships and Opportunities

Potential collaborators include:

- Amateur satellite organizations
- Universities and research institutions
- Commercial CubeSat and launch providers
- NASA JPL (ION-DTN community)

---

## 11. Conclusion

This four-phase approach reduces risk while enabling progressive validation of DTN technologies in increasingly challenging environments. The Engineering Model phase bridges terrestrial testing and flight, ensuring the STM32U585-based flight software is validated on representative hardware before orbital deployment. By leveraging ION-DTN, AX.25 callsign compliance, and the global amateur radio community, the project aims to pioneer accessible deep-space networking and lay the foundation for future interplanetary amateur communication systems.

---

## 12. Link Budget Analysis

### 12.1 LEO CubeSat Link Budget (UHF)

**Configuration:**
- Frequency: 437 MHz (amateur UHF band)
- Altitude: 500 km
- Transmit Power: 2 W (33 dBm)
- Satellite antenna gain: 0 dBi (omnidirectional)
- Ground antenna gain: 12 dBi (Yagi)
- System losses: 2 dB
- Data rate: 9.6 kbps (GMSK/BPSK via STM32U585 IQ baseband)

**Result:** Link margin ≈ 31 dB. Robust link achievable with modest amateur ground stations.

### 12.2 Cislunar Link Budget (S-band)

**Configuration:**
- Frequency: 2.2 GHz (S-band)
- Distance: 384,000 km (Earth–Moon average)
- Transmit Power: 5 W (37 dBm)
- Spacecraft antenna gain: 10 dBi (directional patch)
- Ground antenna gain: 35 dBi (3–5 m dish)
- System losses: 3 dB
- Data rate: 500 bps (BPSK + LDPC/Turbo via STM32U585 IQ baseband)

**Result:** Link margin ≈ 7 dB. Feasible with low data rates, high-gain ground stations, and strong FEC. Viable for Tier 3/4 stations.

### 12.3 Eb/N₀ Summary

| Phase | Data Rate | Eb/N₀ (after losses) | Modulation | FEC |
|-------|-----------|---------------------|------------|-----|
| Terrestrial | 9600 baud | High margin | GFSK/G3RUH | Optional |
| LEO | 9.6 kbps | ~30 dB | GMSK/BPSK | Convolutional/LDPC |
| Cislunar | 500 bps | ~1–2 dB | BPSK | Strong LDPC/Turbo |

---

## 13. Ground Station Tier Model

### Tier 1: Entry-Level Amateur Station
- Equipment: handheld or small Yagi antenna
- Bands: VHF/UHF
- Capability: LEO reception and DTN participation

### Tier 2: Advanced Amateur Station
- Equipment: rotator-mounted Yagi or small dish
- Bands: UHF/S-band
- Capability: full LEO uplink/downlink

### Tier 3: Deep-Space Amateur / Institutional Station
- Equipment: 3–5 m dish or phased array
- Bands: S-band/X-band
- Capability: cislunar communication, low-rate DTN links

### Tier 4: Partner/University Station
- Large dishes, low-noise front ends
- Provides backbone support for cislunar phase

---

## 14. Key Hardware Summary

| Component | Phase | Purpose |
|-----------|-------|---------|
| Raspberry Pi | Terrestrial | Host computer for ground nodes |
| Mobilinkd TNC4 | Terrestrial | USB TNC for AX.25 packet radio |
| Yaesu FT-817 | Terrestrial | VHF/UHF transceiver, 9600 baud data port |
| STM32U585 | EM, LEO, Cislunar | Ultra-low-power OBC (Cortex-M33, TrustZone, crypto) |
| Ettus B200mini | EM only | Lab-grade SDR RF front-end (IQ interface) |
| Flight IQ transceiver IC | LEO, Cislunar | Flight-qualified RF front-end (replaces B200mini) |
| External SPI/QSPI NVM | EM, LEO, Cislunar | 64–256 MB persistent bundle storage |

---

## 15. Key Software Dependencies

| Component | Purpose |
|-----------|---------|
| ION-DTN (NASA JPL) | BPv7, LTP, CGR — primary DTN implementation |
| AX.25 | Link-layer framing with callsign addressing |
| LTP (RFC 5326) | Reliable transfer with deferred ACK |
| UHD (Ettus) | B200mini SDR driver (EM only) |
| SGP4/SDP4 | Orbit propagation for CGR contact prediction |
| STM32 HAL/LL | Hardware abstraction for STM32U585 peripherals |
| LDPC/Turbo codecs | FEC for cislunar phase |

---

## Appendix A: Acronyms

- AX.25: Amateur X.25 link-layer protocol
- BPv7: Bundle Protocol Version 7
- CGR: Contact Graph Routing (used for contact prediction only)
- CLA: Convergence Layer Adapter
- DTN: Delay/Disruption Tolerant Networking
- EM: Engineering Model
- FEC: Forward Error Correction
- GFSK: Gaussian Frequency Shift Keying
- IQ: In-phase/Quadrature (baseband signal representation)
- ION: Interplanetary Overlay Network (NASA JPL DTN implementation)
- LEO: Low Earth Orbit
- LDPC: Low-Density Parity-Check (error correction code)
- LTP: Licklider Transmission Protocol
- NVM: Non-Volatile Memory
- OBC: Onboard Computer
- SDR: Software Defined Radio
- TLE: Two-Line Element (orbital parameter format)
- TNC: Terminal Node Controller
- UHD: USRP Hardware Driver

---

## Appendix B: Future Work

- Optical communication experiments
- Interoperability with professional deep-space networks
- Expansion to Mars-relay simulations
- Relay functionality as a future enhancement (currently out of scope)
