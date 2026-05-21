# AMSAT-UK OSCAR News Article Outline
## RADIANT: An Amateur Radio Pathway to Cislunar Delay-Tolerant Networking

**Target length:** 2 pages (~1,200–1,400 words)  
**Author:** David Johnson, G4DPZ  
**Publication:** AMSAT-UK OSCAR News

---

## 1. Opening / The Vision (120 words)

- Introduce RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed)
- Core ambition: place an amateur-operated DTN node in cislunar space — the first amateur interplanetary-style communication system
- The CubeSat and cislunar missions are the destination; earlier phases build the validated foundation to get there
- Built on NASA Glenn's HDTN (High-rate Delay Tolerant Networking) (https://www.nasa.gov/glenn/glenn-expertise-space-exploration/scan/high-rate-delay-tolerant-networking/), implementing Bundle Protocol v7 (RFC 9171) and Licklider Transmission Protocol (RFC 5326)
- Supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station
- Open-source, community-driven, MIT-licensed

---

## 2. What is DTN and Why It Matters for Space (120 words)

- TCP/IP assumes continuous connectivity — fails when links drop or have multi-second delays
- DTN stores bundles at each node and forwards when the next contact becomes available
- Analogous to packet radio BBS forwarding, formalised as IETF/CCSDS standards used by NASA and ESA
- Essential for: LEO passes (5–10 min windows), GEO links (500ms RTT), cislunar delays (1.3s one-way Earth–Moon)
- LTP provides reliable transfer with deferred acknowledgment — designed for long round-trip times
- Contact Graph Routing schedules transmissions based on predicted contact windows (orbital passes, antenna availability)

---

## 3. The CubeSat Mission — LEO DTN Payload (200 words)

- **Primary goal**: demonstrate ground-to-space DTN ping and store-and-forward messaging from orbit
- **OBC**: STM32U585 ultra-low-power ARM Cortex-M33 (160 MHz, 2 MB flash, 786 KB SRAM)
- **RF**: Flight-qualified IQ transceiver IC interfacing directly with STM32U585 via DAC/ADC
- **Frequency**: UHF 437 MHz, 9.6 kbps GMSK/BPSK — accessible to amateur ground stations worldwide
- **Link budget**: 2W TX, omnidirectional satellite antenna, 12 dBi ground Yagi → ~31 dB margin
- **Operations**: store-and-forward message delivery, DTN ping, priority-based message handling, CGR-predicted contact windows using TLE/SGP4
- **Storage**: 64–256 MB external NVM for persistent bundle store surviving power cycles
- **Power**: 5–10W average, STM32U585 Stop 2 mode (~16 µA) between passes
- **Mass**: 1–3 kg payload
- Distributed amateur ground station network with open-source client software
- Deliverables: flight-proven DTN payload design, operational procedures, public performance dataset

---

## 4. The Cislunar Mission — Deep-Space Amateur DTN (200 words)

- **Primary goal**: extend DTN operations beyond Earth orbit — first amateur cislunar networking node
- **Concept**: hosted payload on cislunar spacecraft or lunar CubeSat (highly elliptical orbit or lunar transfer)
- **RF**: S-band 2.2 GHz at 500 bps, BPSK with strong LDPC/Turbo FEC
- **Link budget**: 5W TX, 10 dBi directional patch antenna, 35 dBi ground dish (3–5m) → ~7 dB margin
- **Delay regime**: 1.3-second one-way light time (Earth–Moon average distance 384,000 km)
- **Storage**: 256 MB–1 GB NVM for long-duration message storage
- **Power**: 10–20W, <5 kg mass
- **Ground segment**: Tier 3/4 stations — 3–5m dishes or phased arrays, low-noise front ends
- **Experiments**: Earth–cislunar DTN ping, delay-tolerant file transfer, network resilience under extreme latency
- **Key challenges**: tight link margins, radiation, limited power/antenna gain, frequency coordination
- **Mitigation**: conservative data rates, robust FEC, phased validation through earlier phases
- Seeking ESA ARTES support and university ground station partnerships

---

## 5. Building Blocks — The Phased Approach (250 words)

Each earlier phase validates critical elements needed for the CubeSat and cislunar missions:

**Phase 1 — Terrestrial Validation (In Progress):**
- Raspberry Pi + Mobilinkd TNC4 + Yaesu FT-817 at 9600 baud G3RUH
- Validates the complete software stack: HDTN, LTP-over-KISS, callsign EIDs
- Proves store-and-forward and DTN ping over real amateur radio links
- Two-node testing by G4DPZ

**Phase 1.5 — QO-100 GEO Satellite (Planned):**
- First space-based DTN demonstration via Es'hail-2 (25.9°E, always visible)
- 2.4 GHz uplink / 10.45 GHz downlink through narrowband transponder
- Validates LTP behaviour with authentic 500ms round-trip space delay
- Proves the architecture over a real space link before committing to orbital hardware

**Phase 2 — CubeSat Engineering Model (Planned):**
- Flight-representative STM32U585 OBC + Ettus B200mini SDR (IQ baseband)
- Simulated orbital passes (5–10 min windows, 4–6 passes/day)
- Power budget profiling, fault injection, thermal/vacuum readiness
- Validates identical flight software on identical flight hardware before launch

**Terrestrial analogue concept:**
- The ground network deliberately mirrors the cislunar communications path: Mission Operations → Ground Gateway → Amateur RF link → Relay Node → Payload endpoint
- Each terrestrial demonstration exercises the same DTN protocols, store-and-forward behaviour, and contact scheduling that a cislunar mission requires
- RF link types include: 9600 baud packet, microwave point-to-point, QO-100 satellite, and EME-inspired operational patterns (moonbounce timing disciplines)
- Goal: build credible operational evidence to support a flight experiment case

**Current achievements:**
- Functioning 3-node cislunar simulation with true packet-level propagation delays (1.3s Moon, 3–12 min Mars)
- CGR computing multi-hop relay paths, LTP managing 2.6s to 24-minute RTTs
- Custom C++17 KISS CLA plugin for HDTN
- Go orchestrator managing HDTN lifecycle, telemetry, and contact plans
- 11 property-based tests, full CI pipeline

---

## 6. Protocol Stack, Regulatory Compliance, and Protocol Publication (200 words)

- Simplified stack: Application → BPv7 → LTP → KISS → G3RUH (9600 baud)
- LTP wrapped directly in KISS framing — eliminates AX.25 layer entirely
- Saves 15 bytes/frame overhead (~10% throughput improvement)
- Station identification: callsigns embedded in DTN Endpoint Identifiers (`dtn://g4dpz-1`)
- Every bundle carries operator callsign in source EID — satisfies identification requirement
- Periodic beacon bundles every 10 minutes for additional compliance
- Beacon payload is plaintext (CALLSIGN, NODE, EID, TIME, GRID) — any station demodulating the signal can identify the transmitter even when the wire format carries opaque numeric `ipn://` EIDs
- Fully compliant with amateur radio regulations
- Dual EID scheme: `ipn://` for compact CGR routing on space links, `dtn://` for callsign metadata

**Protocol definition release:**
- Amateur radio regulations (ITU Radio Regulations, Ofcom, FCC Part 97) require that protocols used on amateur frequencies are published and available for inspection
- RADIANT will publish a formal protocol definition document specifying the `dtn://callsign` EID convention as the mandatory station identification mechanism
- This document defines how callsigns are encoded in the EID URI scheme, the SSID convention (0–15), service demux paths, and beacon format
- Publication ensures any amateur operator or regulator can inspect and verify that every transmission carries a valid callsign
- Follows precedent of APRS, FT8, D-STAR, Winlink: each has a published protocol specification with callsign identification defined
- The protocol definition will be released as an open document alongside the open-source implementation (MIT licence)

---

## 7. How to Get Involved / Call to Action (100 words)

- Ground station tiers from entry-level (Yagi + RPi) to deep-space (3–5m dish)
- Phase 1 participation: Raspberry Pi, TNC4, any 9600 baud radio
- QO-100 participation: standard QO-100 ground station setup
- Seeking: amateur radio clubs, universities, CubeSat teams, microwave experimenters, EME/weak-signal operators, packet radio operators, space networking researchers
- All software open-source (MIT), documentation public
- Website: https://radiant.amsat-uk.org
- Contact: dave@g4dpz.me.uk / G4DPZ
- Invitation to join testing, contribute code, provide ground station time, or collaborate on CubeSat/cislunar hardware

---

## Suggested Figures/Diagrams (for 2-page layout)

1. **Protocol stack diagram** — vertical stack: BPv7 → LTP → KISS → G3RUH (compact, fits in column)
2. **Phase roadmap** — horizontal timeline arrow: Phase 1 → 1.5 → 2 → 3 (LEO CubeSat) → 4 (Cislunar), with CubeSat and Cislunar visually emphasised as destinations
3. **Link budget comparison table** — LEO (31 dB margin, 9.6 kbps) vs Cislunar (7 dB margin, 500 bps)

---

## Notes for Editor

- Article leads with the CubeSat and cislunar missions as the headline goals, then explains how earlier phases build toward them
- All five phases mentioned but emphasis weighted toward LEO and cislunar payloads
- Technical depth appropriate for OSCAR News audience (licensed operators familiar with satellite operations)
- Link budgets included to demonstrate feasibility with amateur equipment
- QO-100 phase mentioned as stepping stone — will resonate with existing QO-100 community
