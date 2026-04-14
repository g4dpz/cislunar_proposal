// Package main demonstrates cislunar S-band IQ CLA usage for deep-space DTN communication
package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla/sband_iq"
	"terrestrial-dtn/pkg/contact"
)

func main() {
	fmt.Println("=== Cislunar Amateur DTN Payload - S-band IQ CLA Example ===")
	fmt.Println()

	// Example 1: S-band configuration for cislunar operations
	fmt.Println("Example 1: S-band Cislunar Configuration")
	fmt.Println(strings.Repeat("-", 50))
	
	sbandConfig := sband_iq.DefaultSBandConfig("W1ABC")
	fmt.Printf("Callsign: %s\n", sbandConfig.Callsign)
	fmt.Printf("Band: S-band (2.2 GHz)\n")
	fmt.Printf("Data Rate: %d bps\n", sbandConfig.DataRate)
	fmt.Printf("TX Power: %.1f W\n", sbandConfig.TXPower)
	fmt.Printf("TX Gain: %.1f dBi (directional patch)\n", sbandConfig.TXGain)
	fmt.Printf("RX Gain: %.1f dBi (ground dish)\n", sbandConfig.RXGain)
	fmt.Printf("FEC: %s (%v)\n", sbandConfig.FECType, sbandConfig.FECEnabled)
	fmt.Printf("Light-time delay: %.3f seconds\n", sbandConfig.LightTimeDelay.Seconds())
	fmt.Printf("LTP timeout: %.1f seconds\n", sbandConfig.LTPTimeout.Seconds())
	fmt.Println()

	// Example 2: X-band configuration for cislunar operations
	fmt.Println("Example 2: X-band Cislunar Configuration")
	fmt.Println(strings.Repeat("-", 50))
	
	xbandConfig := sband_iq.DefaultXBandConfig("K2XYZ")
	fmt.Printf("Callsign: %s\n", xbandConfig.Callsign)
	fmt.Printf("Band: X-band (8.4 GHz)\n")
	fmt.Printf("Data Rate: %d bps\n", xbandConfig.DataRate)
	fmt.Printf("TX Power: %.1f W\n", xbandConfig.TXPower)
	fmt.Printf("TX Gain: %.1f dBi\n", xbandConfig.TXGain)
	fmt.Printf("RX Gain: %.1f dBi\n", xbandConfig.RXGain)
	fmt.Println()

	// Example 3: Create S-band CLA and demonstrate store-and-forward
	fmt.Println("Example 3: Cislunar Store-and-Forward Operation")
	fmt.Println(strings.Repeat("-", 50))
	
	cla, err := sband_iq.New(sbandConfig)
	if err != nil {
		fmt.Printf("Error creating CLA: %v\n", err)
		return
	}

	// Define contact window (Earth-Moon communication window)
	window := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "tier3-ground-station",
		StartTime:  time.Now().Unix(),
		EndTime:    time.Now().Add(30 * time.Minute).Unix(), // 30-minute window
		DataRate:   500,
		LinkType:   contact.LinkTypeSBandIQ,
	}

	fmt.Printf("Opening contact window with %s\n", window.RemoteNode)
	fmt.Printf("Duration: 30 minutes, Data rate: %d bps\n", window.DataRate)
	
	err = cla.Open(window)
	if err != nil {
		fmt.Printf("Error opening CLA: %v\n", err)
		return
	}
	defer cla.Close()

	fmt.Println()

	// Example 4: Send high-priority science data bundle
	fmt.Println("Example 4: Transmitting High-Priority Science Data")
	fmt.Println(strings.Repeat("-", 50))
	
	scienceBundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "cislunar-01.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "mission-control.0"},
		Payload:     []byte("SCIENCE_DATA: Temperature=-180C, Radiation=0.5mSv/day, Position=EM-L2"),
		Priority:    bpa.PriorityCritical,
		Lifetime:    86400, // 24 hours
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	fmt.Printf("Bundle ID: %s\n", scienceBundle.ID)
	fmt.Printf("Destination: %s\n", scienceBundle.Destination)
	fmt.Printf("Priority: %s\n", scienceBundle.Priority)
	fmt.Printf("Payload size: %d bytes\n", len(scienceBundle.Payload))
	
	startTime := time.Now()
	metrics, err := cla.SendBundle(scienceBundle)
	if err != nil {
		fmt.Printf("Error sending bundle: %v\n", err)
		return
	}
	transmitTime := time.Since(startTime)

	fmt.Printf("Transmission complete in %.3f seconds\n", transmitTime.Seconds())
	fmt.Printf("Bytes transferred: %d\n", metrics.BytesTransferred)
	fmt.Printf("Active LTP sessions: %d\n", cla.GetActiveSessions())
	fmt.Println()

	// Example 5: Send telemetry bundle
	fmt.Println("Example 5: Transmitting Telemetry Bundle")
	fmt.Println(strings.Repeat("-", 50))
	
	telemetryBundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "cislunar-01.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    2,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "mission-control.0"},
		Payload:     []byte("TELEMETRY: Battery=85%, Solar=12W, Uptime=72h, Storage=45%"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    3600, // 1 hour
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	fmt.Printf("Bundle ID: %s\n", telemetryBundle.ID)
	fmt.Printf("Priority: %s\n", telemetryBundle.Priority)
	
	metrics, err = cla.SendBundle(telemetryBundle)
	if err != nil {
		fmt.Printf("Error sending bundle: %v\n", err)
		return
	}

	fmt.Printf("Transmission complete\n")
	fmt.Printf("Total bytes transferred: %d\n", metrics.BytesTransferred)
	fmt.Println()

	// Example 6: Link metrics and session management
	fmt.Println("Example 6: Link Metrics and Session Management")
	fmt.Println(strings.Repeat("-", 50))
	
	linkMetrics := cla.LinkMetrics()
	fmt.Printf("RSSI: %d dBm (weak signal expected for deep-space)\n", linkMetrics.RSSI)
	fmt.Printf("SNR: %.1f dB\n", linkMetrics.SNR)
	fmt.Printf("Bit Error Rate: %.6f\n", linkMetrics.BitErrorRate)
	fmt.Printf("Total bytes transferred: %d\n", linkMetrics.BytesTransferred)
	fmt.Printf("Active LTP sessions: %d\n", cla.GetActiveSessions())
	fmt.Println()

	// Example 7: Demonstrate light-time delay impact
	fmt.Println("Example 7: Light-Time Delay Impact")
	fmt.Println(strings.Repeat("-", 50))
	
	oneWayDelay := sbandConfig.LightTimeDelay
	roundTripTime := 2 * oneWayDelay
	
	fmt.Printf("One-way light-time delay: %.3f seconds\n", oneWayDelay.Seconds())
	fmt.Printf("Round-trip time (RTT): %.3f seconds\n", roundTripTime.Seconds())
	fmt.Printf("Earth-Moon distance: ~384,000 km\n")
	fmt.Printf("Speed of light: 299,792 km/s\n")
	fmt.Printf("Calculated delay: 384000 / 299792 = %.3f seconds\n", 384000.0/299792.0)
	fmt.Println()
	fmt.Println("Note: LTP session management accounts for this delay")
	fmt.Printf("LTP timeout configured to: %.1f seconds (allows for RTT + processing)\n", 
		sbandConfig.LTPTimeout.Seconds())
	fmt.Println()

	// Example 8: FEC configuration comparison
	fmt.Println("Example 8: Forward Error Correction (FEC) Comparison")
	fmt.Println(strings.Repeat("-", 50))
	
	fmt.Println("LDPC FEC:")
	fmt.Println("  - Low-Density Parity-Check coding")
	fmt.Println("  - Near Shannon-limit performance")
	fmt.Println("  - Suitable for deep-space links")
	fmt.Println("  - Coding gain: ~6-8 dB")
	fmt.Println()
	
	fmt.Println("Turbo FEC:")
	fmt.Println("  - Turbo coding with iterative decoding")
	fmt.Println("  - Excellent performance for low SNR")
	fmt.Println("  - Used in deep-space missions")
	fmt.Println("  - Coding gain: ~5-7 dB")
	fmt.Println()

	// Example 9: Link budget summary
	fmt.Println("Example 9: Cislunar Link Budget Summary")
	fmt.Println(strings.Repeat("-", 50))
	
	fmt.Println("Transmit Parameters:")
	fmt.Printf("  TX Power: %.1f W (%.1f dBm)\n", sbandConfig.TXPower, 10*math.Log10(sbandConfig.TXPower*1000))
	fmt.Printf("  TX Antenna Gain: %.1f dBi\n", sbandConfig.TXGain)
	fmt.Printf("  EIRP: %.1f dBm\n", 10*math.Log10(sbandConfig.TXPower*1000)+sbandConfig.TXGain)
	fmt.Println()
	
	fmt.Println("Receive Parameters:")
	fmt.Printf("  RX Antenna Gain: %.1f dBi (3-5m dish)\n", sbandConfig.RXGain)
	fmt.Printf("  System Temperature: ~50 K (low-noise)\n")
	fmt.Println()
	
	fmt.Println("Link Parameters:")
	fmt.Printf("  Frequency: %.3f GHz\n", sbandConfig.CenterFreq/1e9)
	fmt.Printf("  Distance: 384,000 km (Earth-Moon)\n")
	fmt.Printf("  Free-space path loss: ~267 dB\n")
	fmt.Printf("  Data rate: %d bps\n", sbandConfig.DataRate)
	fmt.Printf("  Modulation: BPSK\n")
	fmt.Printf("  FEC: %s (coding gain ~6-8 dB)\n", sbandConfig.FECType)
	fmt.Println()
	
	fmt.Println("Expected Performance:")
	fmt.Println("  Link margin: 5-7 dB (positive, link closes)")
	fmt.Println("  BER: <1e-6 with FEC")
	fmt.Println("  Availability: >95% during contact windows")
	fmt.Println()

	// Example 10: Cleanup
	fmt.Println("Example 10: Session Cleanup")
	fmt.Println(strings.Repeat("-", 50))
	
	fmt.Printf("Active sessions before cleanup: %d\n", cla.GetActiveSessions())
	cla.CleanupSessions()
	fmt.Printf("Active sessions after cleanup: %d\n", cla.GetActiveSessions())
	fmt.Println()

	fmt.Println("=== Example Complete ===")
}
