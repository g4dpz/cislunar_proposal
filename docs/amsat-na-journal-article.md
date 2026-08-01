# RADIANT: Bringing NASA's DTN Protocols to Amateur Radio Ground Stations

**David Johnson, G4DPZ**

*How the amateur community can participate in the emerging Solar System Internet — using the same protocols flying on NASA missions today*

---

## The NASA Connection

In 2024, the Korean Pathfinder Lunar Orbiter (KPLO) tested Bundle Protocol at lunar distance as a development objective. The International Space Station has been running Delay-Tolerant Networking experiments for years. NASA's Jet Propulsion Laboratory maintains ION-DTN — a flight-heritage DTN engine that has supported multiple deep-space missions. These are not laboratory curiosities. They are operational protocols in space right now.

The protocols underlying these missions — Bundle Protocol version 7 (RFC 9171), Licklider Transmission Protocol (RFC 5326), and Contact Graph Routing — are openly specified through IETF and CCSDS standards. They form the foundation of what will become the Solar System Internet: a network that functions when round-trip times are measured in seconds, minutes, or hours, and when links appear and disappear according to orbital mechanics.

RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — implements these same standards for amateur radio. The project is supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station, and its architecture directly uses JPL's ION-DTN engine. The question RADIANT poses is straightforward: if agencies are building a Solar System Internet, can amateurs participate as first-class nodes?

---

## The Solar System Internet Vision

In September 2023, the Internet Society's Interplanetary Networking Special Interest Group (IPNSIG) — co-chaired by Vint Cerf, one of the fathers of the internet — published a landmark report on Solar System Internet Architecture and Governance. The report's recommendation is unambiguous: DTN and the Bundle Protocol suite should form the basis for all communications that might have to traverse paths the IP protocol suite cannot support.

The report draws an explicit parallel with 1983, when three disparate networks — ARPANET, SATNET, and PRNET — were unified by TCP/IP to create the internet. Space networking, the report argues, will follow the same trajectory: multiple organisations operating independent networks, interconnecting through standard protocols and peering agreements.

Critically, the IPNSIG report calls for multistakeholder governance. Decisions about space networking architecture should not be made solely by space agencies. Other participants — including the amateur radio community — have a legitimate role. NASA's LunaNet Interoperability Specification mandates the same protocols RADIANT implements. ESA's Moonlight programme builds on identical standards.

But legitimacy requires demonstrated operational competence. Amateur radio operators cannot simply claim a seat at the table — we must earn it by running operational DTN nodes, contributing open-source tools, and generating performance data that validates the protocols in real-world conditions. RADIANT is how the amateur community builds that credibility.

**[FIGURE 1: Architecture diagram showing ION-DTN at the centre with connections to NASA heritage (KPLO, ISS, Deep Space Network) on one side and RADIANT amateur ground stations on the other, highlighting protocol compatibility. Caption: "RADIANT uses the same ION-DTN engine and protocol standards deployed on NASA missions — not a parallel amateur-only invention."]**

---

## What RADIANT Does

RADIANT is open-source ground station software that manages the lifecycle of a DTN node over amateur radio links. At its core, it provides:

**ION-DTN Integration:** Direct lifecycle management of JPL's ION-DTN engine via an HTTP/JSON API. Operators can monitor node health, inspect bundle queues, trigger transmissions, and manage contact plans remotely. A direct ION controller handles operations without abstraction layer overhead when maximum performance is needed.

**Backend-Agnostic Architecture:** While ION-DTN is the primary engine (chosen for NASA compatibility), the architecture supports multiple DTN implementations. Hardy — an independent Rust-based BPv7 implementation — has demonstrated LTP interoperability with ION at 1 MB bundle transfers. µD3TN provides a lightweight alternative for resource-constrained nodes. Operators choose the engine that fits their station.

**Space Link Support:** LTP convergence layers carry bundles over radio links with deferred acknowledgements, specifically designed for round-trip times measured in seconds. KISS framing interfaces directly with TNCs and software modems — no AX.25 layer, just DTN bundles in KISS frames for maximum efficiency.

**Terrestrial Gateway:** TCPCLv3 (RFC 7242) provides standard TCP connectivity for ground-segment interconnection. Any DTN node on the internet can connect to a RADIANT gateway and deliver bundles into the amateur radio network — the same transport protocol used for NASA ground links.

**Contact Graph Routing:** Transmission scheduling based on predicted contact windows. For terrestrial links, this might mean scheduled operating hours. For LEO passes, it means orbital predictions from TLE data. For cislunar operations, it means precise geometric calculations of antenna visibility.

Three operational scenarios have been validated: delayed uplink (bundle storage during link outage), immediate forwarding (real-time relay when path exists), and delayed downlink (spacecraft stores bundles until ground station rises). Each demonstrates a different facet of store-and-forward behaviour essential for space operations.

---

## Interoperability with NASA Systems

RADIANT's architecture deliberately maintains compatibility with agency infrastructure. This is not a parallel amateur-only network — it is designed to interoperate.

ION-DTN is the same engine JPL developed for deep-space missions. By using it directly, RADIANT ground stations speak the same dialect as NASA nodes. LTP interoperability between ION and Hardy proves that multiple independent implementations can exchange bundles reliably — a critical requirement for any networking standard.

Endpoint addressing uses the IPN scheme (`ipn:node.service`) — NASA's addressing convention — alongside the `dtn://` URI scheme for callsign-based identification. Contact Graph Routing implementations follow the same algorithms published by JPL. TCPCLv3 for terrestrial links is identical to the protocol used in NASA's ground-segment architecture.

The LunaNet Interoperability Specification defines how nodes around the Moon will communicate. RADIANT could interoperate with this infrastructure given appropriate frequency coordination and agreements. The protocols are already aligned — what remains is operational coordination between amateur and agency networks.

This interoperability is not academic. As commercial lunar missions multiply, there will be demand for diverse ground station coverage. Amateur operators distributed across the globe can provide geographic diversity that no single agency's ground network matches. Twenty-four-hour pass coverage for a LEO satellite requires many stations — agencies have few, amateurs have many.

**[FIGURE 2: Link budget comparison table showing LEO CubeSat (437 MHz, 9.6 kbps, 2W TX, ground Yagi, 31 dB margin) and Cislunar (S-band, 500 bps, 5W TX, 3-5m dish, 7 dB margin). Caption: "Both RADIANT mission phases close their link budgets with amateur-accessible ground equipment."]**

---

## The Amateur Advantage

The amateur radio community brings unique strengths to space networking that agencies cannot easily replicate.

**Historical precedent:** Amateurs pioneered packet networking before the commercial internet existed. Store-and-forward BBS systems in the 1980s were, in essence, early delay-tolerant networks. OSCAR 1 launched in 1961 — decades before commercial satellite constellations. The amateur community has consistently been first to demonstrate technologies that later became mainstream.

**Geographic diversity:** Amateur ground stations span every continent. For LEO satellite passes lasting 5–10 minutes, continuous coverage requires stations distributed around the globe. A coordinated amateur ground network provides this at near-zero marginal cost — operators volunteer their time and equipment.

**Regulatory transparency:** Amateur radio regulations require published protocols and prohibit encryption. This aligns perfectly with open networking standards. Every protocol RADIANT uses is documented in public IETF RFCs and open-source code. Callsign-based addressing (`dtn://callsign/service`) provides globally unique identifiers without a central registration authority — every licensed amateur already has one.

**Operational culture:** Amateur operators are accustomed to operating with limited power, marginal links, and intermittent propagation. These are precisely the conditions DTN was designed for. The discipline of moonbounce operation — precise scheduling, low signal-to-noise ratios, patience — translates directly to cislunar DTN operations.

---

## The Phased Roadmap

RADIANT progresses through validated phases, each building capability for the next:

**Phase 1 — Terrestrial (In Progress):** 9600-baud packet radio using a Raspberry Pi, Mobilinkd TNC4, and standard VHF/UHF radio. Validates the complete software stack over real amateur links. Two-node testing operational.

**Phase 1.5 — QO-100 GEO Satellite (Planned):** DTN data through Es'hail-2's transponder, introducing authentic 250 ms one-way space delay. Validates LTP behaviour over a real space link — the critical stepping stone.

**Phase 2 — CubeSat Engineering Model (Planned):** Flight-representative hardware (STM32U585 OBC) with simulated orbital passes, power cycling, and fault injection. Proves the flight software before launch.

**Phase 3 — LEO CubeSat (Planned):** Amateur DTN payload at 437 MHz, 9.6 kbps. Approximately 31 dB link margin means any station tracking amateur LEO satellites can participate. Ground-to-space DTN ping and store-and-forward messaging from orbit.

**Phase 4 — Cislunar (Planned):** S-band at 500 bps with strong FEC. A 3–5 metre ground dish provides approximately 7 dB margin. The first amateur networking node beyond Earth orbit — demonstrating that the amateur community can operate in the cislunar environment where Artemis, Gateway, and commercial missions will function.

The ground segment scales with each phase: from packet radio operators at Phase 1, through existing satellite operators at Phase 3, to EME-class stations at Phase 4.

**[FIGURE 3: Phased roadmap timeline showing progression from terrestrial to cislunar, with ground station requirements at each phase. Caption: "Each phase validates protocols and operations for the next, scaling from entry-level packet radio to cislunar-capable ground stations."]**

---

## How AMSAT-NA Members Can Participate

The path to participation is immediate and accessible:

**Packet radio operators:** Any 9600-baud station can join Phase 1 testing today. The software is open-source, runs on a Raspberry Pi, and interfaces with standard TNCs.

**LEO satellite operators:** Phase 3 uses the same UHF equipment and tracking techniques as current amateur satellite work. If you can work AO-91 or ISS, you can work a RADIANT CubeSat.

**EME and microwave operators:** The cislunar phase needs ground stations with capabilities identical to moonbounce — 3–5 metre dishes, low-noise front ends, S-band capability. Your expertise and equipment are directly relevant.

**Software developers:** The codebase uses Rust and Go. Contributions are welcome — from protocol implementation to ground station automation to web interfaces. Everything is MIT-licensed on GitHub.

**CubeSat teams:** AMSAT-NA's engineering experience with the Fox series and other LEO missions is directly applicable. Collaboration on hosted payload opportunities or flight hardware development would accelerate the programme.

Geographic diversity across North America complements the European ground stations already being established. A transatlantic DTN network — amateur stations on both continents relaying bundles through LEO and GEO satellites — is within reach.

---

## Current Status and Partners

RADIANT is operational today. The project has a functioning three-node cislunar simulation demonstrating store-and-forward with realistic propagation delays (1.3 seconds to 12 minutes). Multi-implementation LTP interoperability is proven. A TCPCLv3 gateway is deployed and accepting connections. The production system runs on dedicated infrastructure with automated CI ensuring code quality.

Partners include AMSAT-UK (organisational lead), AMSAT-DL (European ground segment, QO-100 heritage), and Goonhilly Earth Station (professional ground station expertise). Active collaborators include operators contributing UHF ground stations and Hardy BPv7 development. The project is seeking North American partners — ground station operators, CubeSat teams, and universities interested in space networking research.

The amateur radio community built the packet networks that predated the internet. We launched satellites before commercial operators existed. Now the frontier extends beyond Earth, and the protocols that will underpin cislunar and interplanetary communications are standardised, open, and ready for us to implement. RADIANT is how we demonstrate that amateur radio belongs in the Solar System Internet — not as observers, but as operators.

**[FIGURE 4: Screenshot of RADIANT HTTP API showing node health, bundle statistics, and contact plan status. Caption: "The RADIANT management interface provides real-time visibility into DTN node operations — bundle queues, link status, and contact schedules."]**

---

**Contact:**
David Johnson, G4DPZ
Email: dave@g4dpz.me.uk
Website: https://radiant.amsat-uk.org
Source code: https://github.com/g4dpz/cislunar_proposal

---

*David Johnson, G4DPZ, is the project lead for RADIANT. He is Honorary Secretary of AMSAT-UK and a Senior Software Engineer at Goonhilly Earth Station, where he works on commercial lunar communications. He invites AMSAT-NA members to join the project at any level — from ground station participation to software development to CubeSat collaboration.*
