# RADIANT — Radio Amateur Delay-tolerant Interplanetary Networking Testbed

**From amateur packet radio to CubeSat relay to cislunar networking**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8.svg)](https://golang.org/)
[![HDTN](https://img.shields.io/badge/HDTN-2.0-green.svg)](https://github.com/nasa/HDTN)

**Website**: https://radiant.amsat-uk.org

---

## Overview

RADIANT brings Delay-Tolerant Networking (DTN) to amateur radio, enabling store-and-forward messaging across disrupted links from terrestrial ground stations to Low Earth Orbit (LEO) and ultimately to cislunar space.

Built on NASA Glenn's **HDTN** (High-rate Delay Tolerant Networking), this project implements Bundle Protocol version 7 (BPv7) over amateur radio links using LTP wrapped directly in KISS framing, with callsign-embedded DTN Endpoint Identifiers for station identification.

**Supported by**: AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station

---

## What We've Accomplished

### Working DTN Simulation

We have a fully operational 3-node DTN simulation demonstrating store-and-forward relay with true packet-level propagation delays:

```
Ground Station (Earth) → Lunar Orbiter (relay) → Lunar Lander
     nodeId=10              nodeId=20              nodeId=30
```

**Demonstrated scenarios:**
- **Cislunar**: 1.3-second OWLT, 500 bps S-band (Earth-Moon)
- **Mars closest approach**: 3-minute OWLT, 32 kbps X-band
- **Mars average**: 12-minute OWLT, 32 kbps X-band
- **True packet-level delay**: Using HDTN's `udp-delay-sim` proxy for real propagation simulation

**What this proves:**
- LTP deferred acknowledgment with deep-space retransmission timers
- Contact Graph Routing (CGR) computing multi-hop paths
- Store-and-forward relay at the orbiter node
- Bundle delivery across multiple hops with different link characteristics
- LTP session management with 2.6s to 24-minute round-trip times

### HDTN Migration

Migrated from ION-DTN to NASA Glenn's HDTN:
- `pkg/hdtn/` — Lifecycle manager, telemetry collector, contact plan manager
- `pkg/hdtnconfig/` — JSON configuration generation
- `plugins/kiss-cla/` — C++17 KISS CLA plugin for amateur radio TNC interfaces
- 11 property-based tests validating correctness properties
- Full test suite passing

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

HDTN uses numeric `ipn://` addresses for internal routing (CGR requires integer node IDs). The `dtn://` EID with the callsign is stored in each node's configuration via `myDtnEidStr` and appears in bundle metadata, satisfying the regulatory requirement that every transmission carries the operator's callsign.

---

## Running the Cislunar Simulation

### Prerequisites

- HDTN 2.0 built and installed (see below)
- macOS or Linux

### Building HDTN

HDTN must be cloned, built, and symlinked into this project:

```bash
# Clone HDTN alongside this project
cd ~/dev
git clone https://github.com/nasa/HDTN.git
cd HDTN
mkdir build && cd build
cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
sudo make install

# Symlink HDTN source into the project (for config references)
cd ~/dev/cislunar_proposal
ln -s ../HDTN HDTN
```

After installation, `hdtn-one-process`, `bpsendfile`, `bpreceivefile`, and `udp-delay-sim` should be in your PATH.

### Quick Start

```bash
# Start the 3-node simulation
./scripts/run-cislunar-sim.sh

# In another terminal, send a file from Earth to the Lander
rm -f /tmp/hdtn-send/*
echo 'Hello Moon' > /tmp/hdtn-send/lunar.dat
bpsendfile --my-uri-eid=ipn:1.1 --dest-uri-eid=ipn:3.1 \
  --use-bp-version-7 \
  --outducts-config-file=configs/simulation/bping-outducts.json \
  --file-or-folder-path=/tmp/hdtn-send

# Check received files (arrives after ~1.3s propagation delay)
ls cislunar_received/
```

### Adjusting Parameters

Edit `configs/simulation/` JSON files to change:
- `oneWayLightTimeMs` — propagation delay (1300 for Moon, 180000 for Mars)
- `rateBitsPerSec` — link data rate in contact plan
- `DELAY_MS` in run script — true packet delay via proxy

---

## Five-Phase Roadmap

| Phase | Status | Link | Description |
|-------|--------|------|-------------|
| **Phase 1: Terrestrial** | 🔄 In Progress | VHF/UHF 9600 baud | Ground validation with TNC4 + FT-817 |
| **Phase 1.5: QO-100** | 📋 Planned | 2.4/10 GHz | GEO satellite DTN via Es'hail-2 |
| **Phase 2: EM** | 📋 Planned | UHF/S-band | CubeSat flatsat with STM32U585 + B200mini |
| **Phase 3: LEO** | 📋 Planned | UHF 437 MHz | Orbital CubeSat flight |
| **Phase 4: Cislunar** | 📋 Planned | S-band 500 bps | Earth-Moon DTN (seeking ESA ARTES support) |

---

## Project Structure

```
├── cmd/dtn-node/           # Go orchestrator for HDTN
├── pkg/
│   ├── hdtn/               # HDTN lifecycle, telemetry, contact plan
│   ├── hdtnconfig/         # HDTN JSON config generation
│   ├── contact/            # Contact plan manager + CGR
│   ├── cla/                # Convergence layer abstractions
│   └── ...                 # IQ, link budget, store, security
├── kiss/                   # KISS framing (Go reference implementation)
├── plugins/kiss-cla/       # C++17 KISS CLA plugin for HDTN
├── configs/
│   ├── simulation/         # 3-node cislunar simulation configs
│   ├── dtn-node-a.yaml    # Orchestrator config (node A)
│   └── dtn-node-b.yaml    # Orchestrator config (node B)
├── scripts/
│   ├── run-cislunar-sim.sh # Launch 3-node simulation
│   ├── build-hdtn.sh      # Build HDTN with KISS CLA
│   └── start-node-*.sh    # Node startup scripts
├── docs/                   # Phase-specific specs and design docs
├── website/                # Project website (radiant.amsat-uk.org)
└── deploy/                 # Website deployment configs
```

---

## Testing

```bash
# Run all Go tests (includes property-based tests)
go test ./pkg/hdtn/... ./pkg/hdtnconfig/... ./kiss/... ./cmd/dtn-node/...

# Run smoke tests
go test ./test/integration/ -run TestSmoke

# Build verification
go build ./...
```

---

## Key Technologies

- **HDTN** — NASA Glenn's High-rate Delay Tolerant Networking (C++17)
- **BPv7** — Bundle Protocol version 7 (RFC 9171)
- **LTP** — Licklider Transmission Protocol (RFC 5326)
- **CGR** — Contact Graph Routing for scheduled contacts
- **KISS** — TNC serial framing protocol
- **dtn:// EIDs** — Callsign-embedded endpoint identifiers (e.g., dtn://g4dpz-1)

---

## Future Enhancements

### Contact Log — Versioned Contact-Plan and Run-Evidence Logging

A structured logging system that captures both planned (expected) and actual (observed) contact behavior for every DTN session. Designed to enable cross-phase comparison of DTN performance across all mission phases:

- **Versioned log entries** — Each session produces an immutable, schema-versioned JSON record containing the contact plan snapshot, phase metadata, and run evidence
- **Cross-phase comparison** — Normalized metrics (goodput, plan adherence, delivery success ratio) allow direct comparison across terrestrial, QO-100, LEO, and cislunar links
- **Planned vs. actual** — Captures expected contact window parameters alongside observed timing, throughput, and delivery outcomes
- **Phase-aware metadata** — Records link type, frequency band, OWLT, orbital parameters, and modulation for each session's environment
- **Integrates with existing systems** — Pulls contact plan state from the Contact Plan Manager and telemetry from the HDTN REST API automatically
- **Machine-readable and human-readable** — JSON with consistent field ordering; queryable by phase, time range, node pair, and outcome

See [`.kiro/specs/contact-log/requirements.md`](.kiro/specs/contact-log/requirements.md) for the full requirements specification.

### Station Identification Beacon — Amateur Radio Compliance

Periodic transmission of BPv7 bundles containing the operator's callsign and station metadata in plaintext, ensuring compliance with amateur radio regulations requiring station identification at least every 10 minutes:

- **Regulatory compliance** — Transmits callsign in plaintext so any third party demodulating the signal can identify the station, even when wire format only carries opaque numeric ipn:// EIDs
- **Analogous to FT8/WSPR** — Embeds callsign in message payloads using a well-known beacon service number (2048)
- **Independent of data traffic** — Operates on a configurable timer (default 10 minutes) via existing HDTN infrastructure
- **Includes station metadata** — Callsign, Maidenhead grid square, and node type in human-readable payload
- **Cross-phase** — Required for all phases from terrestrial through cislunar

See [`.kiro/specs/station-identification-beacon/requirements.md`](.kiro/specs/station-identification-beacon/requirements.md) for the full requirements specification.

### Test Framework — Requirements-Based Verification (NASA TM Methodology)

A property-based test framework modeled after NASA Glenn's HDTN Test Framework (TM-20240014467 / LEW-20818-1), providing automated verification across all mission phases:

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
- **HDTN-compatible export** — Local node view extracted in NASA HDTN JSON format (source, dest, startTime, endTime, rateBitsPerSec, owlt)

See [`.kiro/specs/multi-node-contact-graph/requirements.md`](.kiro/specs/multi-node-contact-graph/requirements.md) for the full requirements specification.

### Network Orchestrator — Coordinated DTN Operations Platform

The highest-level coordination component in the RADIANT architecture, sitting above the DTN Abstraction Layer and Contact Plan as a Service. Transforms RADIANT from a collection of independent DTN nodes into a platform for coordinating DTN operations, drawing from the GSaaS model where applications express high-level requirements and the infrastructure determines fulfillment:

- **Network topology discovery** — Automatically maintains a graph of all DTN nodes and links derived from CPaaS contacts, with real-time link state tracking (Active, Scheduled, Degraded, Unavailable) and reachability computation
- **High-level delivery API** — Applications submit delivery requests with destination, priority, and QoS requirements (latency, confidence, hop count); the Orchestrator computes optimal paths and handles routing
- **Path computation** — Contact Graph Routing (CGR) over time-varying topology, selecting paths that minimise delivery time while respecting capacity and QoS constraints, with alternative path support for expedited traffic
- **Policy engine** — Service classes (Expedited, Standard, Bulk) with configurable bandwidth allocations and rule-based traffic classification for automatic QoS assignment
- **Security and trust** — Ed25519 node authentication, trust levels (Trusted, Provisional, Untrusted, Revoked) affecting relay eligibility, and certificate management
- **Monitoring and telemetry** — Collects per-node and per-link performance metrics with rolling averages, delivery statistics, network utilisation tracking, and real-time visualisation data via WebSocket
- **Integration** — Consumes contact plans from CPaaS (subscriptions, confidence levels, plan versions), executes routing via DTN Abstraction Layer Engine interface (SendBundle, Health, OnStateChange)
- **Cislunar relevance** — Provides operational concepts applicable to future lunar and deep-space communication architectures

See [`.kiro/specs/radiant-network-orchestrator/requirements.md`](.kiro/specs/radiant-network-orchestrator/requirements.md) for the full requirements specification.

---

## Documentation

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
- **Source**: https://github.com/g4dpz/cislunar_proposal
