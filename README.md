# RADIANT — Radio Amateur Delay-tolerant Interplanetary Networking Testbed

**From amateur packet radio to CubeSat relay to cislunar networking**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Rust](https://img.shields.io/badge/rust-1.78+-orange.svg)](https://www.rust-lang.org/)

**Website**: https://radiant.amsat-uk.org

---

## Overview

RADIANT brings Delay-Tolerant Networking (DTN) to amateur radio, enabling store-and-forward messaging across disrupted links from terrestrial ground stations to Low Earth Orbit (LEO) and ultimately to cislunar space.

The project implements Bundle Protocol version 7 (BPv7) over amateur radio links using LTP wrapped directly in KISS framing, with callsign-embedded DTN Endpoint Identifiers for station identification. The architecture is **DTN-implementation-agnostic** — the abstraction layer supports multiple DTN engines (JPL's ION-DTN, µD3TN, and Hardy) through a common interface, allowing operators to select the engine best suited to their platform and mission phase.

**Supported by**: AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station

---

## What We've Accomplished

### Working DTN Simulation (Lab Environment)

We have a fully operational 3-node DTN simulation demonstrating store-and-forward relay with **simulated** propagation delays (injected via UDP delay proxy over localhost — not real space links):

```
Ground Station (Earth) → Lunar Orbiter (relay) → Lunar Lander
     nodeId=10              nodeId=20              nodeId=30
```

**Simulated scenarios (artificial delay injection, not RF paths):**
- **Cislunar timing**: 1.3-second OWLT injected, 500 bps configured
- **Mars closest approach timing**: 3-minute OWLT injected, 32 kbps configured
- **Mars average timing**: 12-minute OWLT injected, 32 kbps configured

**What this validates:**
- LTP deferred acknowledgment works correctly with deep-space retransmission timers
- Contact Graph Routing (CGR) computes multi-hop paths as expected
- Store-and-forward relay at the orbiter node functions correctly
- Bundle delivery across multiple hops with different configured link characteristics
- LTP session management handles 2.6s to 24-minute round-trip times gracefully

> **Note:** These are software simulations with injected delay, running on a single machine or LAN. They validate the DTN protocol behaviour under representative timing conditions, but do not involve actual RF propagation or space hardware.

### DTN Engine Abstraction

The architecture supports multiple DTN engines through a common interface:
- `crates/radiant-kiss/` — KISS framing library (`no_std` compatible for flight hardware)
- `crates/radiant-cla/` — Convergence layer abstraction trait
- `crates/radiant-contact/` — Contact plan manager + CGR engine
- Property-based tests validating correctness properties
- Full test suite passing

### Cross-Engine Interoperability (ION-DTN ↔ Hardy)

We have demonstrated live BPv7 bundle delivery between two independent DTN implementations over LTP/UDP — proving the abstraction layer can drive heterogeneous DTN nodes from a single canonical configuration:

```
ION-DTN (node 10)  ←─ LTP/UDP ─→  Hardy BPA (node 20)
   ipn:10.1                            ipn:20.1
   udplso :2113                        hardy-ltp-cla :1113
```

**What's working:**
- **ION→Hardy delivery** — 1MB bundle sent via ION's `bpsendfile`, transported over LTP/UDP, received by Hardy's LTP CLA and dispatched to the BPA
- **Hardy→ION delivery** — Bundles injected via Hardy's gRPC Application service, exported by the LTP CLA, received and delivered by ION's `bprecvfile`
- **Abstraction layer config generation** — Both engines configured entirely from a shared canonical YAML model (no hand-written ION admin scripts)
- **Multiple payload sizes verified** — 1KB, 20KB, 100KB, 1MB

**The `radiant-dtn-abstraction` crate** (`radiant-dtn-abstraction/`) provides:
- Vendor-neutral canonical data model (nodes, neighbors, contacts, routing)
- Backend adapters for ION-DTN and Hardy (config gen, lifecycle, hot-reconfig, telemetry)
- Automatic generation of ION admin scripts (`.ionrc`, `.bprc`, `.ltprc`, `.ipnrc`) including loopback entries
- Hardy YAML + LTP CLA config generation
- HTTP/JSON management API (axum) with SSE event streaming
- Engine lifecycle state machine and event bus
- 232 tests (unit, property-based, integration) all passing
- Amateur radio compliant throughout (no BPSec, no encryption over amateur links)

### Protocol Stack

```
┌─────────────────────────────────────┐
│   Application (bpsendfile, bping)   │
├─────────────────────────────────────┤
│   BPv7 (Bundle Protocol)            │
│   EID: dtn://callsign/service       │
├─────────────────────────────────────┤
│   LTP (Licklider Transmission)      │
├─────────────────────────────────────┤
│   KISS (TNC Serial Framing)         │
├─────────────────────────────────────┤
│   USB Serial (TNC4)                 │
├─────────────────────────────────────┤
│   G3RUH GFSK (9600 baud)            │
└─────────────────────────────────────┘
```

### Station Identification (Amateur Radio Compliance)

Amateur radio regulations require station identification in every transmission. RADIANT achieves this by embedding callsigns in DTN Endpoint Identifiers carried in every bundle's metadata:

| Node | ipn:// (routing) | dtn:// (callsign metadata) |
|------|-----------------|---------------------------|
| Ground Station | ipn:10.* | dtn://g4dpz/gs |
| Spacecraft | ipn:20.* | dtn://g4dpz/spacecraft |
| Lander | ipn:30.* | dtn://g4dpz/lander |

DTN engines use numeric `ipn://` addresses for internal routing (CGR requires integer node IDs). The `dtn://` EID with the callsign appears in bundle metadata, satisfying the regulatory requirement that every transmission carries the operator's callsign.

---

## Building

The project uses a DTN-engine-agnostic architecture. The orchestrator, KISS framing, and contact plan management are independent of the underlying DTN engine. Specific engine integration is configured at deployment time.

### Prerequisites

- Rust 1.78+ (stable toolchain)
- macOS or Linux
- A DTN engine installed (ION-DTN, µD3TN, or Hardy)

### Quick Start

```bash
# Build the project
cargo build --workspace

# Run the orchestrator
cargo run -- --config configs/dtn-node-a.yaml
```

---

## Five-Phase Roadmap

| Phase | Status | Link | Description |
|-------|--------|------|-------------|
| **Phase 1: Terrestrial** | 🔄 In Progress | VHF/UHF 9600 baud | Ground validation with TNC4 + FT-817 |
| **Phase 1.5: QO-100** | 📋 Planned | 2.4/10 GHz | GEO satellite DTN via Es'hail-2 |
| **Phase 2: EM** | 📋 Planned | UHF/S-band | CubeSat flatsat with STM32U585 + SDR |
| **Phase 3: LEO** | 📋 Planned | UHF 437 MHz | Orbital CubeSat flight |
| **Phase 4: Cislunar** | 📋 Planned | S-band 500 bps | Earth-Moon DTN (seeking ESA ARTES support) |

---

## Project Structure

```
├── Cargo.toml              # Workspace root
├── crates/
│   ├── radiant-kiss/       # KISS framing (no_std compatible)
│   ├── radiant-cla/        # Convergence layer abstractions
│   ├── radiant-contact/    # Contact plan manager + CGR
│   └── radiant-ffi/        # C-ABI exports for engine plugins
├── radiant-dtn-abstraction/ # DTN abstraction layer (Rust crate)
│   ├── src/                # Core library (model, adapters, API)
│   ├── examples/           # ION↔Hardy interop examples
│   └── tests/              # Property + integration tests
├── src/
│   └── main.rs             # dtn-node orchestrator binary
├── configs/
│   ├── simulation/         # 3-node cislunar simulation configs
│   ├── dtn-node-a.yaml     # Orchestrator config (node A)
│   └── dtn-node-b.yaml     # Orchestrator config (node B)
├── scripts/
│   ├── run-cislunar-sim.sh  # Launch 3-node simulation
│   └── start-node-*.sh     # Node startup scripts
├── docs/                    # Phase-specific specs and design docs
├── website/                 # Project website (radiant.amsat-uk.org)
└── deploy/                  # Website deployment configs
```

---

## Testing

```bash
# Run all tests (includes property-based tests)
cargo test --workspace

# Run clippy for static analysis
cargo clippy --workspace -- -D warnings

# Build verification
cargo build --workspace

# Verify no_std KISS crate compiles for embedded target
cargo build -p radiant-kiss --no-default-features --target thumbv7em-none-eabihf
```

---

## Key Technologies

### Supported DTN Engines (implementation-agnostic architecture)

- **ION-DTN** — JPL's Interplanetary Overlay Network — *reference implementation with flight heritage*
- **µD3TN** — Lightweight, space-tested DTN implementation for microcontrollers and POSIX — *candidate flight software*
- **Hardy** — Modular Rust BPv7 implementation with `no_std` core libraries — *candidate flight software*

### Protocols and Standards

- **BPv7** — Bundle Protocol version 7 (RFC 9171)
- **LTP** — Licklider Transmission Protocol (RFC 5326)
- **CGR** — Contact Graph Routing for scheduled contacts
- **KISS** — TNC serial framing protocol
- **dtn:// EIDs** — Callsign-embedded endpoint identifiers (e.g., dtn://g4dpz-1)

---

## Future Enhancements (Designed, Not Yet Implemented)

The following features have detailed specifications but **no implementation yet**. They represent the design direction for future development phases.

### Contact Log — Versioned Contact-Plan and Run-Evidence Logging

A structured logging system that captures both planned (expected) and actual (observed) contact behavior for every DTN session. Designed to enable cross-phase comparison of DTN performance across all mission phases:

- **Versioned log entries** — Each session produces an immutable, schema-versioned JSON record containing the contact plan snapshot, phase metadata, and run evidence
- **Cross-phase comparison** — Normalized metrics (goodput, plan adherence, delivery success ratio) allow direct comparison across terrestrial, QO-100, LEO, and cislunar links
- **Planned vs. actual** — Captures expected contact window parameters alongside observed timing, throughput, and delivery outcomes
- **Phase-aware metadata** — Records link type, frequency band, OWLT, orbital parameters, and modulation for each session's environment
- **Integrates with existing systems** — Pulls contact plan state from the Contact Plan Manager and telemetry from the DTN engine REST API automatically
- **Machine-readable and human-readable** — JSON with consistent field ordering; queryable by phase, time range, node pair, and outcome

See [`.kiro/specs/contact-log/requirements.md`](.kiro/specs/contact-log/requirements.md) for the full requirements specification.

### Station Identification Beacon — Amateur Radio Compliance

Periodic transmission of BPv7 bundles containing the operator's callsign and station metadata in plaintext, ensuring compliance with amateur radio regulations requiring station identification at least every 10 minutes:

- **Regulatory compliance** — Transmits callsign in plaintext so any third party demodulating the signal can identify the station, even when wire format only carries opaque numeric ipn:// EIDs
- **Analogous to FT8/WSPR** — Embeds callsign in message payloads using a well-known beacon service number (2048)
- **Independent of data traffic** — Operates on a configurable timer (default 10 minutes) via existing DTN infrastructure
- **Includes station metadata** — Callsign, Maidenhead grid square, and node type in human-readable payload
- **Cross-phase** — Required for all phases from terrestrial through cislunar

See [`.kiro/specs/station-identification-beacon/requirements.md`](.kiro/specs/station-identification-beacon/requirements.md) for the full requirements specification.

### Test Framework — Requirements-Based Verification (NASA TM Methodology)

A property-based test framework modeled after NASA Glenn's published Test Framework methodology (TM-20240014467 / LEW-20818-1), providing automated verification across all mission phases:

- **Property-based testing** — Verifies correctness properties hold for all inputs within defined domains using randomized generation (gopter/rapid)
- **Requirements traceability** — Each property test traces to one or more system requirements, enabling requirements-based verification for flight proposals
- **Cross-phase coverage** — Validates the full protocol stack (BPv7 → LTP → KISS → G3RUH) across terrestrial, QO-100, EM, LEO, and cislunar configurations
- **NASA methodology** — Follows NASA Glenn Research Center's published test framework approach
- **CI integration** — Automated test execution in the continuous integration pipeline
- **Supports flight proposals** — Provides verification evidence suitable for regulatory submissions and ESA ARTES proposals

See [`.kiro/specs/test-framework-srs-sdd/requirements.md`](.kiro/specs/test-framework-srs-sdd/requirements.md) for the full requirements specification.

### Multi-Node Contact Graph — Distributed Routing and Plan Distribution

A multi-node contact graph generator with time-dependent routing, enabling distributed ground station networks to route bundles through multiple intermediate nodes (ground relays, LEO satellites, GEO transponders, cislunar payloads) using store-and-forward semantics:

- **Pairwise contact generation** — Automatically computes contact windows between all reachable node pairs: terrestrial (always-on), GEO relay (QO-100 footprint), LEO passes (orbital prediction), and cislunar links
- **Time-dependent Dijkstra routing** — Computes optimal multi-hop paths minimizing end-to-end delivery time while respecting storage constraints at relay nodes and link capacities
- **Four relay scenarios** — Ground relay (route via station with earlier pass), satellite relay (LEO carries bundles between coverage areas), GEO backbone (QO-100 connects stations), multi-station coverage (download to best pass)
- **Storage-constrained routing** — Respects buffer limits at relay nodes (critical for CubeSat NVM); priority preemption for expedited bundles
- **REST API distribution** — Ground stations receive per-node contact plans via `GET /api/contact-plan/{nodeID}` with conditional responses and webhook push notifications
- **OTA distribution** — Space nodes receive plan updates as administrative DTN bundles (≤5KB, expedited priority) during contact windows
- **Bootstrap plans** — Pre-loaded from initial TLE before launch; spacecraft converges to operational plan on first OTA update
- **Plan versioning** — Monotonically increasing versions with latest-wins conflict resolution across multiple uploading ground stations
- **Engine-compatible export** — Local node view extracted in standard JSON format (source, dest, startTime, endTime, rateBitsPerSec, owlt)

See [`.kiro/specs/multi-node-contact-graph/requirements.md`](.kiro/specs/multi-node-contact-graph/requirements.md) for the full requirements specification.

### Network Orchestrator — Coordinated DTN Operations Platform

The highest-level coordination component in the RADIANT architecture, sitting above the DTN Abstraction Layer and Contact Plan as a Service. Transforms RADIANT from a collection of independent DTN nodes into a platform for coordinating DTN operations, drawing from the GSaaS model where applications express high-level requirements and the infrastructure determines fulfillment:

- **Network topology discovery** — Automatically maintains a graph of all DTN nodes and links derived from CPaaS contacts, with real-time link state tracking (Active, Scheduled, Degraded, Unavailable) and reachability computation
- **High-level delivery API** — Applications submit delivery requests with destination, priority, and QoS requirements (latency, confidence, hop count); the Orchestrator computes optimal paths and handles routing
- **Path computation** — Contact Graph Routing (CGR) over time-varying topology, selecting paths that minimise delivery time while respecting capacity and QoS constraints, with alternative path support for expedited traffic
- **Policy engine** — Service classes (Expedited, Standard, Bulk) with configurable bandwidth allocations and rule-based traffic classification for automatic QoS assignment
- **Node trust management** — Trust levels (Trusted, Provisional, Untrusted, Revoked) affecting relay eligibility, based on operator-assigned credentials and network behaviour observation
- **Monitoring and telemetry** — Collects per-node and per-link performance metrics with rolling averages, delivery statistics, network utilisation tracking, and real-time visualisation data via WebSocket
- **Integration** — Consumes contact plans from CPaaS (subscriptions, confidence levels, plan versions), executes routing via DTN Abstraction Layer Engine interface (SendBundle, Health, OnStateChange)
- **Cislunar relevance** — Provides operational concepts applicable to future lunar and deep-space communication architectures

See [`.kiro/specs/radiant-network-orchestrator/requirements.md`](.kiro/specs/radiant-network-orchestrator/requirements.md) for the full requirements specification.

---

## Documentation

- [DTN Abstraction Layer — Interoperability](radiant-dtn-abstraction/docs/INTEROP-ABSTRACTION-LAYER.md)
- [LTP-over-KISS Architecture](docs/LTP-KISS-ARCHITECTURE.md)
- [DTN Callsign EID Configuration](docs/DTN-CALLSIGN-EID-CONFIGURATION.md)
- [Phase 1: Terrestrial DTN](docs/terrestrial-dtn-phase1/)
- [Phase 1.5: QO-100 GEO Satellite](docs/qo-100-geo-satellite-dtn/)
- [Phase 2: CubeSat EM](docs/cubesat-em-phase2/)
- [Phase 3: LEO CubeSat](docs/leo-cubesat-phase3/)
- [Phase 4: Cislunar](docs/cislunar-phase4/)

---

## License

MIT — see [LICENSE](LICENSE)

---

## Contact

- **Website**: https://radiant.amsat-uk.org
- **Email**: dave@g4dpz.me.uk
- **Callsign**: G4DPZ
- **Source**: Private repository (not yet public — contribution agreements pending)
