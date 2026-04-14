# ION-DTN Configuration Generators

This package provides configuration file generators for ION-DTN nodes across all phases of the Cislunar Amateur DTN Payload project.

## Overview

The configuration generators create ION-DTN configuration files (ionrc, ltprc, bprc, ipnrc) for different node types:

- **Terrestrial nodes**: RPi + Mobilinkd TNC4 + FT-817 (Phase 1)
- **Engineering Model (EM) nodes**: STM32U585 + Ettus B200mini SDR (Phase 2)
- **LEO CubeSat nodes**: STM32U585 + Flight IQ Transceiver (Phase 3)

## LEO Configuration Generator

### Features

The LEO configuration generator (`leo.go`) provides:

1. **CGR-Predicted Contact Windows**: Automatically generates contact plans from orbital parameters using ION-DTN's Contact Graph Routing (CGR) engine
2. **Flight IQ Transceiver Support**: Configures for STM32U585 + flight-qualified IQ transceiver (UHF 437 MHz, 9.6 kbps)
3. **TLE/Ephemeris Updates**: Supports re-prediction of contact windows when fresh orbital data is received
4. **Autonomous Store-and-Forward**: Configures for autonomous operation during orbital passes
5. **Telemetry Endpoints**: Optional telemetry configuration for monitoring

### Usage

#### Basic LEO Configuration

```go
import (
    "time"
    "terrestrial-dtn/pkg/ionconfig"
)

// Define orbital parameters (from TLE)
orbitalParams := &ionconfig.OrbitalParameters{
    Epoch:           time.Now(),
    SemiMajorAxisM:  6771000.0,  // ~400 km altitude
    Eccentricity:    0.0005,
    InclinationDeg:  51.6,
    RAANDeg:         45.0,
    ArgPeriapsisDeg: 0.0,
    TrueAnomalyDeg:  0.0,
}

// Define predicted contact windows (from CGR)
contacts := []ionconfig.LEOContact{
    {
        RemoteNodeNumber: 1,
        RemoteCallsign:   "KA1ABC",
        StartTime:        time.Now().Add(10 * time.Minute),
        Duration:         8 * time.Minute,
        DataRate:         9600,
        MaxElevationDeg:  45.0,
        Confidence:       0.95,
    },
}

// Generate configuration
config := ionconfig.LEONodeConfig{
    NodeID:           "leo-cubesat-01",
    NodeNumber:       10,
    Callsign:         "KL0SAT",
    StorageBytes:     128 * 1024 * 1024, // 128 MB NVM
    SRAMBytes:        786 * 1024,         // 786 KB SRAM
    ContactPlan:      contacts,
    OrbitalParams:    orbitalParams,
    TelemetryEnabled: true,
}

err := ionconfig.GenerateLEOConfig(config, "./configs/leo-node")
```

#### With CGR Prediction

```go
import (
    "terrestrial-dtn/pkg/contact"
    "terrestrial-dtn/pkg/ionconfig"
)

// Define ground stations
groundStations := []contact.GroundStationLocation{
    {
        StationID:       "gs-alpha",
        LatitudeDeg:     37.4,
        LongitudeDeg:    -122.1,
        AltitudeM:       100.0,
        MinElevationDeg: 10.0,
    },
}

// Use CGR to predict passes
cgrParams := &contact.OrbitalParameters{
    Epoch:           time.Now().Unix(),
    SemiMajorAxisM:  6771000.0,
    Eccentricity:    0.0005,
    InclinationDeg:  51.6,
    RAANDeg:         45.0,
    ArgPeriapsisDeg: 0.0,
    TrueAnomalyDeg:  0.0,
}

predictedContacts, err := contact.PredictLEOPasses(
    cgrParams,
    groundStations,
    time.Now(),
    time.Now().Add(24 * time.Hour),
    30, // 30-second time step
)

// Convert to LEO contact format and generate config
// (see examples/leo_config_generation.go for full example)
```

#### Updating Contact Plan

When fresh TLE/ephemeris data is received:

```go
// Re-run CGR prediction with updated orbital parameters
newOrbitalParams := &ionconfig.OrbitalParameters{
    Epoch:           newEpoch,
    SemiMajorAxisM:  6771000.0,
    Eccentricity:    0.0005,
    InclinationDeg:  51.6,
    RAANDeg:         46.0, // Updated RAAN
    ArgPeriapsisDeg: 0.0,
    TrueAnomalyDeg:  0.0,
}

// Get new predicted contacts from CGR
newContacts := []ionconfig.LEOContact{ /* ... */ }

// Update contact plan
err := ionconfig.UpdateLEOContactPlan(
    "./configs/leo-node",
    newContacts,
    newOrbitalParams,
)
```

### Generated Files

The LEO configuration generator creates:

1. **node.ionrc**: ION initialization with node ID and memory configuration
2. **node.ltprc**: LTP engine configuration with spans for each predicted contact
3. **node.bprc**: Bundle Protocol configuration with endpoints and outducts
4. **node.ipnrc**: IPN scheme routing plans for each contact
5. **leo.ionconfig**: LEO-specific configuration including:
   - Flight IQ transceiver parameters (UHF 437 MHz, 9.6 kbps)
   - Orbital parameters for CGR re-prediction
   - CGR-predicted contact windows with elevation and confidence
   - Telemetry configuration

### Configuration Parameters

#### LEONodeConfig

- `NodeID`: Unique identifier for the node
- `NodeNumber`: IPN node number (e.g., 10)
- `Callsign`: Amateur radio callsign (e.g., "KL0SAT")
- `StorageBytes`: External NVM capacity (typically 128 MB for STM32U585)
- `SRAMBytes`: Available SRAM (786 KB for STM32U585)
- `ContactPlan`: List of CGR-predicted contact windows
- `OrbitalParams`: TLE-derived orbital parameters for re-prediction
- `TelemetryEnabled`: Enable telemetry endpoint (ipn:X.10)

#### LEOContact

- `RemoteNodeNumber`: Ground station node number
- `RemoteCallsign`: Ground station callsign
- `StartTime`: Contact window start time
- `Duration`: Contact window duration (typically 5-10 minutes for LEO)
- `DataRate`: Link data rate (9600 bps for UHF)
- `MaxElevationDeg`: Peak elevation angle during pass
- `Confidence`: Prediction confidence (0.0-1.0, decreases with time from epoch)

#### OrbitalParameters

- `Epoch`: Reference time for orbital elements
- `SemiMajorAxisM`: Semi-major axis in meters
- `Eccentricity`: Orbital eccentricity (0 = circular, <1 = elliptical)
- `InclinationDeg`: Orbital inclination in degrees
- `RAANDeg`: Right Ascension of Ascending Node in degrees
- `ArgPeriapsisDeg`: Argument of periapsis in degrees
- `TrueAnomalyDeg`: True anomaly at epoch in degrees

## Examples

See `examples/leo_config_generation.go` for a complete example that:
1. Defines orbital parameters for a LEO CubeSat
2. Uses CGR to predict passes over ground stations
3. Generates ION-DTN configuration files
4. Demonstrates TLE/ephemeris update workflow

Run the example:
```bash
go run examples/leo_config_generation.go
```

## Requirements

This implementation satisfies:
- **Requirement 13.1**: LEO flight node configuration for STM32U585 + flight IQ transceiver
- **Requirement 13.3**: Autonomous store-and-forward during orbital passes
- **Requirement 8.5**: CGR-predicted contact windows with TLE/ephemeris updates

## Testing

Run tests:
```bash
go test ./pkg/ionconfig -v
```

Tests cover:
- Basic LEO configuration generation
- CGR-predicted contact window integration
- Configuration with/without orbital parameters
- Contact plan updates
- Telemetry endpoint configuration
