# Contact Plan Manager with CGR-Based Pass Prediction

This package provides contact plan management and CGR (Contact Graph Routing) based orbital pass prediction for LEO satellites.

## Overview

The Contact Plan Manager is responsible for:
- Managing scheduled communication windows between DTN nodes
- Predicting LEO satellite passes over ground stations using orbital mechanics
- Supporting TLE (Two-Line Element) orbital parameter updates
- Providing contact window queries for autonomous store-and-forward operations

## Key Features

### CGR-Based Pass Prediction

The package implements simplified SGP4/SDP4-style orbit propagation for LEO satellites:

- **Keplerian Orbit Propagation**: Propagates satellite position and velocity using classical orbital elements
- **J2 Perturbation**: Accounts for Earth oblateness effects on RAAN precession
- **Line-of-Sight Computation**: Calculates elevation and azimuth from ground stations
- **Doppler Shift Calculation**: Computes frequency shift at carrier frequency (437 MHz UHF)
- **Elevation Threshold Filtering**: Returns only passes exceeding minimum elevation (typically 10-15°)
- **Confidence Degradation**: Assigns confidence values that decrease with time from TLE epoch

### Typical LEO Pass Characteristics

For a 400 km altitude LEO satellite (ISS-like orbit):
- **Pass Duration**: 5-10 minutes
- **Passes per Day**: 4-6 per ground station
- **Data Rate**: 9.6 kbps (UHF 437 MHz)
- **Max Doppler Shift**: ±10 kHz at 437 MHz

## Usage

### Basic Pass Prediction

```go
import (
    "time"
    "terrestrial-dtn/pkg/contact"
)

// Define orbital parameters (from TLE or computed)
params := &contact.OrbitalParameters{
    Epoch:           time.Now().Unix(),
    SemiMajorAxisM:  6771000.0, // ~400 km altitude
    Eccentricity:    0.001,
    InclinationDeg:  51.6,
    RAANDeg:         100.0,
    ArgPeriapsisDeg: 90.0,
    TrueAnomalyDeg:  0.0,
}

// Define ground stations
stations := []contact.GroundStationLocation{
    {
        StationID:       "gs-alpha",
        LatitudeDeg:     42.36,
        LongitudeDeg:    -71.06,
        AltitudeM:       50.0,
        MinElevationDeg: 10.0,
    },
}

// Predict passes over next 24 hours
fromTime := time.Now()
toTime := fromTime.Add(24 * time.Hour)

predicted, err := contact.PredictLEOPasses(params, stations, fromTime, toTime, 30)
if err != nil {
    // Handle error
}

// Use predicted contacts
for _, pc := range predicted {
    fmt.Printf("Pass: %s to %s, max elevation %.1f°\n",
        time.Unix(pc.Window.StartTime, 0),
        time.Unix(pc.Window.EndTime, 0),
        pc.MaxElevationDeg)
}
```

### Using TLE Parameters

```go
// Define TLE parameters (standard satellite orbital data format)
tle := &contact.TLEParameters{
    Epoch:            time.Now(),
    MeanMotionRevDay: 15.5, // ~93 min period
    Eccentricity:     0.001,
    InclinationDeg:   51.6,
    RAANDeg:          100.0,
    ArgPerigeeDeg:    90.0,
    MeanAnomalyDeg:   45.0,
    BStarDrag:        0.0001,
}

// Validate TLE
if err := tle.Validate(); err != nil {
    // Handle invalid TLE
}

// Convert to OrbitalParameters
params := tle.ToOrbitalParameters()

// Use for prediction...
```

### Contact Plan Manager Integration

```go
// Create contact plan manager
cpm := contact.NewContactPlanManager()

// Load initial plan
plan := &contact.ContactPlan{
    PlanID:      1,
    GeneratedAt: time.Now().Unix(),
    ValidFrom:   time.Now().Unix(),
    ValidTo:     time.Now().Add(48 * time.Hour).Unix(),
    Contacts:    []contact.ContactWindow{},
}

err := cpm.LoadPlan(plan)
if err != nil {
    // Handle error
}

// Update plan with CGR predictions
err = cpm.UpdateContactPlanWithPredictions(
    "leo-sat-01",
    params,
    stations,
    fromTime,
    toTime,
)
if err != nil {
    // Handle error
}

// Query next contact with a destination
nextContact, err := cpm.GetNextContact("gs-alpha", time.Now().Unix())
if err != nil {
    // No contact available
}

// Query active contacts
activeContacts := cpm.GetActiveContacts(time.Now().Unix())
```

### Updating Orbital Parameters

```go
// Update orbital parameters (e.g., after receiving fresh TLE data)
err := cpm.UpdateOrbitalParameters("leo-sat-01", newParams)
if err != nil {
    // Handle error
}

// Re-predict contacts with updated parameters
err = cpm.UpdateContactPlanWithPredictions(
    "leo-sat-01",
    newParams,
    stations,
    fromTime,
    toTime,
)
```

## Data Structures

### OrbitalParameters

Classical Keplerian orbital elements:
- `Epoch`: Reference time (Unix seconds)
- `SemiMajorAxisM`: Semi-major axis (meters)
- `Eccentricity`: Orbital eccentricity (0 = circular, <1 = elliptical)
- `InclinationDeg`: Orbital inclination (degrees)
- `RAANDeg`: Right Ascension of Ascending Node (degrees)
- `ArgPeriapsisDeg`: Argument of periapsis (degrees)
- `TrueAnomalyDeg`: True anomaly at epoch (degrees)

### TLEParameters

Two-Line Element set format (standard for satellite tracking):
- `Epoch`: TLE epoch time
- `MeanMotionRevDay`: Mean motion (revolutions per day)
- `Eccentricity`: Orbital eccentricity
- `InclinationDeg`: Inclination (degrees)
- `RAANDeg`: Right Ascension of Ascending Node (degrees)
- `ArgPerigeeDeg`: Argument of perigee (degrees)
- `MeanAnomalyDeg`: Mean anomaly at epoch (degrees)
- `BStarDrag`: Drag term (1/earth radii)

### GroundStationLocation

Ground station geodetic coordinates:
- `StationID`: Unique identifier
- `LatitudeDeg`: Geodetic latitude (degrees)
- `LongitudeDeg`: Geodetic longitude (degrees)
- `AltitudeM`: Altitude above WGS84 ellipsoid (meters)
- `MinElevationDeg`: Minimum elevation for valid contact (degrees)

### PredictedContact

CGR-predicted contact window:
- `Window`: ContactWindow with start/end times, data rate, link type
- `MaxElevationDeg`: Peak elevation angle during pass
- `DopplerShiftHz`: Maximum Doppler shift at carrier frequency
- `Confidence`: Prediction confidence (0.0 to 1.0)

## Implementation Details

### Orbit Propagation

The implementation uses simplified Keplerian propagation with J2 perturbation:

1. **Mean Anomaly Propagation**: Advances mean anomaly using mean motion
2. **Kepler's Equation**: Solves for eccentric anomaly using Newton-Raphson
3. **True Anomaly**: Computes true anomaly from eccentric anomaly
4. **Position/Velocity**: Transforms from perifocal to ECI coordinates
5. **J2 Perturbation**: Accounts for RAAN precession due to Earth oblateness

### Coordinate Transformations

- **ECEF to ECI**: Accounts for Earth rotation (GMST)
- **Geodetic to ECEF**: Converts ground station lat/lon/alt to Cartesian
- **Topocentric SEZ**: Computes elevation and azimuth from ground station

### Pass Detection

The algorithm propagates the orbit forward in time steps (default 30 seconds):

1. For each time step, compute satellite position in ECI
2. Transform ground station position to ECI
3. Compute elevation angle from ground station to satellite
4. Detect rising edge (elevation crosses minimum threshold)
5. Track maximum elevation during pass
6. Detect falling edge (elevation drops below threshold)
7. Create PredictedContact for passes > 60 seconds duration

### Confidence Computation

Confidence decreases exponentially with time from TLE epoch:

```
confidence = exp(-Δt / 10 days)
```

Where Δt is the time difference between prediction and epoch. This reflects increasing uncertainty due to:
- Atmospheric drag variations
- Solar radiation pressure
- Gravitational perturbations
- TLE age

Typical confidence values:
- Same day: 0.95-1.0
- 1 day: 0.85-0.95
- 7 days: 0.4-0.6
- 14 days: 0.2-0.35

## Requirements Validation

This implementation satisfies the following requirements from the design document:

### Requirement 8.1: CGR Contact Prediction
✓ Uses orbital parameters and ground station locations to compute predicted contact windows

### Requirement 8.2: Elevation Threshold
✓ Returns only contacts where maximum elevation ≥ ground station minimum elevation

### Requirement 8.3: Sorted Output
✓ Predicted contacts are sorted by start time in ascending order

### Requirement 8.4: Confidence Degradation
✓ Assigns confidence values that decrease for predictions far from TLE epoch

### Requirement 8.5: TLE/Ephemeris Updates
✓ Supports updating orbital parameters and re-predicting contact windows

### Requirement 8.6: No Overlapping Windows
✓ Pass detection ensures no overlapping windows for the same ground station

### Requirement 8.7: Time Horizon Boundaries
✓ All predicted contacts fall within requested [fromTime, toTime] range

## Testing

Run the test suite:

```bash
go test ./pkg/contact/... -v
```

Run the example:

```bash
go run examples/cgr_pass_prediction.go
```

## Future Enhancements

Potential improvements for production use:

1. **ION-DTN Integration**: Direct integration with ION-DTN's CGR engine
2. **Higher-Order Perturbations**: Add atmospheric drag, solar radiation pressure
3. **SGP4/SDP4 Library**: Use established SGP4 library for improved accuracy
4. **TLE Parsing**: Direct parsing of standard TLE format strings
5. **Cislunar Support**: Extend to lunar orbit propagation
6. **Multi-Satellite**: Optimize for predicting passes of multiple satellites
7. **Real-Time Updates**: Support for real-time TLE updates from tracking networks

## References

- NASA JPL ION-DTN: https://sourceforge.net/projects/ion-dtn/
- SGP4 Orbit Propagation: Vallado, D. A., "Fundamentals of Astrodynamics and Applications"
- TLE Format: https://celestrak.org/NORAD/documentation/tle-fmt.php
- Contact Graph Routing: Burleigh, S., "Contact Graph Routing"
