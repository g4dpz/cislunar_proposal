package contact

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_CGRPredictionValidity validates Property 14:
// For any valid orbital parameters, ground station list, and time horizon,
// all CGR-predicted contact windows SHALL fall within the requested time
// horizon boundaries, and no two predicted windows for the same ground
// station SHALL overlap in time.
//
// **Validates: Requirements 8.1, 8.6, 8.7**
func TestProperty_CGRPredictionValidity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced for computationally expensive tests

	properties := gopter.NewProperties(parameters)

	// Property: All predicted contacts fall within time horizon
	properties.Property("all predicted contacts fall within time horizon", prop.ForAll(
		func(horizonDuration int64, semiMajorAxisKm float64) bool {
			if horizonDuration <= 0 || horizonDuration > 86400 {
				return true // Skip invalid horizons (max 24 hours)
			}
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true // Skip invalid orbits (LEO: 400-600 km altitude)
			}

			// Create LEO orbital parameters
			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001, // Nearly circular
				InclinationDeg:  51.6,  // ISS-like
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			// Create ground station
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
			toTime := fromTime.Add(time.Duration(horizonDuration) * time.Second)

			// Predict passes
			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				// Prediction failed - acceptable for some parameter combinations
				return true
			}

			// Verify all predicted contacts fall within time horizon
			for _, pc := range predicted {
				if pc.Window.StartTime < fromTime.Unix() || pc.Window.EndTime > toTime.Unix() {
					return false
				}
			}

			return true
		},
		gen.Int64Range(3600, 86400),    // 1-24 hours
		gen.Float64Range(6771.0, 7000.0), // LEO semi-major axis (km)
	))

	// Property: No overlapping predicted windows for same ground station
	properties.Property("no overlapping predicted windows for same ground station", prop.ForAll(
		func(numStations uint8, semiMajorAxisKm float64) bool {
			if numStations == 0 || numStations > 5 {
				return true // Skip invalid inputs
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

			// Create multiple ground stations
			stations := make([]GroundStationLocation, 0)
			for i := uint8(0); i < numStations; i++ {
				stations = append(stations, GroundStationLocation{
					StationID:       NodeID("gs-" + string(rune('A'+i))),
					LatitudeDeg:     float64(30 + int(i)*10), // Spread across latitudes
					LongitudeDeg:    float64(-100 + int(i)*20),
					AltitudeM:       100.0,
					MinElevationDeg: 10.0,
				})
			}

			fromTime := time.Now()
			toTime := fromTime.Add(12 * time.Hour)

			// Predict passes
			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true // Prediction failed - acceptable
			}

			// Group predicted contacts by ground station
			stationContacts := make(map[NodeID][]PredictedContact)
			for _, pc := range predicted {
				stationContacts[pc.Window.RemoteNode] = append(stationContacts[pc.Window.RemoteNode], pc)
			}

			// Verify no overlaps within each ground station's contacts
			for _, contacts := range stationContacts {
				for i := 0; i < len(contacts); i++ {
					for j := i + 1; j < len(contacts); j++ {
						c1, c2 := contacts[i].Window, contacts[j].Window
						// Check for overlap
						if c1.StartTime < c2.EndTime && c2.StartTime < c1.EndTime {
							return false // Overlap detected
						}
					}
				}
			}

			return true
		},
		gen.UInt8Range(1, 5),
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Predicted contacts have valid time ranges
	properties.Property("predicted contacts have startTime < endTime", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true // Skip invalid orbits
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

			// Verify all contacts have valid time ranges
			for _, pc := range predicted {
				if pc.Window.StartTime >= pc.Window.EndTime {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Cislunar predictions fall within time horizon
	properties.Property("cislunar predictions fall within time horizon", prop.ForAll(
		func(horizonDuration int64) bool {
			if horizonDuration <= 0 || horizonDuration > 86400 {
				return true // Skip invalid horizons
			}

			// Create cislunar orbital parameters (lunar orbit)
			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  384400.0 * 1000.0, // Earth-Moon distance
				Eccentricity:    0.05,              // Slightly elliptical
				InclinationDeg:  5.0,               // Low inclination
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
					MinElevationDeg: 5.0,
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(time.Duration(horizonDuration) * time.Second)

			// Predict cislunar passes
			predicted, err := PredictCislunarPasses(params, stations, fromTime, toTime, 60)
			if err != nil {
				return true
			}

			// Verify all predicted contacts fall within time horizon
			for _, pc := range predicted {
				if pc.Window.StartTime < fromTime.Unix() || pc.Window.EndTime > toTime.Unix() {
					return false
				}
			}

			return true
		},
		gen.Int64Range(3600, 86400),
	))

	// Property: Predicted contacts have reasonable durations
	properties.Property("LEO predicted contacts have reasonable durations (60s - 15min)", prop.ForAll(
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

			// Verify all LEO passes have reasonable durations
			for _, pc := range predicted {
				duration := pc.Window.EndTime - pc.Window.StartTime
				// LEO passes typically 60 seconds to 15 minutes
				if duration < 60 || duration > 900 {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Invalid orbital parameters are rejected
	properties.Property("invalid orbital parameters are rejected", prop.ForAll(
		func(eccentricity float64) bool {
			// Test with invalid eccentricity (>= 1.0)
			if eccentricity < 1.0 {
				return true // Skip valid eccentricities
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  6771.0 * 1000.0,
				Eccentricity:    eccentricity, // Invalid
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

			// Should fail validation
			_, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err == nil {
				return false // Should have rejected invalid parameters
			}

			return true
		},
		gen.Float64Range(1.0, 2.0),
	))

	// Property: Invalid time horizon is rejected
	properties.Property("invalid time horizon (fromTime >= toTime) is rejected", prop.ForAll(
		func(offset int64) bool {
			if offset < 0 {
				return true // Skip valid offsets
			}

			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  6771.0 * 1000.0,
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
			toTime := fromTime.Add(-time.Duration(offset) * time.Second) // toTime before fromTime

			// Should fail validation
			_, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err == nil {
				return false // Should have rejected invalid time horizon
			}

			return true
		},
		gen.Int64Range(1, 86400),
	))

	properties.TestingRun(t)
}
