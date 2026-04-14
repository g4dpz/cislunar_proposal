package main

import (
	"fmt"
	"log"
	"time"

	"terrestrial-dtn/pkg/ionconfig"
)

func main() {
	fmt.Println("=== Cislunar Amateur DTN Payload - ION-DTN Configuration Generation ===")
	fmt.Println()

	// Example 1: Generate S-band cislunar configuration
	fmt.Println("Example 1: S-band Cislunar Configuration")
	fmt.Println("----------------------------------------")

	epoch := time.Now()
	orbitalParams := &ionconfig.OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  384400000.0, // ~384,400 km (Earth-Moon distance)
		Eccentricity:    0.05,        // Slightly eccentric lunar orbit
		InclinationDeg:  5.0,         // Lunar orbit inclination
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	// Define CGR-predicted cislunar contact windows
	contacts := []ionconfig.CislunarContact{
		{
			RemoteNodeNumber: 100,
			RemoteCallsign:   "W1ABC",
			StartTime:        epoch.Add(2 * time.Hour),
			Duration:         3 * time.Hour, // 3-hour contact window
			DataRate:         500,           // 500 bps S-band with BPSK+LDPC
			MaxElevationDeg:  65.0,
			Confidence:       0.90,
			LightTimeDelay:   1.28, // ~1.28 seconds for Earth-Moon
		},
		{
			RemoteNodeNumber: 101,
			RemoteCallsign:   "K2XYZ",
			StartTime:        epoch.Add(12 * time.Hour),
			Duration:         4 * time.Hour, // 4-hour contact window
			DataRate:         500,
			MaxElevationDeg:  75.0,
			Confidence:       0.85,
			LightTimeDelay:   1.30,
		},
		{
			RemoteNodeNumber: 102,
			RemoteCallsign:   "N3DEF",
			StartTime:        epoch.Add(24 * time.Hour),
			Duration:         2 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  55.0,
			Confidence:       0.75, // Lower confidence for 24h prediction
			LightTimeDelay:   1.32,
		},
	}

	sbandConfig := ionconfig.CislunarNodeConfig{
		NodeID:           "cislunar-sband-01",
		NodeNumber:       10,
		Callsign:         "W0MOON",
		StorageBytes:     512 * 1024 * 1024, // 512 MB for long-duration storage
		SRAMBytes:        786 * 1024,         // 786 KB SRAM (STM32U585)
		Band:             ionconfig.SBand,
		ContactPlan:      contacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	fmt.Printf("Node ID: %s\n", sbandConfig.NodeID)
	fmt.Printf("Node Number: %d\n", sbandConfig.NodeNumber)
	fmt.Printf("Callsign: %s\n", sbandConfig.Callsign)
	fmt.Printf("Band: S-band (2.2 GHz)\n")
	fmt.Printf("Storage: %d MB\n", sbandConfig.StorageBytes/(1024*1024))
	fmt.Printf("SRAM: %d KB\n", sbandConfig.SRAMBytes/1024)
	fmt.Printf("Contact windows: %d\n", len(contacts))
	fmt.Println()

	err := ionconfig.GenerateCislunarConfig(sbandConfig, "configs/cislunar-sband")
	if err != nil {
		log.Fatalf("Failed to generate S-band config: %v", err)
	}

	fmt.Println("Generated S-band configuration files:")
	fmt.Println("  - configs/cislunar-sband/node.ionrc")
	fmt.Println("  - configs/cislunar-sband/node.ltprc")
	fmt.Println("  - configs/cislunar-sband/node.bprc")
	fmt.Println("  - configs/cislunar-sband/node.ipnrc")
	fmt.Println("  - configs/cislunar-sband/sband.ionconfig")
	fmt.Println()

	// Example 2: Generate X-band cislunar configuration
	fmt.Println("Example 2: X-band Cislunar Configuration")
	fmt.Println("----------------------------------------")

	xbandContacts := []ionconfig.CislunarContact{
		{
			RemoteNodeNumber: 200,
			RemoteCallsign:   "W4GHI",
			StartTime:        epoch.Add(4 * time.Hour),
			Duration:         5 * time.Hour, // Longer contact window
			DataRate:         500,
			MaxElevationDeg:  80.0,
			Confidence:       0.92,
			LightTimeDelay:   1.28,
		},
	}

	xbandConfig := ionconfig.CislunarNodeConfig{
		NodeID:           "cislunar-xband-01",
		NodeNumber:       20,
		Callsign:         "K0XRAY",
		StorageBytes:     1024 * 1024 * 1024, // 1 GB storage
		SRAMBytes:        2 * 1024 * 1024,    // 2 MB SRAM (higher-capability processor)
		Band:             ionconfig.XBand,
		ContactPlan:      xbandContacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	fmt.Printf("Node ID: %s\n", xbandConfig.NodeID)
	fmt.Printf("Node Number: %d\n", xbandConfig.NodeNumber)
	fmt.Printf("Callsign: %s\n", xbandConfig.Callsign)
	fmt.Printf("Band: X-band (8.4 GHz)\n")
	fmt.Printf("Storage: %d MB\n", xbandConfig.StorageBytes/(1024*1024))
	fmt.Printf("SRAM: %d MB\n", xbandConfig.SRAMBytes/(1024*1024))
	fmt.Printf("Contact windows: %d\n", len(xbandContacts))
	fmt.Println()

	err = ionconfig.GenerateCislunarConfig(xbandConfig, "configs/cislunar-xband")
	if err != nil {
		log.Fatalf("Failed to generate X-band config: %v", err)
	}

	fmt.Println("Generated X-band configuration files:")
	fmt.Println("  - configs/cislunar-xband/node.ionrc")
	fmt.Println("  - configs/cislunar-xband/node.ltprc")
	fmt.Println("  - configs/cislunar-xband/node.bprc")
	fmt.Println("  - configs/cislunar-xband/node.ipnrc")
	fmt.Println("  - configs/cislunar-xband/xband.ionconfig")
	fmt.Println()

	// Example 3: Using the convenience function with CGR prediction
	fmt.Println("Example 3: CGR-Predicted Configuration")
	fmt.Println("--------------------------------------")

	cgrContacts := []ionconfig.CislunarContact{
		{
			RemoteNodeNumber: 300,
			RemoteCallsign:   "N5JKL",
			StartTime:        epoch.Add(8 * time.Hour),
			Duration:         6 * time.Hour, // Very long contact window
			DataRate:         500,
			MaxElevationDeg:  85.0,
			Confidence:       0.88,
			LightTimeDelay:   1.29,
		},
		{
			RemoteNodeNumber: 301,
			RemoteCallsign:   "W6MNO",
			StartTime:        epoch.Add(20 * time.Hour),
			Duration:         3 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  60.0,
			Confidence:       0.78,
			LightTimeDelay:   1.31,
		},
	}

	err = ionconfig.GenerateCislunarWithCGRPrediction(
		"cislunar-cgr-01",
		30,
		"N0CGR",
		ionconfig.SBand,
		orbitalParams,
		cgrContacts,
		"configs/cislunar-cgr",
	)
	if err != nil {
		log.Fatalf("Failed to generate CGR config: %v", err)
	}

	fmt.Println("Generated CGR-predicted configuration files:")
	fmt.Println("  - configs/cislunar-cgr/node.ionrc")
	fmt.Println("  - configs/cislunar-cgr/node.ltprc")
	fmt.Println("  - configs/cislunar-cgr/node.bprc")
	fmt.Println("  - configs/cislunar-cgr/node.ipnrc")
	fmt.Println("  - configs/cislunar-cgr/sband.ionconfig")
	fmt.Println()

	// Example 4: Display key cislunar configuration parameters
	fmt.Println("Example 4: Key Cislunar Configuration Parameters")
	fmt.Println("------------------------------------------------")
	fmt.Println()

	fmt.Println("S-band Configuration:")
	fmt.Println("  Frequency: 2.2 GHz")
	fmt.Println("  Data Rate: 500 bps")
	fmt.Println("  Modulation: BPSK")
	fmt.Println("  FEC: LDPC (coding gain ~6-8 dB)")
	fmt.Println("  TX Power: 5 W")
	fmt.Println("  TX Gain: 10 dBi (directional patch)")
	fmt.Println("  RX Gain: 35 dBi (3-5m ground dish)")
	fmt.Println()

	fmt.Println("X-band Configuration:")
	fmt.Println("  Frequency: 8.4 GHz")
	fmt.Println("  Data Rate: 500 bps")
	fmt.Println("  Modulation: BPSK")
	fmt.Println("  FEC: Turbo (coding gain ~5-7 dB)")
	fmt.Println("  TX Power: 5 W")
	fmt.Println("  TX Gain: 15 dBi")
	fmt.Println("  RX Gain: 40 dBi")
	fmt.Println()

	fmt.Println("Deep-Space Link Characteristics:")
	fmt.Println("  Distance: ~384,000 km (Earth-Moon)")
	fmt.Println("  Light-time delay: 1-2 seconds (one-way)")
	fmt.Println("  Round-trip time: 2-4 seconds")
	fmt.Println("  LTP timeout: RTT + 10s processing margin")
	fmt.Println("  Contact duration: Hours (vs. minutes for LEO)")
	fmt.Println("  Storage: 512 MB - 1 GB (long-duration message storage)")
	fmt.Println("  Confidence degradation: Faster than LEO (lunar perturbations)")
	fmt.Println()

	fmt.Println("LTP Session Management:")
	fmt.Println("  Segment size: 1400 bytes")
	fmt.Println("  Max block size: 50,000 bytes (larger for deep-space)")
	fmt.Println("  Max sessions: 20 (more for extended windows)")
	fmt.Println("  Max segments: 500 (more for reliability)")
	fmt.Println("  Aggregation size: 5000 bytes (larger for efficiency)")
	fmt.Println()

	fmt.Println("=== Configuration Generation Complete ===")
}
