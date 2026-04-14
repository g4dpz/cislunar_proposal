# Cislunar Amateur DTN Payload

**Delay-Tolerant Networking for Amateur Radio from Earth to the Moon**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.19+-00ADD8.svg)](https://golang.org/)
[![ION-DTN](https://img.shields.io/badge/ION--DTN-4.1.2-green.svg)](https://sourceforge.net/projects/ion-dtn/)

---

## 🌍 → 🌙 Overview

The Cislunar Amateur DTN Payload project brings Delay-Tolerant Networking (DTN) to amateur radio, enabling store-and-forward messaging across disrupted links from terrestrial ground stations to Low Earth Orbit (LEO) and ultimately to cislunar space.

Built on NASA JPL's **ION-DTN** (Interplanetary Overlay Network), this project implements the Bundle Protocol version 7 (BPv7) over amateur radio links, supporting two core operations:

- **Ping**: DTN reachability test (echo request/response)
- **Store-and-Forward**: Reliable message delivery across disrupted links

**Key Features:**
- 📡 AX.25 link-layer framing with callsign addressing (amateur radio compliance)
- 🔐 BPSec integrity protection (HMAC-SHA256, no encryption per regulations)
- 🛰️ Automated orbital pass prediction using Contact Graph Routing (CGR)
- 🔄 Priority-based bundle handling (critical, expedited, normal, bulk)
- 💾 Persistent bundle storage surviving power cycles
- 📊 Real-time telemetry and health monitoring

---

## 🚀 Four-Phase Roadmap

### Phase 1: Terrestrial DTN Validation ✅ **COMPLETE**
**Ground-based validation using commercial amateur radio equipment**

- **Hardware**: Raspberry Pi + Mobilinkd TNC4 + Yaesu FT-817
- **Link**: VHF/UHF at 9600 baud (G3RUH GFSK)
- **Status**: Two-node terrestrial network operational
- **Validated**: Ping, store-and-forward, BPSec integrity, telemetry

[📖 Phase 1 Documentation](docs/terrestrial-dtn-phase1/) | [🎯 Quick Start](#quick-start-phase-1)

---

### Phase 2: CubeSat Engineering Model (EM) 🔧 **IN PROGRESS**
**Ground-based flatsat with flight-representative hardware**

- **Hardware**: STM32U585 OBC + Ettus B200mini SDR + External NVM
- **Link**: UHF 437 MHz / S-band 2.2 GHz
- **Purpose**: Validate flight software, power budget, thermal/vacuum readiness
- **Key**: Identical software stack to flight unit, lab-grade RF front-end

[📖 Phase 2 Documentation](docs/cubesat-em-phase2/)

---

### Phase 3: LEO CubeSat Flight 🛰️ **PLANNED**
**Orbital deployment demonstrating ground-to-space DTN**

- **Hardware**: STM32U585 OBC + Flight IQ transceiver
- **Link**: UHF 437 MHz at 9.6 kbps
- **Operations**: Ground-to-space ping and store-and-forward
- **Community**: Handheld/small Yagi reception for broad participation

[📖 Phase 3 Documentation](docs/leo-cubesat-phase3/)

---

### Phase 4: Cislunar Deep-Space Communication 🌙 **PLANNED**
**Amateur participation in Earth-Moon DTN**

- **Hardware**: STM32U585 or more capable processor
- **Link**: S-band/X-band with LDPC/Turbo coding
- **Range**: Earth-Moon distance (~384,400 km)
- **Infrastructure**: 3-5m dishes for 500 bps cislunar links

[📖 Phase 4 Documentation](docs/cislunar-phase4/)

---

## 🎯 Quick Start (Phase 1)

### Prerequisites

- Linux or macOS (amd64/arm64)
- Go 1.19 or later
- Two Mobilinkd TNC4 terminal node controllers (USB)
- Two Yaesu FT-817 radios configured for 9600 baud
- Amateur radio license (required for transmission)

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/cislunar_proposal.git
cd cislunar_proposal

# Build ION-DTN from source
cd ion-open-source-4.1.2
./configure --prefix=$(pwd)/../ion-install
make && make install
cd ..

# Build dtn-node CLI
go build -o dtn-node ./cmd/dtn-node
```

### Run Two-Node Network

**Terminal 1 (Node A):**
```bash
./dtn-node -config configs/dtn-node-a.yaml
```

**Terminal 2 (Node B):**
```bash
./dtn-node -config configs/dtn-node-b.yaml
```

### Test Connectivity

```bash
# Add ION-DTN binaries to PATH
export PATH=$PATH:$(pwd)/ion-install/bin

# Ping Node B from Node A
bping ipn:1.1 ipn:2.1 -c 5

# Send file from Node A to Node B
bpsendfile ipn:1.1 ipn:2.1 test-message.txt
```

**Expected output:**
```
64 bytes from ipn:2.1: seq=1 time=1234.5 ms
64 bytes from ipn:2.1: seq=2 time=1198.2 ms
...
5 packets transmitted, 5 received, 0% packet loss
```

[📖 Full Setup Guide](docs/terrestrial-dtn-phase1/README.md)

---

## 🏗️ Architecture

### Protocol Stack

```
┌─────────────────────────────────────┐
│   Application (bping, bpsendfile)   │
├─────────────────────────────────────┤
│   BPv7 (Bundle Protocol)            │
├─────────────────────────────────────┤
│   BPSec (Integrity - HMAC-SHA-256)  │
├─────────────────────────────────────┤
│   LTP (Licklider Transmission)      │
├─────────────────────────────────────┤
│   AX.25 (Amateur Radio Link Layer)  │
├─────────────────────────────────────┤
│   KISS (TNC Serial Protocol)        │
├─────────────────────────────────────┤
│   USB Serial (TNC4)                 │
├─────────────────────────────────────┤
│   G3RUH GFSK (9600 baud)            │
└─────────────────────────────────────┘
```

### System Components

- **Node Controller**: Go orchestrator managing node lifecycle
- **Bundle Protocol Agent**: ION-DTN BPA for bundle creation/validation
- **Bundle Store**: Persistent filesystem storage for bundles
- **Contact Plan Manager**: Manual scheduling of communication windows (Phase 1) / CGR-based orbital pass prediction (Phases 2-4)
- **Convergence Layer Adapter**: AX.25/KISS interface to TNC4 (Phase 1) / IQ baseband interface (Phases 2-4)
- **Telemetry Collector**: Health monitoring and statistics

---

## 📦 Project Structure

```
.
├── cmd/
│   ├── dtn-node/               # Phase 1: Terrestrial DTN CLI
│   ├── em-node/                # Phase 2: Engineering Model CLI
│   ├── leo-node/               # Phase 3: LEO CubeSat CLI
│   └── cislunar-node/          # Phase 4: Cislunar CLI
├── pkg/
│   ├── ion/                    # ION-DTN Go wrapper
│   ├── contact/                # Contact plan manager + CGR
│   ├── security/               # BPSec + rate limiting
│   ├── iq/                     # IQ baseband processing
│   ├── linkbudget/             # Link budget calculations
│   └── store/                  # Bundle storage
├── configs/
│   ├── node-a/                 # Node A ION-DTN configs
│   ├── node-b/                 # Node B ION-DTN configs
│   ├── leo-cubesat/            # LEO satellite configs
│   └── cislunar-*/             # Cislunar configs
├── docs/
│   ├── terrestrial-dtn-phase1/ # Phase 1 documentation
│   ├── cubesat-em-phase2/      # Phase 2 documentation
│   ├── leo-cubesat-phase3/     # Phase 3 documentation
│   └── cislunar-phase4/        # Phase 4 documentation
├── examples/                   # Working examples
└── ax25/                       # AX.25 frame handling
```

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/contact
go test ./pkg/security
```

### Property-Based Tests

The project uses property-based testing to validate correctness properties:

```bash
# Run property tests
go test -v ./pkg/store -run Property
go test -v ./pkg/contact -run Property

# Run with more iterations
go test -v ./pkg/store -run Property -gopter.minSuccessfulTests=1000
```

### Integration Tests

```bash
# End-to-end integration test (requires hardware)
./scripts/test-e2e-integration.sh node-a

# Extended duration test (1 hour)
./scripts/test-extended-duration.sh node-a 60
```

---

## 📊 Performance Characteristics

### Phase 1 (Terrestrial)
- **Link Rate**: 9600 baud (1200 bytes/sec theoretical)
- **Effective Throughput**: ~960 bytes/sec (with AX.25/LTP overhead)
- **Ping RTT**: 1000-1500 ms (typical)
- **Max Bundle Size**: 64 KB (configurable)

### Phase 3 (LEO CubeSat)
- **Link Rate**: 9.6 kbps UHF
- **Pass Duration**: 5-10 minutes
- **Passes per Day**: 4-6 per ground station
- **Max Doppler Shift**: ±10 kHz at 437 MHz

### Phase 4 (Cislunar)
- **Link Rate**: 500 bps S-band (with LDPC)
- **One-Way Light Time**: ~1.3 seconds (Earth-Moon)
- **Round-Trip Time**: ~2.6 seconds + processing
- **Range**: Up to 384,400 km

---

## 🔐 Security

### BPSec Integrity Protection

All bundles can be protected with HMAC-SHA256 integrity blocks:

- ✅ Bundle origin authentication
- ✅ Tamper detection
- ✅ Pre-shared key management
- ❌ No encryption (amateur radio regulations require unencrypted transmissions)

### Rate Limiting

Configurable rate limiting prevents bundle flooding attacks:

- Per-source endpoint rate limits
- Maximum bundle size enforcement
- Storage capacity protection

### Hardware Security (Phases 2-4)

STM32U585-based nodes leverage hardware security features:

- **Hardware Crypto Accelerator**: AES-256, SHA-256, PKA
- **TrustZone**: Secure key storage isolated from application code
- **Secure Boot**: Verified firmware loading

---

## 📚 Documentation

### Phase-Specific Guides

- [Phase 1: Terrestrial DTN](docs/terrestrial-dtn-phase1/)
  - [Requirements](docs/terrestrial-dtn-phase1/requirements.md)
  - [Design](docs/terrestrial-dtn-phase1/design.md)
  - [Tasks](docs/terrestrial-dtn-phase1/tasks.md)
  - [Test Guides](docs/terrestrial-dtn-phase1/)

- [Phase 2: Engineering Model](docs/cubesat-em-phase2/)
- [Phase 3: LEO CubeSat](docs/leo-cubesat-phase3/)
- [Phase 4: Cislunar](docs/cislunar-phase4/)

### Package Documentation

- [ION-DTN Wrapper](pkg/ion/README.md)
- [Contact Plan Manager + CGR](pkg/contact/README.md)
- [Security Package](pkg/security/README.md)
- [Core Infrastructure](pkg/README.md)

### External Resources

- [ION-DTN Documentation](https://sourceforge.net/projects/ion-dtn/)
- [RFC 9171: Bundle Protocol Version 7](https://www.rfc-editor.org/rfc/rfc9171.html)
- [RFC 9172: Bundle Protocol Security (BPSec)](https://www.rfc-editor.org/rfc/rfc9172.html)
- [RFC 5326: Licklider Transmission Protocol (LTP)](https://www.rfc-editor.org/rfc/rfc5326.html)
- [AX.25 Link Access Protocol](http://www.ax25.net/AX25.2.2-Jul%2098-2.pdf)

---

## 🤝 Contributing

This is a research project exploring DTN for amateur radio. Contributions are welcome!

### Areas for Contribution

- 🐛 Bug reports and fixes
- 📝 Documentation improvements
- 🧪 Additional test coverage
- 🔧 Hardware integration (TNC4, B200mini, STM32U585)
- 📡 RF link optimization
- 🛰️ Orbital mechanics improvements

### Development Setup

```bash
# Clone repository
git clone https://github.com/yourusername/cislunar_proposal.git
cd cislunar_proposal

# Install dependencies
go mod download

# Run tests
go test ./...

# Build all CLIs
go build -o dtn-node ./cmd/dtn-node
go build -o em-node ./cmd/em-node
go build -o leo-node ./cmd/leo-node
go build -o cislunar-node ./cmd/cislunar-node
```

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **NASA JPL** for ION-DTN
- **Mobilinkd** for TNC4 hardware
- **Ettus Research** for B200mini SDR
- **Amateur radio community** for AX.25 and KISS protocols
- **STMicroelectronics** for STM32U585 ultra-low-power MCU

---

## 📞 Contact

For questions, issues, or collaboration:

- **GitHub Issues**: [Report a bug or request a feature](https://github.com/yourusername/cislunar_proposal/issues)
- **Discussions**: [Join the conversation](https://github.com/yourusername/cislunar_proposal/discussions)

---

## 🌟 Project Status

| Phase | Status | Hardware | Link | Documentation |
|-------|--------|----------|------|---------------|
| **Phase 1: Terrestrial** | ✅ Complete | RPi + TNC4 + FT-817 | VHF/UHF 9600 baud | [📖 Docs](docs/terrestrial-dtn-phase1/) |
| **Phase 2: EM** | 🔧 In Progress | STM32U585 + B200mini | UHF/S-band | [📖 Docs](docs/cubesat-em-phase2/) |
| **Phase 3: LEO** | 📋 Planned | STM32U585 + IQ Radio | UHF 437 MHz | [📖 Docs](docs/leo-cubesat-phase3/) |
| **Phase 4: Cislunar** | 📋 Planned | STM32U585+ | S/X-band | [📖 Docs](docs/cislunar-phase4/) |

---

<p align="center">
  <strong>Bringing Delay-Tolerant Networking to Amateur Radio</strong><br>
  From Earth to the Moon 🌍 → 🌙
</p>
