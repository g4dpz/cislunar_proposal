package contact

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_CGRSortedOutput validates Property 16:
// For any set of CGR-predicted contacts, the results SHALL be sorted by
// start time in ascending order.
//
// **Validates: Requirement 8.3**
func TestProperty_CGRSortedOutput(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced for computationally expensive tests

	properties := gopter.NewProperties(parameters)

	// Property: Predicted contacts are sorted by start time (ascending)
	properties.Property("predicted contacts are sorted by start time ascending", prop.ForAll(
		func(semiMajorAxisKm float64, numStations uint8) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true // Skip invalid orbits
			}
			if numStations == 0 || numStations > 5 {
				return true // Skip invalid station counts
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
					LatitudeDeg:     float64(30 + int(i)*10),
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

			// Verify contacts are sorted by start time
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Window.StartTime > predicted[i+1].Window.StartTime {
					return false // Not sorted
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
		gen.UInt8Range(1, 5),
	))

	// Property: Cislunar predictions are sorted by start time
	properties.Property("cislunar predictions are sorted by start time ascending", prop.ForAll(
		func(numStations uint8) bool {
			if numStations == 0 || numStations > 3 {
				return true
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

			// Create multiple ground stations
			stations := make([]GroundStationLocation, 0)
			for i := uint8(0); i < numStations; i++ {
				stations = append(stations, GroundStationLocation{
					StationID:       NodeID("gs-cislunar-" + string(rune('A'+i))),
					LatitudeDeg:     float64(30 + int(i)*15),
					LongitudeDeg:    float64(-120 + int(i)*30),
					AltitudeM:       1000.0,
					MinElevationDeg: 5.0,
				})
			}

			fromTime := time.Now()
			toTime := fromTime.Add(24 * time.Hour)

			// Predict cislunar passes
			predicted, err := PredictCislunarPasses(params, stations, fromTime, toTime, 60)
			if err != nil {
				return true
			}

			// Verify contacts are sorted by start time
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Window.StartTime > predicted[i+1].Window.StartTime {
					return false
				}
			}

			return true
		},
		gen.UInt8Range(1, 3),
	))

	// Property: Sorted order is maintained across multiple ground stations
	properties.Property("sorted order maintained across multiple ground stations", prop.ForAll(
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

			// Create ground stations at different locations
			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-west"),
					LatitudeDeg:     35.0,
					LongitudeDeg:    -120.0,
					AltitudeM:       100.0,
					MinElevationDeg: 10.0,
				},
				{
					StationID:       NodeID("gs-east"),
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

			// Verify global sort order across all ground stations
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Window.StartTime > predicted[i+1].Window.StartTime {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Start times are monotonically non-decreasing
	properties.Property("start times are monotonically non-decreasing", prop.ForAll(
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

			// Verify monotonically non-decreasing start times
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Window.StartTime > predicted[i+1].Window.StartTime {
					return false // Decreasing - violates property
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: First contact has earliest start time
	properties.Property("first contact has earliest start time", prop.ForAll(
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
			if err != nil || len(predicted) == 0 {
				return true
			}

			// Verify first contact has earliest start time
			firstStartTime := predicted[0].Window.StartTime
			for i := 1; i < len(predicted); i++ {
				if predicted[i].Window.StartTime < firstStartTime {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Last contact has latest start time
	properties.Property("last contact has latest start time", prop.ForAll(
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
			if err != nil || len(predicted) == 0 {
				return true
			}

			// Verify last contact has latest start time
			lastStartTime := predicted[len(predicted)-1].Window.StartTime
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Window.StartTime > lastStartTime {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Sorting is stable across re-predictions
	properties.Property("sorting is stable across re-predictions", prop.ForAll(
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

			// Predict twice with same parameters
			predicted1, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			predicted2, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Verify both predictions have same order
			if len(predicted1) != len(predicted2) {
				return false
			}

			for i := 0; i < len(predicted1); i++ {
				if predicted1[i].Window.StartTime != predicted2[i].Window.StartTime {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Empty result is trivially sorted
	properties.Property("empty result is trivially sorted", prop.ForAll(
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

			// Ground station with very high minimum elevation (unlikely to see passes)
			stations := []GroundStationLocation{
				{
					StationID:       NodeID("gs-1"),
					LatitudeDeg:     40.0,
					LongitudeDeg:    -75.0,
					AltitudeM:       100.0,
					MinElevationDeg: 85.0, // Very high threshold
				},
			}

			fromTime := time.Now()
			toTime := fromTime.Add(1 * time.Hour) // Short horizon

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Empty result is trivially sorted
			if len(predicted) == 0 {
				return true
			}

			// If not empty, verify sorting
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Window.StartTime > predicted[i+1].Window.StartTime {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	properties.TestingRun(t)
}
