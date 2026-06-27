# Observations from the IPNSIG Solar System Internet Report

## Applicability to RADIANT

Source: IPNSIG, "Solar System Internet Architecture and Governance", September 2023
Authors: Vint Cerf, Scott Burleigh, Scott Pace, Jim Green, Laura DeNardis, et al.
Published by: Internet Society Interplanetary Chapter (IPNSIG)
Licence: Creative Commons Attribution-NonCommercial 4.0

---

## 1. RADIANT as a Pathfinder for the Solar System Internet

The report describes a three-phase evolution of space networking:

- **Today**: Space agency-sustained, point-to-point communication systems
- **Transitional**: Commercial and amateur networks emerge alongside agency backbones; public-private partnerships
- **Future**: Commercial entities provide networking services; multiple stakeholders interconnect via peering agreements

RADIANT sits squarely in the "transitional" phase. We are building an amateur SSI node before the commercial ones exist. The report explicitly anticipates that networks beyond space agency control will emerge — RADIANT is one of them.

The historical parallel is direct: "In 1983, three disparate networks (Arpanet, SATNET, PRNET) joined together to form a single network using TCP/IP, giving birth to the operational Internet. Establishing a shared and interoperable network in space will follow a similar trajectory."

RADIANT is to the Solar System Internet what early university networks were to the terrestrial Internet.

---

## 2. BPv7 + LTP + CGR as the Recommended Architecture

The report recommends:

> "DTN and the BP suite should be the basis for all communications that might have to traverse paths that the IP protocol suite cannot support. This includes all paths spanning interplanetary distances as well as all paths that might experience disruptions in communications due to antenna pointing and/or scheduling constraints."

RADIANT implements exactly this:
- BPv7 (RFC 9171) as the networking layer
- LTP (RFC 5326) for reliable delivery over delayed/disrupted links
- Contact Graph Routing for time-scheduled forwarding

We are not making a speculative protocol choice. We are implementing the architecture that NASA, ESA, CCSDS, IETF, and the Internet Society have converged on.

---

## 3. Multistakeholder Governance

The report argues strongly that SSI governance should be multistakeholder — not just space agencies:

> "For the long-term sustainability of the Solar System Internet, policy decisions should be made collaboratively amongst all stakeholders, suggesting the Multistakeholder governance model inherited from the Internet."

RADIANT already embodies multistakeholder participation:
- Amateur radio operators (globally distributed, self-governing)
- Open-source software developers
- Academic collaborators
- Professional organisations (AMSAT-UK, AMSAT-DL, Goonhilly)

The report calls for stakeholders beyond space agencies to be involved from the earliest stages. RADIANT is how the amateur radio community earns its seat at that table.

---

## 4. IPv6 Locally, BP/LTP Between Worlds

The report recommends:
- IPv6 for local networks on celestial bodies (Moon surface, Mars surface)
- BP/LTP for long-haul links between them
- IPv6 as the only IP version in space (no IPv4, to avoid NAT complexity)

This maps directly to RADIANT's architecture:
- Local IP networks at ground stations
- BP/LTP over the RF links between nodes
- The DTN abstraction layer bridges the two domains

The report also discusses application-layer gateways (HTTP over BP, Email over BP) — relevant for future RADIANT applications.

---

## 5. Contact Graph Routing Validated

The report identifies routing as the most-studied problem in interplanetary networking and states key principles:

1. **Autonomy and Automation** — routing must not require continuous human intervention from Earth
2. **Standards** — standard methods for propagating routing information
3. **Interoperability** — common inter-regional routing procedures
4. **Scalability** — must support thousands or millions of nodes

RADIANT's multi-node contact graph implementation directly addresses all four:
- Automated contact window computation from orbital predictions
- Standard JSON export format compatible with multiple DTN engines
- REST API and OTA distribution for plan sharing (interoperability)
- Time-dependent Dijkstra routing (scalable)

---

## 6. Security Without Encryption — A Legitimate Posture

The report discusses BPSec (RFC 9172) and notes:

> "The authenticity of the bundle source can be provided by digital signatures while confidentiality of the payload can be achieved by separately encrypting the payload block."

For amateur radio, we use the first without the second. This is not a compromise — it's a legitimate security posture that aligns with the report's architecture:

- Integrity protection: HMAC appended as metadata, verifiable by any node
- Source authentication: digital signatures prove who sent a bundle
- No payload encryption: amateur radio regulations require transmissions to be in the clear

The report also notes that security policies should be "transparent and published" — amateur radio regulations already enforce this by requiring all protocols to be publicly documented.

The report's recommendation for "a minimum set of security policies and interoperable security implementations" is something RADIANT can contribute to by demonstrating integrity-without-confidentiality as a viable operational model.

---

## 7. Identifier Allocation — Callsign-EIDs Solve a Governance Problem

The report discusses BP identifier allocation at length:
- Currently managed by IANA and SANA
- May need a "Space Internet Registry" (SIR) for space networks
- Policies for fair and consistent distribution are needed

RADIANT's callsign-embedded EIDs (dtn://g4dpz/service) elegantly solve this:
- Amateur radio callsigns are already globally unique
- They are already regulated and allocated by national authorities
- They already have an international coordination framework (ITU)
- No new registry needed — we leverage existing infrastructure

This is a practical demonstration of the report's principle of "hierarchical management" and "fair and consistent resource allocation" using existing governance mechanisms.

---

## 8. The "Network of Networks" Vision

The report's core vision:

> The SSI will be a "network of networks" — like the terrestrial Internet. Multiple commercial networks will interconnect based on private-private partnerships or agreements, similar to the peering concept in today's Internet.

RADIANT could be one of those networks. An amateur DTN network that peers with agency networks (NASA DSN, ESA ESTRACK) via standard BP/LTP. The long-term vision:

- RADIANT ground stations as autonomous systems in the SSI
- Peering with commercial lunar service providers via standard protocols
- Amateur relay satellites as additional routing options in the contact graph
- Like early ISPs peering with university networks in the 1990s

---

## 9. Time and Coordinate Systems

The report discusses:
- Need for common time standards (Barycentric Coordinate Time / TCB, Barycentric Dynamical Time / TDB)
- Clock synchronisation across the SSI
- Potential need for UTC(Moon) or UTC(Mars)
- Relationship between time precision and position determination

For RADIANT:
- Contact plans require accurate time for scheduling
- Phase 1-3: GPS-derived UTC is sufficient
- Phase 4 (cislunar): relativistic effects become relevant (small but non-zero)
- The report notes that "maintaining precise time is important not only for SSI network operation, but also for determining position"
- RADIANT's contact plan versioning and distribution system will need to account for time reference consistency across nodes

---

## 10. Transparency as a Design Principle

The report's transparency principle:

> "To enable collaboration, it is essential to publicly share key pieces of information. Network connectivity details, such as link availability periods, supported communication protocols, bandwidth and usage of key identifiers play a vital role in facilitating interconnections and interoperability between networks."

RADIANT already does this:
- All protocols publicly documented
- Open-source software
- Contact plans shareable via REST API
- Station identification in every transmission
- Amateur radio regulations enforce transparency by default

The report adds: "Transparency is also required in the processes used for governance and in making the decisions that will define SSI architecture and governance."

RADIANT's development process (open GitHub, public specs, AMSAT collaboration) aligns with this.

---

## 11. The KPLO Precedent

The report notes that the Bundle Protocol is "flying at the Moon today on the Korea Pathfinder Lunar Orbiter (KPLO) as a Development Test Objective." This is significant for RADIANT:

- BP/LTP is not theoretical — it's operational in cislunar space right now
- The protocol stack we're implementing has been space-qualified
- RADIANT's Phase 4 goal (cislunar DTN node) is building on proven technology

---

## 12. Spectrum Considerations

The report discusses ITU-R spectrum allocation for lunar communications:
- The US proposed a new agenda item at WRC-23 for lunar/cislunar spectrum
- Work proceeding toward WRC-27 recommendations
- Space services categorised by GSO/NGSO, E-S/S-E/S-S

For RADIANT:
- Phase 1-3: existing amateur radio allocations (VHF/UHF/microwave)
- Phase 4: may need coordination with ITU processes for cislunar amateur allocations
- The amateur radio service already has space allocations (amateur-satellite service)
- RADIANT operates within existing regulatory frameworks but should track ITU developments

---

## Summary: How RADIANT Aligns with SSI Principles

| SSI Principle | RADIANT Implementation |
|---|---|
| Collaboration | AMSAT-UK, AMSAT-DL, Goonhilly, open-source community |
| Fair resource allocation | Callsign-EIDs leverage existing amateur radio identifier system |
| Transparency | Open source, public protocols, regulatory requirement for plaintext |
| Interoperability | DTN-engine-agnostic abstraction layer, standard BP/LTP/CGR |
| Security | BPSec integrity without payload encryption, published policies |
| Multistakeholder governance | Amateur operators, developers, academics, professional orgs |
| Standards-based | RFC 9171, RFC 5326, CCSDS profiles |
| Scalable architecture | Multi-node contact graph, distributed ground stations |
| Autonomy and automation | Automated contact plan computation, CGR routing |

---

## Actionable Items for RADIANT

1. **Reference the SSI report** in project documentation and proposals — it provides institutional backing for our architecture choices

2. **Engage with IPNSIG** — consider joining as a contributing organisation; RADIANT is a practical demonstration of their vision

3. **Track IETF TVR Working Group** — Time-Variant Routing is directly relevant to our contact graph routing work

4. **Monitor WRC-27 outcomes** — spectrum decisions for lunar communications may affect Phase 4 planning

5. **Document our security model** formally — integrity-without-confidentiality as a BPSec profile, publishable as a contribution to the SSI security architecture discussion

6. **Propose callsign-EID scheme** to IPNSIG/CCSDS as a model for amateur radio BP identifier allocation — it solves a real governance problem elegantly

7. **Consider SANA registration** for RADIANT node numbers — formalise our place in the BP identifier space

8. **Time synchronisation architecture** — document how RADIANT nodes maintain time consistency, especially as we move toward cislunar operations

---

## References

- IPNSIG, "Solar System Internet Architecture and Governance", September 2023
  https://www.ipnsig.org/
- RFC 9171 — Bundle Protocol Version 7
- RFC 5326 — Licklider Transmission Protocol
- RFC 9172 — Bundle Protocol Security (BPSec)
- RFC 4838 — Delay-Tolerant Networking Architecture
- CCSDS 730.1-G-1 — Solar System Internetwork Architecture (2014)
- CCSDS 734.1-B-1 — Licklider Transmission Protocol (space profile)
- NASA/TP-20210021073/Rev.4 — LunaNet Interoperability Specification
- NASA HDTN — https://github.com/nasa/HDTN (historical reference)
- ION-DTN — https://sourceforge.net/projects/ion-dtn/
- IETF DTN WG — https://datatracker.ietf.org/wg/dtn/about/
- IETF TVR WG — https://datatracker.ietf.org/wg/tvr/about/
- SANA — https://sanaregistry.org/
- IANA BP registries — https://www.iana.org/assignments/bundle/bundle.xhtml
