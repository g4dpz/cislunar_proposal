# Extending the RADIANT Vision: From Amateur Radio DTN to Network Orchestration

The RADIANT project brings Delay-Tolerant Networking (DTN) to amateur radio — building practical experience with the protocols that will underpin future cislunar communications. Phase 1 terrestrial validation is underway using BPv7 over amateur links, with the roadmap extending through GEO satellite demonstrations, a LEO CubeSat flight, and ultimately deep-space communication.

As the architecture matures, we've been thinking about what comes next.

## Current Roadmap

The project progresses through five phases:

- **Phase 1: Terrestrial DTN Validation** (in progress) — Two-node ground network using Raspberry Pi, Mobilinkd TNC4, and Yaesu FT-817 at 9600 baud
- **Phase 1.5: QO-100 GEO Satellite DTN** — First space-based demonstration via the Es'hail-2 amateur transponder
- **Phase 2: CubeSat Engineering Model** — Flight-representative flatsat with STM32U585 OBC and SDR
- **Phase 3: LEO CubeSat Flight** — Orbital deployment demonstrating ground-to-space DTN
- **Phase 4: Cislunar Deep-Space** — Earth-Moon DTN with 3-5m dishes, seeking ESA ARTES support

## Planned Features

Alongside the phased hardware roadmap, several cross-cutting capabilities are in design (specifications written, implementation not yet started):

- **Multi-Node Contact Graph** — Time-dependent routing across ground stations, LEO, GEO, and cislunar nodes with automated plan distribution
- **Contact Log** — Versioned logging of planned vs. actual contact behaviour for cross-phase performance comparison
- **Station Identification Beacon** — Regulatory-compliant callsign beacons in BPv7 bundles every 10 minutes
- **Test Framework** — Property-based verification modelled after NASA Glenn's HDTN Test Framework methodology

## Future Vision: Three Layers of Evolution

Looking further ahead, we're exploring how RADIANT could evolve beyond a single DTN implementation:

**DTN Abstraction Layer** — A common interface supporting multiple DTN engines (HDTN, µD3TN, Hardy, ION-DTN, ESA DTN). Swap engines via configuration. Compare performance. Applications remain unchanged. µD3TN is a particularly strong candidate for flight software given its microcontroller heritage and space-tested track record. Hardy brings memory-safe Rust with `no_std` core libraries suitable for embedded targets.

**Contact Plan as a Service (CPaaS)** — A centralised service treating contact information as a shared network resource. Orbital predictions, conflict resolution, multi-format export, and OTA distribution to spacecraft — all independent of which DTN engine each node runs.

**Network Orchestrator** — The coordination layer tying it all together. Applications express high-level requirements (destination, priority, delivery confidence) and the orchestrator handles topology, routing, and resource allocation.

## Building on Solid Foundations

These ideas extend our current objectives, not replace them. The immediate work remains: validate BPv7 over amateur radio, demonstrate store-and-forward across real satellite links, build operational experience.

**Today:** Amateur Radio + DTN
**Tomorrow:** Amateur Radio + DTN + Network Orchestration

Interested in collaborating? RADIANT is an open project seeking partners from the amateur radio, space networking, and DTN research communities.

---

*A collaboration involving AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station.*

#DTN #SpaceNetworking #AmateurRadio #Cislunar #BPv7 #AMSAT #DelayTolerantNetworking
