# AMSAT Journal (AMSAT-NA) Article Outline
## RADIANT: Bringing NASA's DTN Protocols to Amateur Radio Ground Stations

**Target length:** ~2,000 words  
**Author:** David Johnson, G4DPZ  
**Publication:** The AMSAT Journal (AMSAT-NA)

---

## 1. Opening — The NASA Connection (200 words)

- JPL's ION-DTN has flight heritage on multiple deep-space missions
- The Korean Pathfinder Lunar Orbiter (KPLO) tested Bundle Protocol at the Moon in 2024
- DTN is running on the International Space Station today
- These are not experimental — they are operational protocols in space
- RADIANT implements the same standards (BPv7 RFC 9171, LTP RFC 5326) for amateur radio
- The question: if agencies are building a Solar System Internet, can amateurs participate?

---

## 2. The Solar System Internet Vision (250 words)

- IPNSIG (Internet Society, co-chaired by Vint Cerf) published the Solar System Internet governance report in 2023
- Key quote: "DTN and the BP suite should be the basis for all communications that might have to traverse paths that the IP protocol suite cannot support"
- The 1983 analogy: three disparate networks (Arpanet, SATNET, PRNET) joined via TCP/IP — space networking will follow the same trajectory
- LunaNet Interoperability Specification mandates the same protocols RADIANT implements
- Multistakeholder governance: the report explicitly calls for non-agency participants
- Amateur radio operators are a legitimate stakeholder — but only with demonstrated operational competence
- RADIANT is how the amateur community earns its seat at the table

---

## 3. What RADIANT Does (300 words)

- Open-source ground station software managing DTN node lifecycle
- HTTP/JSON API for remote monitoring and control
- Backend-agnostic: currently supports ION-DTN (JPL), architecture allows Hardy, µD3TN
- TCPCLv3 terrestrial gateway — any DTN node can connect via standard TCP
- LTP convergence layer for space links (proven interop with Hardy at 1MB)
- Store-and-forward with Contact Graph Routing
- Three demonstrated scenarios: delayed uplink (bundle storage during link outage), immediate forwarding, delayed downlink
- Direct ION-DTN controller for operations without abstraction layer overhead
- All code MIT licensed, open source on GitHub

---

## 4. Interoperability with NASA Systems (300 words)

- RADIANT uses ION-DTN — the same DTN engine JPL developed for deep-space missions
- LTP interoperability proven between ION and Hardy (independent Rust BPv7 implementation)
- TCPCLv3 (RFC 7242) for terrestrial connectivity — same protocol used for ground links
- Contact Graph Routing implementations follow the same algorithms as NASA's CGR
- EID addressing uses the IPN scheme (ipn:node.service) — compatible with NASA's addressing conventions
- LunaNet compliance: RADIANT could interoperate with NASA's lunar infrastructure given appropriate coordination
- The architecture deliberately maintains compatibility — not building a parallel incompatible network

---

## 5. The Amateur Advantage (200 words)

- Amateurs pioneered packet networking before the commercial Internet
- OSCAR 1 launched in 1961 — decades before commercial constellations
- Store-and-forward BBS systems in the 1980s were early DTN
- Amateur ground stations provide global geographic diversity at near-zero marginal cost
- 24/7 pass coverage requires many stations — agencies have few, amateurs have many
- Callsign-based addressing (dtn://callsign/service) provides globally unique identifiers without a central authority
- Amateur regulations require transparency: published protocols, no encryption — aligns perfectly with open networking standards

---

## 6. The Phased Roadmap (250 words)

- Phase 1: Terrestrial (9600 baud, Raspberry Pi + TNC) — validated
- Phase 1.5: QO-100 GEO satellite (250ms delay, always visible) — planned
- Phase 2: CubeSat engineering model (STM32U585, flight-representative hardware) — planned
- Phase 3: LEO CubeSat (437 MHz, 9.6 kbps, 31 dB link margin) — amateur ground stations worldwide
- Phase 4: Cislunar (S-band, 500 bps, 7 dB margin, 3-5m ground dish) — the destination
- Each phase validates protocols, operations, and hardware for the next
- The ground segment scales with each phase: from packet radio operators to EME-class stations

---

## 7. How AMSAT-NA Members Can Participate (200 words)

- Any 9600-baud packet station can participate in Phase 1 testing
- QO-100 operators (limited from NA, but possible via remote stations)
- LEO satellite operators: same equipment as current amateur satellite work
- Software developers: Rust, Go contributions welcome
- EME operators: cislunar phases need the same ground station capabilities
- Ground station network: geographic diversity across NA complements European stations
- AMSAT-NA's LEO fleet experience directly relevant to Phase 3 CubeSat
- Potential collaboration on CubeSat hosted payload (AMSAT engineering resources)
- Contact: dave@g4dpz.me.uk / radiant.amsat-uk.org

---

## 8. Partners and Status (150 words)

- AMSAT-UK (lead), AMSAT-DL, Goonhilly Earth Station
- Collaborators: Lars (UHF ground station, thesis topics), Loïc (FOSM-1 satellite, Hardy developer)
- Production VM deployed and operational
- Seeking NA partners: ground station operators, CubeSat teams, universities
- Code: github.com/g4dpz/cislunar_proposal
- Website: radiant.amsat-uk.org

---

## Suggested Figures

1. Architecture diagram showing ION-DTN integration with NASA protocol heritage
2. Link budget comparison table (LEO vs Cislunar) — showing feasibility with amateur equipment
3. Phased roadmap timeline
4. Screenshot of radiant-ion HTTP API (health check, bundle stats)

---

## Notes for Editor

- Angle for NA audience: NASA connection, interoperability with agency systems, AMSAT-NA participation path
- Technical depth appropriate for AMSAT Journal readers (licensed operators, many with engineering background)
- Emphasise that this uses the same protocols as ISS/KPLO — not a parallel amateur-only invention
- The "seat at the table" argument resonates with AMSAT-NA's history of agency collaboration (ARISS, P3D, Fox series)
