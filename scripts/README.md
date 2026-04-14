# Test Scripts for Terrestrial DTN Phase 1

This directory contains test scripts for validating the terrestrial DTN Phase 1 implementation.

## Overview

The test scripts follow a progressive validation approach, testing individual components before full system integration:

1. **Task 7**: ION-DTN bping (ping functionality)
2. **Task 9**: Store-and-forward operations
3. **Task 11**: BPSec integrity protection
4. **Task 15.1**: End-to-end integration using Go wrapper
5. **Task 15.2**: Extended duration stability test

## Test Scripts

### Node Lifecycle Scripts

#### `start-node-a.sh`
Starts Node A (Engine 1) with ION-DTN configuration.
```bash
./scripts/start-node-a.sh
```

#### `start-node-b.sh`
Starts Node B (Engine 2) with ION-DTN configuration.
```bash
./scripts/start-node-b.sh
```

#### `stop-node.sh`
Stops ION-DTN cleanly using `ionstop`.
```bash
./scripts/stop-node.sh
```

### Task 7: ION-DTN bping Test

#### `test-ion-bping.sh`
Tests ION-DTN ping functionality over KISS CLA.

**Usage:**
```bash
# On Node A
./scripts/test-ion-bping.sh node-a

# On Node B
./scripts/test-ion-bping.sh node-b
```

**What it tests:**
- ION-DTN initialization
- Process verification (rfxclock, ltpkisscli, ltpkissclo)
- bping in both directions (A→B and B→A)
- Round-trip time measurement

**Documentation:** [TASK7_BPING_TEST_GUIDE.md](../docs/terrestrial-dtn-phase1/TASK7_BPING_TEST_GUIDE.md)

### Task 9: Store-and-Forward Test

#### `test-ion-store-forward.sh`
Tests store-and-forward functionality with multiple sub-tests.

**Usage:**
```bash
# On Node A
./scripts/test-ion-store-forward.sh node-a

# On Node B
./scripts/test-ion-store-forward.sh node-b
```

**What it tests:**
- Task 9.1: bpsendfile / bprecvfile
- Task 9.2: Store-and-forward with delayed contact
- Task 9.3: Priority-based delivery
- Task 9.4: Bundle lifetime expiry

**Sub-scripts:**
- `test-task-9.1.sh` - File transfer test
- `test-task-9.2.sh` - Delayed contact test
- `test-task-9.3.sh` - Priority delivery test
- `test-task-9.4.sh` - Lifetime expiry test

**Documentation:** [TASK9_STORE_FORWARD_TEST_GUIDE.md](../docs/terrestrial-dtn-phase1/TASK9_STORE_FORWARD_TEST_GUIDE.md)

### Task 11: BPSec Integrity Test

#### `test-ion-bpsec.sh`
Tests BPSec integrity protection (HMAC-SHA-256).

**Usage:**
```bash
# On Node A
./scripts/test-ion-bpsec.sh node-a

# On Node B
./scripts/test-ion-bpsec.sh node-b
```

**What it tests:**
- Task 11.1: BPSec configuration
- Task 11.2: Bundle integrity verification
- Task 11.3: Integrity failure detection (optional)

**Documentation:** [TASK11_BPSEC_TEST_GUIDE.md](../docs/terrestrial-dtn-phase1/TASK11_BPSEC_TEST_GUIDE.md)

### Task 15.1: End-to-End Integration Test

#### `test-e2e-integration.sh`
**NEW** - Full end-to-end integration test using the Go wrapper (`dtn-node` CLI).

**Usage:**
```bash
# On Node A
./scripts/test-e2e-integration.sh node-a

# On Node B
./scripts/test-e2e-integration.sh node-b
```

**What it tests:**
1. Build `dtn-node` CLI
2. Start nodes using Go wrapper
3. Verify ION-DTN processes
4. Query telemetry via HTTP endpoint
5. Run bping tests in both directions
6. Send files in both directions
7. Verify telemetry accuracy
8. Graceful shutdown

**Requirements validated:** All requirements (end-to-end system)

**Documentation:** [TASK15_E2E_INTEGRATION_GUIDE.md](../docs/terrestrial-dtn-phase1/TASK15_E2E_INTEGRATION_GUIDE.md)

### Task 15.2: Extended Duration Test

#### `test-extended-duration.sh`
**NEW** - Extended duration stability test (1+ hours).

**Usage:**
```bash
# On Node A (60 minutes)
./scripts/test-extended-duration.sh node-a 60

# On Node B (120 minutes)
./scripts/test-extended-duration.sh node-b 120
```

**What it tests:**
- Node stability over extended duration (1+ hours)
- Periodic bundle exchanges (every 5 minutes)
- Memory leak detection (storage growth monitoring)
- Process crash detection and recovery
- Telemetry accuracy over time

**Test parameters:**
- Default duration: 60 minutes
- Bundle interval: 5 minutes
- Telemetry interval: 1 minute

**Outputs:**
- Main log: `test-logs/extended-duration-node-a-YYYYMMDD-HHMMSS.log`
- Telemetry log: `test-logs/telemetry-node-a-YYYYMMDD-HHMMSS.log`

**Requirements validated:**
- Requirement 15.1: Node performance
- Requirement 13.1-13.2: Telemetry
- Requirement 14.3: Error recovery
- Requirement 2: Bundle persistence

**Documentation:** [TASK15_E2E_INTEGRATION_GUIDE.md](../docs/terrestrial-dtn-phase1/TASK15_E2E_INTEGRATION_GUIDE.md)

## Test Progression

Follow this order for comprehensive validation:

1. **Basic connectivity**: Task 7 (bping)
2. **Store-and-forward**: Task 9 (file transfers, priority, lifetime)
3. **Security**: Task 11 (BPSec integrity)
4. **Integration**: Task 15.1 (end-to-end with Go wrapper)
5. **Stability**: Task 15.2 (extended duration)

## Prerequisites

### Hardware
- Two Linux or macOS hosts
- Two Mobilinkd TNC4 (USB connection)
- Two Yaesu FT-817 radios (9600 baud)

### Software
- ION-DTN installed in `ion-install/`
- Go 1.19+ (for Task 15)
- Configuration files in `configs/`

### Network
- Node A: `ipn:1.1`, telemetry port 8080
- Node B: `ipn:2.1`, telemetry port 8081

## Common Issues

### ION-DTN won't start
- Check configuration files in `configs/node-a/` or `configs/node-b/`
- Verify ION binaries exist: `ls ion-install/bin/ionadmin`
- Review `ion.log` for errors
- Ensure no previous ION instance is running: `ionstop`

### TNC4 not detected
- Check USB connection: `ls /dev/tty.usbmodem*` (macOS) or `ls /dev/ttyACM*` (Linux)
- Verify device path in `kiss.ionconfig`
- Check TNC4 power and USB cable

### bping fails
- Verify both nodes are running
- Check contact plan allows communication
- Verify radios are powered on and configured
- Check TNC4 connections on both nodes

### File transfer fails
- Ensure receiver is started before sender
- Check bundle lifetime is sufficient
- Verify contact window is active
- Review LTP statistics for retransmissions

## Log Files

### ION-DTN logs
- `ion.log` - ION-DTN system log
- `dtn-node-node-a.log` - Go wrapper log (Task 15)
- `dtn-node-node-b.log` - Go wrapper log (Task 15)

### Test logs
- `test-logs/extended-duration-*.log` - Extended duration test logs
- `test-logs/telemetry-*.log` - Telemetry snapshots

### Telemetry files
- `telemetry-node-a.json` - Node A telemetry (updated every 10s)
- `telemetry-node-b.json` - Node B telemetry (updated every 10s)

## Telemetry Endpoints

### Node A
- Health: `http://localhost:8080/health`
- Contacts: `http://localhost:8080/contacts`
- Active contacts: `http://localhost:8080/contacts/active`

### Node B
- Health: `http://localhost:8081/health`
- Contacts: `http://localhost:8081/contacts`
- Active contacts: `http://localhost:8081/contacts/active`

## Test Data

Test files are created in `test-data/`:
- `testfile-*.txt` - Test files for store-and-forward
- `received-*.txt` - Received files
- `e2e-test-*.txt` - End-to-end test files
- `extended-test-*.txt` - Extended duration test files

## Cleanup

After testing:
```bash
# Stop ION-DTN
./scripts/stop-node.sh

# Clean up test data
rm -rf test-data/

# Clean up logs
rm -rf test-logs/

# Remove telemetry files
rm -f telemetry-node-*.json
```

## References

- [Requirements Document](../docs/terrestrial-dtn-phase1/requirements.md)
- [Design Document](../docs/terrestrial-dtn-phase1/design.md)
- [Tasks Document](../docs/terrestrial-dtn-phase1/tasks.md)
- [DTN Node Integration Guide](../docs/terrestrial-dtn-phase1/DTN_NODE_INTEGRATION_GUIDE.md)
