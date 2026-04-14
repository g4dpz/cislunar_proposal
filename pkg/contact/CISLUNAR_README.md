# Cislunar CGR-Based Pass Prediction

This document describes the cislunar orbital parameter support and pass prediction functionality added to the `pkg/contact` package.

## Overview

The contact package now supports both LEO (Low Earth Orbit) and cislunar orbital pass prediction using ION-DTN's Contact Graph Routing (CGR) methodology. The system automatically determines orbit type based on semi-major axis and applies appropriate propagation algorithms.

## Key Features

### 1. Automatic Orbit Type Detection

The system automatically determines whether an orbit is LEO or cislunar based on the semi-major axis:

- **LEO**: Semi-major axis < 8,000 km (altitude < ~1,600 km)
- **Cislunar**: Semi-major axis ≥ 8,000 km (includes lunar orbit, Earth-Moon L-points, etc.)

```go
params := &OrbitalParameters{
    SemiMajorAxisM: 384400000.0, // ~384,400 km (Earth-Moon distance)
    // ... other parameters
}

orbitType := params.DetermineOrbitType() // Returns OrbitTypeCislunar
```

### 2. Cislunar Orbit Propagation

The `PropagateCislunarOrbit` function implements simplified two-body propagation suitable for cislunar orbits:

- Uses Earth's gravitational parameter for propagation
- Accounts for much larger orbital radii (Earth-Moon distance ~384,400 km vs LEO ~500 km)
- Longer orbital periods (20-30 days vs 90-100 minutes for LEO)
- Simplified model suitable for contact prediction over short horizons (days)

```go
state, err := PropagateCislunarOrbit(params, targetTime)
// Returns position and velocity in ECI coordinates
```

### 3. Light-Time Delay Computation

Cislunar distances result in significant light-time delays (1-2 seconds one-way):

```go
delay := ComputeLightTimeDelay(384400.0) // Earth-Moon distance in km
// Returns ~1.282 seconds one-way delay
// Round-trip time: ~2.564 seconds
```

### 4. Cislunar Pass Prediction

The `PredictCislunarPasses` function generates contact windows for cislunar payloads:

**Key Differences from LEO:**
- **Longer contact windows**: Hours instead of minutes (10-12 hours typical)
- **Slower time step**: 60 seconds (vs 30 seconds for LEO)
- **Lower data rate**: 500 bps S-band with BPSK+LDPC (vs 9600 bps UHF for LEO)
- **Faster confidence degradation**: Due to lunar perturbations
- **Lower minimum elevation**: 5° typical (vs 10° for LEO)

```go
predicted, err := PredictCislunarPasses(
    params,
    stations,
    fromTime,
    toTime,
    60, // 60-second time step
)
```

### 5. Confidence Degradation

Cislunar orbits experience faster confidence degradation due to:
- Lunar gravitational perturbations
- Earth-Moon system dynamics
- Longer propagation times

**Confidence Model:**
- LEO: `confidence = exp(-t / 10 days)`
- Cislunar: `confidence = exp(-t / 5 days)` (degrades 2x faster)

| Time from Epoch | LEO Confidence | Cislunar Confidence |
|-----------------|----------------|---------------------|
| 0 days          | 1.000          | 1.000               |
| 1 day           | 0.905          | 0.819               |
| 3 days          | 0.741          | 0.549               |
| 7 days          | 0.497          | 0.247               |

**Recommendation**: Update cislunar ephemeris data more frequently than LEO (every 2-3 days vs weekly for LEO).

### 6. Unified Pass Prediction Interface

The `PredictPasses` function automatically selects the appropriate propagation method:

```go
// Automatically uses LEO or cislunar propagation based on semi-major axis
predicted, err := PredictPasses(params, stations, fromTime, toTime, timeStep)
```

## Cislunar Contact Window Characteristics

### Data Rate and Link Type
- **Data Rate**: 500 bps (vs 9600 bps for LEO)
- **Link Type**: `LinkTypeSBandIQ` (S-band 2.2 GHz with IQ baseband)
- **Modulation**: BPSK + LDPC/Turbo coding for deep-space links

### Contact Duration
- **Typical Duration**: 10-12 hours (vs 5-10 minutes for LEO)
- **Minimum Duration**: 5 minutes (300 seconds)
- **Reason**: Much slower angular velocity at lunar distance

### Ground Station Requirements
- **Tier 3 Stations**: 3-5m dish antennas required
- **Minimum Elevation**: 5° (lower than LEO's 10°)
- **Frequency**: S-band 2.2 GHz (vs UHF 437 MHz for LEO)

## Usage Examples

### Basic Cislunar Pass Prediction

```go
package main

import (
    "fmt"
    "time"
    "terrestrial-dtn/pkg/contact"
)

func main() {
    // Define cislunar orbital parameters
    params := &contact.OrbitalParameters{
        Epoch:           time.Now().Unix(),
        SemiMajorAxisM:  384400000.0, // ~384,400 km
        Eccentricity:    0.05,
        InclinationDeg:  5.0,
        RAANDeg:         0.0,
        ArgPeriapsisDeg: 0.0,
        TrueAnomalyDeg:  0.0,
    }

    // Define Tier 3 ground station
    stations := []contact.GroundStationLocation{
        {
            StationID:       "gs-tier3",
            LatitudeDeg:     35.0,
            LongitudeDeg:    -106.0,
            AltitudeM:       1500.0,
            MinElevationDeg: 5.0,
        },
    }

    // Predict passes over 7 days
    fromTime := time.Now()
    toTime := fromTime.Add(7 * 24 * time.Hour)

    predicted, err := contact.PredictCislunarPasses(
        params,
        stations,
        fromTime,
        toTime,
        60, // 60-second time step
    )
    if err != nil {
        panic(err)
    }

    // Process predicted contacts
    for _, pc := range predicted {
        fmt.Printf("Pass: %s to %s\n",
            time.Unix(pc.Window.StartTime, 0),
            time.Unix(pc.Window.EndTime, 0))
        fmt.Printf("  Duration: %.1f hours\n",
            float64(pc.Window.EndTime-pc.Window.StartTime)/3600.0)
        fmt.Printf("  Max Elevation: %.1f°\n", pc.MaxElevationDeg)
        fmt.Printf("  Confidence: %.2f\n", pc.Confidence)
    }
}
```

### Using ContactPlanManager

```go
cpm := contact.NewContactPlanManager()

// Load initial plan
plan := &contact.ContactPlan{
    PlanID:      1,
    GeneratedAt: time.Now().Unix(),
    ValidFrom:   time.Now().Unix(),
    ValidTo:     time.Now().Add(7 * 24 * time.Hour).Unix(),
    Contacts:    []contact.ContactWindow{},
}
cpm.LoadPlan(plan)

// Predict contacts (automatically detects cislunar orbit)
predicted, err := cpm.PredictContacts(
    params,
    stations,
    time.Now().Unix(),
    time.Now().Add(7*24*time.Hour).Unix(),
)
```

## Implementation Details

### Orbital Propagation

The cislunar propagation uses simplified two-body Keplerian dynamics:

1. **Mean Motion**: Computed using Earth's gravitational parameter
2. **Mean Anomaly**: Propagated forward in time
3. **Kepler's Equation**: Solved using Newton-Raphson iteration
4. **True Anomaly**: Computed from eccentric anomaly
5. **Position/Velocity**: Transformed from perifocal to ECI coordinates

**Simplifications:**
- No J2 perturbation (negligible at cislunar distances)
- No lunar gravity (simplified for short-term prediction)
- Two-body problem (Earth-centric)

**Validity**: Suitable for contact prediction over horizons of days to weeks. For longer horizons or high-precision requirements, use full ephemeris data.

### Constants

```go
const (
    moonRadiusKm    = 1737.4      // Moon mean radius in km
    moonMuKm3PerS2  = 4902.8      // Moon gravitational parameter
    earthMoonDistKm = 384400.0    // Mean Earth-Moon distance in km
    speedOfLightKmS = 299792.458  // Speed of light in km/s
)
```

## Testing

Comprehensive tests are provided in `cislunar_test.go`:

- `TestDetermineOrbitType`: Orbit type classification
- `TestPropagateCislunarOrbit`: Orbital propagation accuracy
- `TestComputeLightTimeDelay`: Light-time delay computation
- `TestPredictCislunarPasses`: Full pass prediction
- `TestComputeCislunarConfidence`: Confidence degradation
- `TestCislunarConfidenceDegradation`: Comparison with LEO
- `TestPredictPasses_AutomaticSelection`: Automatic orbit type selection
- `TestCislunarOrbitalPeriod`: Orbital period validation

Run tests:
```bash
go test ./pkg/contact/... -v -run Cislunar
```

## Requirements Validation

This implementation satisfies the following requirements from the cislunar-amateur-dtn-payload spec:

- **8.1**: CGR-based contact prediction using orbital parameters ✓
- **8.2**: Elevation threshold filtering (5° minimum for cislunar) ✓
- **8.3**: Sorted output by start time ✓
- **8.4**: Confidence degradation over time ✓
- **8.5**: Re-computation on fresh orbital parameters ✓
- **8.6**: No overlapping predicted windows ✓
- **8.7**: Time horizon boundary enforcement ✓

## Future Enhancements

Potential improvements for production use:

1. **Full Ephemeris Integration**: Use JPL SPICE kernels for high-precision lunar orbit propagation
2. **Lunar Gravity**: Include lunar gravitational perturbations for longer-term predictions
3. **Earth-Moon L-Points**: Explicit support for Lagrange point station-keeping orbits
4. **Solar Radiation Pressure**: Account for SRP effects on cislunar trajectories
5. **Eclipse Prediction**: Compute Earth/Moon shadow entry/exit times
6. **Link Budget Integration**: Incorporate range-dependent link margin calculations

## References

- ION-DTN Contact Graph Routing documentation
- Vallado, D. A. (2013). *Fundamentals of Astrodynamics and Applications*
- NASA JPL SPICE Toolkit for high-precision ephemeris
- Amateur Radio Cislunar Communication: S-band link budgets and modulation schemes
