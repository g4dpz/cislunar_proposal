# radiant-ion

ION-DTN configuration management and lifecycle control daemon for the RADIANT project.

Reads a canonical YAML network configuration and generates ION admin scripts, manages ION's lifecycle (start/stop/restart), collects telemetry, and exposes an HTTP/JSON API for remote management.

## Prerequisites

ION-DTN binaries must be installed and available on `$PATH`:

- `ionadmin` — ION node administration
- `bpadmin` — Bundle Protocol administration
- `ltpadmin` — Licklider Transmission Protocol administration
- `ipnadmin` — IPN scheme routing administration
- `ionstop` — Stop all ION daemons
- `bpstats` — Bundle statistics
- `bplist` — List queued bundles
- `ltpinfo` — LTP span information

## Usage

```bash
# Generate ION admin scripts from a YAML config
radiant-ion generate configs/my-node.yaml --output-dir ./ion-scripts/

# Generate configs and start ION
radiant-ion start configs/my-node.yaml

# Stop ION
radiant-ion stop

# Show ION health and bundle statistics
radiant-ion status

# Start ION with the HTTP/JSON management API
radiant-ion serve configs/my-node.yaml --port 3000
```

### Environment Variables

- `RUST_LOG` — Controls log verbosity (default: `info`). Example: `RUST_LOG=debug`

## Configuration

The YAML configuration follows the canonical `NetworkConfiguration` model from `radiant-dtn-abstraction`. See that crate's documentation for the full schema.

### Example

```yaml
version: "1.0"
backend: ion-dtn

local_node:
  node_number: 10
  name: "G4DPZ Ground Station"
  endpoint_id:
    Ipn:
      node_number: 10
      service_number: 0
  callsign_eid:
    Dtn:
      authority: "g4dpz-1"
      path: "gs"
  services:
    - service_number: 1
      description: "Bundle delivery"
    - service_number: 2047
      description: "Echo"

neighbors:
  - node_number: 20
    name: "Lunar Orbiter"
    links:
      - LtpUdp:
          id: "ltp-to-orbiter"
          local_engine_id: 10
          remote_engine_id: 20
          remote_host: "192.168.1.20"
          remote_port: 1113
          local_port: 2113
          mtu: 1360

contact_plan:
  contacts:
    - source_node: 10
      dest_node: 20
      start_time: 1700000000
      end_time: 1700003600
      rate_bps: 9600
      confidence: 0.95
  ranges:
    - source_node: 10
      dest_node: 20
      owlt_secs: 1.3

routing:
  strategy: Cgr
  static_routes: []
```

## HTTP/JSON API Endpoints

When running in `serve` mode, the following endpoints are available:

### Configuration
- `POST /config` — Validate and store canonical config
- `GET /config` — Retrieve current canonical config
- `POST /config/preview` — Generate backend config without deploying
- `POST /config/deploy` — Generate and deploy backend config

### Lifecycle
- `POST /lifecycle/start` — Start DTN engine
- `POST /lifecycle/stop` — Stop DTN engine
- `POST /lifecycle/restart` — Restart DTN engine
- `GET /lifecycle/state` — Get current engine state
- `GET /lifecycle/health` — Health check
- `GET /lifecycle/version` — Engine version query

### Runtime Administration (hot reconfiguration)
- `POST /runtime/contacts` — Add contact
- `DELETE /runtime/contacts` — Remove contact
- `POST /runtime/neighbors` — Add neighbor
- `DELETE /runtime/neighbors` — Remove neighbor
- `POST /runtime/links/:id/enable` — Enable link
- `POST /runtime/links/:id/disable` — Disable link

### Monitoring
- `GET /stats` — Bundle statistics
- `GET /stats/links` — Per-neighbor link state
- `GET /capabilities` — Backend capability set
- `GET /adapters` — List registered adapters

### Events
- `GET /events` — Server-Sent Events stream (real-time notifications)

## Amateur Radio Compliance

This tool is designed for use within the amateur radio service:

- **No encryption** is configured by default. Bundle payloads remain unencrypted per ITU Radio Regulations (Article 25) and national regulations (e.g., FCC Part 97.113).
- No BPSec directives are generated in ION admin scripts.
- Callsign-embedded EIDs (`dtn://callsign-ssid/service`) are supported as first-class citizens.
- All protocols and data formats used over-the-air are publicly documented.

## Building

```bash
cargo build -p radiant-ion
```

Or from this directory:

```bash
cargo build --release
```

## License

MIT
