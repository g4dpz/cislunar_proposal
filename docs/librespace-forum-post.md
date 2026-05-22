# LibreSpace Community Forum Post

**Category:** General / Projects (or whichever category fits best on community.libre.space)

**Title:** RADIANT — Open-Source Delay-Tolerant Networking for Amateur Radio (Terrestrial to Cislunar)

---

Hi everyone,

I'd like to introduce **RADIANT** (Radio Amateur Delay-tolerant Interplanetary Networking Testbed) — an open-source project bringing Delay-Tolerant Networking to amateur radio, with the long-term goal of operating a DTN node in cislunar space.

## What is it?

RADIANT implements the Bundle Protocol v7 (RFC 9171) and Licklider Transmission Protocol over amateur radio links, built on NASA Glenn Research Center's [HDTN](https://www.nasa.gov/glenn/glenn-expertise-space-exploration/scan/high-rate-delay-tolerant-networking/) software stack. The system uses store-and-forward messaging — bundles are stored at each node and forwarded when the next contact window becomes available. Think of it as modern packet radio BBS forwarding, formalised into IETF/CCSDS standards suitable for spaceflight.

The protocol stack is deliberately simple:

```
Application (bping, bpsendfile)
BPv7 (Bundle Protocol) — EID: dtn://callsign/service
LTP (Licklider Transmission Protocol)
KISS (TNC Serial Framing)
G3RUH GFSK (9600 baud)
```

We eliminate AX.25 entirely — LTP is wrapped directly in KISS framing, saving ~15 bytes/frame overhead.

## Phased Roadmap

| Phase | Description | Status |
|-------|-------------|--------|
| **Phase 1** | Terrestrial DTN validation (RPi + TNC4 + FT-817, 9600 baud VHF/UHF) | **Active** |
| **Phase 1.5** | QO-100 GEO satellite DTN (first space-based demo via Es'hail-2) | Planned |
| **Phase 2** | CubeSat Engineering Model (STM32U585 + Ettus B200mini SDR) | Planned |
| **Phase 3** | LEO CubeSat flight (UHF 437 MHz, 9.6 kbps) | Planned |
| **Phase 4** | Cislunar deep-space (S-band 2.2 GHz, 500 bps, 3-5m dishes) | Planned |

Each phase validates technologies needed for the next. The terrestrial network deliberately mirrors the cislunar communications path: Mission Operations → Ground Gateway → Amateur RF Link → Relay Node → Payload Endpoint.

## Current Status

Phase 1 is operational with two nodes. We've demonstrated:

- Store-and-forward delivery and DTN ping over real amateur radio links
- 3-node cislunar simulation with true packet-level propagation delays (1.3s Moon, 3-12 min Mars)
- Contact Graph Routing computing multi-hop relay paths
- LTP managing RTTs up to 24 minutes
- Custom C++17 KISS convergence layer adapter plugin for HDTN
- Go-based orchestration system with property-based testing (35+ correctness properties)
- Full CI pipeline

## Why LibreSpace?

There's natural overlap between RADIANT and the LibreSpace ecosystem:

- **SatNOGS ground stations** could potentially serve as DTN ground segment nodes for the LEO CubeSat phase
- **Open-source satellite philosophy** — RADIANT is MIT-licensed, all documentation public
- **Distributed ground station network** — we're building exactly this for DTN, and SatNOGS has proven the model works
- **CubeSat community** — we're targeting a CubeSat-class payload and would welcome collaboration with teams who have flight experience

We're also interested in whether anyone in the community has experience with:
- DTN or store-and-forward protocols on spacecraft
- S-band amateur allocations and frequency coordination
- CubeSat payload hosting opportunities
- QO-100 narrowband transponder experimentation

## Links

- **Website:** https://radiant.amsat-uk.org
- **GitHub:** https://github.com/g4dpz/cislunar_proposal
- **Supported by:** AMSAT-UK, AMSAT-DL, Goonhilly Earth Station

The project is actively seeking collaboration from CubeSat teams, ground station operators, and anyone interested in space networking. All contributions welcome — from code to ground station time to flight hardware partnerships.

Happy to answer any questions about the architecture, protocol choices, or roadmap.

73,
David, G4DPZ
