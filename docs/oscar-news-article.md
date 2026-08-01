# RADIANT: An Amateur Radio Pathway to Cislunar Delay-Tolerant Networking

**David Johnson, G4DPZ**

*Building amateur radio's first space networking nodes — from a CubeSat in low Earth orbit to a relay near the Moon*

---

A few years from now, an amateur ground station in southern England transmits a short data bundle toward a CubeSat passing overhead at 437 MHz. The satellite receives it, stores it in onboard memory, and continues along its orbit. Twenty minutes later, over central Europe, the spacecraft forwards that bundle to another ground station — one that was below the horizon when the message was first sent. The data has travelled via space, stored and forwarded by an intelligent node that understood the network topology and chose the optimal delivery path.

This is the goal of RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed. We are building the amateur radio infrastructure to make it real, progressing through validated phases toward two headline missions: a DTN payload on a LEO CubeSat, and ultimately an amateur networking node in cislunar space.

RADIANT is an open-source collaboration between AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station, implementing the same Delay-Tolerant Networking protocols that NASA and ESA deploy for deep-space communications — Bundle Protocol version 7 (RFC 9171) and Licklider Transmission Protocol (RFC 5326). The architecture is backend-agnostic, currently supporting ION-DTN (NASA JPL's flight-heritage engine), with the design accommodating Hardy (an independent Rust BPv7 implementation) and µD3TN.

---

## Why DTN for Amateur Satellites?

Conventional internet protocols assume continuous connectivity. TCP's three-way handshake requires multiple round trips before data flows — workable when latency is milliseconds, but unacceptable when a LEO satellite is visible for only ten minutes, or when the Moon is 1.3 light-seconds away.

DTN solves this with a store-and-forward model familiar to anyone who operated packet radio BBS networks in the 1980s. Data is packaged into "bundles," stored at each network node, and forwarded when the next communication link becomes available. LTP (Licklider Transmission Protocol) handles reliable delivery with deferred acknowledgements — designed specifically for links where round-trip times are measured in seconds rather than milliseconds. Contact Graph Routing schedules transmissions based on predicted orbital passes, much like a railway timetable.

These are not experimental protocols. The International Space Station runs DTN today. The Korean Pathfinder Lunar Orbiter (KPLO) tested Bundle Protocol at lunar distance in 2024. ESA's Moonlight programme builds on the same standards. RADIANT brings these operational protocols to amateur radio.

**[FIGURE 1: Protocol stack diagram — vertical stack showing Application → BPv7 → LTP → KISS → G3RUH (9600 baud). Caption: "The RADIANT protocol stack eliminates AX.25 entirely, carrying DTN bundles directly in KISS frames for maximum efficiency."]**

---

## The CubeSat Mission: LEO DTN Payload

The primary near-term goal is an amateur DTN payload aboard a LEO CubeSat, demonstrating ground-to-space DTN ping and store-and-forward messaging from orbit.

The payload design centres on an STM32U585 ultra-low-power ARM Cortex-M33 microcontroller (160 MHz, 2 MB flash, 786 KB SRAM) with a flight-qualified IQ transceiver operating at 437 MHz UHF. Data rates of 9.6 kbps GMSK/BPSK keep the system accessible to amateur ground stations using modest equipment — a UHF Yagi antenna and an SDR or TNC.

The link budget is comfortable: 2 watts transmit power from the spacecraft, an omnidirectional satellite antenna, and a 12 dBi ground Yagi yield approximately 31 dB of margin. This means any station currently tracking amateur LEO satellites has sufficient equipment to participate.

Onboard, 64–256 MB of external non-volatile memory provides persistent bundle storage surviving power cycles and orbital eclipses. Contact Graph Routing uses TLE/SGP4 orbital predictions to schedule transmissions during ground station passes (typically 5–10 minute windows, 4–6 passes per day). Between passes, the STM32U585 enters Stop 2 mode drawing approximately 16 µA — critical for the 5–10 watt power budget.

The deliverables from Phase 3 are a flight-proven DTN payload design, published operational procedures, and a public performance dataset demonstrating store-and-forward over amateur satellite links.

---

## The Cislunar Mission: Beyond Earth Orbit

Phase 4 extends DTN operations to cislunar space — the region between Earth and the Moon where Artemis, Lunar Gateway, and commercial lunar missions will operate. This would be the first amateur-operated interplanetary-style communication system.

The concept is a hosted payload on a cislunar spacecraft or lunar CubeSat, operating on S-band (2.2 GHz) at 500 bits per second with strong LDPC/Turbo forward error correction. The link budget is tighter: 5 watts transmit power, a 10 dBi directional patch antenna on the spacecraft, and a 35 dBi ground dish (3–5 metres) provide approximately 7 dB of margin.

The 1.3-second one-way light time to the Moon creates a 2.6-second minimum round-trip — well within LTP's design envelope. DTN's store-and-forward model means data accumulates across multiple contact windows; 500 bps is slow, but messages get through reliably.

Ground segment requirements for the cislunar phase are comparable to EME (moonbounce) stations: 3–5 metre dishes with low-noise front ends on S-band. University ground stations and larger amateur installations would form the Tier 3/4 ground network.

**[FIGURE 2: Link budget comparison table — two columns: LEO CubeSat (437 MHz, 9.6 kbps, 31 dB margin, ground Yagi) vs Cislunar (S-band, 500 bps, 7 dB margin, 3–5m dish). Caption: "Both missions close their link budgets with amateur-accessible equipment."]**

---

## Building Blocks: The Phased Approach

Each earlier phase validates critical elements needed for the orbital missions.

**Phase 1 — Terrestrial Validation (In Progress):** A Raspberry Pi, Mobilinkd TNC4, and Yaesu FT-817 at 9600 baud G3RUH. This validates the complete software stack — LTP-over-KISS, callsign EIDs, store-and-forward, and DTN ping — over real amateur radio links. Two-node testing is underway at G4DPZ.

**Phase 1.5 — QO-100 (Planned):** DTN data through Es'hail-2's narrowband transponder. The approximately 250 ms one-way delay provides a genuine space environment, validating LTP's deferred acknowledgement mechanism over an authentic space link before committing to orbital hardware. This phase will resonate with the large QO-100 operator community — standard narrowband ground stations are all that is required.

**Phase 2 — CubeSat Engineering Model (Planned):** A ground-based flatsat using the flight-representative STM32U585 OBC with an Ettus B200mini SDR for IQ baseband. Simulated orbital passes, power budget profiling, fault injection, and thermal readiness testing — validating identical flight software on identical flight hardware before launch.

The ground network deliberately mirrors the cislunar communications path: Mission Operations → Ground Gateway → Amateur RF link → Relay Node → Payload endpoint. Every terrestrial demonstration exercises the same protocols, store-and-forward behaviour, and contact scheduling that the orbital missions require.

---

## Protocol Stack and Regulatory Compliance

RADIANT carries DTN bundles directly in KISS frames, eliminating the AX.25 layer entirely. This saves 15 bytes per frame (approximately 10% throughput improvement) without sacrificing functionality.

Station identification is built into the protocol: callsigns are embedded in DTN Endpoint Identifiers (`dtn://g4dpz-1/service`), meaning every bundle carries the operator's callsign as its source address. Periodic plaintext beacons every ten minutes transmit callsign, grid locator, node type, and EID — any station demodulating the signal can identify the transmitter even without full DTN decoding.

There is no encryption anywhere in the system. All data travels in the clear, fully compliant with ITU Radio Regulations Article 25 and national amateur radio rules. All protocols are publicly documented through IETF RFCs and open-source code. A formal protocol definition document specifying the callsign EID convention will be published, following the precedent of APRS, FT8, D-STAR, and Winlink.

---

## Current Achievements

RADIANT is not a paper exercise. The project has a functioning three-node cislunar simulation with true packet-level propagation delays (1.3 seconds for Earth–Moon, configurable to 12 minutes for Mars scenarios). Contact Graph Routing computes multi-hop relay paths while LTP manages round-trip times from 2.6 seconds to 24 minutes. Multi-implementation LTP interoperability is proven between ION-DTN and Hardy at 1 MB bundle transfers. A TCPCLv3 terrestrial gateway allows any DTN node on the internet to connect and deliver bundles.

The entire codebase is open-source under the MIT licence, with automated CI testing ensuring reliability.

---

## Get Involved

RADIANT welcomes participation at every level. Phase 1 requires only a 9600-baud packet station. Phase 1.5 uses standard QO-100 narrowband setups. The LEO phase needs UHF satellite operators — the same equipment you already use. Software developers working in Rust or Go are welcome to contribute. Microwave and EME operators bring directly relevant experience for the cislunar phases.

We are seeking amateur radio clubs, universities, CubeSat teams, and anyone interested in space networking to join the effort.

**[FIGURE 3: Phased roadmap timeline — horizontal arrow showing Phase 1 → 1.5 → 2 → 3 (LEO CubeSat) → 4 (Cislunar), with Phases 3 and 4 visually emphasised as destination milestones. Caption: "The RADIANT roadmap: each phase validates technology for the next, building toward orbital DTN nodes."]**

---

**Contact:**
David Johnson, G4DPZ
Email: dave@g4dpz.me.uk
Website: https://radiant.amsat-uk.org
Source code: https://github.com/g4dpz/cislunar_proposal

---

*David Johnson, G4DPZ, is the project lead for RADIANT. He is Honorary Secretary of AMSAT-UK and a Senior Software Engineer at Goonhilly Earth Station, where he works on commercial lunar communications.*
