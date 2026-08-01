# RADIANT: Implementing Delay-Tolerant Networking Over Amateur Radio Links for Cislunar Communication

**Authors:** David Johnson (G4DPZ)  
**Affiliation:** AMSAT-UK / Goonhilly Earth Station  
**Date:** June 2025  
**Version:** 2.0 — For IPNSIG TD WG Zotero Library  
**Contact:** dave@g4dpz.me.uk  
**Website:** https://radiant.amsat-uk.org

---

## Abstract

RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed) is an operational project implementing Bundle Protocol version 7 (RFC 9171), Licklider Transmission Protocol (RFC 5326), and Contact Graph Routing over amateur radio links. The project demonstrates a phased pathway from terrestrial ground station validation through Low Earth Orbit CubeSat operations to cislunar networking. RADIANT uses a DTN-engine-agnostic architecture supporting multiple implementations (JPL's ION-DTN, Hardy, and µD3TN) through a common abstraction layer, with a custom KISS convergence layer enabling LTP transport directly over amateur radio TNC hardware and a TCPCLv3 gateway for terrestrial interconnection. Station identification is achieved through callsign-embedded DTN Endpoint Identifiers, providing regulatory compliance without protocol overhead. Multi-implementation interoperability has been proven through successful 1 MB LTP bundle transfers between ION-DTN and Hardy. Three validated store-and-forward scenarios are operational, and a European ground segment spanning UK, German, and French nodes with QO-100 GEO backbone connectivity is deployed. This paper describes the system architecture, protocol design, operational results, the European ground segment topology, and the five-phase roadmap toward an amateur-operated cislunar DTN node. The project represents the first non-agency operational DTN stakeholder in the Solar System Internet ecosystem.

**Keywords:** DTN, BPv7, LTP, amateur radio, cislunar, ION-DTN, µD3TN, Hardy, KISS, TCPCLv3, store-and-forward, Contact Graph Routing, CubeSat, Solar System Internet

---

## 1. Introduction

The Solar System Internet (SSI) vision articulated by IPNSIG describes an evolution from agency-sustained point-to-point communication toward a network of networks with multiple stakeholders [1]. Delay-Tolerant Networking (DTN) and the Bundle Protocol suite have been identified as the foundation for all communications traversing paths that IP cannot support, including interplanetary distances and paths experiencing disruption due to scheduling or orbital constraints [1].

Amateur radio has historically served as a proving ground for networking protocols — from packet radio and AX.25 in the 1980s to the digital modes ecosystem today. RADIANT extends this tradition by implementing the DTN protocol stack recommended for the SSI over amateur radio infrastructure, creating an accessible experimental platform for store-and-forward networking across disrupted links.

The project addresses a gap in the current DTN ecosystem: while BP/LTP implementations exist for space agency missions (notably the Korea Pathfinder Lunar Orbiter, which flies BP as a Development Test Objective [1]) and for high-rate research networks, no implementation targets the constrained, regulatory-compliant environment of amateur radio with its bandwidth limitations, identification requirements, and prohibition on encrypted payloads.

RADIANT is now operational. Multi-implementation LTP interoperability has been proven, a TCPCLv3 gateway is deployed and accepting connections, and three validated store-and-forward scenarios demonstrate the architecture under representative conditions. The project represents the first non-agency operational DTN stakeholder — building operational SSI nodes before commercial ones exist.

### 1.1 Objectives

1. Demonstrate BPv7/LTP operations over amateur radio links from terrestrial to cislunar distances
2. Validate Contact Graph Routing and store-and-forward relay across increasing latency regimes
3. Develop an open-source, DTN-engine-agnostic architecture suitable for constrained amateur platforms
4. Prove multi-implementation interoperability over real and simulated RF links
5. Provide operational evidence supporting a flight experiment case for an amateur DTN CubeSat and cislunar payload
6. Contribute to the SSI multistakeholder ecosystem by demonstrating that amateur operators can participate in interplanetary networking using standard protocols

### 1.2 Relationship to SSI Architecture

RADIANT implements the architecture recommended by the IPNSIG SSI report [1]:
- BPv7 (RFC 9171) as the networking layer
- LTP (RFC 5326) for reliable delivery over delayed/disrupted links
- Contact Graph Routing for time-scheduled forwarding
- Integrity protection without payload encryption (compatible with BPSec RFC 9172)

The project demonstrates that amateur radio's existing governance mechanisms — globally unique callsigns allocated by national authorities under ITU coordination — provide a natural solution to the BP identifier allocation problem discussed in the SSI report. No new registry is required for amateur DTN identifier management.

### 1.3 The Amateur Radio Stakeholder Case

The SSI governance report identifies a three-phase evolution: agency-sustained point-to-point links today, a transitional period where commercial and amateur networks emerge alongside agency backbones, and a future of interconnected networks peering via standard protocols. RADIANT positions amateur radio squarely in that transitional phase.

The amateur radio community brings capabilities that agencies cannot replicate:

- **Geographic diversity at scale.** A LEO satellite at 500 km altitude is visible to any single ground station for approximately 10 minutes per pass. Continuous store-and-forward coverage requires globally distributed stations. Amateur operators span every continent.
- **Pre-existing globally unique identifier allocation.** Callsigns are globally unique, government-issued, internationally coordinated through the ITU, and require no new registry infrastructure. RADIANT embeds callsigns directly in `dtn://` EIDs.
- **Regulatory transparency as a feature.** ITU Article 25 compliance means all protocols are publicly documented, all transmissions are unencrypted, and all data formats are available for inspection. This is structural alignment with the SSI report's transparency principle.
- **Operational culture matched to DTN conditions.** Amateur operators routinely work with limited power, marginal links, intermittent propagation, and precise timing. Moonbounce operation — scheduling transmissions to sub-second precision, decoding signals at noise floor, tolerating 2.6-second round trips — is functionally identical to cislunar DTN operations.

The governance model requires demonstrated operational competence, not merely claimed intent. RADIANT is how the amateur community builds that credibility through verifiable results.

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

The architecture decouples ground station management from the underlying DTN engine. A Rust-based orchestrator manages engine lifecycle, telemetry collection, contact plan distribution, and the HTTP/JSON API layer. The supported engines are:

- **ION-DTN** (JPL) — Primary engine. Selected for NASA compatibility and extensive flight heritage on deep-space missions. Direct lifecycle management via ION's administrative interface. Provides BPv7, LTP, and CGR.
- **Hardy** (Rust) — Independent modular BPv7 implementation with `no_std` core targeting embedded platforms. Proven LTP interoperability with ION at 1 MB bundle transfers. Candidate flight software.
- **µD3TN** (D3TN GmbH) — Lightweight implementation from TU Dresden / D3TN, suitable for microcontroller platforms and resource-constrained environments. Relevant for CubeSat flight software and European academic collaboration.

Operators select the engine appropriate to their station. The abstraction layer presents a consistent management interface regardless of backend — bundle injection, status queries, contact plan updates, and beacon scheduling all use the same API.

A direct ION controller mode bypasses the abstraction layer when maximum performance is needed, providing raw access to ION's BPA for high-throughput scenarios.

### 2.3 Convergence Layers

**RF Links (KISS CL):** LTP segments are wrapped directly in KISS frames — no AX.25 layer. The KISS CLA module bridges any DTN engine's LTP output with amateur radio TNC hardware via serial KISS framing. At 9600 baud, eliminating AX.25 yields approximately 10% throughput improvement.

**Terrestrial Interconnection (TCPCLv3):** RFC 7242 provides standard TCP connectivity for ground-segment gateway operations. Any DTN node on the internet can connect to a RADIANT gateway and inject bundles into the amateur RF network. This is the same transport protocol used in NASA's ground segment architecture — interoperability is immediate. The TCPCLv3 gateway is deployed and operational.

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

DTN engines typically use numeric `ipn://` addresses internally for Contact Graph Routing (which requires integer node IDs). The dual-EID scheme provides:
- `ipn://` for compact CBOR-encoded addresses on bandwidth-constrained space links (4 bytes wire size)
- `dtn://callsign-ssid/service` for human-readable, callsign-bearing addresses on terrestrial segments (~16–26 bytes)

Both schemes coexist. A node responds to both `ipn:10.1` and `dtn://g4dpz-1`. CGR operates on numeric node IDs from contact plans; the `dtn://` scheme provides the regulatory identification overlay.

---

## 3. LTP-over-KISS: Protocol Design

### 3.1 Rationale

Traditional amateur packet radio uses AX.25 as the link-layer protocol, providing callsign-based addressing and error detection. However, when carrying DTN bundles:

1. AX.25 addressing is redundant — DTN EIDs already identify source and destination
2. AX.25's error detection (CRC-16) is redundant — LTP provides its own integrity mechanisms
3. The 18+ byte per-frame overhead is significant at 9600 baud (1200 bytes/sec)
4. DTN bundles are opaque to AX.25 — there is no interoperability with traditional packet radio regardless

Eliminating AX.25 and wrapping LTP directly in KISS frames preserves the TNC hardware interface while removing unnecessary protocol complexity. This follows the same architectural logic that led CCSDS to define space-specific convergence layers rather than mandating terrestrial link protocols.

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

For orbital phases, contact windows are computed from TLE/ephemeris data using SGP4/SDP4 propagation. For terrestrial and GEO phases, contacts may be always-on or scheduled around antenna availability. Contact plan computation from TLE/SGP4 orbital data is implemented and operational.

### 4.2 Store-and-Forward Relay

The 3-node cislunar simulation demonstrates the core DTN value proposition:

```
Ground Station (Earth)  →  Lunar Orbiter (relay)  →  Lunar Lander
     nodeId=10                  nodeId=20                nodeId=30
```

When the ground station transmits a bundle destined for the lander, CGR determines that the orbiter is the next hop. The orbiter stores the bundle and forwards it to the lander when the next contact window opens. If any link is disrupted, bundles persist in storage until the link becomes available — no data is lost due to temporary disconnection.

### 4.3 Validated Scenarios

Three operational scenarios have been validated on production infrastructure:

1. **Delayed uplink** — Bundle held at ground station during link outage, delivered when contact resumes
2. **Immediate forwarding** — Real-time relay when end-to-end path exists
3. **Delayed downlink** — Remote node stores bundles for future delivery when ground station becomes available

### 4.4 Simulation Results

The simulation validates DTN protocol behaviour under representative timing conditions using a UDP delay proxy to inject propagation delays:

- **Cislunar timing** (1.3 s OWLT): LTP sessions complete correctly with 2.6 s round-trip; CGR routes through relay node; bundles delivered end-to-end
- **Mars closest approach** (3 min OWLT): LTP retransmission timers handle 6-minute RTT gracefully
- **Mars average** (12 min OWLT): 24-minute round-trip times managed without session timeout; store-and-forward relay functions correctly

These are software simulations with injected delay, not RF propagation paths. They validate protocol behaviour, not link budget performance. The simulation environment is open-source with automated CI validating every change.

---

## 5. Five-Phase Roadmap

### Phase 1: Terrestrial Validation (Operational)

- **Hardware**: Raspberry Pi + Mobilinkd TNC4 + Yaesu FT-817
- **Link**: VHF/UHF 9600 baud G3RUH GFSK
- **Validates**: Complete software stack (LTP-over-KISS, callsign EIDs), store-and-forward, DTN ping over real amateur radio links
- **Status**: Two-node testing operational. Three validated scenarios (delayed uplink, immediate forwarding, delayed downlink) confirmed on production infrastructure.

### Phase 1.5: QO-100 GEO Satellite (Planned)

- **Link**: Es'hail-2 narrowband transponder (2.4 GHz up / 10.45 GHz down)
- **Delay**: 250 ms one-way (500 ms RTT) — sufficient to exercise LTP deferred acknowledgement authentically
- **Equipment**: Standard QO-100 ground station (60–90 cm dish)
- **Validates**: LTP behaviour with authentic space delay; first space-based amateur DTN demonstration
- **Significance**: Always-available link enables rapid iteration without orbital tracking constraints. The critical stepping stone between terrestrial millisecond delays and cislunar 1.3-second delays.

### Phase 2: CubeSat Engineering Model (Planned)

- **Hardware**: STM32U585 (ARM Cortex-M33, 160 MHz, 786 KB SRAM, 2 MB flash) + Ettus B200mini SDR
- **Validates**: Flight software on flight-representative hardware; power budget; simulated orbital passes (5–10 min windows); fault injection
- **Significance**: Bridges terrestrial testing and flight commitment with identical hardware/software validation

### Phase 3: LEO CubeSat (Planned)

| Parameter | Value |
|-----------|-------|
| Frequency | 437 MHz UHF |
| Data rate | 9.6 kbps GMSK/BPSK |
| TX power | 2 W (33 dBm) |
| Spacecraft antenna | 0 dBi omnidirectional |
| Ground antenna | 12 dBi Yagi |
| Link margin | ~31 dB |

Any station currently tracking amateur LEO satellites can participate. Ground-to-space DTN ping and store-and-forward messaging from orbit. First amateur DTN payload in orbit.

### Phase 4: Cislunar (Planned)

| Parameter | Value |
|-----------|-------|
| Frequency | 2.2 GHz S-band |
| Data rate | 500 bps BPSK |
| TX power | 5 W (37 dBm) |
| Spacecraft antenna | 10 dBi directional patch |
| Ground antenna | 35 dBi (3–5 m dish) |
| FEC | LDPC/Turbo |
| One-way delay | 1.3 s |
| Link margin | ~7 dB |

First amateur networking node beyond Earth orbit. Ground segment requires EME-class stations (3–5 m dishes). The ESA Moonlight LCNS programme — with the primary contract awarded to Telespazio in October 2024 [12] — establishes the commercial cislunar communication infrastructure alongside which RADIANT would operate as an independent amateur stakeholder. Seeking ESA ARTES support for amateur participation.

---

## 6. Ground Segment Architecture

The ground segment extends the FUNcube distributed reception model into active DTN participation. Rather than passively receiving telemetry, amateur stations become store-and-forward nodes routing bundles through the network.

### 6.1 European Ground Segment Topology

**Current topology:**

| Node | Location | Role | Connectivity |
|------|----------|------|--------------|
| G4DPZ | UK | Primary development and gateway station | TCPCLv3, UHF RF |
| AMSAT-DL | Germany | European DTN node, UHF ground station | TCPCLv3, UHF RF |
| French node | France | Hardy BPA nodes on 44net, FOSM-1 satellite link | 44net, TCPCLv3 |
| QO-100 backbone | GEO (Es'hail-2) | Persistent European-wide connectivity | Always-on GEO relay |

The geographic distribution is deliberate: a LEO satellite visible from Germany but not the UK delivers bundles to the German station, which forwards them via QO-100 or TCPCLv3 to the rest of the network. Multiple European stations multiply available contact time per orbit.

### 6.2 TCPCLv3 Gateway

Deployed and accepting connections. Any internet-connected DTN node can deliver bundles into the amateur RF network via standard TCPCLv3 (RFC 7242). Tested with both ION-DTN and Hardy clients. This provides immediate interoperability with any conformant DTN implementation worldwide.

### 6.3 Contact Graph Routing in the Ground Segment

CGR computes optimal paths across the distributed ground segment. When a direct path is unavailable, bundles route via QO-100, terrestrial TCPCLv3 links, or intermediate nodes — automatically, without operator intervention. For LEO phases, contact plans derive from TLE/SGP4 orbit propagation. For terrestrial and GEO links, contact plans reflect scheduled operating windows and always-available paths.

---

## 7. Implementation Status and Operational Results

### 7.1 Production Deployment

RADIANT is not a proposal — it is operational. The following components are deployed on production infrastructure:

- ION-DTN node with HTTP/JSON management API
- TCPCLv3 gateway deployed and accepting connections
- Direct ION controller for ground station management
- Automated CI pipeline with property-based tests validating every commit
- Open-source codebase (MIT licence) publicly available
- Project website operational (https://radiant.amsat-uk.org)

### 7.2 Interoperability Results

Multi-implementation interoperability is a fundamental requirement for any networking standard. RADIANT has demonstrated:

**ION-DTN ↔ Hardy LTP interoperability:** Successful 1 MB bundle transfers between JPL's ION-DTN and Hardy's independent Rust LTP implementation. This proves that two independently-developed implementations can exchange LTP segments and deliver BPv7 bundles reliably — the same validation that TCP required across BSD, System V, and commercial implementations in the 1980s.

**Three-node cislunar simulation:** Store-and-forward with true packet-level propagation delays — 1.3 seconds for Earth-Moon paths, configurable up to 12 minutes for Mars scenarios. CGR computing multi-hop relay paths across the simulated topology.

**TCPCLv3 gateway interoperability:** Standard RFC 7242 implementation accepting connections from any conformant DTN node. Tested with both ION and Hardy clients.

These results are reproducible. Anyone can clone the repository and run the full three-node cislunar scenario locally.

### 7.3 Validated Store-and-Forward Scenarios

Three scenarios confirmed operational:

1. **Delayed uplink** — Bundle held at ground station during link outage, delivered when contact resumes
2. **Immediate forwarding** — Real-time relay when end-to-end path exists
3. **Delayed downlink** — Remote node stores bundles for future delivery when ground station becomes available

### 7.4 Software Components

| Component | Status | Notes |
|-----------|--------|-------|
| DTN-engine-agnostic abstraction layer (Rust) | Operational | Supports ION-DTN, Hardy, µD3TN |
| KISS framing library (`no_std`) | Complete | Flight-hardware compatible |
| 3-node cislunar simulation | Operational | CGR + store-and-forward |
| LTP-over-KISS convergence layer | Complete | Protocol design and implementation |
| Callsign-EID configuration | Complete | Dual EID scheme (ipn:// + dtn://) |
| ION-DTN integration | Operational | Direct controller mode available |
| Hardy integration | Operational | 1 MB LTP interop proven |
| TCPCLv3 gateway | Deployed | Accepting external connections |
| Contact plan computation (TLE/SGP4) | Implemented | Automated orbital predictions |
| Property-based test suite | Complete | CI-validated |

### 7.5 Planned

- QO-100 demonstration (Phase 1.5)
- µD3TN integration for CubeSat flight software
- Engineering model on STM32U585 (Phase 2)
- Flight CubeSat payload (Phase 3)
- Cislunar mission (Phase 4)

---

## 8. Alignment with SSI Principles

RADIANT directly implements several principles identified in the IPNSIG Solar System Internet Architecture and Governance report [1]:

| SSI Principle | RADIANT Implementation |
|---|---|
| Standards-based protocols | BPv7 (RFC 9171), LTP (RFC 5326), CGR, TCPCLv3 (RFC 7242) |
| Multistakeholder governance | Amateur operators, AMSAT orgs, open-source community, Goonhilly |
| Transparency | Open source (MIT), public protocols, regulatory requirement for plaintext |
| Interoperability | DTN-engine-agnostic abstraction; proven multi-implementation LTP interop |
| Fair identifier allocation | Callsign-EIDs leverage existing ITU-coordinated amateur callsign system |
| Security without confidentiality | BPSec integrity (HMAC) without payload encryption |
| Autonomy and automation | Automated contact plan computation, CGR routing, store-and-forward |

The project demonstrates that the amateur radio community can participate as a stakeholder in the SSI using standard protocols and existing governance mechanisms. RADIANT is to the Solar System Internet what early university and amateur networks were to the terrestrial Internet — a proving ground for protocols and operational concepts before commercial deployment.

The IETF Time-Variant Routing (TVR) Working Group [6] is developing standards directly relevant to contact graph approaches used in DTN. RADIANT's contact plan architecture and CGR implementation provide operational experience with time-variant routing that can inform TVR standardisation work, and the project will adopt TVR standards as they mature.

---

## 9. Future Directions

### 9.1 Contact Plan as a Service

As RADIANT matures beyond individual DTN nodes, the architecture is evolving toward network orchestration. The key concept under development: **Contact Plan as a Service (CPaaS)**.

Today, contact plans are managed within individual DTN implementations. CPaaS treats contact information as a shared network resource managed independently of the DTN engine:

```
Mission Planning / Applications
            │
            ▼
   RADIANT Contact Service
            │
     ┌──────┼──────┐
     │      │      │
     ▼      ▼      ▼
  ION-DTN  Hardy  µD3TN
```

The contact service becomes the authoritative source for network opportunities. Individual DTN implementations remain responsible for bundle forwarding. Applications express delivery requirements — destination, priority, delivery confidence — and the distributed ground segment determines the optimal path.

This mirrors the evolution in commercial Ground Station as a Service platforms where users request communication opportunities rather than access to specific antennas. For RADIANT, an operator would request "deliver this bundle to the lunar node with 95% confidence" and the European ground segment — distributed across multiple stations and countries — would coordinate delivery automatically.

### 9.2 Network Orchestration Layer

A longer-term extension of CPaaS introduces a Network Orchestration Layer managing:

- **Contact information** — scheduled contacts, predicted visibility windows, link characteristics
- **Network topology** — reachability information, network maps, relay relationships
- **Policy management** — traffic priorities, service classes, resource allocation
- **Security services** — node identities, trust relationships, authentication policies
- **Monitoring and telemetry** — link performance, network utilisation, delivery statistics

This positions RADIANT as a coordination platform for amateur space networking rather than merely a collection of independent nodes. The DTN implementation abstraction layer already operational today is the architectural foundation for this evolution.

### 9.3 DTN Implementation Interoperability Testing

RADIANT's engine-agnostic architecture enables systematic cross-implementation testing over real RF links. The interoperability test plan will validate:

- Bundle exchange between engines over LTP
- EID resolution consistency across implementations
- CGR contact plan interpretation
- LTP session behaviour under identical timing constraints
- Correct handling of BPSec integrity blocks across heterogeneous nodes

The amateur radio environment provides unique testing conditions: real RF propagation, different engines at different ground stations, constrained bandwidth exercising LTP edge cases, and flight payloads potentially using a different engine than ground infrastructure.

### 9.4 Other Planned Work

1. **Protocol definition publication**: Formal specification of the `dtn://callsign` EID convention as the station identification mechanism for amateur DTN
2. **SANA registration**: Formalise RADIANT node numbers in the BP identifier space
3. **Multi-node contact graph**: Distributed routing across ground station networks with time-dependent Dijkstra path computation
4. **Security model**: Document integrity-without-confidentiality as a BPSec profile contribution to SSI security architecture
5. **Flight software selection**: Final evaluation of µD3TN and Hardy for STM32U585 platform constraints
6. **Ground station network expansion**: From current European topology to global amateur ground station participation

---

## 10. Partners and Collaboration Model

**AMSAT-UK** — Organisational lead. Provides the institutional framework, access to the global amateur satellite community, and editorial platform for community engagement.

**AMSAT-DL** — European ground segment and QO-100 heritage. Built the amateur transponders on Es'hail-2 that Phase 1.5 will use. Hosts European DTN nodes and brings deep amateur satellite engineering expertise (Phase 3D, P3D, QO-100).

**Goonhilly Earth Station** — Professional ground station expertise. Potential access to large-aperture antennas for cislunar phases. Provides commercial lunar communications experience (Intuitive Machines IM-1, IM-2 support) directly relevant to RADIANT's technical challenges.

The collaboration model is deliberately multistakeholder: amateur operators contributing ground station time, open-source developers contributing code, academic institutions providing research context, and professional organisations providing engineering rigour. Active contributions include UHF ground station operators across Europe, Hardy BPv7 development, and French 44net operators running DTN nodes connected to the FOSM-1 satellite payload currently in orbit.

---

## 11. Related Work

- **KPLO DTN DTO**: Korea Pathfinder Lunar Orbiter flies BP as a Development Test Objective, demonstrating the protocol suite in cislunar space [1]
- **ION-DTN**: JPL's reference implementation used on multiple deep-space missions [2]
- **µD3TN**: Lightweight DTN for microcontrollers, space-tested [3]
- **Hardy**: Modular Rust BPv7 implementation with `no_std` core [4]
- **LunaNet**: NASA's interoperability specification for lunar communications [5]
- **IETF TVR WG**: Time-Variant Routing working group, developing standards for contact graph approaches [6]
- **ORI (Open Research Institute)**: Open-source modem and signal chain work (DVB-S2, LDPC) potentially complementary for higher-rate phases [7]
- **ESA Moonlight LCNS**: Lunar Communications and Navigation Services programme, primary contract awarded to Telespazio (October 2024) [12]

---

## 12. Conclusion

RADIANT demonstrates that the DTN protocol architecture recommended for the Solar System Internet can be implemented and operated over amateur radio infrastructure. The project has moved beyond demonstration into production deployment: multi-implementation LTP interoperability is proven, a TCPCLv3 gateway is operational, three validated store-and-forward scenarios confirm the architecture, and a European ground segment with QO-100 backbone connectivity is deployed.

The LTP-over-KISS protocol design eliminates unnecessary protocol layers while maintaining full regulatory compliance through callsign-embedded Endpoint Identifiers. The five-phase roadmap progressively validates the architecture from terrestrial links through to cislunar distances, with each phase building operational evidence for the next.

Amateur radio operators are running operational BPv7/LTP nodes using the same protocols specified in LunaNet. The addressing scheme solves the identifier allocation problem the SSI report identifies — using governance infrastructure that has functioned reliably for nearly a century. The regulatory environment enforces the transparency the SSI report recommends.

The historical parallel holds: early university networks were not competitors to ARPANET — they were the ecosystem that made the internet possible. RADIANT is building the equivalent ecosystem for the Solar System Internet. Not a parallel amateur-only network, but interoperable infrastructure that validates protocols, generates operational data, and demonstrates that diverse stakeholders can participate meaningfully in space networking.

The project contributes to the SSI ecosystem as the first non-agency operational DTN stakeholder, demonstrating that the multistakeholder governance model is not aspirational but achievable. As the ESA Moonlight programme establishes commercial cislunar infrastructure and the IETF TVR Working Group standardises time-variant routing, RADIANT provides the amateur community's operational foundation for participation in both.

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

[12] ESA, "Moonlight: Lunar Communications and Navigation Services," ESA/Telespazio contract award, October 2024. Available: https://www.esa.int/Applications/Connectivity_and_Secure_Communications/Moonlight

[13] IETF, "TCP Convergence-Layer Protocol Version 3," RFC 7242, June 2014.

---

## Appendix A: Acronyms

| Acronym | Expansion |
|---------|-----------|
| BPv7 | Bundle Protocol Version 7 |
| CGR | Contact Graph Routing |
| CLA | Convergence Layer Adapter |
| CPaaS | Contact Plan as a Service |
| DTN | Delay/Disruption Tolerant Networking |
| EID | Endpoint Identifier |
| FEC | Forward Error Correction |
| GFSK | Gaussian Frequency Shift Keying |
| KISS | Keep It Simple, Stupid (TNC framing protocol) |
| LCNS | Lunar Communications and Navigation Services |
| LDPC | Low-Density Parity-Check |
| LTP | Licklider Transmission Protocol |
| OWLT | One-Way Light Time |
| SSI | Solar System Internet |
| TCPCLv3 | TCP Convergence-Layer Protocol Version 3 |
| TNC | Terminal Node Controller |
| TVR | Time-Variant Routing |

---

## Appendix B: Project Links

- **Website**: https://radiant.amsat-uk.org
- **Source**: https://github.com/g4dpz/cislunar_proposal
- **AMSAT-UK**: https://amsat-uk.org
- **AMSAT-DL**: https://amsat-dl.org
- **ION-DTN**: https://sourceforge.net/projects/ion-dtn/
- **µD3TN**: https://gitlab.com/d3tn/ud3tn
- **Hardy**: https://github.com/hardybp/hardy
- **IPNSIG**: https://www.ipnsig.org
- **IETF TVR WG**: https://datatracker.ietf.org/wg/tvr/about/
- **Author contact**: dave@g4dpz.me.uk / G4DPZ
