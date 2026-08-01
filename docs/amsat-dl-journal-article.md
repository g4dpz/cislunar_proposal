# RADIANT: Delay-Tolerant Networking for Amateur Radio — A European Ground Segment

**David Johnson, G4DPZ**

*An open-source collaboration between AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station — building on QO-100 heritage to create amateur radio's first space networking infrastructure*

---

AMSAT-DL built the amateur transponders on Es'hail-2. Those transponders — now serving thousands of operators as QO-100 — represent a remarkable achievement: a permanent amateur radio presence in geostationary orbit, providing continuous connectivity across Europe, Africa, and parts of Asia. But QO-100's transponders are mirrors. They receive signals and retransmit them in real time. The satellite does not understand, store, or route the data passing through it.

RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — aims to take the next step. Rather than simply reflecting signals, RADIANT creates intelligent network nodes that receive data, understand its destination, store it when necessary, and forward it when the optimal path becomes available. It is the difference between a telephone exchange and a post office: the sender and receiver need not be active at the same time.

This open-source project, supported jointly by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station, implements the same Delay-Tolerant Networking protocols that NASA and ESA deploy for deep-space communications — Bundle Protocol version 7 (RFC 9171) and Licklider Transmission Protocol (RFC 5326). The architecture supports multiple DTN engines: ION-DTN from NASA JPL, Hardy (a Rust-based BPv7 implementation), and µD3TN.

For AMSAT-DL members, RADIANT builds directly on QO-100 heritage. Phase 1.5 of the project sends DTN data through the very transponders AMSAT-DL engineered — using QO-100 as a stepping stone between terrestrial radio links and cislunar space.

**[FIGURE 1: Network architecture diagram showing European ground segment — UK node (Goonhilly/G4DPZ), German node (AMSAT-DL), QO-100 link between them, and additional European stations. Caption: "The RADIANT European ground segment leverages AMSAT-DL's QO-100 infrastructure and distributed continental coverage."]**

---

## Why DTN — and Why QO-100 Is the Ideal Testbed

Delay-Tolerant Networking was designed for environments where conventional internet protocols fail: links with long propagation delays, intermittent connectivity, and limited bandwidth. Space is the canonical example. TCP's handshake protocol — requiring multiple round trips before data flows — cannot function when signals take 1.3 seconds to reach the Moon or 12 minutes to reach Mars.

DTN's answer is store-and-forward. Data is packaged into "bundles," held at each node, and forwarded when the next communication link opens. LTP (Licklider Transmission Protocol) provides reliable delivery with deferred acknowledgements — it sends data and waits patiently for confirmation, however long the round trip takes.

QO-100 provides the ideal environment to validate these protocols over a real space link. The approximately 250 milliseconds one-way delay (500 ms round-trip) to geostationary orbit is short enough to iterate quickly, but long enough to exercise LTP's deferred acknowledgement mechanisms authentically. Unlike a LEO satellite with 10-minute pass windows, QO-100 is always available — operators can test, debug, and refine without waiting for the next orbital pass.

AMSAT-DL's transponder infrastructure makes this possible. Phase 1.5 will carry DTN bundles through QO-100's narrowband transponder using standard ground station equipment — the same setups hundreds of European operators already use daily. No new hardware required. Just new software and a new protocol stack.

This is the critical stepping stone. If DTN works reliably through QO-100 at 500 ms round-trip, the same protocols will work at 2.6 seconds (Moon) and beyond.

---

## Current Technical Status

RADIANT is operational today, not merely a proposal. The project has demonstrated:

- **ION-DTN lifecycle management** via HTTP API — deployed on production infrastructure
- **Multi-implementation LTP interoperability** between ION-DTN and Hardy, proven at 1 MB bundle transfers
- **TCPCLv3 terrestrial gateway** — any DTN node on the internet can connect and deliver bundles into the network
- **Store-and-forward with scheduled contacts** — three timing scenarios validated: delayed uplink (bundle held until link opens), immediate forwarding, and delayed downlink (node stores bundle for future delivery)
- **Direct ION controller** for ground station operations without abstraction overhead
- **Three-node cislunar simulation** with true packet-level propagation delays (1.3s Moon, 3–12 minutes Mars)

The architecture is backend-agnostic. ION-DTN provides NASA compatibility and flight heritage. Hardy offers a modern Rust implementation with proven interoperability. µD3TN provides a lightweight option for constrained platforms. Operators choose the engine that fits their station requirements.

All code is open-source under the MIT licence, with automated CI ensuring quality with every change.

---

## The European Ground Segment

RADIANT envisions a distributed amateur ground station network spanning Europe — and AMSAT-DL's role is central to this.

The concept extends the FUNcube model: rather than passive telemetry reception, amateur stations become active participants in a store-and-forward network. When a LEO CubeSat passes over Germany but not England, the German station receives bundles and forwards them via terrestrial links (or QO-100) to the rest of the network. Geographic diversity across Europe provides near-continuous coverage for LEO passes.

The European ground segment currently includes:

- **UK node** — G4DPZ station with Goonhilly Earth Station providing professional ground segment expertise and potential access to large-aperture antennas for cislunar phases
- **German node** — AMSAT-DL hosting a European DTN node, with contributing operators providing UHF ground stations for LEO DTN reception
- **French node** — contributing operators running Hardy BPA nodes on 44net, with connections to the FOSM-1 satellite payload currently in orbit
- **QO-100 backbone** — persistent European-wide connectivity through the geostationary transponder, available 24 hours a day without orbital tracking

Contact Graph Routing computes optimal paths across this distributed ground segment. If a bundle needs to reach a station in Germany but the direct path is unavailable, CGR routes it via QO-100 or through an intermediate terrestrial node — automatically, without operator intervention.

This distributed European infrastructure is not merely convenient — it is essential for the LEO CubeSat mission. A single ground station sees a LEO satellite for perhaps 10 minutes per pass, 4–6 times per day. Multiple stations across Europe multiply the available contact time, enabling higher data throughput and faster message delivery.

**[FIGURE 2: Protocol stack diagram — BPv7 → LTP → KISS → Radio modem, with annotations showing callsign EIDs and no encryption. Caption: "The RADIANT protocol stack: space-grade DTN protocols delivered through familiar amateur radio interfaces."]**

---

## The Phased Roadmap

RADIANT progresses through validated phases:

**Phase 1 — Terrestrial Validation (In Progress):** 9600-baud packet radio using standard equipment (Raspberry Pi, TNC, VHF/UHF radio). Validates the complete software stack over real amateur RF links.

**Phase 1.5 — QO-100 (Planned):** DTN through AMSAT-DL's geostationary transponder. The first space-based amateur DTN demonstration. Validates LTP over authentic space delay using equipment operators already own.

**Phase 2 — CubeSat Engineering Model (Planned):** Flight-representative hardware (STM32U585 OBC) tested on the ground with simulated orbital passes, power cycling, and fault injection.

**Phase 3 — LEO CubeSat (Planned):** Amateur DTN payload at 437 MHz, 9.6 kbps, with 31 dB link margin. Accessible to any UHF satellite ground station.

**Phase 4 — Cislunar (Planned):** S-band DTN node in the Earth-Moon region. 500 bps with strong FEC, requiring 3–5 metre ground dishes — comparable to EME station capabilities.

Each phase validates technology for the next. QO-100 specifically proves that LTP and store-and-forward work over a real space link before the project commits to flight hardware — a pragmatic, risk-reducing approach.

---

## Protocol Stack and Amateur Compliance

The protocol stack is streamlined: BPv7 → LTP → KISS → radio modem. No AX.25 layer — bundles travel directly in KISS frames, saving overhead and improving throughput.

Station identification is embedded in every transmission through DTN Endpoint Identifiers: `dtn://g4dpz/service`, `dtn://dl0abc/service`. Every bundle carries the operator's callsign as its source address. Periodic plaintext beacons provide additional identification.

There is no encryption. All data travels in the clear, fully compliant with ITU Radio Regulations and national amateur radio rules across Europe. All protocols are publicly documented through IETF RFCs and open-source code, satisfying the requirement that amateur protocols must be available for public inspection.

---

## Call to Action for AMSAT-DL Members

RADIANT needs European operators at every level:

**QO-100 operators:** You already have the ground station for Phase 1.5. Standard narrowband equipment is sufficient. Help validate DTN over a real space link.

**UHF satellite operators:** Your tracking antennas and receivers are ready for the LEO CubeSat phase. Ground station participation means your station becomes an active network node, not just a passive receiver.

**Software developers:** The project uses Rust and Go. Contributions are welcome — from DTN engine integration to ground station automation to web interfaces.

**Microwave experimenters:** S-band capability is needed for the cislunar phase. If you operate on 2.3 GHz or 2.4 GHz, your experience and equipment are directly relevant.

**CubeSat and university groups:** RADIANT provides a practical framework for teaching space communications, networking protocols, and systems engineering. Collaboration on hardware development or ground station hosting is welcomed.

AMSAT-DL's heritage in amateur satellite engineering — from Phase 3D to QO-100 — demonstrates what the European amateur community can achieve. RADIANT is the next step: from building transponders that mirror signals to building intelligent nodes that participate in a space network.

---

**Contact:**
David Johnson, G4DPZ
Email: dave@g4dpz.me.uk
Website: https://radiant.amsat-uk.org
Source code: https://github.com/g4dpz/cislunar_proposal

---

*David Johnson, G4DPZ, is the project lead for RADIANT. He is Honorary Secretary of AMSAT-UK and a Senior Software Engineer at Goonhilly Earth Station. RADIANT welcomes AMSAT-DL members to join the project — from QO-100 testing to ground station operations to CubeSat collaboration.*
