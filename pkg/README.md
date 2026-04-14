# Cislunar Amateur DTN Payload - Core Infrastructure

This directory contains the core DTN infrastructure shared across all phases of the cislunar amateur DTN payload project.

## Package Overview

### `pkg/bpa` - Bundle Protocol Agent
Go wrapper around ION-DTN bundle operations. Handles:
- Bundle creation and validation
- Ping echo request/response handling
- Bundle type management (data, ping request, ping response)
- Priority-based queue management

**Key Files:**
- `types.go` - Core data types (Bundle, BundleID, EndpointID, Priority, BundleType)
- `bpa.go` - BPA implementation
- `bpa_property_test.go` - Property-based tests for bundle validation and ping correctness

### `pkg/store` - Bundle Store
Go wrapper around ION-DTN bundle storage. Handles:
- Persistent bundle storage
- Priority-ordered retrieval
- Capacity management with eviction policy
- Atomic store/delete operations

**Key Files:**
- `store.go` - Bundle store implementation
- `store_property_test.go` - Property-based tests for store operations, priority ordering, and eviction

### `pkg/contact` - Contact Plan Manager
Manages scheduled communication windows and CGR-based contact prediction. Handles:
- Contact window scheduling
- Active contact queries
- CGR integration for orbital pass prediction (placeholder for ION-DTN CGR)
- Direct contact lookup (no multi-hop routing)

**Key Files:**
- `types.go` - Contact plan data types (ContactWindow, OrbitalParameters, PredictedContact)
- `manager.go` - Contact plan manager implementation

### `pkg/cla` - Convergence Layer Adapter
Abstraction for AX.25/LTP protocol stack across all radio links. Provides:
- Uniform interface for bundle transmission
- Support for multiple CLA types (VHF/UHF TNC, UHF IQ, S-band IQ, X-band IQ)
- Link quality monitoring (RSSI, SNR, BER)

**Key Files:**
- `cla.go` - CLA interface definition
- `mock_cla.go` - Mock CLA implementation for testing

### `pkg/node` - Node Controller
Top-level orchestrator that ties together BPA, store, contact plan, and CLA. Handles:
- Autonomous operation cycle (store-check-deliver loop)
- Contact window execution
- Ping request/response handling
- Health monitoring and telemetry collection

**Key Files:**
- `types.go` - Node configuration and health/statistics types
- `controller.go` - Node controller implementation

## Architecture

The core infrastructure follows a layered architecture:

```
┌─────────────────────────────────────┐
│      Node Controller (pkg/node)     │  ← Top-level orchestrator
├─────────────────────────────────────┤
│  BPA (pkg/bpa)  │  Contact Manager  │  ← Bundle operations & scheduling
│                 │  (pkg/contact)    │
├─────────────────┴───────────────────┤
│      Bundle Store (pkg/store)       │  ← Persistent storage
├─────────────────────────────────────┤
│         CLA (pkg/cla)               │  ← AX.25/LTP abstraction
├─────────────────────────────────────┤
│           ION-DTN                   │  ← NASA JPL's DTN implementation
└─────────────────────────────────────┘
```

## ION-DTN Integration

This Go code **wraps** ION-DTN, it does not reimplement it. ION-DTN provides:
- BPv7 bundle protocol
- LTP (Licklider Transmission Protocol)
- Bundle storage and persistence
- Priority handling and lifetime enforcement
- BPSec security (RFC 9172)
- CGR (Contact Graph Routing) for orbital pass prediction

Our Go code provides:
- Configuration file generation for ION-DTN
- Node orchestration and lifecycle management
- Telemetry collection and health monitoring
- Hardware interfaces (TNC4, B200mini SDR, IQ transceivers)
- Integration testing

## Property-Based Testing

The core infrastructure includes property-based tests using [gopter](https://github.com/leanovate/gopter) to validate correctness properties:

### Bundle Store Properties
- **Property 1**: Bundle Store/Retrieve Round-Trip (Requirement 2.2)
- **Property 3**: Priority Ordering Invariant (Requirements 2.3, 5.3)
- **Property 4**: Eviction Policy Ordering (Requirements 2.4, 2.5)
- **Property 5**: Store Capacity Bound (Requirement 2.6)

### Bundle Protocol Agent Properties
- **Property 2**: Bundle Validation Correctness (Requirements 1.1, 1.2, 1.3)
- **Property 7**: Ping Echo Correctness (Requirements 4.1, 4.2)

## Running Tests

```bash
# Run all tests
go test ./...

# Run property-based tests only
go test -v ./pkg/store -run Property
go test -v ./pkg/bpa -run Property

# Run with more iterations for thorough testing
go test -v ./pkg/store -run Property -gopter.minSuccessfulTests=1000
```

## Dependencies

```bash
# Install dependencies
go mod download

# Property-based testing library
go get github.com/leanovate/gopter
```

## Next Steps

This completes Task 1 (Core DTN Infrastructure). The next tasks involve:
- Task 3: Phase 1 - Terrestrial DTN Validation (RPi + TNC4 + FT-817)
- Task 5: Phase 2 - CubeSat Engineering Model (STM32U585 + B200mini)
- Task 7: Phase 3 - LEO CubeSat Flight (STM32U585 + flight IQ transceiver)
- Task 9: Phase 4 - Cislunar Deep-Space Communication

See `.kiro/specs/cislunar-amateur-dtn-payload/tasks.md` for the complete implementation plan.
