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
│   EID: dtn://callsign-ssid          │
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
