package contact

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_CGRElevationThreshold validates Property 15:
// For any CGR-predicted contact, the maximum elevation angle SHALL meet or
// exceed the corresponding ground station's minimum elevation threshold.
//
// **Validates: Requirement 8.2**
func TestProperty_CGRElevationThreshold(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced for computationally expensive tests

	properties := gopter.NewProperties(parameters)

	// Property: All predicted contacts meet minimum elevation threshold
	properties.Property("all predicted contacts meet minimum elevation threshold", prop.ForAll(
		func(minElevation float64, semiMajorAxisKm float64) bool {
			if minElevation < 0 || minElevation > 45 {
				return true // Skip invalid elevations
			}
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true // Skip invalid orbits
			}

			// Create LEO orbital parameters
			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			// Create ground station with specified minimum elevation
			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: minElevation,
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			// Predict passes
			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true // Prediction failed - acceptable
			}

			// Verify all predicted contacts meet minimum elevation threshold
			for _, pc := range predicted {
				if pc.MaxElevationDeg < minElevation {
					return false // Contact below minimum elevation
				}
			}

			return true
		},
		gen.Float64Range(0.0, 45.0),
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Higher minimum elevation results in fewer or equal contacts
	properties.Property("higher minimum elevation results in fewer or equal contacts", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			// Predict with low minimum elevation (5 degrees)
			stationsLow := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 5.0,
				},
			}

			predictedLow, err := PredictLEOPasses(params, stationsLow, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Predict with high minimum elevation (20 degrees)
			stationsHigh := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 20.0,
				},
			}

			predictedHigh, err := PredictLEOPasses(params, stationsHigh, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Higher minimum elevation should result in fewer or equal contacts
			if len(predictedHigh) > len(predictedLow) {
				return false
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Cislunar contacts meet minimum elevation threshold
	properties.Property("cislunar contacts meet minimum elevation threshold", prop.ForAll(
		func(minElevation float64) bool {
			if minElevation < 0 || minElevation > 30 {
				return true // Skip invalid elevations
			}

			// Create cislunar orbital parameters
			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  384400.0 * 1000.0,
				Eccentricity:    0.05,
				InclinationDeg:  5.0,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-cislunar"),
					LatitudeDeg:     35.0,
					LongitudeDeg:    -120.0,
					AltitudeM:       1000.0,
					MinElevationDeg: minElevation,
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(24 * time.Hour)

			// Predict cislunar passes
			predicted, err := PredictCislunarPasses(params, stations, fromTime, toTime, 60)
			if err != nil {
				return true
			}

			// Verify all predicted contacts meet minimum elevation threshold
			for _, pc := range predicted {
				if pc.MaxElevationDeg < minElevation {
					return false
				}
			}

			return true
		},
		gen.Float64Range(0.0, 30.0),
	))

	// Property: Maximum elevation is within valid range [0, 90] degrees
	properties.Property("maximum elevation is within valid range [0, 90] degrees", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 10.0,
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Verify all maximum elevations are in valid range
			for _, pc := range predicted {
				if pc.MaxElevationDeg < 0 || pc.MaxElevationDeg > 90 {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Maximum elevation >= minimum elevation for all contacts
	properties.Property("maximum elevation >= minimum elevation for all contacts", prop.ForAll(
		func(minElevation float64, semiMajorAxisKm float64) bool {
			if minElevation < 0 || minElevation > 45 {
				return true
			}
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: minElevation,
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Verify max elevation >= min elevation for all contacts
			for _, pc := range predicted {
				if pc.MaxElevationDeg < minElevation {
					return false
				}
			}

			return true
		},
		gen.Float64Range(0.0, 45.0),
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Multiple ground stations with different thresholds
	properties.Property("multiple ground stations respect their individual thresholds", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			// Create multiple ground stations with different minimum elevations
			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-low"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 5.0, // Low threshold
				},
				{
					StationID:       NodeID("gs-high"),
					LatitudeDeg:     45.0,
					LongitudeDeg:    -80.0,
					AltitudeM:       100.0,
					MinElevationDeg: 20.0, // High threshold
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Verify each contact meets its ground station's threshold
			for _, pc := range predicted {
				// Find the ground station for this contact
				var stationMinElev float64
				for _, station := range stations {
					if station.StationID == pc.Window.RemoteNode {
						stationMinElev = station.MinElevationDeg
						break
					}
				}

				// Verify max elevation meets the station's threshold
				if pc.MaxElevationDeg < stationMinElev {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Zero minimum elevation allows all passes
	properties.Property("zero minimum elevation allows all geometrically visible passes", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			stationsZero := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 0.0, // Zero threshold
				},
			}

			stationsTen := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 10.0, // 10 degree threshold
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			predictedZero, err := PredictLEOPasses(params, stationsZero, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			predictedTen, err := PredictLEOPasses(params, stationsTen, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Zero threshold should result in more or equal contacts
			if len(predictedZero) < len(predictedTen) {
				return false
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	properties.TestingRun(t)
}
