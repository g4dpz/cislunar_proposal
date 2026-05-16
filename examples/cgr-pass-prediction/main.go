package main

import (
	"fmt"
	"time"

	"terrestrial-dtn/pkg/contact"
)

// Example demonstrating CGR-based LEO pass prediction
func main() {
	fmt.Println("=== CGR-Based LEO Pass Prediction Example ===\n")

	// Define ISS-like orbital parameters
	// These would typically come from TLE data
	params := &contact.OrbitalParameters{
		Epoch:           time.Now().Unix(),
		SemiMajorAxisM:  6771000.0, // ~400 km altitude
		Eccentricity:    0.001,     // Nearly circular
		InclinationDeg:  51.6,      // ISS inclination
		RAANDeg:         100.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	fmt.Printf("Orbital Parameters:\n")
	fmt.Printf("  Epoch: %s\n", time.Unix(params.Epoch, 0).Format(time.RFC3339))
	fmt.Printf("  Semi-major axis: %.1f km\n", params.SemiMajorAxisM/1000.0)
	fmt.Printf("  Eccentricity: %.4f\n", params.Eccentricity)
	fmt.Printf("  Inclination: %.1f°\n", params.InclinationDeg)
	fmt.Printf("  RAAN: %.1f°\n", params.RAANDeg)
	fmt.Printf("  Arg of Perigee: %.1f°\n\n", params.ArgPeriapsisDeg)

	// Define ground stations
	stations := []contact.GroundStationLocation{
		{
			StationID:       "gs-boston",
			LatitudeDeg:     42.36,
			LongitudeDeg:    -71.06,
			AltitudeM:       50.0,
			MinElevationDeg: 10.0,
		},
		{
			StationID:       "gs-houston",
			LatitudeDeg:     29.76,
			LongitudeDeg:    -95.37,
			AltitudeM:       30.0,
			MinElevationDeg: 10.0,
		},
		{
			StationID:       "gs-seattle",
			LatitudeDeg:     47.61,
			LongitudeDeg:    -122.33,
			AltitudeM:       100.0,
			MinElevationDeg: 10.0,
		},
	}

	fmt.Printf("Ground Stations:\n")
	for _, gs := range stations {
		fmt.Printf("  %s: %.2f°N, %.2f°E, min elevation %.0f°\n",
			gs.StationID, gs.LatitudeDeg, gs.LongitudeDeg, gs.MinElevationDeg)
	}
	fmt.Println()

	// Predict passes over next 24 hours
	fromTime := time.Unix(params.Epoch, 0)
	toTime := fromTime.Add(24 * time.Hour)

	fmt.Printf("Predicting passes from %s to %s...\n\n",
		fromTime.Format("2006-01-02 15:04 MST"),
		toTime.Format("2006-01-02 15:04 MST"))

	predicted, err := contact.PredictLEOPasses(params, stations, fromTime, toTime, 30)
	if err != nil {
		fmt.Printf("Error predicting passes: %v\n", err)
		return
	}

	fmt.Printf("Found %d passes:\n\n", len(predicted))

	// Group passes by station
	passesByStation := make(map[contact.NodeID][]contact.PredictedContact)
	for _, pc := range predicted {
		stationID := pc.Window.RemoteNode
		passesByStation[stationID] = append(passesByStation[stationID], pc)
	}

	// Display passes for each station
	for _, station := range stations {
		passes := passesByStation[station.StationID]
		fmt.Printf("Station: %s (%d passes)\n", station.StationID, len(passes))
		fmt.Println("  " + "─────────────────────────────────────────────────────────────────────────────")

		if len(passes) == 0 {
			fmt.Println("  No passes predicted")
			fmt.Println()
			continue
		}

		for i, pc := range passes {
			startTime := time.Unix(pc.Window.StartTime, 0)
			endTime := time.Unix(pc.Window.EndTime, 0)
			duration := endTime.Sub(startTime)

			fmt.Printf("  Pass %d:\n", i+1)
			fmt.Printf("    Start:         %s\n", startTime.Format("2006-01-02 15:04:05 MST"))
			fmt.Printf("    End:           %s\n", endTime.Format("2006-01-02 15:04:05 MST"))
			fmt.Printf("    Duration:      %.1f minutes\n", duration.Minutes())
			fmt.Printf("    Max Elevation: %.1f°\n", pc.MaxElevationDeg)
			fmt.Printf("    Max Doppler:   %.0f Hz (at 437 MHz)\n", pc.DopplerShiftHz)
			fmt.Printf("    Data Rate:     %d bps\n", pc.Window.DataRate)
			fmt.Printf("    Confidence:    %.2f\n", pc.Confidence)
			fmt.Println()
		}
	}

	// Summary statistics
	fmt.Println("=== Summary ===")
	totalPasses := len(predicted)
	if totalPasses > 0 {
		var totalDuration time.Duration
		var maxElevation float64
		var maxDoppler float64

		for _, pc := range predicted {
			duration := time.Unix(pc.Window.EndTime, 0).Sub(time.Unix(pc.Window.StartTime, 0))
			totalDuration += duration

			if pc.MaxElevationDeg > maxElevation {
				maxElevation = pc.MaxElevationDeg
			}
			if pc.DopplerShiftHz > maxDoppler {
				maxDoppler = pc.DopplerShiftHz
			}
		}

		avgDuration := totalDuration / time.Duration(totalPasses)
		fmt.Printf("Total passes:        %d\n", totalPasses)
		fmt.Printf("Average duration:    %.1f minutes\n", avgDuration.Minutes())
		fmt.Printf("Total contact time:  %.1f minutes\n", totalDuration.Minutes())
		fmt.Printf("Highest elevation:   %.1f°\n", maxElevation)
		fmt.Printf("Max Doppler shift:   %.0f Hz\n", maxDoppler)
	}

	// Example: Using ContactPlanManager
	fmt.Println("\n=== Contact Plan Manager Integration ===")
	cpm := contact.NewContactPlanManager()

	// Load initial plan
	plan := &contact.ContactPlan{
		PlanID:      1,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   fromTime.Unix(),
		ValidTo:     toTime.Unix(),
		Contacts:    []contact.ContactWindow{},
	}

	err = cpm.LoadPlan(plan)
	if err != nil {
		fmt.Printf("Error loading plan: %v\n", err)
		return
	}

	// Update plan with CGR predictions
	err = cpm.UpdateContactPlanWithPredictions(
		"leo-sat-01",
		params,
		stations,
		fromTime,
		toTime,
	)
	if err != nil {
		fmt.Printf("Error updating plan: %v\n", err)
		return
	}

	fmt.Printf("Contact plan updated with %d predicted contacts\n", len(predicted))
	fmt.Println("Ready for autonomous store-and-forward operations!")
}
