# RADIANT: Implementing Delay-Tolerant Networking Over Amateur Radio Links for Cislunar Communication

**Authors:** David Johnson (G4DPZ)  
**Affiliation:** AMSAT-UK  
**Date:** June 2026  
**Version:** 1.0 — Draft for IPNSIG TD WG Zotero Library  
**Contact:** dave@g4dpz.me.uk  
**Website:** https://radiant.amsat-uk.org

---

## Abstract

RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed) is an open-source project implementing Bundle Protocol version 7 (RFC 9171) and Licklider Transmission Protocol (RFC 5326) over amateur radio links. The project demonstrates a phased pathway from terrestrial ground station validation through Low Earth Orbit CubeSat operations to cislunar networking. RADIANT uses a DTN-engine-agnostic architecture supporting multiple implementations (JPL's ION-DTN, µD3TN, and Hardy) through a common abstraction layer, with a custom KISS convergence layer enabling LTP transport directly over amateur radio TNC hardware. Station identification is achieved through callsign-embedded DTN Endpoint Identifiers, providing regulatory compliance without protocol overhead. This paper describes the system architecture, the LTP-over-KISS protocol design, the DTN-engine-agnostic abstraction layer, current implementation status, and the five-phase roadmap toward an amateur-operated cislunar DTN node. A functioning 3-node simulation demonstrates store-and-forward relay with Contact Graph Routing across simulated cislunar delays. The project is supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station.

**Keywords:** DTN, BPv7, LTP, amateur radio, cislunar, ION-DTN, µD3TN, Hardy, KISS, store-and-forward, Contact Graph Routing, CubeSat

---

## 1. Introduction

The Solar System Internet (SSI) vision articulated by IPNSIG describes an evolution from agency-sustained point-to-point communication toward a network of networks with multiple stakeholders [1]. Delay-Tolerant Networking (DTN) and the Bundle Protocol suite have been identified as the foundation for all communications traversing paths that IP cannot support, including interplanetary distances and paths experiencing disruption due to scheduling or orbital constraints [1].

Amateur radio has historically served as a proving ground for networking protocols — from packet radio and AX.25 in the 1980s to the digital modes ecosystem today. RADIANT extends this tradition by implementing the DTN protocol stack recommended for the SSI over amateur radio infrastructure, creating an accessible experimental platform for store-and-forward networking across disrupted links.

The project addresses a gap in the current DTN ecosystem: while BP/LTP implementations exist for space agency missions (notably the Korea Pathfinder Lunar Orbiter, which flies BP as a Development Test Objective [1]) and for high-rate research networks, no implementation targets the constrained, regulatory-compliant environment of amateur radio with its bandwidth limitations, identification requirements, and prohibition on encrypted payloads.

### 1.1 Objectives

1. Demonstrate BPv7/LTP operations over amateur radio links from terrestrial to cislunar distances
2. Validate Contact Graph Routing and store-and-forward relay across increasing latency regimes
3. Develop an open-source, DTN-engine-agnostic architecture suitable for constrained amateur platforms
4. Provide operational evidence supporting a flight experiment case for an amateur DTN CubeSat and cislunar payload
5. Contribute to the SSI multistakeholder ecosystem by demonstrating that amateur operators can participate in interplanetary networking using standard protocols

### 1.2 Relationship to SSI Architecture

RADIANT implements the architecture recommended by the IPNSIG SSI report [1]:
- BPv7 (RFC 9171) as the networking layer
- LTP (RFC 5326) for reliable delivery over delayed/disrupted links
- Contact Graph Routing for time-scheduled forwarding
- Integrity protection without payload encryption (compatible with BPSec RFC 9172)

The project demonstrates that amateur radio's existing governance mechanisms — globally unique callsigns allocated by national authorities under ITU coordination — provide a natural solution to the BP identifier allocation problem discussed in the SSI report.

---

## 2. System Architecture

### 2.1 Protocol Stack

RADIANT employs a simplified protocol stack that eliminates the traditional AX.25 layer used in amateur packet radio, wrapping LTP segments directly in KISS (Keep It Simple, Stupid) framing:

```
┌─────────────────────────────────────┐
│   Application (bping, bpsendfile)   │
├─────────────────────────────────────┤
│   BPv7 (Bundle Protocol v7)        │
│   EID: dtn://callsign/service      │
├─────────────────────────────────────┤
│   LTP (Licklider Transmission)     │
├─────────────────────────────────────┤
│   KISS (TNC Serial Framing)        │
├─────────────────────────────────────┤
│   G3RUH GFSK (9600 baud)           │
└─────────────────────────────────────┘
```

This design provides several advantages over a traditional AX.25-based approach:

- **Reduced overhead**: KISS framing adds 3 bytes per frame versus 18+ bytes for AX.25 headers and FCS, yielding approximately 10% throughput improvement at 9600 baud
- **Eliminated semantic mismatch**: AX.25 addressing is redundant when DTN provides its own addressing and routing
- **Simplified implementation**: no frame construction, CRC-16 calculation, or address encoding required at the link layer
- **Native DTN semantics**: the protocol stack operates end-to-end without protocol translation

### 2.2 DTN-Engine-Agnostic Abstraction Layer

The architecture is designed to support multiple DTN engines through a common interface:

- **ION-DTN** (JPL) — the reference implementation with extensive flight heritage on deep-space missions
- **µD3TN** (D3TN GmbH) — lightweight implementation suitable for microcontroller platforms (candidate flight software)
- **Hardy** (Rust) — modular BPv7 implementation with `no_std` core (candidate flight software)

This approach allows operators to select the engine best suited to their platform constraints and mission phase — ION-DTN for ground stations requiring proven reliability, µD3TN or Hardy for constrained flight hardware on the STM32U585.

### 2.3 Engine Integration Architecture

The implementation uses a common abstraction layer that interfaces with any conformant BPv7/LTP engine:

- **KISS CLA Module**: A convergence layer adapter that bridges any DTN engine's LTP output with amateur radio TNC hardware via serial KISS framing
- **Orchestrator** (`dtn-node`): Manages DTN engine lifecycle, telemetry collection, and contact plan updates
- **Configuration Generation**: Programmatic generation of engine-specific configurations from a common schema

### 2.4 Callsign-Embedded Endpoint Identifiers

Amateur radio regulations (ITU Radio Regulations Article 25, FCC Part 97, Ofcom licence conditions) require station identification in every transmission. RADIANT satisfies this by embedding operator callsigns in DTN Endpoint Identifiers:

```
dtn://<callsign>-<ssid>/<service>

Examples:
  dtn://g4dpz-1          Primary station
  dtn://g4dpz-1/mail     Mail service endpoint
  dtn://g4dpz-1/beacon   Beacon service
```

Every bundle's primary block carries source and destination EIDs containing callsigns. Additionally, periodic beacon bundles (every 10 minutes) transmit plaintext identification payloads containing callsign, Maidenhead grid locator, and node type.

DTN engines typically use numeric `ipn://` addresses internally for Contact Graph Routing (which requires integer node IDs). The dual-EID scheme provides compact routing via `ipn://` while maintaining callsign identification via `dtn://` in bundle metadata.

This approach leverages the existing global callsign allocation infrastructure — no new registry is required for amateur DTN identifier management.

---

## 3. LTP-over-KISS: Protocol Design

### 3.1 Rationale

Traditional amateur packet radio uses AX.25 as the link-layer protocol, providing callsign-based addressing and error detection. However, when carrying DTN bundles:

1. AX.25 addressing is redundant — DTN EIDs already identify source and destination
2. AX.25's error detection (CRC-16) is redundant — LTP provides its own integrity mechanisms
3. The 18+ byte per-frame overhead is significant at 9600 baud (1200 bytes/sec)
4. DTN bundles are opaque to AX.25 — there is no interoperability with traditional packet radio regardless

Eliminating AX.25 and wrapping LTP directly in KISS frames preserves the TNC hardware interface while removing unnecessary protocol complexity.

### 3.2 Frame Structure

```
[FEND] [CMD] [LTP segment bytes...] [FEND]

FEND = 0xC0 (frame boundary)
CMD  = 0x00 (data frame)
```

Byte stuffing follows the standard KISS specification:
- `0xC0` in data → `0xDB 0xDC`
- `0xDB` in data → `0xDB 0xDD`

### 3.3 LTP Configuration

LTP is configured with parameters appropriate to each mission phase:

| Phase | One-Way Light Time | Data Rate | LTP MTU | Retransmission Timer |
|-------|-------------------|-----------|---------|---------------------|
| Terrestrial | <1 ms | 9600 bps | 512 B | 2 s |
| QO-100 (GEO) | 250 ms | 2400 bps | 256 B | 5 s |
| LEO | 2–8 ms | 9600 bps | 512 B | 10 s |
| Cislunar | 1300 ms | 500 bps | 128 B | 30 s |

LTP's deferred acknowledgment mechanism is essential for links with long round-trip times — it allows continuous transmission without waiting for per-segment acknowledgment, batching reception reports at checkpoint boundaries.

### 3.4 Regulatory Compliance

| Requirement | Mechanism | Status |
|-------------|-----------|--------|
| Station identification | Callsign in DTN EID (every bundle) + periodic beacon | Compliant |
| No encryption | All payloads transmitted in the clear | Compliant |
| Published protocol | BPv7 (RFC 9171), LTP (RFC 5326), KISS spec, open-source implementation | Compliant |
| Unobscured meaning | All data formats publicly documented; encoding for FEC permitted | Compliant |

The approach follows established precedent: APRS, FT8/WSPR, D-STAR, and Winlink all use published protocols with callsign identification in message structures rather than relying solely on the link layer.

---

## 4. Contact Graph Routing and Store-and-Forward

### 4.1 Contact Plans

RADIANT uses time-dependent contact plans specifying communication windows between node pairs. Each contact entry defines:

```json
{
  "source": 10,
  "dest": 20,
  "startTime": 0,
  "endTime": 600,
  "rateBitsPerSec": 9600,
  "oneWayLightTimeMs": 1300
}
```

For orbital phases, contact windows are computed from TLE/ephemeris data using SGP4/SDP4 propagation. For terrestrial and GEO phases, contacts may be always-on or scheduled around antenna availability.

### 4.2 Store-and-Forward Relay

The 3-node cislunar simulation demonstrates the core DTN value proposition:

```
Ground Station (Earth)  →  Lunar Orbiter (relay)  →  Lunar Lander
     nodeId=10                  nodeId=20                nodeId=30
```

When the ground station transmits a bundle destined for the lander, CGR determines that the orbiter is the next hop. The orbiter stores the bundle and forwards it to the lander when the next contact window opens. If any link is disrupted, bundles persist in storage until the link becomes available — no data is lost due to temporary disconnection.

### 4.3 Simulation Results

The simulation validates DTN protocol behaviour under representative timing conditions using a UDP delay proxy to inject propagation delays:

- **Cislunar timing** (1.3 s OWLT): LTP sessions complete correctly with 2.6 s round-trip; CGR routes through relay node; bundles delivered end-to-end
- **Mars closest approach** (3 min OWLT): LTP retransmission timers handle 6-minute RTT gracefully
- **Mars average** (12 min OWLT): 24-minute round-trip times managed without session timeout; store-and-forward relay functions correctly

These are software simulations with injected delay, not RF propagation paths. They validate protocol behaviour, not link budget performance.

---

## 5. Five-Phase Roadmap

### Phase 1: Terrestrial Validation (In Progress)

- **Hardware**: Raspberry Pi + Mobilinkd TNC4 + Yaesu FT-817
- **Link**: VHF/UHF 9600 baud G3RUH GFSK
- **Validates**: Complete software stack (LTP-over-KISS, callsign EIDs), store-and-forward, DTN ping over real amateur radio links
- **Status**: Two-node testing operational

### Phase 1.5: QO-100 GEO Satellite (Planned)

- **Link**: Es'hail-2 narrowband transponder (2.4 GHz up / 10.45 GHz down)
- **Validates**: LTP behaviour with authentic 500 ms round-trip space delay; first space-based amateur DTN demonstration
- **Significance**: Proves architecture over a real space link before committing to orbital hardware

### Phase 2: CubeSat Engineering Model (Planned)

- **Hardware**: STM32U585 (ARM Cortex-M33, 160 MHz, 786 KB SRAM) + Ettus B200mini SDR
- **Validates**: Flight software on flight-representative hardware; power budget; simulated orbital passes (5–10 min windows); fault injection
- **Significance**: Bridges terrestrial testing and flight commitment

### Phase 3: LEO CubeSat (Planned)

- **Frequency**: UHF 437 MHz, 9.6 kbps
- **Link budget**: 2W TX, omni antenna → 12 dBi ground Yagi = ~31 dB margin
- **Operations**: Ground-to-space DTN ping, store-and-forward messaging, CGR-predicted contact windows
- **Significance**: First amateur DTN payload in orbit

### Phase 4: Cislunar (Planned)

- **Frequency**: S-band 2.2 GHz, 500 bps BPSK + LDPC/Turbo FEC
- **Link budget**: 5W TX, 10 dBi patch → 35 dBi ground dish (3–5 m) = ~7 dB margin
- **Delay regime**: 1.3 s one-way light time (Earth–Moon)
- **Ground segment**: Tier 3/4 stations with 3–5 m dishes
- **Significance**: First amateur cislunar DTN node; seeking ESA ARTES support

---

## 6. Alignment with SSI Principles

RADIANT directly implements several principles identified in the IPNSIG Solar System Internet Architecture and Governance report [1]:

| SSI Principle | RADIANT Implementation |
|---|---|
| Standards-based protocols | BPv7 (RFC 9171), LTP (RFC 5326), CGR |
| Multistakeholder governance | Amateur operators, AMSAT orgs, open-source community |
| Transparency | Open source (MIT), public protocols, regulatory requirement for plaintext |
| Interoperability | DTN-engine-agnostic abstraction; standard BP/LTP/CGR |
| Fair identifier allocation | Callsign-EIDs leverage existing ITU-coordinated amateur callsign system |
| Security without confidentiality | BPSec integrity (HMAC) without payload encryption |
| Autonomy and automation | Automated contact plan computation, CGR routing |

The project demonstrates that the amateur radio community can participate as a stakeholder in the SSI using standard protocols and existing governance mechanisms. RADIANT is to the Solar System Internet what early university and amateur networks were to the terrestrial Internet — a proving ground for protocols and operational concepts before commercial deployment.

---

## 7. Implementation Status

### Completed

- DTN-engine-agnostic abstraction layer (Rust)
- KISS framing library (`no_std` compatible for flight hardware)
- 3-node cislunar simulation with CGR and store-and-forward relay
- LTP-over-KISS protocol design and implementation
- Callsign-EID configuration
- Property-based tests validating correctness properties
- Continuous integration pipeline
- Project website (https://radiant.amsat-uk.org)

### In Progress

- Phase 1 terrestrial two-node testing (G4DPZ)
- ION-DTN integration via abstraction layer
- Hardy evaluation for flight software

### Planned

- QO-100 demonstration (Phase 1.5)
- Engineering model on STM32U585 (Phase 2)
- Flight CubeSat payload (Phase 3)
- Cislunar mission (Phase 4)

---

## 8. Related Work

- **KPLO DTN DTO**: Korea Pathfinder Lunar Orbiter flies BP as a Development Test Objective, demonstrating the protocol suite in cislunar space [1]
- **ION-DTN**: JPL's reference implementation used on multiple missions [2]
- **µD3TN**: Lightweight DTN for microcontrollers, space-tested [3]
- **Hardy**: Modular Rust BPv7 implementation with `no_std` core [4]
- **LunaNet**: NASA's interoperability specification for lunar communications [5]
- **IETF TVR WG**: Time-Variant Routing working group, directly relevant to contact graph approaches [6]
- **ORI (Open Research Institute)**: Open-source modem and signal chain work (DVB-S2, LDPC) potentially complementary for higher-rate phases [7]

---

## 9. Future Directions

### 9.1 DTN Implementation Interoperability

A key objective for RADIANT is to serve as an independent interoperability testbed for DTN implementations. The project's engine-agnostic architecture is designed to support multiple BPv7/LTP implementations behind a common interface, enabling cross-implementation testing over real RF links. The implementations under consideration are:

- **ION-DTN** (NASA JPL) — the reference implementation with extensive flight heritage
- **µD3TN** (D3TN GmbH) — lightweight, space-tested implementation suitable for constrained platforms
- **ESA-DTN** — ESA's implementation for European missions
- **Hardy** — modular Rust BPv7 implementation with `no_std` core libraries for microcontroller targets

The amateur radio environment provides unique interoperability testing conditions:

1. Real RF links with authentic propagation characteristics (not loopback or LAN)
2. Different ground stations can run different engines — some operators may prefer ION, others µD3TN
3. Flight payloads may use a different engine than ground infrastructure (Hardy or µD3TN on an STM32U585, ION-DTN on ground stations)
4. Constrained bandwidth (9600 baud to 500 bps) exercises edge cases in LTP segmentation and session management that high-rate links do not expose
5. BPv7 and LTP are wire-format standards (RFC 9171, RFC 5326) — conformant implementations should interoperate regardless of which engine generated the bundle

The interoperability test plan will validate: bundle exchange between engines over LTP, EID resolution consistency, CGR contact plan interpretation, LTP session behaviour under identical timing constraints, and correct handling of BPSec integrity blocks across heterogeneous nodes.

This work directly supports the SSI vision of a network of networks where different operators and agencies deploy different implementations that must interoperate seamlessly via standard protocols.

### 9.2 Other Planned Work

1. **Protocol definition publication**: Formal specification of the `dtn://callsign` EID convention as the station identification mechanism for amateur DTN, following precedent of published amateur digital mode specifications
2. **SANA registration**: Formalise RADIANT node numbers in the BP identifier space
3. **Multi-node contact graph**: Distributed routing across ground station networks with time-dependent Dijkstra path computation
4. **Security model**: Document integrity-without-confidentiality as a BPSec profile contribution to SSI security architecture
5. **Flight software selection**: Evaluate µD3TN and Hardy for STM32U585 platform constraints
6. **Ground station network**: Expand from single-operator testing to distributed amateur ground station participation

---

## 10. Conclusion

RADIANT demonstrates that the DTN protocol architecture recommended for the Solar System Internet can be implemented over amateur radio infrastructure, providing a low-cost, accessible platform for store-and-forward networking research across disrupted links. The LTP-over-KISS protocol design eliminates unnecessary protocol layers while maintaining full regulatory compliance through callsign-embedded Endpoint Identifiers. The five-phase roadmap progressively validates the architecture from terrestrial links through to cislunar distances, building operational evidence for flight experiments.

The project contributes to the SSI ecosystem by showing that amateur radio operators — with their globally coordinated identifier system, regulatory framework requiring transparency, and tradition of protocol experimentation — are natural participants in interplanetary networking. As the SSI evolves from agency-sustained systems toward a multistakeholder network of networks, RADIANT provides one pathway for the amateur community to contribute operationally.

All software is open-source (MIT licence). Contributions, ground station participation, and collaboration are welcomed.

---

## References

[1] IPNSIG, "Solar System Internet Architecture and Governance," Internet Society Interplanetary Chapter, September 2023. Available: https://www.ipnsig.org/

[2] S. Burleigh, "Interplanetary Overlay Network (ION) Design and Operation," NASA JPL, 2020.

[3] D-3TN GmbH, "µD3TN — Lightweight DTN Implementation." Available: https://gitlab.com/d3tn/ud3tn

[4] Hardy BPv7, "Modular Rust BPv7 Implementation." Available: https://github.com/hardybp/hardy

[5] NASA, "LunaNet Interoperability Specification," NASA/TP-20210021073/Rev.4, 2023.

[6] IETF, "Time-Variant Routing (TVR) Working Group." Available: https://datatracker.ietf.org/wg/tvr/about/

[7] Open Research Institute, "ORI Projects." Available: https://www.openresearch.institute/

[8] IETF, "Bundle Protocol Version 7," RFC 9171, January 2022.

[9] IETF, "Licklider Transmission Protocol," RFC 5326, September 2008.

[10] IETF, "Bundle Protocol Security (BPSec)," RFC 9172, January 2022.

[11] CCSDS, "Solar System Internetwork Architecture," CCSDS 730.1-G-1, 2014.

---

## Appendix A: Acronyms

| Acronym | Expansion |
|---------|-----------|
| BPv7 | Bundle Protocol Version 7 |
| CGR | Contact Graph Routing |
| CLA | Convergence Layer Adapter |
| DTN | Delay/Disruption Tolerant Networking |
| EID | Endpoint Identifier |
| FEC | Forward Error Correction |
| GFSK | Gaussian Frequency Shift Keying |
| KISS | Keep It Simple, Stupid (TNC framing protocol) |
| LDPC | Low-Density Parity-Check |
| LTP | Licklider Transmission Protocol |
| OWLT | One-Way Light Time |
| SSI | Solar System Internet |
| TNC | Terminal Node Controller |

---

## Appendix B: Project Links

- **Website**: https://radiant.amsat-uk.org
- **AMSAT-UK**: https://amsat-uk.org
- **ION-DTN**: https://sourceforge.net/projects/ion-dtn/
- **µD3TN**: https://gitlab.com/d3tn/ud3tn
- **IPNSIG**: https://www.ipnsig.org
- **Author contact**: dave@g4dpz.me.uk / G4DPZ
