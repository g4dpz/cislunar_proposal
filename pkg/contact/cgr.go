package contact

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// CGR-based contact prediction for LEO orbital passes
// This implements simplified SGP4/SDP4-style orbit propagation for LEO satellites
// to predict ground station contact windows

const (
	earthRadiusKm      = 6371.0  // Earth mean radius in km
	earthMuKm3PerS2    = 398600.4418 // Earth gravitational parameter (km^3/s^2)
	earthRotationRadS  = 7.2921159e-5 // Earth rotation rate (rad/s)
	degreesToRadians   = math.Pi / 180.0
	radiansToDegrees   = 180.0 / math.Pi
	
	// Cislunar constants
	moonRadiusKm       = 1737.4  // Moon mean radius in km
	moonMuKm3PerS2     = 4902.8  // Moon gravitational parameter (km^3/s^2)
	earthMoonDistKm    = 384400.0 // Mean Earth-Moon distance in km
	speedOfLightKmS    = 299792.458 // Speed of light in km/s
)

// TLEParameters represents Two-Line Element set orbital parameters
// This is the standard format for distributing satellite orbital data
type TLEParameters struct {
	Epoch            time.Time // TLE epoch time
	MeanMotionRevDay float64   // Mean motion (revolutions per day)
	Eccentricity     float64   // Orbital eccentricity
	InclinationDeg   float64   // Inclination (degrees)
	RAANDeg          float64   // Right Ascension of Ascending Node (degrees)
	ArgPerigeeDeg    float64   // Argument of perigee (degrees)
	MeanAnomalyDeg   float64   // Mean anomaly at epoch (degrees)
	BStarDrag        float64   // Drag term (1/earth radii)
}

// Validate checks if TLE parameters are valid for LEO
func (tle *TLEParameters) Validate() error {
	if tle.Eccentricity < 0 || tle.Eccentricity >= 1.0 {
		return fmt.Errorf("eccentricity must be in range [0, 1)")
	}
	if tle.InclinationDeg < 0 || tle.InclinationDeg > 180 {
		return fmt.Errorf("inclination must be in range [0, 180] degrees")
	}
	if tle.MeanMotionRevDay <= 0 {
		return fmt.Errorf("mean motion must be positive")
	}
	// LEO typically has mean motion > 11 rev/day (period < 130 min)
	if tle.MeanMotionRevDay < 11.0 {
		return fmt.Errorf("mean motion too low for LEO (< 11 rev/day)")
	}
	return nil
}

// ToOrbitalParameters converts TLE to OrbitalParameters format
func (tle *TLEParameters) ToOrbitalParameters() *OrbitalParameters {
	// Compute semi-major axis from mean motion
	// n = sqrt(mu / a^3), so a = (mu / n^2)^(1/3)
	meanMotionRadS := tle.MeanMotionRevDay * 2.0 * math.Pi / 86400.0
	semiMajorAxisKm := math.Pow(earthMuKm3PerS2/(meanMotionRadS*meanMotionRadS), 1.0/3.0)
	
	return &OrbitalParameters{
		Epoch:           tle.Epoch.Unix(),
		SemiMajorAxisM:  semiMajorAxisKm * 1000.0,
		Eccentricity:    tle.Eccentricity,
		InclinationDeg:  tle.InclinationDeg,
		RAANDeg:         tle.RAANDeg,
		ArgPeriapsisDeg: tle.ArgPerigeeDeg,
		TrueAnomalyDeg:  tle.MeanAnomalyDeg, // Approximation for low eccentricity
	}
}

// Vector3D represents a 3D position or velocity vector
type Vector3D struct {
	X, Y, Z float64
}

// Magnitude returns the magnitude of the vector
func (v Vector3D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Subtract returns v - other
func (v Vector3D) Subtract(other Vector3D) Vector3D {
	return Vector3D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// Dot returns the dot product of v and other
func (v Vector3D) Dot(other Vector3D) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// OrbitalState represents position and velocity at a specific time
type OrbitalState struct {
	Time     time.Time
	Position Vector3D // ECI coordinates (km)
	Velocity Vector3D // ECI coordinates (km/s)
}

// PropagateOrbit propagates orbital elements forward in time using simplified Keplerian propagation
// This is a simplified SGP4-style propagator suitable for LEO passes (5-10 min duration)
func PropagateOrbit(params *OrbitalParameters, targetTime time.Time) (*OrbitalState, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	epoch := time.Unix(params.Epoch, 0)
	deltaT := targetTime.Sub(epoch).Seconds()

	// Convert to km for calculations
	semiMajorAxisKm := params.SemiMajorAxisM / 1000.0
	
	// Compute mean motion (rad/s)
	meanMotion := math.Sqrt(earthMuKm3PerS2 / (semiMajorAxisKm * semiMajorAxisKm * semiMajorAxisKm))
	
	// Propagate mean anomaly
	meanAnomalyRad := params.TrueAnomalyDeg * degreesToRadians
	meanAnomalyRad += meanMotion * deltaT
	meanAnomalyRad = math.Mod(meanAnomalyRad, 2.0*math.Pi)
	
	// Solve Kepler's equation for eccentric anomaly (Newton-Raphson)
	eccentricAnomalyRad := solveKeplersEquation(meanAnomalyRad, params.Eccentricity)
	
	// Compute true anomaly
	trueAnomalyRad := 2.0 * math.Atan2(
		math.Sqrt(1.0+params.Eccentricity)*math.Sin(eccentricAnomalyRad/2.0),
		math.Sqrt(1.0-params.Eccentricity)*math.Cos(eccentricAnomalyRad/2.0),
	)
	
	// Compute position in orbital plane
	radius := semiMajorAxisKm * (1.0 - params.Eccentricity*math.Cos(eccentricAnomalyRad))
	
	// Position in perifocal coordinates
	posPerifocal := Vector3D{
		X: radius * math.Cos(trueAnomalyRad),
		Y: radius * math.Sin(trueAnomalyRad),
		Z: 0.0,
	}
	
	// Velocity in perifocal coordinates
	h := math.Sqrt(earthMuKm3PerS2 * semiMajorAxisKm * (1.0 - params.Eccentricity*params.Eccentricity))
	velPerifocal := Vector3D{
		X: -(earthMuKm3PerS2 / h) * math.Sin(trueAnomalyRad),
		Y: (earthMuKm3PerS2 / h) * (params.Eccentricity + math.Cos(trueAnomalyRad)),
		Z: 0.0,
	}
	
	// Transform to ECI coordinates
	// Account for RAAN precession due to J2 perturbation (simplified)
	raanRad := params.RAANDeg * degreesToRadians
	raanRad += computeRAANPrecession(params) * deltaT
	
	incRad := params.InclinationDeg * degreesToRadians
	argPerigeeRad := params.ArgPeriapsisDeg * degreesToRadians
	
	posECI := perifocalToECI(posPerifocal, raanRad, incRad, argPerigeeRad)
	velECI := perifocalToECI(velPerifocal, raanRad, incRad, argPerigeeRad)
	
	return &OrbitalState{
		Time:     targetTime,
		Position: posECI,
		Velocity: velECI,
	}, nil
}

// solveKeplersEquation solves Kepler's equation M = E - e*sin(E) for E using Newton-Raphson
func solveKeplersEquation(meanAnomaly, eccentricity float64) float64 {
	E := meanAnomaly // Initial guess
	tolerance := 1e-8
	maxIterations := 10
	
	for i := 0; i < maxIterations; i++ {
		f := E - eccentricity*math.Sin(E) - meanAnomaly
		fPrime := 1.0 - eccentricity*math.Cos(E)
		delta := f / fPrime
		E -= delta
		
		if math.Abs(delta) < tolerance {
			break
		}
	}
	
	return E
}

// computeRAANPrecession computes RAAN precession rate due to J2 perturbation (rad/s)
func computeRAANPrecession(params *OrbitalParameters) float64 {
	// J2 perturbation coefficient
	j2 := 1.08263e-3
	
	semiMajorAxisKm := params.SemiMajorAxisM / 1000.0
	incRad := params.InclinationDeg * degreesToRadians
	
	// RAAN precession rate (rad/s)
	n := math.Sqrt(earthMuKm3PerS2 / (semiMajorAxisKm * semiMajorAxisKm * semiMajorAxisKm))
	raanDot := -1.5 * j2 * n * math.Pow(earthRadiusKm/semiMajorAxisKm, 2.0) * math.Cos(incRad) /
		math.Pow(1.0-params.Eccentricity*params.Eccentricity, 2.0)
	
	return raanDot
}

// perifocalToECI transforms a vector from perifocal to ECI coordinates
func perifocalToECI(vec Vector3D, raan, inc, argPerigee float64) Vector3D {
	// Rotation matrix from perifocal to ECI
	cosRaan := math.Cos(raan)
	sinRaan := math.Sin(raan)
	cosInc := math.Cos(inc)
	sinInc := math.Sin(inc)
	cosArgP := math.Cos(argPerigee)
	sinArgP := math.Sin(argPerigee)
	
	// Combined rotation matrix elements
	r11 := cosRaan*cosArgP - sinRaan*sinArgP*cosInc
	r12 := -cosRaan*sinArgP - sinRaan*cosArgP*cosInc
	r21 := sinRaan*cosArgP + cosRaan*sinArgP*cosInc
	r22 := -sinRaan*sinArgP + cosRaan*cosArgP*cosInc
	r31 := sinArgP * sinInc
	r32 := cosArgP * sinInc
	
	return Vector3D{
		X: r11*vec.X + r12*vec.Y,
		Y: r21*vec.X + r22*vec.Y,
		Z: r31*vec.X + r32*vec.Y,
	}
}

// GroundStationECEF represents ground station position in ECEF coordinates
type GroundStationECEF struct {
	Position Vector3D // ECEF coordinates (km)
	Station  GroundStationLocation
}

// LatLonToECEF converts geodetic coordinates to ECEF
func LatLonToECEF(station GroundStationLocation) GroundStationECEF {
	latRad := station.LatitudeDeg * degreesToRadians
	lonRad := station.LongitudeDeg * degreesToRadians
	altKm := station.AltitudeM / 1000.0
	
	// WGS84 ellipsoid parameters
	a := earthRadiusKm // Simplified: use mean radius
	
	cosLat := math.Cos(latRad)
	sinLat := math.Sin(latRad)
	cosLon := math.Cos(lonRad)
	sinLon := math.Sin(lonRad)
	
	r := a + altKm
	
	return GroundStationECEF{
		Position: Vector3D{
			X: r * cosLat * cosLon,
			Y: r * cosLat * sinLon,
			Z: r * sinLat,
		},
		Station: station,
	}
}

// ECEFToECI converts ECEF coordinates to ECI at a given time
func ECEFToECI(ecef Vector3D, t time.Time) Vector3D {
	// Greenwich Mean Sidereal Time (simplified)
	gmst := computeGMST(t)
	
	cosGMST := math.Cos(gmst)
	sinGMST := math.Sin(gmst)
	
	return Vector3D{
		X: cosGMST*ecef.X - sinGMST*ecef.Y,
		Y: sinGMST*ecef.X + cosGMST*ecef.Y,
		Z: ecef.Z,
	}
}

// computeGMST computes Greenwich Mean Sidereal Time (radians)
func computeGMST(t time.Time) float64 {
	// Julian date
	jd := float64(t.Unix())/86400.0 + 2440587.5
	
	// Days since J2000.0
	d := jd - 2451545.0
	
	// GMST in hours
	gmstHours := 18.697374558 + 24.06570982441908*d
	
	// Convert to radians and normalize
	gmstRad := math.Mod(gmstHours*15.0*degreesToRadians, 2.0*math.Pi)
	if gmstRad < 0 {
		gmstRad += 2.0 * math.Pi
	}
	
	return gmstRad
}

// ComputeElevationAzimuth computes elevation and azimuth from ground station to satellite
func ComputeElevationAzimuth(satECI Vector3D, gsECEF GroundStationECEF, t time.Time) (elevationDeg, azimuthDeg float64) {
	// Convert ground station to ECI
	gsECI := ECEFToECI(gsECEF.Position, t)
	
	// Range vector from ground station to satellite
	rangeVec := satECI.Subtract(gsECI)
	
	// Convert to topocentric coordinates (SEZ: South-East-Zenith)
	latRad := gsECEF.Station.LatitudeDeg * degreesToRadians
	lonRad := gsECEF.Station.LongitudeDeg * degreesToRadians
	gmst := computeGMST(t)
	lst := gmst + lonRad // Local Sidereal Time
	
	cosLat := math.Cos(latRad)
	sinLat := math.Sin(latRad)
	cosLST := math.Cos(lst)
	sinLST := math.Sin(lst)
	
	// Transform to SEZ
	south := -sinLat*cosLST*rangeVec.X - sinLat*sinLST*rangeVec.Y + cosLat*rangeVec.Z
	east := -sinLST*rangeVec.X + cosLST*rangeVec.Y
	zenith := cosLat*cosLST*rangeVec.X + cosLat*sinLST*rangeVec.Y + sinLat*rangeVec.Z
	
	// Compute elevation and azimuth
	rangeDistance := math.Sqrt(south*south + east*east + zenith*zenith)
	elevationRad := math.Asin(zenith / rangeDistance)
	azimuthRad := math.Atan2(east, south)
	
	if azimuthRad < 0 {
		azimuthRad += 2.0 * math.Pi
	}
	
	return elevationRad * radiansToDegrees, azimuthRad * radiansToDegrees
}

// ComputeDopplerShift computes Doppler shift at carrier frequency
func ComputeDopplerShift(satState *OrbitalState, gsECEF GroundStationECEF, carrierFreqHz float64) float64 {
	// Convert ground station to ECI
	gsECI := ECEFToECI(gsECEF.Position, satState.Time)
	
	// Range vector
	rangeVec := satState.Position.Subtract(gsECI)
	rangeDistance := rangeVec.Magnitude()
	
	// Unit range vector
	rangeUnit := Vector3D{
		X: rangeVec.X / rangeDistance,
		Y: rangeVec.Y / rangeDistance,
		Z: rangeVec.Z / rangeDistance,
	}
	
	// Radial velocity (positive = receding, negative = approaching)
	radialVelocity := satState.Velocity.Dot(rangeUnit)
	
	// Doppler shift: f_observed = f_transmitted * (1 - v/c)
	// For v << c: Δf ≈ -f * v/c
	speedOfLight := 299792.458 // km/s
	dopplerShift := -carrierFreqHz * radialVelocity / speedOfLight
	
	return dopplerShift
}

// PredictLEOPasses predicts LEO satellite passes over ground stations
// Returns contact windows where elevation exceeds minimum threshold
func PredictLEOPasses(
	params *OrbitalParameters,
	stations []GroundStationLocation,
	fromTime, toTime time.Time,
	timeStepSeconds int,
) ([]PredictedContact, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid orbital parameters: %w", err)
	}
	
	if fromTime.After(toTime) {
		return nil, fmt.Errorf("fromTime must be before toTime")
	}
	
	if timeStepSeconds <= 0 {
		timeStepSeconds = 30 // Default 30-second time step
	}
	
	// Validate all ground stations
	for i, station := range stations {
		if err := station.Validate(); err != nil {
			return nil, fmt.Errorf("invalid ground station %d: %w", i, err)
		}
	}
	
	// Convert ground stations to ECEF
	gsECEF := make([]GroundStationECEF, len(stations))
	for i, station := range stations {
		gsECEF[i] = LatLonToECEF(station)
	}
	
	// Track contact state for each ground station
	type contactState struct {
		inContact       bool
		contactStart    time.Time
		maxElevation    float64
		maxDoppler      float64
		elevationSamples []float64
	}
	
	stationStates := make(map[NodeID]*contactState)
	for _, station := range stations {
		stationStates[station.StationID] = &contactState{
			inContact: false,
			elevationSamples: make([]float64, 0),
		}
	}
	
	var predictedContacts []PredictedContact
	contactID := uint64(1)
	
	// Propagate orbit and check visibility
	currentTime := fromTime
	for currentTime.Before(toTime) {
		// Propagate satellite position
		satState, err := PropagateOrbit(params, currentTime)
		if err != nil {
			return nil, fmt.Errorf("orbit propagation failed at %v: %w", currentTime, err)
		}
		
		// Check visibility from each ground station
		for i, gs := range gsECEF {
			station := stations[i]
			state := stationStates[station.StationID]
			
			elevation, _ := ComputeElevationAzimuth(satState.Position, gs, currentTime)
			
			// Check if satellite is above minimum elevation
			isVisible := elevation >= station.MinElevationDeg
			
			if isVisible && !state.inContact {
				// Rising edge: contact starts
				state.inContact = true
				state.contactStart = currentTime
				state.maxElevation = elevation
				state.elevationSamples = []float64{elevation}
				
				// Compute Doppler shift at UHF 437 MHz
				doppler := ComputeDopplerShift(satState, gs, 437e6)
				state.maxDoppler = math.Abs(doppler)
				
			} else if isVisible && state.inContact {
				// Still in contact: update max elevation
				if elevation > state.maxElevation {
					state.maxElevation = elevation
				}
				state.elevationSamples = append(state.elevationSamples, elevation)
				
				doppler := ComputeDopplerShift(satState, gs, 437e6)
				if math.Abs(doppler) > state.maxDoppler {
					state.maxDoppler = math.Abs(doppler)
				}
				
			} else if !isVisible && state.inContact {
				// Falling edge: contact ends
				state.inContact = false
				
				// Create predicted contact window
				duration := currentTime.Sub(state.contactStart).Seconds()
				
				// Only include passes with reasonable duration (> 60 seconds)
				if duration >= 60 {
					// Compute confidence based on time from epoch
					confidence := computeConfidence(params.Epoch, state.contactStart.Unix())
					
					predictedContacts = append(predictedContacts, PredictedContact{
						Window: ContactWindow{
							ContactID:  contactID,
							RemoteNode: station.StationID,
							StartTime:  state.contactStart.Unix(),
							EndTime:    currentTime.Unix(),
							DataRate:   9600, // 9.6 kbps for LEO UHF
							LinkType:   LinkTypeUHFIQ,
						},
						MaxElevationDeg: state.maxElevation,
						DopplerShiftHz:  state.maxDoppler,
						Confidence:      confidence,
					})
					contactID++
				}
				
				// Reset state
				state.maxElevation = 0
				state.maxDoppler = 0
				state.elevationSamples = nil
			}
		}
		
		currentTime = currentTime.Add(time.Duration(timeStepSeconds) * time.Second)
	}
	
	// Sort by start time
	sort.Slice(predictedContacts, func(i, j int) bool {
		return predictedContacts[i].Window.StartTime < predictedContacts[j].Window.StartTime
	})
	
	return predictedContacts, nil
}

// UpdateContactPlanWithPredictions updates a contact plan with CGR-predicted passes
func (cpm *ContactPlanManager) UpdateContactPlanWithPredictions(
	spaceNodeID NodeID,
	params *OrbitalParameters,
	stations []GroundStationLocation,
	fromTime, toTime time.Time,
) error {
	// Predict passes
	predicted, err := PredictLEOPasses(
		params,
		stations,
		fromTime,
		toTime,
		30, // 30-second time step
	)
	if err != nil {
		return fmt.Errorf("pass prediction failed: %w", err)
	}
	
	cpm.mu.Lock()
	defer cpm.mu.Unlock()
	
	if cpm.plan == nil {
		return fmt.Errorf("no contact plan loaded")
	}
	
	// Update orbital parameters
	cpm.orbitalParams[spaceNodeID] = params
	
	// Add predicted contacts to plan
	cpm.plan.PredictedContacts = predicted
	cpm.plan.OrbitalData = params
	
	// Also add to regular contacts for scheduling
	for _, pc := range predicted {
		cpm.plan.Contacts = append(cpm.plan.Contacts, pc.Window)
	}
	
	// Re-validate plan
	return cpm.plan.Validate()
}

// PropagateCislunarOrbit propagates cislunar orbital elements forward in time
// This implements simplified lunar orbit propagation for cislunar payloads
// Accounts for Earth-Moon system dynamics and longer orbital periods
func PropagateCislunarOrbit(params *OrbitalParameters, targetTime time.Time) (*OrbitalState, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	epoch := time.Unix(params.Epoch, 0)
	deltaT := targetTime.Sub(epoch).Seconds()

	// Convert to km for calculations
	semiMajorAxisKm := params.SemiMajorAxisM / 1000.0
	
	// For cislunar orbits, we use a simplified two-body propagation around Earth
	// In reality, cislunar orbits are highly perturbed by lunar gravity
	// This is a simplified model suitable for contact prediction over short horizons (days)
	
	// Compute mean motion (rad/s) - using Earth's gravitational parameter
	// For lunar orbit, the semi-major axis is measured from Earth center
	meanMotion := math.Sqrt(earthMuKm3PerS2 / (semiMajorAxisKm * semiMajorAxisKm * semiMajorAxisKm))
	
	// Propagate mean anomaly
	meanAnomalyRad := params.TrueAnomalyDeg * degreesToRadians
	meanAnomalyRad += meanMotion * deltaT
	meanAnomalyRad = math.Mod(meanAnomalyRad, 2.0*math.Pi)
	
	// Solve Kepler's equation for eccentric anomaly (Newton-Raphson)
	eccentricAnomalyRad := solveKeplersEquation(meanAnomalyRad, params.Eccentricity)
	
	// Compute true anomaly
	trueAnomalyRad := 2.0 * math.Atan2(
		math.Sqrt(1.0+params.Eccentricity)*math.Sin(eccentricAnomalyRad/2.0),
		math.Sqrt(1.0-params.Eccentricity)*math.Cos(eccentricAnomalyRad/2.0),
	)
	
	// Compute position in orbital plane
	radius := semiMajorAxisKm * (1.0 - params.Eccentricity*math.Cos(eccentricAnomalyRad))
	
	// Position in perifocal coordinates
	posPerifocal := Vector3D{
		X: radius * math.Cos(trueAnomalyRad),
		Y: radius * math.Sin(trueAnomalyRad),
		Z: 0.0,
	}
	
	// Velocity in perifocal coordinates
	h := math.Sqrt(earthMuKm3PerS2 * semiMajorAxisKm * (1.0 - params.Eccentricity*params.Eccentricity))
	velPerifocal := Vector3D{
		X: -(earthMuKm3PerS2 / h) * math.Sin(trueAnomalyRad),
		Y: (earthMuKm3PerS2 / h) * (params.Eccentricity + math.Cos(trueAnomalyRad)),
		Z: 0.0,
	}
	
	// Transform to ECI coordinates
	// For cislunar orbits, RAAN precession is much slower due to larger semi-major axis
	raanRad := params.RAANDeg * degreesToRadians
	// Simplified: no J2 precession for cislunar (negligible effect at this distance)
	
	incRad := params.InclinationDeg * degreesToRadians
	argPerigeeRad := params.ArgPeriapsisDeg * degreesToRadians
	
	posECI := perifocalToECI(posPerifocal, raanRad, incRad, argPerigeeRad)
	velECI := perifocalToECI(velPerifocal, raanRad, incRad, argPerigeeRad)
	
	return &OrbitalState{
		Time:     targetTime,
		Position: posECI,
		Velocity: velECI,
	}, nil
}

// ComputeLightTimeDelay computes one-way light-time delay for a given distance
func ComputeLightTimeDelay(distanceKm float64) float64 {
	return distanceKm / speedOfLightKmS
}

// PredictCislunarPasses predicts cislunar payload passes over ground stations
// Returns contact windows where the payload is visible from ground stations
// Accounts for 1-2 second light-time delay and confidence degradation
func PredictCislunarPasses(
	params *OrbitalParameters,
	stations []GroundStationLocation,
	fromTime, toTime time.Time,
	timeStepSeconds int,
) ([]PredictedContact, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid orbital parameters: %w", err)
	}
	
	if fromTime.After(toTime) {
		return nil, fmt.Errorf("fromTime must be before toTime")
	}
	
	if timeStepSeconds <= 0 {
		timeStepSeconds = 60 // Default 60-second time step for cislunar (slower dynamics)
	}
	
	// Validate all ground stations
	for i, station := range stations {
		if err := station.Validate(); err != nil {
			return nil, fmt.Errorf("invalid ground station %d: %w", i, err)
		}
	}
	
	// Convert ground stations to ECEF
	gsECEF := make([]GroundStationECEF, len(stations))
	for i, station := range stations {
		gsECEF[i] = LatLonToECEF(station)
	}
	
	// Track contact state for each ground station
	type contactState struct {
		inContact       bool
		contactStart    time.Time
		maxElevation    float64
		maxDoppler      float64
		lightTimeDelay  float64
		elevationSamples []float64
	}
	
	stationStates := make(map[NodeID]*contactState)
	for _, station := range stations {
		stationStates[station.StationID] = &contactState{
			inContact: false,
			elevationSamples: make([]float64, 0),
		}
	}
	
	var predictedContacts []PredictedContact
	contactID := uint64(1)
	
	// Propagate orbit and check visibility
	currentTime := fromTime
	for currentTime.Before(toTime) {
		// Propagate cislunar payload position
		satState, err := PropagateCislunarOrbit(params, currentTime)
		if err != nil {
			return nil, fmt.Errorf("orbit propagation failed at %v: %w", currentTime, err)
		}
		
		// Check visibility from each ground station
		for i, gs := range gsECEF {
			station := stations[i]
			state := stationStates[station.StationID]
			
			elevation, _ := ComputeElevationAzimuth(satState.Position, gs, currentTime)
			
			// Compute range for light-time delay
			gsECI := ECEFToECI(gs.Position, currentTime)
			rangeVec := satState.Position.Subtract(gsECI)
			rangeKm := rangeVec.Magnitude()
			lightTimeDelay := ComputeLightTimeDelay(rangeKm)
			
			// Check if satellite is above minimum elevation
			isVisible := elevation >= station.MinElevationDeg
			
			if isVisible && !state.inContact {
				// Rising edge: contact starts
				state.inContact = true
				state.contactStart = currentTime
				state.maxElevation = elevation
				state.lightTimeDelay = lightTimeDelay
				state.elevationSamples = []float64{elevation}
				
				// Compute Doppler shift at S-band 2.2 GHz
				doppler := ComputeDopplerShift(satState, gs, 2.2e9)
				state.maxDoppler = math.Abs(doppler)
				
			} else if isVisible && state.inContact {
				// Still in contact: update max elevation
				if elevation > state.maxElevation {
					state.maxElevation = elevation
				}
				state.elevationSamples = append(state.elevationSamples, elevation)
				
				// Update light-time delay (changes as range changes)
				if lightTimeDelay > state.lightTimeDelay {
					state.lightTimeDelay = lightTimeDelay
				}
				
				doppler := ComputeDopplerShift(satState, gs, 2.2e9)
				if math.Abs(doppler) > state.maxDoppler {
					state.maxDoppler = math.Abs(doppler)
				}
				
			} else if !isVisible && state.inContact {
				// Falling edge: contact ends
				state.inContact = false
				
				// Create predicted contact window
				duration := currentTime.Sub(state.contactStart).Seconds()
				
				// Cislunar passes can be much longer (hours instead of minutes)
				// Only include passes with reasonable duration (> 5 minutes)
				if duration >= 300 {
					// Compute confidence based on time from epoch
					// Cislunar orbits degrade faster due to lunar perturbations
					confidence := computeCislunarConfidence(params.Epoch, state.contactStart.Unix())
					
					predictedContacts = append(predictedContacts, PredictedContact{
						Window: ContactWindow{
							ContactID:  contactID,
							RemoteNode: station.StationID,
							StartTime:  state.contactStart.Unix(),
							EndTime:    currentTime.Unix(),
							DataRate:   500, // 500 bps for cislunar S-band
							LinkType:   LinkTypeSBandIQ,
						},
						MaxElevationDeg: state.maxElevation,
						DopplerShiftHz:  state.maxDoppler,
						Confidence:      confidence,
					})
					contactID++
				}
				
				// Reset state
				state.maxElevation = 0
				state.maxDoppler = 0
				state.lightTimeDelay = 0
				state.elevationSamples = nil
			}
		}
		
		currentTime = currentTime.Add(time.Duration(timeStepSeconds) * time.Second)
	}
	
	// Sort by start time
	sort.Slice(predictedContacts, func(i, j int) bool {
		return predictedContacts[i].Window.StartTime < predictedContacts[j].Window.StartTime
	})
	
	return predictedContacts, nil
}

// computeCislunarConfidence calculates prediction confidence for cislunar orbits
// Confidence decreases faster than LEO due to lunar perturbations
func computeCislunarConfidence(epoch, predictionTime int64) float64 {
	// Time difference in days
	timeDiff := float64(predictionTime-epoch) / 86400.0

	// Confidence decreases faster for cislunar orbits
	// After 3 days, confidence is ~0.5
	// After 7 days, confidence is ~0.25
	confidence := math.Exp(-timeDiff / 5.0)

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// PredictPasses is a unified interface that automatically selects LEO or cislunar propagation
// based on the orbital parameters
func PredictPasses(
	params *OrbitalParameters,
	stations []GroundStationLocation,
	fromTime, toTime time.Time,
	timeStepSeconds int,
) ([]PredictedContact, error) {
	orbitType := params.DetermineOrbitType()
	
	switch orbitType {
	case OrbitTypeLEO:
		return PredictLEOPasses(params, stations, fromTime, toTime, timeStepSeconds)
	case OrbitTypeCislunar:
		return PredictCislunarPasses(params, stations, fromTime, toTime, timeStepSeconds)
	default:
		return nil, fmt.Errorf("unknown orbit type")
	}
}
