package contact

import (
	"math"
	"testing"
	"time"
)

func TestTLEParametersValidation(t *testing.T) {
	tests := []struct {
		name    string
		tle     TLEParameters
		wantErr bool
	}{
		{
			name: "valid LEO TLE",
			tle: TLEParameters{
				Epoch:            time.Now(),
				MeanMotionRevDay: 15.5,
				Eccentricity:     0.001,
				InclinationDeg:   51.6,
				RAANDeg:          100.0,
				ArgPerigeeDeg:    90.0,
				MeanAnomalyDeg:   45.0,
				BStarDrag:        0.0001,
			},
			wantErr: false,
		},
		{
			name: "invalid eccentricity",
			tle: TLEParameters{
				Epoch:            time.Now(),
				MeanMotionRevDay: 15.5,
				Eccentricity:     1.5,
				InclinationDeg:   51.6,
			},
			wantErr: true,
		},
		{
			name: "invalid mean motion (too low for LEO)",
			tle: TLEParameters{
				Epoch:            time.Now(),
				MeanMotionRevDay: 5.0,
				Eccentricity:     0.001,
				InclinationDeg:   51.6,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tle.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TLEParameters.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTLEToOrbitalParameters(t *testing.T) {
	tle := TLEParameters{
		Epoch:            time.Now(),
		MeanMotionRevDay: 15.5, // ~93 min period, typical LEO
		Eccentricity:     0.001,
		InclinationDeg:   51.6,
		RAANDeg:          100.0,
		ArgPerigeeDeg:    90.0,
		MeanAnomalyDeg:   45.0,
		BStarDrag:        0.0001,
	}

	params := tle.ToOrbitalParameters()

	// Check semi-major axis is reasonable for LEO (~6800-7000 km from Earth center)
	semiMajorAxisKm := params.SemiMajorAxisM / 1000.0
	if semiMajorAxisKm < 6700 || semiMajorAxisKm > 7200 {
		t.Errorf("Semi-major axis %f km out of LEO range", semiMajorAxisKm)
	}

	// Check other parameters are preserved
	if params.Eccentricity != tle.Eccentricity {
		t.Errorf("Eccentricity mismatch: got %f, want %f", params.Eccentricity, tle.Eccentricity)
	}
	if params.InclinationDeg != tle.InclinationDeg {
		t.Errorf("Inclination mismatch: got %f, want %f", params.InclinationDeg, tle.InclinationDeg)
	}
}

func TestPropagateOrbit(t *testing.T) {
	// ISS-like orbit parameters
	params := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  6771000.0, // ~400 km altitude
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         100.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	epoch := time.Unix(params.Epoch, 0)

	// Propagate forward 1 orbit period (~93 minutes)
	targetTime := epoch.Add(93 * time.Minute)

	state, err := PropagateOrbit(params, targetTime)
	if err != nil {
		t.Fatalf("PropagateOrbit failed: %v", err)
	}

	// Check position magnitude is reasonable (should be near semi-major axis)
	positionMagnitude := state.Position.Magnitude()
	expectedRadius := params.SemiMajorAxisM / 1000.0
	tolerance := 100.0 // 100 km tolerance

	if math.Abs(positionMagnitude-expectedRadius) > tolerance {
		t.Errorf("Position magnitude %f km differs from expected %f km by more than %f km",
			positionMagnitude, expectedRadius, tolerance)
	}

	// Check velocity magnitude is reasonable for LEO (~7.7 km/s)
	velocityMagnitude := state.Velocity.Magnitude()
	if velocityMagnitude < 7.0 || velocityMagnitude > 8.5 {
		t.Errorf("Velocity magnitude %f km/s out of LEO range", velocityMagnitude)
	}
}

func TestLatLonToECEF(t *testing.T) {
	// Test with known location: equator at prime meridian
	station := GroundStationLocation{
		StationID:       "test-gs",
		LatitudeDeg:     0.0,
		LongitudeDeg:    0.0,
		AltitudeM:       0.0,
		MinElevationDeg: 10.0,
	}

	ecef := LatLonToECEF(station)

	// At equator, prime meridian: X should be ~Earth radius, Y and Z should be ~0
	if math.Abs(ecef.Position.X-earthRadiusKm) > 1.0 {
		t.Errorf("ECEF X coordinate %f differs from Earth radius %f", ecef.Position.X, earthRadiusKm)
	}
	if math.Abs(ecef.Position.Y) > 1.0 {
		t.Errorf("ECEF Y coordinate %f should be near 0", ecef.Position.Y)
	}
	if math.Abs(ecef.Position.Z) > 1.0 {
		t.Errorf("ECEF Z coordinate %f should be near 0", ecef.Position.Z)
	}
}

func TestComputeElevationAzimuth(t *testing.T) {
	// Satellite directly overhead at 400 km altitude
	station := GroundStationLocation{
		StationID:       "test-gs",
		LatitudeDeg:     0.0,
		LongitudeDeg:    0.0,
		AltitudeM:       0.0,
		MinElevationDeg: 10.0,
	}

	gsECEF := LatLonToECEF(station)
	testTime := time.Now()

	// Satellite position: directly above ground station at 400 km
	satECI := ECEFToECI(Vector3D{
		X: (earthRadiusKm + 400.0),
		Y: 0.0,
		Z: 0.0,
	}, testTime)

	elevation, _ := ComputeElevationAzimuth(satECI, gsECEF, testTime)

	// Elevation should be close to 90 degrees (directly overhead)
	if math.Abs(elevation-90.0) > 5.0 {
		t.Errorf("Elevation %f degrees should be near 90 degrees for overhead satellite", elevation)
	}
}

func TestPredictLEOPasses(t *testing.T) {
	// ISS-like orbit
	params := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  6771000.0, // ~400 km altitude
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         100.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	// Ground station at mid-latitude
	stations := []GroundStationLocation{
		{
			StationID:       "gs-alpha",
			LatitudeDeg:     40.0,
			LongitudeDeg:    -75.0,
			AltitudeM:       100.0,
			MinElevationDeg: 10.0,
		},
	}

	fromTime := time.Unix(params.Epoch, 0)
	toTime := fromTime.Add(24 * time.Hour)

	predicted, err := PredictLEOPasses(params, stations, fromTime, toTime, 30)
	if err != nil {
		t.Fatalf("PredictLEOPasses failed: %v", err)
	}

	// Should predict multiple passes over 24 hours (typically 4-6 for LEO)
	if len(predicted) < 2 {
		t.Errorf("Expected at least 2 passes in 24 hours, got %d", len(predicted))
	}

	// Check each predicted contact
	for i, pc := range predicted {
		// Duration should be 5-10 minutes for LEO
		duration := pc.Window.EndTime - pc.Window.StartTime
		if duration < 60 || duration > 900 {
			t.Errorf("Pass %d duration %d seconds out of expected range (60-900s)", i, duration)
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

		// Data rate should be 9600 bps for LEO UHF
		if pc.Window.DataRate != 9600 {
			t.Errorf("Pass %d data rate %d, expected 9600", i, pc.Window.DataRate)
		}
	}

	// Check passes are sorted by start time
	for i := 1; i < len(predicted); i++ {
		if predicted[i].Window.StartTime < predicted[i-1].Window.StartTime {
			t.Errorf("Passes not sorted by start time: pass %d starts before pass %d", i, i-1)
		}
	}
}

func TestUpdateContactPlanWithPredictions(t *testing.T) {
	cpm := NewContactPlanManager()

	// Load initial plan
	plan := &ContactPlan{
		PlanID:      1,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   time.Now().Unix(),
		ValidTo:     time.Now().Add(48 * time.Hour).Unix(),
		Contacts:    []ContactWindow{},
	}

	err := cpm.LoadPlan(plan)
	if err != nil {
		t.Fatalf("LoadPlan failed: %v", err)
	}

	// ISS-like orbit
	params := &OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  6771000.0,
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         100.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	stations := []GroundStationLocation{
		{
			StationID:       "gs-alpha",
			LatitudeDeg:     40.0,
			LongitudeDeg:    -75.0,
			AltitudeM:       100.0,
			MinElevationDeg: 10.0,
		},
	}

	fromTime := time.Unix(params.Epoch, 0)
	toTime := fromTime.Add(24 * time.Hour)

	err = cpm.UpdateContactPlanWithPredictions("leo-sat-01", params, stations, fromTime, toTime)
	if err != nil {
		t.Fatalf("UpdateContactPlanWithPredictions failed: %v", err)
	}

	// Check that predicted contacts were added
	if len(cpm.plan.PredictedContacts) == 0 {
		t.Error("No predicted contacts added to plan")
	}

	// Check that contacts were also added to regular contacts list
	if len(cpm.plan.Contacts) == 0 {
		t.Error("No contacts added to plan")
	}

	// Check orbital data was stored
	if cpm.plan.OrbitalData == nil {
		t.Error("Orbital data not stored in plan")
	}
}

func TestComputeConfidence(t *testing.T) {
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
			wantRange:      [2]float64{0.85, 0.95},
		},
		{
			name:           "7 days from epoch",
			predictionTime: epoch + 7*86400,
			wantRange:      [2]float64{0.4, 0.6},
		},
		{
			name:           "14 days from epoch",
			predictionTime: epoch + 14*86400,
			wantRange:      [2]float64{0.2, 0.35},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := computeConfidence(epoch, tt.predictionTime)

			if confidence < tt.wantRange[0] || confidence > tt.wantRange[1] {
				t.Errorf("computeConfidence() = %f, want range [%f, %f]",
					confidence, tt.wantRange[0], tt.wantRange[1])
			}

			// Confidence should always be in [0, 1]
			if confidence < 0 || confidence > 1.0 {
				t.Errorf("Confidence %f out of valid range [0, 1]", confidence)
			}
		})
	}
}

func TestVector3DOperations(t *testing.T) {
	v1 := Vector3D{X: 1.0, Y: 2.0, Z: 3.0}
	v2 := Vector3D{X: 4.0, Y: 5.0, Z: 6.0}

	// Test magnitude
	mag := v1.Magnitude()
	expected := math.Sqrt(1.0 + 4.0 + 9.0)
	if math.Abs(mag-expected) > 1e-10 {
		t.Errorf("Magnitude() = %f, want %f", mag, expected)
	}

	// Test subtract
	diff := v2.Subtract(v1)
	if diff.X != 3.0 || diff.Y != 3.0 || diff.Z != 3.0 {
		t.Errorf("Subtract() = %+v, want {3, 3, 3}", diff)
	}

	// Test dot product
	dot := v1.Dot(v2)
	expectedDot := 1.0*4.0 + 2.0*5.0 + 3.0*6.0
	if math.Abs(dot-expectedDot) > 1e-10 {
		t.Errorf("Dot() = %f, want %f", dot, expectedDot)
	}
}
