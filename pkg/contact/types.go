package contact

import (
	"fmt"

	"terrestrial-dtn/pkg/bpa"
)

// LinkType represents the type of communication link
type LinkType int

const (
	LinkTypeVHF LinkType = iota
	LinkTypeUHFTNC
	LinkTypeUHFIQB200
	LinkTypeUHFIQ
	LinkTypeSBandIQ
	LinkTypeXBandIQ
)

func (lt LinkType) String() string {
	switch lt {
	case LinkTypeVHF:
		return "vhf"
	case LinkTypeUHFTNC:
		return "uhf_tnc"
	case LinkTypeUHFIQB200:
		return "uhf_iq_b200"
	case LinkTypeUHFIQ:
		return "uhf_iq"
	case LinkTypeSBandIQ:
		return "sband_iq"
	case LinkTypeXBandIQ:
		return "xband_iq"
	default:
		return "unknown"
	}
}

// NodeID represents a unique node identifier
type NodeID string

func (n NodeID) String() string {
	return string(n)
}

// ContactWindow represents a scheduled communication window
type ContactWindow struct {
	ContactID  uint64
	RemoteNode NodeID
	StartTime  int64 // Unix epoch seconds
	EndTime    int64 // Unix epoch seconds
	DataRate   int64 // bits per second
	LinkType   LinkType
}

// IsActive checks if the contact window is active at the given time
func (cw *ContactWindow) IsActive(currentTime int64) bool {
	return cw.StartTime <= currentTime && currentTime < cw.EndTime
}

// Duration returns the duration of the contact window in seconds
func (cw *ContactWindow) Duration() int64 {
	return cw.EndTime - cw.StartTime
}

// OrbitalParameters represents orbital elements for CGR-based contact prediction
type OrbitalParameters struct {
	Epoch            int64   // reference epoch (Unix seconds)
	SemiMajorAxisM   float64 // semi-major axis in meters
	Eccentricity     float64 // orbital eccentricity (0 = circular, <1 = elliptical)
	InclinationDeg   float64 // orbital inclination in degrees
	RAANDeg          float64 // right ascension of ascending node in degrees
	ArgPeriapsisDeg  float64 // argument of periapsis in degrees
	TrueAnomalyDeg   float64 // true anomaly at epoch in degrees
}

// OrbitType represents the type of orbit (LEO or cislunar)
type OrbitType int

const (
	OrbitTypeLEO OrbitType = iota
	OrbitTypeCislunar
)

// DetermineOrbitType determines if the orbit is LEO or cislunar based on semi-major axis
func (op *OrbitalParameters) DetermineOrbitType() OrbitType {
	semiMajorAxisKm := op.SemiMajorAxisM / 1000.0
	
	// LEO: semi-major axis < 8000 km (altitude < ~1600 km)
	// Cislunar: semi-major axis >= 8000 km (includes lunar orbit, L-points, etc.)
	if semiMajorAxisKm < 8000.0 {
		return OrbitTypeLEO
	}
	return OrbitTypeCislunar
}

// Validate checks if the orbital parameters are valid
func (op *OrbitalParameters) Validate() error {
	if op.Eccentricity < 0 || op.Eccentricity >= 1.0 {
		return fmt.Errorf("eccentricity must be in range [0, 1)")
	}
	if op.SemiMajorAxisM <= 6371000 { // Earth radius ~6371 km
		return fmt.Errorf("semi-major axis must be greater than Earth radius")
	}
	if op.InclinationDeg < 0 || op.InclinationDeg > 180 {
		return fmt.Errorf("inclination must be in range [0, 180] degrees")
	}
	return nil
}

// GroundStationLocation represents a ground station for contact prediction
type GroundStationLocation struct {
	StationID       NodeID
	LatitudeDeg     float64 // geodetic latitude in degrees
	LongitudeDeg    float64 // geodetic longitude in degrees
	AltitudeM       float64 // altitude above WGS84 ellipsoid in meters
	MinElevationDeg float64 // minimum elevation angle for valid contact
}

// Validate checks if the ground station location is valid
func (gs *GroundStationLocation) Validate() error {
	if gs.LatitudeDeg < -90 || gs.LatitudeDeg > 90 {
		return fmt.Errorf("latitude must be in range [-90, 90] degrees")
	}
	if gs.LongitudeDeg < -180 || gs.LongitudeDeg > 180 {
		return fmt.Errorf("longitude must be in range [-180, 180] degrees")
	}
	if gs.MinElevationDeg < 0 || gs.MinElevationDeg > 90 {
		return fmt.Errorf("minimum elevation must be in range [0, 90] degrees")
	}
	return nil
}

// PredictedContact represents a CGR-predicted contact window
type PredictedContact struct {
	Window          ContactWindow
	MaxElevationDeg float64 // peak elevation angle during pass
	DopplerShiftHz  float64 // max Doppler shift at carrier frequency
	Confidence      float64 // prediction confidence (0.0 to 1.0)
}

// ContactPlan represents a complete contact plan
type ContactPlan struct {
	PlanID            uint64
	GeneratedAt       int64 // Unix epoch seconds
	ValidFrom         int64 // Unix epoch seconds
	ValidTo           int64 // Unix epoch seconds
	Contacts          []ContactWindow
	PredictedContacts []PredictedContact
	OrbitalData       *OrbitalParameters // optional, for space nodes
}

// Validate checks if the contact plan is valid
func (cp *ContactPlan) Validate() error {
	if cp.ValidFrom >= cp.ValidTo {
		return fmt.Errorf("validFrom must be less than validTo")
	}

	// Check all contacts fall within valid range
	for i, contact := range cp.Contacts {
		if contact.StartTime < cp.ValidFrom || contact.EndTime > cp.ValidTo {
			return fmt.Errorf("contact %d falls outside valid time range", i)
		}
		if contact.StartTime >= contact.EndTime {
			return fmt.Errorf("contact %d has invalid time range", i)
		}
		if contact.DataRate <= 0 {
			return fmt.Errorf("contact %d has invalid data rate", i)
		}
	}

	// Check for overlapping contacts on the same link
	for i := 0; i < len(cp.Contacts); i++ {
		for j := i + 1; j < len(cp.Contacts); j++ {
			c1, c2 := cp.Contacts[i], cp.Contacts[j]
			// Check if same remote node and overlapping time
			if c1.RemoteNode == c2.RemoteNode {
				if c1.StartTime < c2.EndTime && c2.StartTime < c1.EndTime {
					return fmt.Errorf("contacts %d and %d overlap for node %s", i, j, c1.RemoteNode)
				}
			}
		}
	}

	// Validate predicted contacts
	for i, pc := range cp.PredictedContacts {
		if pc.Confidence < 0 || pc.Confidence > 1.0 {
			return fmt.Errorf("predicted contact %d has invalid confidence value", i)
		}
	}

	return nil
}

// DirectContactEntry represents a direct contact route to a destination
type DirectContactEntry struct {
	Destination           bpa.EndpointID
	ViaContact            ContactWindow
	EstimatedDeliveryTime int64
	Confidence            float64
}
