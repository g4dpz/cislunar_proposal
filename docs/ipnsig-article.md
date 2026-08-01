# RADIANT: Amateur Radio Builds Operational DTN Infrastructure for the Solar System Internet

**David Johnson, G4DPZ**  
**AMSAT-UK / Goonhilly Earth Station**

*Project URL: https://radiant.amsat-uk.org*  
*Full technical paper: "RADIANT: Implementing Delay-Tolerant Networking Over Amateur Radio Links for Cislunar Communication" — available in the IPNSIG TD WG Zotero Library*

---

## What Is RADIANT?

RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed) is an operational project running BPv7, LTP, and Contact Graph Routing over amateur radio links. Not a simulation. Not a proposal. Deployed, tested, and producing results.

The project implements the exact protocol stack recommended by the IPNSIG Solar System Internet report — Bundle Protocol version 7 (RFC 9171), Licklider Transmission Protocol (RFC 5326), and CGR — over amateur radio infrastructure. The goal: demonstrate that amateur operators can function as first-class participants in the Solar System Internet using standard, interoperable protocols.

The SSI governance report describes a three-phase evolution: agency point-to-point links today, a transitional period where commercial and amateur networks emerge alongside agency backbones, and a future of interconnected networks peering via standard protocols. RADIANT positions amateur radio in that transitional phase — building operational SSI nodes now, generating the operational data and interoperability evidence that validates the multistakeholder model.

We believe RADIANT represents the first non-agency operational DTN stakeholder. The project is supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station.

---

## What Has It Proven?

Several key results are now confirmed:

**Multi-implementation LTP interoperability.** Successful 1 MB bundle transfers between JPL's ION-DTN and Hardy (an independent Rust BPv7/LTP implementation). Two independently-developed codebases exchanging LTP segments and delivering bundles reliably. This is the kind of cross-implementation validation that TCP went through in the 1980s — and it works. BPv7 and LTP are wire-format standards; conformant implementations should interoperate regardless of origin. We have proven that they do.

**Three validated store-and-forward scenarios.** Delayed uplink (bundle held during outage, delivered when contact resumes), immediate forwarding (real-time relay), and delayed downlink (remote node stores for future delivery). All operational on production infrastructure. These aren't theoretical — they run on deployed ION-DTN nodes with automated CI validating every commit.

**TCPCLv3 gateway deployed.** Standard RFC 7242 TCP convergence layer accepting connections from any conformant DTN node on the internet. Bundles injected via TCPCLv3 route into the amateur RF network. Tested with both ION-DTN and Hardy clients. This means any internet-connected DTN node worldwide can reach into the RADIANT amateur network today.

**Three-node cislunar simulation.** Store-and-forward with true packet-level propagation delays — 1.3 seconds for Earth-Moon, up to 12 minutes for Mars scenarios. CGR computing multi-hop relay paths, LTP managing round-trip times from 2.6 seconds to 24 minutes. The simulation is open-source and reproducible — clone the repository and run it locally.

**European ground segment operational.** UK node (G4DPZ), German node (AMSAT-DL), French nodes (Hardy on 44net, connected to FOSM-1 satellite payload), and QO-100 as persistent GEO backbone. Geographic distribution means a LEO satellite visible from Germany but not the UK still delivers bundles into the network — the ground segment routes around single-point visibility constraints automatically.

---

## Architecture in Brief

RADIANT's architecture is DTN-engine-agnostic. A Rust-based orchestrator manages whichever engine an operator selects:

- **ION-DTN (JPL)** — primary engine, NASA flight heritage
- **Hardy** — modular Rust implementation, proven LTP interop with ION
- **µD3TN** — lightweight, designed for constrained platforms and CubeSat flight software

The RF convergence layer wraps LTP directly in KISS frames — no AX.25 — eliminating 15+ bytes of redundant overhead per frame and yielding ~10% throughput improvement at 9600 baud. Callsigns are embedded in DTN Endpoint Identifiers (`dtn://g4dpz-1/service`), satisfying ITU Article 25 identification requirements in every bundle without any additional protocol mechanism.

No encryption. All payloads in the clear, all protocols publicly documented. BPSec integrity (HMAC) available for authentication without obscuring content. This isn't a limitation — it's structural alignment with the SSI transparency principle.

The full technical paper details the frame structures, LTP configuration tables, CGR JSON format, and link budgets for each phase.

---

## Where Is It Going?

The five-phase roadmap progresses from current terrestrial operations through to a cislunar node:

**Phase 1 (Operational):** 9600-baud terrestrial packet radio. Complete stack validated with three confirmed scenarios.

**Phase 1.5 (Next):** DTN through QO-100's GEO transponder. 250 ms one-way delay exercises LTP deferred acknowledgement over a real space link with authentic propagation characteristics. Always-available — no orbital tracking needed — enabling rapid iteration on protocol behaviour under genuine space delay.

**Phase 2:** CubeSat engineering model on STM32U585 flight hardware (ARM Cortex-M33, 786 KB SRAM). Simulated orbital passes, power cycling, fault injection. Validates identical flight software on identical flight hardware before committing to orbit.

**Phase 3:** LEO CubeSat. First amateur DTN payload in orbit. 437 MHz UHF, 9.6 kbps, ~31 dB link margin. Any station currently tracking amateur LEO satellites can participate as a ground node. Store-and-forward messaging from orbit using the same CGR contact plans computed from TLE/SGP4 propagation.

**Phase 4:** Cislunar. S-band, 500 bps BPSK with LDPC/Turbo FEC, 1.3-second one-way delay, ~7 dB margin with EME-class ground dishes. First amateur networking node beyond Earth orbit. The ESA Moonlight LCNS programme — primary contract awarded to Telespazio in October 2024 — is establishing commercial cislunar communication infrastructure. RADIANT would operate alongside it as an independent amateur stakeholder, demonstrating that the SSI can support diverse participants from the start.

---

## Contact Plan as a Service

As the ground segment grows beyond individual nodes to a distributed European network, managing contact plans within each DTN engine independently doesn't scale. RADIANT is developing the concept of Contact Plan as a Service (CPaaS) — treating contact information as a shared network resource managed independently of the underlying DTN implementation.

Today's model: each ION-DTN or Hardy node maintains its own contact plan. Tomorrow's model: a centralised contact service computes and distributes optimal paths, and individual engines consume those plans as configuration.

The vision: applications express delivery requirements (destination, priority, confidence level) and the distributed European ground segment determines the optimal path automatically. An operator requests "deliver this bundle to the lunar node with 95% confidence" and the network — spanning UK, German, and French stations with QO-100 backbone — coordinates delivery without manual intervention.

This mirrors commercial Ground Station as a Service evolution (AWS Ground Station, Azure Orbital) where users request communication opportunities rather than access to specific hardware. For amateur DTN, the service layer would manage TLE-derived LEO contact windows, always-on GEO paths, scheduled operating periods, and inter-station relay topology as a unified resource.

Longer term, a Network Orchestration Layer would manage contact schedules, network topology, traffic policy, security services, and performance monitoring across the entire amateur DTN network. The engine-agnostic abstraction layer operational today is the architectural foundation — CPaaS is its logical extension as the node count grows.

---

## Why Should IPNSIG Care?

The SSI report asserts that multistakeholder governance is necessary. RADIANT provides empirical evidence that it works.

We're running operational BPv7/LTP nodes using the same protocols specified in LunaNet. Our callsign-EID addressing solves the identifier allocation problem the SSI report identifies — using governance infrastructure that has functioned for nearly a century. Our regulatory environment enforces the transparency the report recommends. The IETF TVR Working Group is standardising time-variant routing approaches that align directly with our CGR implementation and contact plan architecture.

The amateur radio community brings geographic diversity (stations on every continent for LEO coverage), a pre-existing globally unique identifier system (ITU-coordinated callsigns requiring no new registry), regulatory transparency (all protocols published, all payloads unencrypted), and an operational culture already matched to DTN conditions (moonbounce operators tolerate 2.6-second round trips as routine).

Amateur radio built packet networks before the commercial internet existed. We launched satellites before commercial constellations. We are now building operational DTN nodes that implement the architecture IPNSIG helped define. The protocols are standardised. The implementations interoperate. The ground segment is deploying across Europe.

What remains is the coordination between amateur and agency networks that turns independent nodes into an internetwork. We look forward to that coordination.

---

## Get Involved

The project is open-source (MIT licence) and welcomes participation:

- **Ground station operators** — run a DTN node at your station
- **Developers** — contribute to the Rust orchestrator, KISS CLA, or contact plan service
- **Researchers** — the three-node simulation and interoperability testbed are available for experimentation

Full technical details — frame structures, LTP configuration tables, CGR JSON format, link budgets, and ground segment topology — are in the companion paper available through the IPNSIG TD WG Zotero Library.

---

**Contact:**  
David Johnson, G4DPZ  
Email: dave@g4dpz.me.uk  
Website: https://radiant.amsat-uk.org  
Source: https://github.com/g4dpz/cislunar_proposal

---

*David Johnson, G4DPZ, is project lead for RADIANT, Honorary Secretary of AMSAT-UK, and a Senior Software Engineer at Goonhilly Earth Station where he works on commercial lunar communications.*
