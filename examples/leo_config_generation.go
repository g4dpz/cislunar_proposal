package main

import (
	"fmt"
	"log"
	"time"

	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/ionconfig"
)

// Example: Generate ION-DTN configuration for a LEO CubeSat with CGR-predicted passes
func main() {
	fmt.Println("=== LEO CubeSat ION-DTN Configuration Generator ===")
	fmt.Println()

	// Define orbital parameters (TLE-derived)
	// Example: ISS-like orbit at ~400 km altitude
	epoch := time.Now()
	orbitalParams := &ionconfig.OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  6771000.0, // ~400 km altitude LEO
		Eccentricity:    0.0005,    // Nearly circular
		InclinationDeg:  51.6,      // ISS-like inclination
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	fmt.Printf("Orbital Parameters:\n")
	fmt.Printf("  Epoch: %v\n", orbitalParams.Epoch.Format(time.RFC3339))
	fmt.Printf("  Semi-major axis: %.0f m (%.0f km altitude)\n", 
		orbitalParams.SemiMajorAxisM, orbitalParams.SemiMajorAxisM-6371000)
	fmt.Printf("  Eccentricity: %.4f\n", orbitalParams.Eccentricity)
	fmt.Printf("  Inclination: %.1f°\n", orbitalParams.InclinationDeg)
	fmt.Println()

	// Define ground stations
	groundStations := []contact.GroundStationLocation{
		{
			StationID:       "gs-alpha",
			LatitudeDeg:     37.4,  // Northern California
			LongitudeDeg:    -122.1,
			AltitudeM:       100.0,
			MinElevationDeg: 10.0, // 10° minimum elevation
		},
		{
			StationID:       "gs-bravo",
			LatitudeDeg:     42.3, // Boston area
			LongitudeDeg:    -71.1,
			AltitudeM:       50.0,
			MinElevationDeg: 10.0,
		},
	}

	fmt.Printf("Ground Stations:\n")
	for _, gs := range groundStations {
		fmt.Printf("  %s: %.2f°N, %.2f°E (min elevation: %.0f°)\n",
			gs.StationID, gs.LatitudeDeg, gs.LongitudeDeg, gs.MinElevationDeg)
	}
	fmt.Println()

	// Use CGR to predict contact windows over next 24 hours
	fmt.Println("Predicting contact windows using CGR...")
	
	// Convert orbital parameters to contact package format
	cgrParams := &contact.OrbitalParameters{
		Epoch:           epoch.Unix(),
		SemiMajorAxisM:  orbitalParams.SemiMajorAxisM,
		Eccentricity:    orbitalParams.Eccentricity,
		InclinationDeg:  orbitalParams.InclinationDeg,
		RAANDeg:         orbitalParams.RAANDeg,
		ArgPeriapsisDeg: orbitalParams.ArgPeriapsisDeg,
		TrueAnomalyDeg:  orbitalParams.TrueAnomalyDeg,
	}

	fromTime := epoch
	toTime := epoch.Add(24 * time.Hour)

	predictedContacts, err := contact.PredictLEOPasses(
		cgrParams,
		groundStations,
		fromTime,
		toTime,
		30, // 30-second time step
	)
	if err != nil {
		log.Fatalf("CGR prediction failed: %v", err)
	}

	fmt.Printf("Found %d predicted passes over next 24 hours\n", len(predictedContacts))
	fmt.Println()

	// Convert predicted contacts to LEO contact format
	leoContacts := make([]ionconfig.LEOContact, len(predictedContacts))
	for i, pc := range predictedContacts {
		// Map node ID to node number (simplified)
		nodeNumber := 1
		if pc.Window.RemoteNode == "gs-bravo" {
			nodeNumber = 2
		}

		// Map node ID to callsign (simplified)
		callsign := "KA1ABC"
		if pc.Window.RemoteNode == "gs-bravo" {
			callsign = "KB2XYZ"
		}

		leoContacts[i] = ionconfig.LEOContact{
			RemoteNodeNumber: nodeNumber,
			RemoteCallsign:   callsign,
			StartTime:        time.Unix(pc.Window.StartTime, 0),
			Duration:         time.Duration(pc.Window.EndTime-pc.Window.StartTime) * time.Second,
			DataRate:         int(pc.Window.DataRate),
			MaxElevationDeg:  pc.MaxElevationDeg,
			Confidence:       pc.Confidence,
		}

		fmt.Printf("Pass %d:\n", i+1)
		fmt.Printf("  Ground Station: %s (%s)\n", pc.Window.RemoteNode, callsign)
		fmt.Printf("  Start: %v\n", leoContacts[i].StartTime.Format("15:04:05"))
		fmt.Printf("  Duration: %.1f minutes\n", leoContacts[i].Duration.Minutes())
		fmt.Printf("  Max Elevation: %.1f°\n", pc.MaxElevationDeg)
		fmt.Printf("  Data Rate: %d bps\n", pc.Window.DataRate)
		fmt.Printf("  Confidence: %.2f\n", pc.Confidence)
		fmt.Println()
	}

	// Generate ION-DTN configuration files
	fmt.Println("Generating ION-DTN configuration files...")
	
	outputDir := "./configs/leo-cubesat"
	err = ionconfig.GenerateLEOWithCGRPrediction(
		"leo-cubesat-01",
		10,
		"KL0SAT",
		orbitalParams,
		leoContacts,
		outputDir,
	)
	if err != nil {
		log.Fatalf("Configuration generation failed: %v", err)
	}

	fmt.Printf("✓ Configuration files generated in %s\n", outputDir)
	fmt.Println()
	fmt.Println("Generated files:")
	fmt.Println("  - node.ionrc    (ION initialization)")
	fmt.Println("  - node.ltprc    (LTP configuration)")
	fmt.Println("  - node.bprc     (Bundle Protocol configuration)")
	fmt.Println("  - node.ipnrc    (IPN scheme configuration)")
	fmt.Println("  - leo.ionconfig (LEO-specific configuration)")
	fmt.Println()
	fmt.Println("The LEO node is now configured for autonomous store-and-forward")
	fmt.Println("during predicted orbital passes.")
	fmt.Println()
	fmt.Println("To update with fresh TLE/ephemeris data:")
	fmt.Println("  1. Receive new TLE from ground station")
	fmt.Println("  2. Re-run CGR prediction with updated orbital parameters")
	fmt.Println("  3. Call UpdateLEOContactPlan() to refresh contact windows")
}
