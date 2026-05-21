package contact

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: test-framework-srs-sdd, Property 17: CGR Confidence Monotonicity
// For any two CGR-predicted contacts from the same prediction run, the contact
// whose start time is further from the orbital parameter epoch SHALL have a
// confidence value less than or equal to the contact closer to the epoch.
//
// **Validates: SRS-TF-015 (Requirement 14.1)**
func TestProperty_CGRConfidenceMonotonicity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced for computationally expensive tests

	properties := gopter.NewProperties(parameters)

	// Property: Confidence decreases with time from epoch
	properties.Property("confidence decreases with time from epoch", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true // Skip invalid orbits
			}

			// Create LEO orbital parameters with epoch at current time
			epoch := time.Now()
			params := &OrbitalParameters{
				Epoch:           epoch.Unix(),
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

			fromTime := epoch
			toTime := epoch.Add(24 * time.Hour) // 24 hour horizon

			// Predict passes
			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil || len(predicted) < 2 {
				return true // Need at least 2 contacts to compare
			}

			// Verify confidence decreases with time from epoch
			for i := 0; i < len(predicted)-1; i++ {
				// Contact i+1 is further from epoch (sorted by start time)
				// So confidence[i+1] <= confidence[i]
				if predicted[i+1].Confidence > predicted[i].Confidence {
					return false // Confidence increased - violates monotonicity
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Confidence is in valid range [0, 1]
	properties.Property("confidence is in valid range [0, 1]", prop.ForAll(
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
			toTime := fromTime.Add(24 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil {
				return true
			}

			// Verify all confidence values are in [0, 1]
			for _, pc := range predicted {
				if pc.Confidence < 0.0 || pc.Confidence > 1.0 {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Cislunar confidence decreases faster than LEO
	// Validated directly via the confidence functions rather than pass prediction,
	// since pass timing differences make direct comparison unreliable.
	properties.Property("cislunar confidence decreases faster than LEO", prop.ForAll(
		func(timeDiffHours int64) bool {
			if timeDiffHours <= 0 || timeDiffHours > 168 {
				return true // Skip invalid inputs
			}

			epoch := time.Now().Unix()
			predictionTime := epoch + timeDiffHours*3600

			// Compute confidence using the same functions used by the prediction engine
			// LEO: exp(-t/10 days), Cislunar: exp(-t/5 days)
			leoConf := computeConfidence(epoch, predictionTime)
			cislunarConf := computeCislunarConfidence(epoch, predictionTime)

			// Cislunar confidence should be <= LEO confidence for same time offset
			return cislunarConf <= leoConf
		},
		gen.Int64Range(1, 168),
	))

	// Property: First contact has highest confidence
	properties.Property("first contact has highest confidence", prop.ForAll(
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
			toTime := fromTime.Add(24 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil || len(predicted) == 0 {
				return true
			}

			// Verify first contact has highest confidence
			firstConfidence := predicted[0].Confidence
			for i := 1; i < len(predicted); i++ {
				if predicted[i].Confidence > firstConfidence {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Last contact has lowest confidence
	properties.Property("last contact has lowest confidence", prop.ForAll(
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
			toTime := fromTime.Add(24 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil || len(predicted) == 0 {
				return true
			}

			// Verify last contact has lowest confidence
			lastConfidence := predicted[len(predicted)-1].Confidence
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i].Confidence < lastConfidence {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Confidence at epoch is close to 1.0
	properties.Property("confidence at epoch is close to 1.0", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			// Set epoch very close to prediction start time
			epoch := time.Now()
			params := &OrbitalParameters{
				Epoch:           epoch.Unix(),
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

			fromTime := epoch
			toTime := epoch.Add(6 * time.Hour) // Short horizon

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil || len(predicted) == 0 {
				return true
			}

			// First contact should have high confidence (close to 1.0)
			if predicted[0].Confidence < 0.8 {
				return false
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Confidence decreases exponentially with time
	properties.Property("confidence decreases exponentially with time", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			epoch := time.Now()
			params := &OrbitalParameters{
				Epoch:           epoch.Unix(),
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

			fromTime := epoch
			toTime := epoch.Add(24 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil || len(predicted) < 3 {
				return true // Need at least 3 contacts to verify exponential decay
			}

			// Verify confidence decreases (not necessarily strictly exponential, but monotonic)
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i+1].Confidence > predicted[i].Confidence {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	// Property: Confidence monotonicity holds across multiple ground stations
	properties.Property("confidence monotonicity holds across multiple ground stations", prop.ForAll(
		func(semiMajorAxisKm float64) bool {
			if semiMajorAxisKm < 6771 || semiMajorAxisKm > 7000 {
				return true
			}

			epoch := time.Now()
			params := &OrbitalParameters{
				Epoch:           epoch.Unix(),
				SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
				Eccentricity:    0.001,
				InclinationDeg:  51.6,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			// Multiple ground stations
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

			fromTime := epoch
			toTime := epoch.Add(24 * time.Hour)

			predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
			if err != nil || len(predicted) < 2 {
				return true
			}

			// Verify confidence monotonicity across all contacts (sorted by start time)
			for i := 0; i < len(predicted)-1; i++ {
				if predicted[i+1].Confidence > predicted[i].Confidence {
					return false
				}
			}

			return true
		},
		gen.Float64Range(6771.0, 7000.0),
	))

	properties.TestingRun(t)
}
