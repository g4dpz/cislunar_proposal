// Package linkbudget implements RF link budget computation for DTN links.
// Computes free-space path loss (FSPL), received power, and link margin
// for LEO UHF and cislunar S-band configurations.
package linkbudget

import (
	"fmt"
	"math"
)

// LinkBudgetParams contains all parameters for link budget computation
type LinkBudgetParams struct {
	TxPowerDBm       float64 // Transmit power in dBm
	TxAntennaGainDBi float64 // Transmit antenna gain in dBi
	RxAntennaGainDBi float64 // Receive antenna gain in dBi
	FrequencyHz      float64 // Carrier frequency in Hz
	DistanceM        float64 // Link distance in meters
	SystemLossDB     float64 // System losses (cable, pointing, atmospheric) in dB
	DataRateBps      float64 // Data rate in bits per second
	RequiredEbN0DB   float64 // Required Eb/N0 for target BER in dB
	NoiseDensityDBmHz float64 // Noise density in dBm/Hz (typically -174 dBm/Hz)
}

// LinkBudgetResult contains the computed link budget results
type LinkBudgetResult struct {
	FSPL          float64 // Free-space path loss in dB
	ReceivedPower float64 // Received power in dBm
	NoisePower    float64 // Noise power in dBm
	EbN0          float64 // Eb/N0 in dB
	LinkMargin    float64 // Link margin in dB (positive = link closes)
}

// ComputeLinkBudget computes the link budget for the given parameters.
// Returns the link budget result with FSPL, received power, and link margin.
//
// Free-space path loss (FSPL) formula:
//   FSPL (dB) = 20*log10(distance_m) + 20*log10(frequency_Hz) - 147.55
//
// Link budget:
//   Received Power (dBm) = TX Power + TX Gain - FSPL + RX Gain - System Loss
//   Noise Power (dBm) = Noise Density + 10*log10(Data Rate)
//   Eb/N0 (dB) = Received Power - Noise Power
//   Link Margin (dB) = Eb/N0 - Required Eb/N0
//
// A positive link margin indicates the link closes (sufficient SNR).
func ComputeLinkBudget(params LinkBudgetParams) (*LinkBudgetResult, error) {
	// Validate parameters
	if params.DistanceM <= 0 {
		return nil, fmt.Errorf("distance must be positive, got %.2f m", params.DistanceM)
	}
	if params.FrequencyHz <= 0 {
		return nil, fmt.Errorf("frequency must be positive, got %.2f Hz", params.FrequencyHz)
	}
	if params.DataRateBps <= 0 {
		return nil, fmt.Errorf("data rate must be positive, got %.2f bps", params.DataRateBps)
	}

	// Compute free-space path loss (FSPL)
	// FSPL (dB) = 20*log10(d) + 20*log10(f) - 147.55
	fspl := 20*math.Log10(params.DistanceM) + 20*math.Log10(params.FrequencyHz) - 147.55

	// Compute received power
	// Received Power = TX Power + TX Gain - FSPL + RX Gain - System Loss
	receivedPower := params.TxPowerDBm + params.TxAntennaGainDBi - fspl +
		params.RxAntennaGainDBi - params.SystemLossDB

	// Compute noise power
	// Noise Power = Noise Density + 10*log10(Data Rate)
	noisePower := params.NoiseDensityDBmHz + 10*math.Log10(params.DataRateBps)

	// Compute Eb/N0
	// Eb/N0 = Received Power - Noise Power
	ebN0 := receivedPower - noisePower

	// Compute link margin
	// Link Margin = Eb/N0 - Required Eb/N0
	linkMargin := ebN0 - params.RequiredEbN0DB

	return &LinkBudgetResult{
		FSPL:          fspl,
		ReceivedPower: receivedPower,
		NoisePower:    noisePower,
		EbN0:          ebN0,
		LinkMargin:    linkMargin,
	}, nil
}

// LEOUHFParams returns the standard LEO UHF link budget parameters
// (2W TX, omni antenna, 437 MHz, 500 km, 9.6 kbps, Yagi ground antenna)
func LEOUHFParams() LinkBudgetParams {
	return LinkBudgetParams{
		TxPowerDBm:       33.0,  // 2W
		TxAntennaGainDBi: 0.0,   // omni
		RxAntennaGainDBi: 12.0,  // Yagi
		FrequencyHz:      437e6, // UHF 437 MHz
		DistanceM:        500e3, // 500 km
		SystemLossDB:     2.0,   // cable, pointing losses
		DataRateBps:      9600.0,
		RequiredEbN0DB:   10.0,       // GMSK/BPSK
		NoiseDensityDBmHz: -174.0,     // standard thermal noise
	}
}

// CislunarSBandParams returns the standard cislunar S-band link budget parameters
// (5W TX, 10 dBi patch, 2.2 GHz, 384,000 km, 500 bps, 35 dBi ground dish, BPSK + LDPC)
func CislunarSBandParams() LinkBudgetParams {
	return LinkBudgetParams{
		TxPowerDBm:       37.0,  // 5W
		TxAntennaGainDBi: 10.0,  // directional patch
		RxAntennaGainDBi: 35.0,  // 3-5m dish
		FrequencyHz:      2.2e9, // S-band 2.2 GHz
		DistanceM:        384e6, // Earth-Moon distance
		SystemLossDB:     3.0,   // cable, pointing, atmospheric losses
		DataRateBps:      500.0,
		RequiredEbN0DB:   2.0,   // BPSK + strong LDPC
		NoiseDensityDBmHz: -174.0,
	}
}

// String returns a human-readable summary of the link budget result
func (r *LinkBudgetResult) String() string {
	return fmt.Sprintf("FSPL: %.1f dB, RX Power: %.1f dBm, Noise: %.1f dBm, Eb/N0: %.1f dB, Margin: %.1f dB",
		r.FSPL, r.ReceivedPower, r.NoisePower, r.EbN0, r.LinkMargin)
}

// LinkCloses returns true if the link closes (positive margin)
func (r *LinkBudgetResult) LinkCloses() bool {
	return r.LinkMargin > 0
}
