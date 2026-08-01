# AMSAT-DL Journal Article Outline
## RADIANT: Delay-Tolerant Networking für den Amateurfunk — Ein europäisches Bodensegment

**Target length:** ~1,500 words  
**Author:** David Johnson, G4DPZ  
**Publication:** AMSAT-DL Journal  
**Language:** English (with German abstract if required by editor)

---

## 1. Opening (150 words)

- RADIANT is an open-source project bringing space-grade DTN to amateur radio
- Joint effort between AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station
- Emphasis on AMSAT-DL's role: QO-100 heritage, European ground segment partner
- Built on the same protocols NASA and ESA are deploying (BPv7, LTP, CGR)
- Goal: amateur operators running DTN ground station nodes — from terrestrial links to cislunar space

---

## 2. Why DTN Matters — The QO-100 Connection (200 words)

- AMSAT-DL built the amateur transponders on Es'hail-2 (QO-100)
- QO-100 provides a real space link with ~250ms one-way delay
- DTN's store-and-forward model is designed exactly for this environment
- Phase 1.5 of RADIANT: DTN data through QO-100 narrowband transponder
- LTP (Licklider Transmission Protocol) handles the deferred acknowledgements over the space delay
- QO-100 is the ideal stepping stone between terrestrial millisecond delays and cislunar 1.3s delays
- AMSAT-DL's transponder infrastructure makes this possible

---

## 3. Current Technical Status (300 words)

- ION-DTN lifecycle management via HTTP API — deployed and operational
- Multi-implementation LTP interoperability proven (ION ↔ Hardy, 1MB bundle transfer)
- TCPCLv3 terrestrial gateway — any DTN node on the internet can connect and deliver bundles
- Store-and-forward demonstrated with scheduled contact windows
- Three timing scenarios validated: delayed uplink, immediate forwarding, delayed downlink
- Direct ION controller for ground station operations
- Backend-agnostic architecture: supports ION-DTN, Hardy, µD3TN
- All open source, MIT licensed

---

## 4. The European Ground Segment (300 words)

- RADIANT envisions a distributed amateur ground station network across Europe
- AMSAT-DL's role: hosting a European DTN node alongside the UK node
- Lars (DB[call]) contributing a UHF ground station for LEO DTN reception
- Loïc (F[call]) running a Hardy BPA node on 44net, with FOSM-1 satellite payload in orbit
- European stations provide geographic diversity for LEO pass coverage
- QO-100 enables persistent European-wide connectivity (always visible from EU)
- Contact Graph Routing computes paths across the distributed ground segment
- The FUNcube model extended: not just passive telemetry reception, but active store-and-forward participation

---

## 5. The Phased Roadmap (250 words)

- Phase 1: Terrestrial validation (9600 baud packet radio) — in progress
- Phase 1.5: QO-100 narrowband DTN — data through AMSAT-DL's geostationary transponder
- Phase 2: CubeSat engineering model (STM32U585)
- Phase 3: LEO CubeSat payload (437 MHz, 9.6 kbps)
- Phase 4: Cislunar DTN node
- Each phase validates technology needed for the next
- QO-100 phase specifically validates LTP behaviour over authentic space delay before committing to orbital hardware

---

## 6. Protocol Stack and Amateur Compliance (150 words)

- BPv7 → LTP → KISS → radio modem
- Callsigns embedded in DTN Endpoint Identifiers (dtn://callsign/service)
- No encryption — fully compliant with ITU and national regulations
- All protocols publicly documented (IETF RFCs, open source)
- Protocol definition document to be published

---

## 7. Call to Action for AMSAT-DL Members (150 words)

- QO-100 operators: ideal test platform for Phase 1.5
- UHF satellite operators: ground station participation in LEO phase
- Software developers: Rust, Go, open-source contributions welcome
- Microwave experimenters: S-band for cislunar phases
- Seeking: European ground station volunteers, CubeSat collaboration partners
- Contact: dave@g4dpz.me.uk / radiant.amsat-uk.org

---

## Suggested Figures

1. Network architecture diagram showing European ground segment (UK + DE nodes + QO-100)
2. Protocol stack diagram
3. Phased roadmap with QO-100 phase highlighted
