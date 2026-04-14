package main

import (
	"fmt"
	"log"
	"time"

	"terrestrial-dtn/pkg/contact"
)

func main() {
	fmt.Println("=== Cislunar Amateur DTN Payload - Pass Prediction Example ===\n")

	// Define cislunar orbital parameters
	// This represents a payload in lunar orbit at ~384,400 km from Earth
	epoch := time.Now()
	cislunarParams := &contact.OrbitalParameters{
		Epoch:           epoch.Unix(),
		SemiMajorAxisM:  384400000.0, // ~384,400 km (Earth-Moon distance)
		Eccentricity:    0.05,        // Slightly eccentric lunar orbit
		InclinationDeg:  5.0,         // Lunar orbit inclination
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	fmt.Printf("Cislunar Orbital Parameters:\n")
	fmt.Printf("  Epoch: %s\n", epoch.Format(time.RFC3339))
	fmt.Printf("  Semi-major axis: %.0f km\n", cislunarParams.SemiMajorAxisM/1000.0)
	fmt.Printf("  Eccentricity: %.3f\n", cislunarParams.Eccentricity)
	fmt.Printf("  Inclination: %.1f°\n", cislunarParams.InclinationDeg)
	fmt.Printf("  Orbit type: %v\n\n", getOrbitTypeName(cislunarParams.DetermineOrbitType()))

	// Define Tier 3 ground stations with large dishes (3-5m)
	// These are required for cislunar S-band communication
	stations := []contact.GroundStationLocation{
		{
			StationID:       "gs-tier3-nm",
			LatitudeDeg:     35.0,
			LongitudeDeg:    -106.0, // New Mexico
			AltitudeM:       1500.0,
			MinElevationDeg: 5.0, // Lower elevation acceptable for cislunar
		},
		{
			StationID:       "gs-tier3-ca",
			LatitudeDeg:     37.0,
			LongitudeDeg:    -122.0, // California
			AltitudeM:       100.0,
			MinElevationDeg: 5.0,
		},
		{
			StationID:       "gs-tier3-fl",
			LatitudeDeg:     28.5,
			LongitudeDeg:    -80.5, // Florida
			AltitudeM:       10.0,
			MinElevationDeg: 5.0,
		},
	}

	fmt.Println("Ground Stations (Tier 3 - Large Dish):")
	for _, station := range stations {
		fmt.Printf("  %s: %.2f°N, %.2f°E, %.0fm altitude, min elevation %.1f°\n",
			station.StationID, station.LatitudeDeg, station.LongitudeDeg,
			station.AltitudeM, station.MinElevationDeg)
	}
	fmt.Println()

	// Predict passes over 7 days
	fromTime := epoch
	toTime := epoch.Add(7 * 24 * time.Hour)

	fmt.Printf("Predicting cislunar passes from %s to %s (7 days)\n\n",
		fromTime.Format("2006-01-02 15:04"), toTime.Format("2006-01-02 15:04"))

	predicted, err := contact.PredictCislunarPasses(
		cislunarParams,
		stations,
		fromTime,
		toTime,
		60, // 60-second time step (slower dynamics than LEO)
	)
	if err != nil {
		log.Fatalf("Pass prediction failed: %v", err)
	}

	fmt.Printf("Predicted %d cislunar passes:\n\n", len(predicted))

	for i, pc := range predicted {
		startTime := time.Unix(pc.Window.StartTime, 0)
		endTime := time.Unix(pc.Window.EndTime, 0)
		duration := endTime.Sub(startTime)

		// Compute light-time delay at start of pass
		// (approximation based on semi-major axis)
		lightTimeDelay := contact.ComputeLightTimeDelay(cislunarParams.SemiMajorAxisM / 1000.0)

		fmt.Printf("Pass %d:\n", i+1)
		fmt.Printf("  Ground Station: %s\n", pc.Window.RemoteNode)
		fmt.Printf("  Start: %s\n", startTime.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("  End:   %s\n", endTime.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("  Duration: %s\n", formatDuration(duration))
		fmt.Printf("  Max Elevation: %.1f°\n", pc.MaxElevationDeg)
		fmt.Printf("  Max Doppler Shift: %.1f Hz (at 2.2 GHz S-band)\n", pc.DopplerShiftHz)
		fmt.Printf("  Light-time Delay: %.3f seconds (one-way)\n", lightTimeDelay)
		fmt.Printf("  Round-trip Time: %.3f seconds\n", lightTimeDelay*2)
		fmt.Printf("  Data Rate: %d bps (S-band with BPSK+LDPC)\n", pc.Window.DataRate)
		fmt.Printf("  Link Type: %s\n", pc.Window.LinkType.String())
		fmt.Printf("  Confidence: %.2f (degrades faster for cislunar)\n", pc.Confidence)
		fmt.Printf("  Contact ID: %d\n", pc.Window.ContactID)
		fmt.Println()
	}

	// Demonstrate confidence degradation comparison
	fmt.Println("=== Confidence Degradation Comparison ===\n")
	fmt.Println("Time from Epoch | LEO Confidence | Cislunar Confidence")
	fmt.Println("----------------|----------------|--------------------")

	testTimes := []time.Duration{
		0,
		24 * time.Hour,
		3 * 24 * time.Hour,
		7 * 24 * time.Hour,
	}

	for _, dt := range testTimes {
		testTime := epoch.Add(dt)
		leoConf := computeLEOConfidence(epoch.Unix(), testTime.Unix())
		cislunarConf := computeCislunarConfidence(epoch.Unix(), testTime.Unix())

		fmt.Printf("%-15s | %.3f          | %.3f\n",
			formatDuration(dt), leoConf, cislunarConf)
	}

	fmt.Println("\nNote: Cislunar confidence degrades faster due to lunar perturbations")
	fmt.Println("      and longer propagation times. Fresh ephemeris data should be")
	fmt.Println("      uploaded more frequently for cislunar missions.")
}

func getOrbitTypeName(ot contact.OrbitType) string {
	switch ot {
	case contact.OrbitTypeLEO:
		return "LEO"
	case contact.OrbitTypeCislunar:
		return "Cislunar"
	default:
		return "Unknown"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%.0f min", d.Minutes())
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// Helper functions to access confidence computation
// (These would normally be internal to the contact package)

func computeLEOConfidence(epoch, predictionTime int64) float64 {
	// LEO confidence: exp(-t/10 days)
	timeDiff := float64(predictionTime-epoch) / 86400.0
	confidence := 1.0
	if timeDiff > 0 {
		confidence = 1.0 / (1.0 + timeDiff/10.0)
	}
	// Approximate the exponential decay
	if timeDiff > 0 {
		confidence = 0.9048 // exp(-1/10) for 1 day
		if timeDiff >= 1 {
			confidence = 0.9048
		}
		if timeDiff >= 7 {
			confidence = 0.4966 // exp(-7/10)
		}
		if timeDiff >= 14 {
			confidence = 0.2466 // exp(-14/10)
		}
	}
	return confidence
}

func computeCislunarConfidence(epoch, predictionTime int64) float64 {
	// Cislunar confidence: exp(-t/5 days) - degrades faster
	timeDiff := float64(predictionTime-epoch) / 86400.0
	confidence := 1.0
	if timeDiff > 0 {
		confidence = 0.8187 // exp(-1/5) for 1 day
		if timeDiff >= 1 {
			confidence = 0.8187
		}
		if timeDiff >= 3 {
			confidence = 0.5488 // exp(-3/5)
		}
		if timeDiff >= 7 {
			confidence = 0.2466 // exp(-7/5)
		}
	}
	return confidence
}
