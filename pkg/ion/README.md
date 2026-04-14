# ION-DTN Go Orchestration Wrapper

This package provides a Go wrapper for managing ION-DTN node lifecycle, telemetry collection, and contact plan management for the terrestrial DTN Phase 1 implementation.

## Components

### 1. Node Lifecycle Management (`lifecycle.go`)

Manages ION-DTN node startup, shutdown, and health monitoring.

**Features:**
- Start ION-DTN with automatic initialization of ionadmin, ltpadmin, bpadmin, ipnadmin, bpsecadmin
- Graceful shutdown with ionstop
- Process health checking (IsRunning)
- Automatic environment setup (PATH, DYLD_LIBRARY_PATH, LD_LIBRARY_PATH)
- Configuration file management (copies kiss.ionconfig to working directory)

**Example:**
```go
lifecycle, err := ion.NewNodeLifecycle(ion.NodeConfig{
    NodeID:     "node-a",
    NodeNumber: 1,
    ConfigDir:  "./configs/node-a",
    IONInstall: "./ion-install",
})

if err := lifecycle.Start(); err != nil {
    log.Fatal(err)
}

// Wait for ION to be ready
if err := lifecycle.WaitForReady(30 * time.Second); err != nil {
    log.Printf("Warning: %v", err)
}

// Check if running
if lifecycle.IsRunning() {
    log.Println("ION-DTN is running")
}

// Graceful shutdown
lifecycle.Stop()
```

### 2. Telemetry Collection (`telemetry.go`)

Collects and exposes ION-DTN telemetry data via JSON.

**Features:**
- Query Bundle Protocol statistics (bundles stored, sent, received, storage usage)
- Query LTP statistics (sessions, segments, retransmissions)
- Query contact plan information
- Calculate node health status
- Export telemetry as JSON
- Save telemetry to file

**Telemetry Data Structure:**
```go
type Telemetry struct {
    Timestamp      time.Time
    NodeID         string
    NodeNumber     int
    BundleProtocol BPTelemetry
    LTP            LTPTelemetry
    ContactPlan    ContactTelemetry
    Health         HealthStatus
}
```

**Example:**
```go
collector := ion.NewTelemetryCollector(lifecycle)

telemetry, err := collector.Collect()
if err != nil {
    log.Fatal(err)
}

// Print as JSON
jsonStr, _ := telemetry.ToJSONString()
fmt.Println(jsonStr)

// Save to file
telemetry.SaveToFile("telemetry.json")
```

### 3. Contact Plan Management (`contactplan.go`)

Manages ION-DTN contact plans with YAML/JSON support and runtime updates.

**Features:**
- Load contact plans from YAML or JSON files
- Validate contact plan integrity
- Generate ionadmin commands from contact plan
- Apply contact plans to running ION-DTN nodes
- Add/remove/update contacts at runtime
- Query active contacts
- Find next contact with a specific node
- Save contact plans to YAML/JSON

**Contact Plan Structure:**
```yaml
plan_id: "terrestrial-always-on"
valid_from: 0
valid_to: 2147483647

contacts:
  - id: "node-a-to-node-b"
    start_time: 0
    end_time: 2147483647
    from_node: 1
    to_node: 2
    data_rate: 9600
    confidence: 1.0

ranges:
  - start_time: 0
    from_node: 1
    to_node: 2
    distance: 1
```

**Example:**
```go
cpm := ion.NewContactPlanManager(lifecycle)

// Load from YAML
if err := cpm.LoadFromYAML("contact-plan.yaml"); err != nil {
    log.Fatal(err)
}

// Apply to ION-DTN
if err := cpm.Apply(); err != nil {
    log.Fatal(err)
}

// Add a new contact at runtime
contact := ion.Contact{
    ID:        "new-contact",
    StartTime: time.Now().Unix(),
    EndTime:   time.Now().Unix() + 3600,
    FromNode:  1,
    ToNode:    2,
    DataRate:  9600,
}
cpm.AddContact(contact)

// Get active contacts
active := cpm.GetActiveContacts(time.Now().Unix())
fmt.Printf("Active contacts: %d\n", len(active))
```

## CLI Tool: `dtn-node`

A unified command-line interface for operating terrestrial DTN nodes.

### Usage

```bash
# Start node with config file
./dtn-node -config configs/dtn-node-a.yaml

# Start node with command-line flags
./dtn-node -node-id node-a -node-number 1 -config-dir ./configs/node-a

# Show version
./dtn-node -version
```

### Configuration File

YAML format:
```yaml
node_id: "node-a"
node_number: 1
callsign: "G4DPZ-1"
config_dir: "./configs/node-a"
ion_install: "./ion-install"
contact_plan_file: "./configs/terrestrial-contact-plan.yaml"
telemetry_port: 8080
telemetry_file: "./telemetry-node-a.json"
health_interval: 10
```

JSON format is also supported.

### Features

1. **Automatic ION-DTN Startup**: Initializes all ION-DTN subsystems
2. **Contact Plan Loading**: Loads and applies contact plans from YAML/JSON
3. **Health Monitoring**: Periodic health checks with configurable interval
4. **Telemetry HTTP Server**: Exposes telemetry via HTTP endpoints
5. **Graceful Shutdown**: Handles Ctrl+C for clean shutdown

### HTTP Endpoints

When telemetry server is enabled:

- `GET /health` - Current node health and telemetry
- `GET /contacts` - List all contacts in the plan
- `GET /contacts/active` - List currently active contacts

Example:
```bash
# Get node health
curl http://localhost:8080/health

# Get active contacts
curl http://localhost:8080/contacts/active
```

### Example: Running Two Nodes

Terminal 1 (Node A):
```bash
./dtn-node -config configs/dtn-node-a.yaml
```

Terminal 2 (Node B):
```bash
./dtn-node -config configs/dtn-node-b.yaml
```

Terminal 3 (Test ping):
```bash
# Wait for nodes to start, then test
./ion-install/bin/bping ipn:1.1 ipn:2.1 -c 5
```

## Requirements Validation

This implementation validates the following requirements from the terrestrial-dtn-phase1 spec:

- **Requirement 14.3**: Node lifecycle management (Start, Stop, IsRunning, graceful shutdown)
- **Requirement 13.1**: Telemetry collection (uptime, storage, bundles, contacts)
- **Requirement 13.2**: Telemetry statistics (bundles sent/received, bytes, latency, contacts)
- **Requirement 13.3**: Telemetry exposure via local interface (JSON file and HTTP)
- **Requirement 13.4**: Telemetry query response within 1 second
- **Requirement 7.1**: Contact plan time-tagged schedule maintenance
- **Requirement 7.2**: Active contacts query
- **Requirement 7.3**: Next contact lookup
- **Requirement 7.6**: Contact plan persistence to filesystem
- **Requirement 7.7**: Contact plan loading from configuration file

## Architecture

```
┌─────────────────────────────────────────┐
│         dtn-node CLI                    │
│  (cmd/dtn-node/main.go)                 │
└─────────────────┬───────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
    ▼             ▼             ▼
┌─────────┐  ┌──────────┐  ┌──────────┐
│Lifecycle│  │Telemetry │  │ Contact  │
│ Manager │  │Collector │  │Plan Mgr  │
└────┬────┘  └────┬─────┘  └────┬─────┘
     │            │             │
     └────────────┼─────────────┘
                  │
                  ▼
         ┌────────────────┐
         │   ION-DTN      │
         │  (C binaries)  │
         └────────────────┘
```

## Testing

The wrapper can be tested with the existing ION-DTN configuration:

```bash
# Build the CLI
go build -o dtn-node ./cmd/dtn-node

# Test Node A
./dtn-node -config configs/dtn-node-a.yaml

# In another terminal, test Node B
./dtn-node -config configs/dtn-node-b.yaml

# Test telemetry
curl http://localhost:8080/health | jq .

# Test ping
./ion-install/bin/bping ipn:1.1 ipn:2.1 -c 5
```

## Integration with Existing Code

This wrapper integrates with the existing terrestrial-dtn codebase:

- Uses existing ION-DTN configuration files in `configs/node-a/` and `configs/node-b/`
- Works with existing startup scripts (`scripts/start-node-a.sh`, etc.)
- Compatible with existing ION-DTN binaries in `ion-install/bin/`
- Powers the `cmd/dtn-node/main.go` CLI tool for Phase 1 operations

## Future Enhancements

Potential improvements for future phases:

1. WebSocket support for real-time telemetry streaming
2. Prometheus metrics export
3. Automated contact plan generation from orbital predictions (for space segments)
4. Integration with the custom BPv7 implementation in `pkg/bpa/`
5. Support for multiple simultaneous contact windows
6. Advanced error recovery and automatic restart
7. Configuration hot-reload without restart
