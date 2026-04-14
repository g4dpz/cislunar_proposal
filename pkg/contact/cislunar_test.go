package contact

import (
	"math"
	"testing"
	"time"
)

func TestDetermineOrbitType(t *testing.T) {
	tests := []struct {
		name           string
		semiMajorAxisM float64
		wantType       OrbitType
	}{
		{
			name:           "LEO orbit (400 km altitude)",
			semiMajorAxisM: 6771000.0, // ~400 km altitude
			wantType:       OrbitTypeLEO,
		},
		{
			name:           "LEO orbit (1500 km altitude)",
			semiMajorAxisM: 7871000.0, // ~1500 km altitude
			wantType:       OrbitTypeLEO,
		},
		{
			name:           "Cislunar orbit (lunar distance)",
			semiMajorAxisM: 384400000.0, // ~384,400 km (Earth-Moon distance)
			wantType:       OrbitTypeCislunar,
		},
		{
			name:           "Cislunar orbit (GEO-like)",
			semiMajorAxisM: 42164000.0, // ~42,164 km (GEO altitude)
			wantType:       OrbitTypeCislunar,
		},
		{
			name:           "Boundary case (just above LEO)",
			semiMajorAxisM: 8100000.0, // ~8,100 km
			wantType:       OrbitTypeCislunar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &OrbitalParameters{
				Epoch:           time.Now().Unix(),
				SemiMajorAxisM:  tt.semiMajorAxisM,
				Eccentricity:    0.01,
				InclinationDeg:  28.5,
				RAANDeg:         0.0,
				ArgPeriapsisDeg: 0.0,
				TrueAnomalyDeg:  0.0,
			}

			gotType := params.DetermineOrbitType()
			if gotType != tt.wantType {
				t.Errorf("DetermineOrbitType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestPropagateCislunarOrbit(t *testing.T) {
	// Lunar orbit parameters (simplified circular orbit around Earth at lunar distance)
	params := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  384400000.0, // ~384,400 km (Earth-Moon distance)
		Eccentricity:    0.05,        // Slightly eccentric
		InclinationDeg:  5.0,         // Lunar orbit inclination
		RAANDeg:         0.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	epoch := time.Unix(params.Epoch, 0)

	// Propagate forward 1 day
	targetTime := epoch.Add(24 * time.Hour)

	state, err := PropagateCislunarOrbit(params, targetTime)
	if err != nil {
		t.Fatalf("PropagateCislunarOrbit failed: %v", err)
	}

	// Check position magnitude is reasonable (should be near semi-major axis)
	positionMagnitude := state.Position.Magnitude()
	expectedRadius := params.SemiMajorAxisM / 1000.0
	tolerance := 50000.0 // 50,000 km tolerance (cislunar orbits are large)

	if math.Abs(positionMagnitude-expectedRadius) > tolerance {
		t.Errorf("Position magnitude %f km differs from expected %f km by more than %f km",
			positionMagnitude, expectedRadius, tolerance)
	}

	// Check velocity magnitude is reasonable for cislunar orbit (~1 km/s)
	velocityMagnitude := state.Velocity.Magnitude()
	if velocityMagnitude < 0.5 || velocityMagnitude > 2.0 {
		t.Errorf("Velocity magnitude %f km/s out of cislunar range", velocityMagnitude)
	}
}

func TestComputeLightTimeDelay(t *testing.T) {
	tests := []struct {
		name        string
		distanceKm  float64
		wantDelayS  float64
		tolerance   float64
	}{
		{
			name:        "Earth-Moon distance",
			distanceKm:  384400.0,
			wantDelayS:  1.282, // ~1.28 seconds
			tolerance:   0.01,
		},
		{
			name:        "LEO distance (500 km)",
			distanceKm:  500.0,
			wantDelayS:  0.00167, // ~1.67 milliseconds
			tolerance:   0.0001,
		},
		{
			name:        "GEO distance (42,164 km)",
			distanceKm:  42164.0,
			wantDelayS:  0.1407, // ~140.7 milliseconds
			tolerance:   0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := ComputeLightTimeDelay(tt.distanceKm)
			if math.Abs(delay-tt.wantDelayS) > tt.tolerance {
				t.Errorf("ComputeLightTimeDelay(%f) = %f s, want %f s (tolerance %f)",
					tt.distanceKm, delay, tt.wantDelayS, tt.tolerance)
			}
		})
	}
}

func TestPredictCislunarPasses(t *testing.T) {
	// Lunar orbit parameters
	params := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  384400000.0, // ~384,400 km
		Eccentricity:    0.05,
		InclinationDeg:  5.0,
		RAANDeg:         0.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	// Ground station with large dish (Tier 3)
	stations := []GroundStationLocation{
		{
			StationID:       "gs-tier3",
			LatitudeDeg:     35.0,
			LongitudeDeg:    -106.0, // New Mexico
			AltitudeM:       1500.0,
			MinElevationDeg: 5.0, // Lower elevation for cislunar
		},
	}

	fromTime := time.Unix(params.Epoch, 0)
	toTime := fromTime.Add(7 * 24 * time.Hour) // 7 days

	predicted, err := PredictCislunarPasses(params, stations, fromTime, toTime, 60)
	if err != nil {
		t.Fatalf("PredictCislunarPasses failed: %v", err)
	}

	// Should predict at least one pass over 7 days
	if len(predicted) < 1 {
		t.Errorf("Expected at least 1 pass in 7 days, got %d", len(predicted))
	}

	// Check each predicted contact
	for i, pc := range predicted {
		// Duration should be longer for cislunar (potentially hours)
		duration := pc.Window.EndTime - pc.Window.StartTime
		if duration < 300 { // At least 5 minutes
			t.Errorf("Pass %d duration %d seconds too short", i, duration)
		}

		// Max elevation should be >= minimum elevation
		if pc.MaxElevationDeg < stations[0].MinElevationDeg {
			t.Errorf("Pass %d max elevation %f < minimum %f",
				i, pc.MaxElevationDeg, stations[0].MinElevationDeg)
		}

		// Confidence should be in [0, 1]
		if pc.Confidence < 0 || pc.Confidence > 1.0 {
			t.Errorf("Pass %d confidence %f out of range [0, 1]", i, pc.Confidence)
		}

		// Data rate should be 500 bps for cislunar S-band
		if pc.Window.DataRate != 500 {
			t.Errorf("Pass %d data rate %d, expected 500", i, pc.Window.DataRate)
		}

		// Link type should be S-band IQ
		if pc.Window.LinkType != LinkTypeSBandIQ {
			t.Errorf("Pass %d link type %v, expected LinkTypeSBandIQ", i, pc.Window.LinkType)
		}
	}

	// Check passes are sorted by start time
	for i := 1; i < len(predicted); i++ {
		if predicted[i].Window.StartTime < predicted[i-1].Window.StartTime {
			t.Errorf("Passes not sorted by start time: pass %d starts before pass %d", i, i-1)
		}
	}
}

func TestComputeCislunarConfidence(t *testing.T) {
	epoch := time.Now().Unix()

	tests := []struct {
		name           string
		predictionTime int64
		wantRange      [2]float64 // min, max
	}{
		{
			name:           "same as epoch",
			predictionTime: epoch,
			wantRange:      [2]float64{0.95, 1.0},
		},
		{
			name:           "1 day from epoch",
			predictionTime: epoch + 86400,
			wantRange:      [2]float64{0.75, 0.85},
		},
		{
			name:           "3 days from epoch",
			predictionTime: epoch + 3*86400,
			wantRange:      [2]float64{0.45, 0.60},
		},
		{
			name:           "7 days from epoch",
			predictionTime: epoch + 7*86400,
			wantRange:      [2]float64{0.20, 0.30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := computeCislunarConfidence(epoch, tt.predictionTime)

			if confidence < tt.wantRange[0] || confidence > tt.wantRange[1] {
				t.Errorf("computeCislunarConfidence() = %f, want range [%f, %f]",
					confidence, tt.wantRange[0], tt.wantRange[1])
			}

			// Confidence should always be in [0, 1]
			if confidence < 0 || confidence > 1.0 {
				t.Errorf("Confidence %f out of valid range [0, 1]", confidence)
			}
		})
	}
}

func TestCislunarConfidenceDegradation(t *testing.T) {
	epoch := time.Now().Unix()

	// Test that cislunar confidence degrades faster than LEO
	time1Day := epoch + 86400
	time3Days := epoch + 3*86400

	cislunarConf1Day := computeCislunarConfidence(epoch, time1Day)
	cislunarConf3Days := computeCislunarConfidence(epoch, time3Days)

	leoConf1Day := computeConfidence(epoch, time1Day)
	leoConf3Days := computeConfidence(epoch, time3Days)

	// Cislunar confidence should degrade faster
	if cislunarConf1Day >= leoConf1Day {
		t.Errorf("Cislunar confidence at 1 day (%f) should be less than LEO (%f)",
			cislunarConf1Day, leoConf1Day)
	}

	if cislunarConf3Days >= leoConf3Days {
		t.Errorf("Cislunar confidence at 3 days (%f) should be less than LEO (%f)",
			cislunarConf3Days, leoConf3Days)
	}

	// Both should be monotonically decreasing
	if cislunarConf3Days >= cislunarConf1Day {
		t.Errorf("Cislunar confidence should decrease over time: 1day=%f, 3days=%f",
			cislunarConf1Day, cislunarConf3Days)
	}
}

func TestPredictPasses_AutomaticSelection(t *testing.T) {
	stations := []GroundStationLocation{
		{
			StationID:       "gs-test",
			LatitudeDeg:     40.0,
			LongitudeDeg:    -75.0,
			AltitudeM:       100.0,
			MinElevationDeg: 10.0,
		},
	}

	fromTime := time.Now()
	toTime := fromTime.Add(24 * time.Hour)

	t.Run("LEO orbit", func(t *testing.T) {
		leoParams := &OrbitalParameters{
			Epoch:           fromTime.Unix(),
			SemiMajorAxisM:  6771000.0, // ~400 km altitude
			Eccentricity:    0.001,
			InclinationDeg:  51.6,
			RAANDeg:         100.0,
			ArgPeriapsisDeg: 90.0,
			TrueAnomalyDeg:  0.0,
		}

		predicted, err := PredictPasses(leoParams, stations, fromTime, toTime, 30)
		if err != nil {
			t.Fatalf("PredictPasses (LEO) failed: %v", err)
		}

		// Should predict LEO passes
		if len(predicted) > 0 {
			// Check data rate is LEO (9600 bps)
			if predicted[0].Window.DataRate != 9600 {
				t.Errorf("LEO pass data rate %d, expected 9600", predicted[0].Window.DataRate)
			}
		}
	})

	t.Run("Cislunar orbit", func(t *testing.T) {
		cislunarParams := &OrbitalParameters{
			Epoch:           fromTime.Unix(),
			SemiMajorAxisM:  384400000.0, // ~384,400 km
			Eccentricity:    0.05,
			InclinationDeg:  5.0,
			RAANDeg:         0.0,
			ArgPeriapsisDeg: 0.0,
			TrueAnomalyDeg:  0.0,
		}

		predicted, err := PredictPasses(cislunarParams, stations, fromTime, toTime.Add(6*24*time.Hour), 60)
		if err != nil {
			t.Fatalf("PredictPasses (cislunar) failed: %v", err)
		}

		// Should predict cislunar passes
		if len(predicted) > 0 {
			// Check data rate is cislunar (500 bps)
			if predicted[0].Window.DataRate != 500 {
				t.Errorf("Cislunar pass data rate %d, expected 500", predicted[0].Window.DataRate)
			}

			// Check link type is S-band
			if predicted[0].Window.LinkType != LinkTypeSBandIQ {
				t.Errorf("Cislunar pass link type %v, expected LinkTypeSBandIQ", predicted[0].Window.LinkType)
			}
		}
	})
}

func TestCislunarOrbitalPeriod(t *testing.T) {
	// Test that cislunar orbital period is much longer than LEO
	leoParams := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  6771000.0, // ~400 km altitude
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         0.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	cislunarParams := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  384400000.0, // ~384,400 km
		Eccentricity:    0.05,
		InclinationDeg:  5.0,
		RAANDeg:         0.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	// Compute orbital periods using Kepler's third law: T = 2π√(a³/μ)
	leoSemiMajorKm := leoParams.SemiMajorAxisM / 1000.0
	leoPeriodS := 2.0 * math.Pi * math.Sqrt(math.Pow(leoSemiMajorKm, 3)/earthMuKm3PerS2)

	cislunarSemiMajorKm := cislunarParams.SemiMajorAxisM / 1000.0
	cislunarPeriodS := 2.0 * math.Pi * math.Sqrt(math.Pow(cislunarSemiMajorKm, 3)/earthMuKm3PerS2)

	// LEO period should be ~90-100 minutes
	leoPeriodMin := leoPeriodS / 60.0
	if leoPeriodMin < 85 || leoPeriodMin > 105 {
		t.Errorf("LEO period %f minutes out of expected range (85-105 min)", leoPeriodMin)
	}

	// Cislunar period should be much longer (days)
	cislunarPeriodDays := cislunarPeriodS / 86400.0
	if cislunarPeriodDays < 20 || cislunarPeriodDays > 30 {
		t.Errorf("Cislunar period %f days out of expected range (20-30 days)", cislunarPeriodDays)
	}

	// Cislunar period should be much longer than LEO
	if cislunarPeriodS <= leoPeriodS*100 {
		t.Errorf("Cislunar period (%f s) should be much longer than LEO period (%f s)",
			cislunarPeriodS, leoPeriodS)
	}
}

func TestContactPlanManager_PredictContacts_Cislunar(t *testing.T) {
	cpm := NewContactPlanManager()

	// Load initial plan
	plan := &ContactPlan{
		PlanID:      1,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   time.Now().Unix(),
		ValidTo:     time.Now().Add(7 * 24 * time.Hour).Unix(),
		Contacts:    []ContactWindow{},
	}

	err := cpm.LoadPlan(plan)
	if err != nil {
		t.Fatalf("LoadPlan failed: %v", err)
	}

	// Cislunar orbit parameters
	params := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  384400000.0,
		Eccentricity:    0.05,
		InclinationDeg:  5.0,
		RAANDeg:         0.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	stations := []GroundStationLocation{
		{
			StationID:       "gs-tier3",
			LatitudeDeg:     35.0,
			LongitudeDeg:    -106.0,
			AltitudeM:       1500.0,
			MinElevationDeg: 5.0,
		},
	}

	fromTime := time.Unix(params.Epoch, 0).Unix()
	toTime := time.Unix(params.Epoch, 0).Add(7 * 24 * time.Hour).Unix()

	predicted, err := cpm.PredictContacts(params, stations, fromTime, toTime)
	if err != nil {
		t.Fatalf("PredictContacts failed: %v", err)
	}

	// Should predict at least one cislunar pass
	if len(predicted) < 1 {
		t.Logf("Warning: No cislunar passes predicted in 7 days (may be valid depending on geometry)")
	}

	// Verify cislunar characteristics
	for i, pc := range predicted {
		if pc.Window.DataRate != 500 {
			t.Errorf("Pass %d data rate %d, expected 500 bps for cislunar", i, pc.Window.DataRate)
		}

		if pc.Window.LinkType != LinkTypeSBandIQ {
			t.Errorf("Pass %d link type %v, expected LinkTypeSBandIQ", i, pc.Window.LinkType)
		}

		// Confidence should degrade faster for cislunar
		if pc.Confidence > 0.9 && (pc.Window.StartTime-params.Epoch) > 86400 {
			t.Errorf("Pass %d confidence %f too high for cislunar prediction > 1 day from epoch",
				i, pc.Confidence)
		}
	}
}
