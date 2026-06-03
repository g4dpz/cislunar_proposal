# RADIANT: A Delay-Tolerant Networking Payload for the Future GEO Amateur Satellite

**Proposal for the Future GEO Meeting, Ham Radio Friedrichshafen 2026**

**Author:** David Johnson, G4DPZ  
**Affiliation:** RADIANT Project (AMSAT-UK / AMSAT-DL / Goonhilly Earth Station)  
**Contact:** dave@g4dpz.me.uk  
**Website:** https://radiant.amsat-uk.org

---

## Abstract

We propose a Delay-Tolerant Networking (DTN) payload as a first-class experiment on the next-generation GEO amateur satellite. The payload would provide the amateur radio community with an operational store-and-forward messaging backbone using Bundle Protocol version 7 (BPv7) and Licklider Transmission Protocol (LTP), enabling disruption-tolerant communications across the satellite's coverage footprint. Unlike conventional bent-pipe transponders, a DTN payload stores bundles onboard and forwards them when the destination ground station is available, transforming the GEO satellite from a real-time relay into an intelligent message routing node. This capability is directly relevant to future cislunar and deep-space amateur missions where continuous connectivity is impossible.

---

## 1. Introduction

The amateur radio community has operated successfully through the QO-100 (Es'hail-2) satellite since 2019, demonstrating that GEO amateur transponders serve a global user base with always-on connectivity. The Future GEO project presents an opportunity to advance beyond analogue and digital voice/data transponders toward networking protocols designed for the space environment.

Delay-Tolerant Networking is the networking architecture adopted by NASA, ESA, and JAXA for deep-space communications. DTN's store-and-forward model — where data bundles are held at intermediate nodes until the next link becomes available — is the foundation of the Solar System Internet architecture currently under development by space agencies worldwide. By deploying a DTN payload on a GEO amateur satellite, the amateur community gains operational experience with the same protocols that will underpin future lunar and interplanetary communications.

RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed) has been developing this capability since 2024, with a phased roadmap progressing from terrestrial validation through GEO to cislunar space. The project is supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station. The architecture is DTN-implementation-agnostic, supporting multiple BPv7 engines (including HDTN, ION-DTN, µD3TN, and Hardy) through a common abstraction layer.

---

## 2. Scientific and Technical Rationale

### 2.1 Why DTN on a GEO Satellite?

A GEO satellite with a DTN payload provides unique capabilities that a conventional transponder cannot:

**Store-and-forward messaging:** Ground stations transmit bundles to the satellite, which stores them onboard and delivers them when the destination station is active. This decouples sender and receiver in time — operators need not be online simultaneously.

**Protocol validation at space distances:** The GEO orbit introduces approximately 250 ms one-way light time (500 ms round-trip), providing an authentic space delay environment for validating LTP's deferred acknowledgment mechanisms. This is the critical stepping stone between terrestrial links (millisecond RTT) and cislunar distances (1.3-second one-way).

**Always-on DTN backbone:** Unlike LEO satellites with 5–10 minute pass windows, a GEO DTN node provides continuous availability across its coverage footprint. Ground stations can deposit and retrieve bundles at any time, creating a persistent message store accessible from Europe, Africa, the Middle East, and parts of Asia and South America.

**Cislunar mission rehearsal:** Operating a DTN node in GEO exercises the same protocols, contact management, and operational procedures required for a future cislunar amateur mission. The GEO payload serves as a flight-proven reference for lunar DTN operations.

**Educational platform:** DTN is increasingly taught in university space communications courses. An operational GEO DTN node provides a real-world laboratory for students and researchers studying delay-tolerant protocols.

### 2.2 Relevance to ESA Programmes

ESA's Moonlight initiative is deploying a lunar communications and navigation infrastructure. The DTN architecture (BPv7/LTP) is central to ESA's approach to cislunar networking. An amateur GEO DTN payload demonstrates community readiness to participate in this broader ecosystem and validates interoperability concepts at reduced cost and risk.

The CCSDS (Consultative Committee for Space Data Systems) has standardised BPv7 and LTP for space networking. An amateur implementation operating on a GEO satellite provides independent validation of these standards in an operational environment.

---

## 3. Payload Description

### 3.1 Functional Overview

The DTN payload operates as an autonomous store-and-forward node in geostationary orbit. It receives BPv7 bundles from ground stations via the satellite uplink, stores them in onboard non-volatile memory, and forwards them to destination ground stations via the downlink when those stations are reachable.

Core operations:
- **Store-and-forward:** Accept bundles from any ground station, store onboard, deliver to addressed destination
- **DTN ping:** Echo request/response for reachability testing (expected RTT: 500–600 ms single hop, ~1000 ms ground-to-ground via satellite)
- **Priority queuing:** Expedited, normal, and bulk service classes with configurable bandwidth allocation
- **Bundle lifetime management:** Automatic expiry and deletion of aged bundles to manage storage
- **Telemetry:** Periodic beacon bundles reporting payload health, storage utilisation, and link metrics

### 3.2 Protocol Stack

| Layer | Protocol | Standard |
|-------|----------|----------|
| Application | Store-and-Forward, Ping | — |
| Network | BPv7 (Bundle Protocol version 7) | RFC 9171 / CCSDS |
| Transport | LTP (Licklider Transmission Protocol) | RFC 5326 |
| Modem | Digital (BPSK/QPSK + FEC) | — |
| Physical | RF Interface (satellite transponder) | — |

On the ground segment, LTP segments are carried over KISS framing to interface with TNC hardware.

### 3.3 Transponder Requirements

The DTN payload requires a dedicated digital channel allocation within the satellite transponder:

- **Bandwidth:** 5–20 kHz (depending on data rate and modulation)
- **Mode:** Digital (BPSK or QPSK with FEC)
- **Access:** Dedicated time slot or frequency allocation for DTN traffic
- **Uplink/downlink:** Separate frequencies for full-duplex DTN operation
- **Power allocation:** Sufficient transponder power for the DTN channel to close the link budget with Tier 1/2 amateur ground stations

The payload could operate within a narrowband digital sub-band of the transponder, coexisting with analogue and other digital modes.

### 3.4 Ground Segment

Ground stations access the DTN payload using open-source software running on standard amateur radio equipment:

| Tier | Equipment | Capability |
|------|-----------|------------|
| Tier 1 | Small dish (60–90 cm) + SDR | Send/receive bundles, DTN ping |
| Tier 2 | Larger dish + dedicated modem | Higher throughput, automated operation |
| Tier 3 | 3–5 m dish | Gateway to terrestrial DTN network |

The ground station software is already under development as part of the RADIANT project and will be released as open-source (MIT licence) before the satellite launch.

---

## 4. Station Identification and Regulatory Compliance

All transmissions through the DTN payload comply with amateur radio regulations:

- **Callsign in every bundle:** The DTN Endpoint Identifier (EID) scheme embeds the operator's callsign in every transmitted bundle: `dtn://G4DPZ/service`
- **Periodic beacon:** The payload transmits a plaintext identification beacon every 10 minutes containing callsign, satellite name, and operational status
- **No encryption:** All bundle payloads are transmitted in the clear. No cryptographic operations are applied to user data. CRC integrity checks are used for error detection.
- **Published protocol:** The DTN-over-amateur-radio protocol specification will be published as an open document, satisfying the requirement that protocols used on amateur frequencies are publicly available for inspection
- **Satellite command exception:** Telecommand (TT&C) for the DTN payload processor may use encryption for spacecraft safety, as permitted under ITU Radio Regulations for amateur satellite command and control

---

## 5. RADIANT Project Status and Readiness

### 5.1 Current Achievements

The RADIANT project has completed significant development work that directly applies to the GEO payload:

- **Working 3-node DTN simulation** demonstrating store-and-forward relay with true packet-level propagation delays (1.3 s Earth–Moon, 3–12 min Mars scenarios)
- **DTN-implementation-agnostic architecture** — abstraction layer supporting multiple BPv7 engines (HDTN, ION-DTN, µD3TN, Hardy) with custom KISS Convergence Layer Adapter
- **Go orchestrator** managing DTN engine lifecycle, telemetry collection, and contact plan management
- **Property-based test suite** — 11 correctness properties validated, CI pipeline operational
- **QO-100 Phase 1.5 design complete** — full requirements and technical design for GEO DTN operation via Es'hail-2, including link budget analysis, LTP timeout adaptation, and always-on contact model
- **Station identification system** — callsign-embedded EIDs with periodic beacon, fully compliant with amateur regulations
- **Protocol stack validated** — BPv7 / LTP / KISS framing operational end-to-end

### 5.2 Phased Roadmap to GEO

| Phase | Status | Description | Relevance to GEO Payload |
|-------|--------|-------------|--------------------------|
| 1: Terrestrial | In progress | VHF/UHF 9600 baud ground validation | Software stack validation |
| 1.5: QO-100 | Designed | DTN via Es'hail-2 (current GEO) | Direct GEO DTN experience |
| 2: CubeSat EM | Planned | Flight hardware ground testing | Flight processor validation |
| 3: LEO CubeSat | Planned | Orbital DTN demonstration | Flight heritage |
| **GEO Payload** | **Proposed** | **Future GEO satellite** | **This proposal** |
| 4: Cislunar | Planned | Earth–Moon DTN | Ultimate mission goal |

### 5.3 Flight Software Candidates

The DTN-implementation-agnostic architecture allows selection of the most appropriate DTN engine for the GEO payload:

- **µD3TN** — Lightweight, space-tested DTN implementation designed for microcontrollers and POSIX systems. Strong candidate for resource-constrained flight processors.
- **Hardy** — Modular Rust BPv7 implementation with `no_std` core libraries. Memory-safe, suitable for safety-critical flight software.
- **HDTN** — NASA Glenn's high-rate implementation. Proven in simulation, suitable if a more capable processor is available.
- **ION-DTN** — JPL's Interplanetary Overlay Network. Mature, widely deployed in space missions.

---

## 6. Operational Concept

### 6.1 Nominal Operations

1. Ground station A transmits a BPv7 bundle addressed to ground station B via the GEO satellite uplink
2. The satellite DTN payload receives the bundle, validates it, and stores it in onboard NVM
3. When ground station B is active and requests delivery (or on a scheduled basis), the payload transmits the stored bundle via the downlink
4. Ground station B acknowledges receipt via LTP; the payload deletes the delivered bundle

### 6.2 Multi-Hop Networking

With a DTN payload onboard, the GEO satellite becomes a node in a wider DTN network:

```
Ground Station A --> GEO DTN Payload --> Ground Station B
                          |
                          +--> LEO CubeSat (Phase 3, future)
                          |
                          +--> Cislunar Node (Phase 4, future)
```

The GEO payload can serve as a relay between ground stations and future LEO or cislunar DTN nodes, providing a persistent backbone for the amateur DTN network.

### 6.3 Capacity Planning

Assuming a 5 kbps dedicated digital channel:

| Metric | Value |
|--------|-------|
| Raw throughput | 5 kbps = 540 KB/hour |
| Usable throughput (after protocol overhead) | ~400 KB/hour |
| Bundle storage (256 MB NVM) | ~650 average bundles (400 KB each) |
| Bundle storage (1 GB NVM) | ~2,600 average bundles |
| Typical message size | 1–10 KB (text, small files) |
| Messages stored (256 MB, 5 KB avg) | ~50,000 messages |
| Bundle lifetime | Configurable (default: 24 hours) |

---

## 7. Benefits to the Amateur Community

1. **First operational amateur DTN backbone** — transforms amateur satellite communications from real-time-only to store-and-forward capable
2. **Global message store** — operators deposit messages for later retrieval, enabling asynchronous communication across time zones
3. **Cislunar pathfinder** — validates the exact protocols needed for future lunar amateur missions
4. **Educational resource** — universities and students gain access to a real DTN node for research and coursework
5. **Interoperability demonstration** — proves amateur DTN can coexist with and complement existing transponder modes
6. **Open-source ecosystem** — all software freely available, lowering the barrier to participation
7. **Standards validation** — independent amateur implementation of CCSDS BPv7/LTP standards

---

## 8. Collaboration and Contributions

### 8.1 What RADIANT Brings

- Complete DTN software stack (open-source, MIT licence)
- Ground station software for all tiers
- Protocol specification and documentation
- Operational experience from Phases 1–3
- Property-based test framework for flight software verification
- Integration with AMSAT-UK and AMSAT-DL ground station networks

### 8.2 What We Need from the Future GEO Project

- Dedicated digital channel allocation (5–20 kHz bandwidth)
- Payload processor slot or hosted computing resource
- Power allocation (2–5 W active, <100 µW idle)
- Mass allocation (< 1 kg)
- Telemetry/telecommand interface for payload management
- Early engagement in satellite system design to ensure interface compatibility

### 8.3 Development Timeline

| Milestone | Target | Dependency |
|-----------|--------|------------|
| Phase 1 terrestrial validation complete | 2026 | None |
| QO-100 DTN demonstration (Phase 1.5) | 2026–2027 | QO-100 access |
| Flight software selection and qualification | 2027–2028 | Processor selection |
| GEO payload engineering model | 2028–2029 | Interface specification |
| GEO payload flight model delivery | TBD | Satellite schedule |

---

## 9. Conclusion

A DTN payload on the next-generation GEO amateur satellite would be a landmark achievement for the amateur radio community — the first operational store-and-forward networking node in geostationary orbit using internationally standardised space networking protocols. It advances amateur radio from real-time transponder access toward intelligent, delay-tolerant networking that mirrors the architecture being deployed by space agencies for lunar and deep-space communications.

The RADIANT project has the software, the design, and the community support to deliver this payload. We are ready to collaborate with the Future GEO project to make amateur DTN from geostationary orbit a reality.

---

## References

1. RFC 9171 — Bundle Protocol Version 7 (BPv7), IETF, 2022
2. RFC 5326 — Licklider Transmission Protocol (LTP), IETF, 2008
3. CCSDS 734.2-B-1 — CCSDS Bundle Protocol Specification, 2020
4. RADIANT Project — https://radiant.amsat-uk.org
5. ESA Moonlight — https://www.esa.int/Applications/Connectivity_and_Secure_Communications/Moonlight
6. QO-100 (Es'hail-2) — https://amsat-dl.org/eshail-2-amsat-phase-4a
7. ITU Radio Regulations, Article 25 — Amateur Radio Service

---

## Appendix A: Link Budget Summary (GEO DTN Channel)

| Parameter | Uplink | Downlink |
|-----------|--------|----------|
| Frequency | 2.4 GHz | 10.45 GHz |
| Ground antenna gain | 15–22 dBi (60–90 cm dish) | 20–25 dBi (60–90 cm dish) |
| Ground TX power | 5–10 W | — |
| Path loss (36,000 km) | ~188 dB | ~206 dB |
| Required Eb/N₀ | ~10 dB (BPSK, rate ½ FEC) | ~10 dB |
| Link margin | > 3 dB | > 3 dB |
| Data rate | 1–10 kbps | 1–10 kbps |

*Detailed link budget analysis available in the RADIANT QO-100 Phase 1.5 design document.*

## Appendix B: Acronyms

- **BPv7** — Bundle Protocol version 7
- **CCSDS** — Consultative Committee for Space Data Systems
- **CGR** — Contact Graph Routing
- **DTN** — Delay-Tolerant Networking
- **EID** — Endpoint Identifier
- **ESA** — European Space Agency
- **FEC** — Forward Error Correction
- **GEO** — Geostationary Earth Orbit
- **HDTN** — High-rate Delay Tolerant Networking
- **KISS** — TNC serial framing protocol
- **LTP** — Licklider Transmission Protocol
- **NVM** — Non-Volatile Memory
- **QO-100** — Qatar-OSCAR 100 (Es'hail-2 amateur transponder)
- **RADIANT** — Radio Amateur Delay-tolerant Interplanetary Networking Testbed
- **RTT** — Round-Trip Time
- **TT&C** — Telemetry, Tracking, and Command
