# dtn-node - Terrestrial DTN Node CLI

A unified command-line interface for operating ION-DTN terrestrial nodes with integrated lifecycle management, telemetry collection, and contact plan management.

## Features

- **Automated ION-DTN Startup**: Initializes ionadmin, ltpadmin, bpadmin, ipnadmin
- **Contact Plan Management**: Load and apply contact plans from YAML/JSON files
- **Health Monitoring**: Periodic health checks with configurable intervals
- **Telemetry Collection**: Query ION-DTN statistics and expose via HTTP/JSON
- **Graceful Shutdown**: Clean shutdown on Ctrl+C or SIGTERM
- **Configuration Files**: YAML or JSON configuration support

## Installation

```bash
# Build from source
go build -o dtn-node ./cmd/dtn-node

# Or use the build script
make dtn-node
```

## Usage

### Basic Usage

```bash
# Start with configuration file
./dtn-node -config configs/dtn-node-a.yaml

# Start with command-line flags
./dtn-node \
  -node-id node-a \
  -node-number 1 \
  -config-dir ./configs/node-a \
  -ion-install ./ion-install \
  -telemetry-port 8080

# Show version
./dtn-node -version
```

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-config` | Path to configuration file (YAML or JSON) | - |
| `-node-id` | Node identifier (overrides config) | - |
| `-node-number` | ION node number (overrides config) | - |
| `-config-dir` | ION-DTN configuration directory (overrides config) | - |
| `-ion-install` | Path to ION-DTN installation | `./ion-install` |
| `-telemetry-port` | HTTP port for telemetry endpoint | `8080` |
| `-version` | Show version information | - |

## Configuration File

### YAML Format

```yaml
# Node identification
node_id: "node-a"
node_number: 1
callsign: "G4DPZ-1"

# Paths
config_dir: "./configs/node-a"
ion_install: "./ion-install"
contact_plan_file: "./configs/terrestrial-contact-plan.yaml"

# Telemetry
telemetry_port: 8080
telemetry_file: "./telemetry-node-a.json"
health_interval: 10  # seconds
```

### JSON Format

```json
{
  "node_id": "node-a",
  "node_number": 1,
  "callsign": "G4DPZ-1",
  "config_dir": "./configs/node-a",
  "ion_install": "./ion-install",
  "contact_plan_file": "./configs/terrestrial-contact-plan.yaml",
  "telemetry_port": 8080,
  "telemetry_file": "./telemetry-node-a.json",
  "health_interval": 10
}
```

### Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `node_id` | string | Yes | Unique node identifier |
| `node_number` | int | Yes | ION-DTN node number (must be positive) |
| `callsign` | string | No | Amateur radio callsign |
| `config_dir` | string | Yes | Path to ION-DTN config files (node.ionrc, etc.) |
| `ion_install` | string | Yes | Path to ION-DTN installation directory |
| `contact_plan_file` | string | No | Path to contact plan file (YAML or JSON) |
| `telemetry_port` | int | No | HTTP port for telemetry server (0 to disable) |
| `telemetry_file` | string | No | Path to save telemetry JSON snapshots |
| `health_interval` | int | No | Health check interval in seconds (default: 10) |

## Contact Plan File

### YAML Format

```yaml
plan_id: "terrestrial-always-on"
valid_from: 0
valid_to: 2147483647  # Max int32

contacts:
  - id: "node-a-to-node-b"
    start_time: 0
    end_time: 2147483647
    from_node: 1
    to_node: 2
    data_rate: 9600  # bits per second
    confidence: 1.0

  - id: "node-b-to-node-a"
    start_time: 0
    end_time: 2147483647
    from_node: 2
    to_node: 1
    data_rate: 9600
    confidence: 1.0

ranges:
  - start_time: 0
    from_node: 1
    to_node: 2
    distance: 1  # kilometers
```

## HTTP Telemetry Endpoints

When `telemetry_port` is configured, the following HTTP endpoints are available:

### GET /health

Returns current node health and telemetry.

**Response:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "node_id": "node-a",
  "node_number": 1,
  "bundle_protocol": {
    "bundles_stored": 5,
    "bundles_received": 10,
    "bundles_sent": 8,
    "bundles_forwarded": 3,
    "bundles_expired": 0,
    "bytes_received": 1024,
    "bytes_sent": 2048,
    "storage_used_bytes": 512,
    "storage_quota_bytes": 10485760
  },
  "ltp": {
    "sessions_active": 1,
    "sessions_completed": 5,
    "sessions_failed": 0,
    "segments_sent": 20,
    "segments_received": 18,
    "retransmissions": 2
  },
  "contact_plan": {
    "contacts_active": 1,
    "contacts_completed": 0,
    "contacts_missed": 0
  },
  "health": {
    "running": true,
    "uptime_seconds": 300,
    "storage_percent": 4.88,
    "error_count": 0
  }
}
```

### GET /contacts

Returns all contacts in the loaded contact plan.

**Response:**
```json
{
  "contacts": [
    {
      "id": "node-a-to-node-b",
      "start_time": 0,
      "end_time": 2147483647,
      "from_node": 1,
      "to_node": 2,
      "data_rate": 9600,
      "confidence": 1.0
    }
  ]
}
```

### GET /contacts/active

Returns currently active contacts.

**Response:**
```json
{
  "active_contacts": [
    {
      "id": "node-a-to-node-b",
      "start_time": 0,
      "end_time": 2147483647,
      "from_node": 1,
      "to_node": 2,
      "data_rate": 9600,
      "confidence": 1.0
    }
  ],
  "current_time": 1705315800
}
```

## Example: Running Two Nodes

### Terminal 1: Start Node A

```bash
./dtn-node -config configs/dtn-node-a.yaml
```

Output:
```
Starting DTN node: node-a (node 1)
  Callsign: G4DPZ-1
  Config directory: /path/to/configs/node-a
  ION install: /path/to/ion-install
  Telemetry port: 8080
Starting ION-DTN...
Waiting for ION-DTN to be ready...
ION-DTN is ready
Loading contact plan from ./configs/terrestrial-contact-plan.yaml...
Contact plan loaded successfully
Contact plan applied to ION-DTN
Telemetry server started on http://localhost:8080
Node is running. Press Ctrl+C to stop.
Health: uptime=10s, storage=0.0%, bundles_stored=0, bundles_sent=0, bundles_received=0
```

### Terminal 2: Start Node B

```bash
./dtn-node -config configs/dtn-node-b.yaml
```

### Terminal 3: Test Communication

```bash
# Test ping
./ion-install/bin/bping ipn:1.1 ipn:2.1 -c 5

# Query telemetry
curl http://localhost:8080/health | jq .

# Check active contacts
curl http://localhost:8080/contacts/active | jq .

# Send a file
./ion-install/bin/bpsendfile ipn:1.1 ipn:2.1 testfile.txt

# Receive on Node B
./ion-install/bin/bprecvfile ipn:2.1 1
```

## Health Monitoring

The node automatically monitors health at the configured interval (default: 10 seconds) and logs:

- Uptime in seconds
- Storage utilization percentage
- Bundles stored, sent, received
- LTP sessions and segments
- Contact statistics

Example log output:
```
Health: uptime=60s, storage=2.3%, bundles_stored=3, bundles_sent=5, bundles_received=4
```

If `telemetry_file` is configured, health snapshots are saved to the file in JSON format.

## Graceful Shutdown

The node handles shutdown signals gracefully:

1. Receives SIGINT (Ctrl+C) or SIGTERM
2. Logs shutdown message
3. Stops ION-DTN with `ionstop`
4. Waits for clean shutdown
5. Exits

Example:
```
^CReceived signal interrupt, shutting down...
Stopping ION-DTN...
Node stopped
```

## Troubleshooting

### ION-DTN fails to start

**Problem:** `ionadmin failed: exit status 1`

**Solutions:**
- Check that `config_dir` contains valid ION-DTN config files
- Verify `ion_install` path is correct
- Ensure no other ION-DTN instance is running (`ionstop`)
- Check file permissions on config files

### Telemetry collection fails

**Problem:** `Failed to collect telemetry: bpadmin failed`

**Solutions:**
- Verify ION-DTN is running: `ps aux | grep ion`
- Check ION-DTN logs in the working directory
- Ensure ION-DTN binaries are in `ion_install/bin/`

### Contact plan not applied

**Problem:** `Failed to apply contact plan: ionadmin failed`

**Solutions:**
- Validate contact plan file syntax (YAML/JSON)
- Check that `valid_from < valid_to`
- Ensure all contacts fall within valid time range
- Verify node numbers match ION-DTN configuration

### Port already in use

**Problem:** `bind: address already in use`

**Solutions:**
- Change `telemetry_port` in config file
- Stop other process using the port
- Use `-telemetry-port 0` to disable HTTP server

## Integration with ION-DTN

The CLI integrates seamlessly with ION-DTN:

1. **Configuration Files**: Uses standard ION-DTN config files:
   - `node.ionrc` - ION initialization
   - `node.ltprc` - LTP configuration
   - `node.bprc` - Bundle Protocol configuration
   - `node.ipnrc` - IPN routing
   - `kiss.ionconfig` - KISS CLA configuration

2. **Binaries**: Executes ION-DTN binaries:
   - `ionadmin` - ION administration
   - `ltpadmin` - LTP administration
   - `bpadmin` - BP administration
   - `ipnadmin` - IPN administration
   - `ionstop` - Clean shutdown

3. **Environment**: Automatically sets:
   - `PATH` - Includes ION bin directory
   - `DYLD_LIBRARY_PATH` - macOS library path
   - `LD_LIBRARY_PATH` - Linux library path

## Requirements Validation

This CLI validates the following requirements from terrestrial-dtn-phase1:

- **Task 13.1**: Node lifecycle (Start, Stop, IsRunning, graceful shutdown)
- **Task 13.2**: Telemetry collection (query ION-DTN, parse output, expose JSON/HTTP)
- **Task 13.3**: Contact plan management (load YAML/JSON, generate ionadmin commands, runtime updates)
- **Task 13.4**: Unified CLI (single entry point, config parsing, health monitoring)

## See Also

- [ION-DTN Go Wrapper Documentation](../../pkg/ion/README.md)
- [Terrestrial DTN Phase 1 Design](../../docs/terrestrial-dtn-phase1/design.md)
- [ION-DTN Configuration Guide](../../configs/README.md)
