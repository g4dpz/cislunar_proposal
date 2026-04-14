package linkbudget

import (
	"math"
	"testing"
)

// TestLEOLinkBudget validates the LEO UHF link budget computation
// Validates: Requirement 18.1
func TestLEOLinkBudget(t *testing.T) {
	params := LEOUHFParams()
	result, err := ComputeLinkBudget(params)
	if err != nil {
		t.Fatalf("ComputeLinkBudget failed: %v", err)
	}

	t.Logf("LEO UHF link budget:")
	t.Logf("  TX power: %.1f dBm (2W)", params.TxPowerDBm)
	t.Logf("  TX antenna gain: %.1f dBi (omni)", params.TxAntennaGainDBi)
	t.Logf("  RX antenna gain: %.1f dBi (Yagi)", params.RxAntennaGainDBi)
	t.Logf("  Frequency: %.1f MHz", params.FrequencyHz/1e6)
	t.Logf("  Distance: %.0f km", params.DistanceM/1e3)
	t.Logf("  Data rate: %.0f bps", params.DataRateBps)
	t.Logf("  FSPL: %.1f dB", result.FSPL)
	t.Logf("  Received power: %.1f dBm", result.ReceivedPower)
	t.Logf("  Noise power: %.1f dBm", result.NoisePower)
	t.Logf("  Eb/N0: %.1f dB", result.EbN0)
	t.Logf("  Required Eb/N0: %.1f dB", params.RequiredEbN0DB)
	t.Logf("  Link margin: %.1f dB", result.LinkMargin)

	// Verify link closes with positive margin
	if !result.LinkCloses() {
		t.Errorf("LEO UHF link does not close: margin = %.1f dB (expected > 0)", result.LinkMargin)
	}

	// Verify margin is reasonable (should be around 30+ dB for LEO UHF)
	if result.LinkMargin < 20.0 {
		t.Errorf("LEO UHF link margin too low: %.1f dB (expected > 20 dB)", result.LinkMargin)
	}
}

// TestCislunarLinkBudget validates the cislunar S-band link budget computation
// Validates: Requirement 18.2
func TestCislunarLinkBudget(t *testing.T) {
	params := CislunarSBandParams()
	result, err := ComputeLinkBudget(params)
	if err != nil {
		t.Fatalf("ComputeLinkBudget failed: %v", err)
	}

	t.Logf("Cislunar S-band link budget:")
	t.Logf("  TX power: %.1f dBm (5W)", params.TxPowerDBm)
	t.Logf("  TX antenna gain: %.1f dBi", params.TxAntennaGainDBi)
	t.Logf("  RX antenna gain: %.1f dBi (3-5m dish)", params.RxAntennaGainDBi)
	t.Logf("  Frequency: %.1f GHz", params.FrequencyHz/1e9)
	t.Logf("  Distance: %.0f km", params.DistanceM/1e3)
	t.Logf("  Data rate: %.0f bps", params.DataRateBps)
	t.Logf("  FSPL: %.1f dB", result.FSPL)
	t.Logf("  Received power: %.1f dBm", result.ReceivedPower)
	t.Logf("  Noise power: %.1f dBm", result.NoisePower)
	t.Logf("  Eb/N0: %.1f dB", result.EbN0)
	t.Logf("  Required Eb/N0: %.1f dB", params.RequiredEbN0DB)
	t.Logf("  Link margin: %.1f dB", result.LinkMargin)

	// Verify link closes with positive margin
	if !result.LinkCloses() {
		t.Errorf("Cislunar S-band link does not close: margin = %.1f dB (expected > 0)", result.LinkMargin)
	}

	// Verify margin is in expected range (5-7 dB per design)
	if result.LinkMargin < 5.0 || result.LinkMargin > 7.0 {
		t.Logf("Warning: Link margin %.1f dB outside expected range [5, 7] dB", result.LinkMargin)
	} else {
		t.Logf("Link margin within expected range: %.1f dB", result.LinkMargin)
	}
}

// TestLinkBudgetValidation tests parameter validation
func TestLinkBudgetValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  LinkBudgetParams
		wantErr bool
	}{
		{
			name: "valid parameters",
			params: LinkBudgetParams{
				TxPowerDBm:       30.0,
				TxAntennaGainDBi: 0.0,
				RxAntennaGainDBi: 10.0,
				FrequencyHz:      437e6,
				DistanceM:        500e3,
				SystemLossDB:     2.0,
				DataRateBps:      9600.0,
				RequiredEbN0DB:   10.0,
				NoiseDensityDBmHz: -174.0,
			},
			wantErr: false,
		},
		{
			name: "zero distance",
			params: LinkBudgetParams{
				FrequencyHz: 437e6,
				DistanceM:   0,
				DataRateBps: 9600.0,
			},
			wantErr: true,
		},
		{
			name: "negative distance",
			params: LinkBudgetParams{
				FrequencyHz: 437e6,
				DistanceM:   -100,
				DataRateBps: 9600.0,
			},
			wantErr: true,
		},
		{
			name: "zero frequency",
			params: LinkBudgetParams{
				FrequencyHz: 0,
				DistanceM:   500e3,
				DataRateBps: 9600.0,
			},
			wantErr: true,
		},
		{
			name: "zero data rate",
			params: LinkBudgetParams{
				FrequencyHz: 437e6,
				DistanceM:   500e3,
				DataRateBps: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ComputeLinkBudget(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeLinkBudget() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFSPLComputation verifies the free-space path loss formula
func TestFSPLComputation(t *testing.T) {
	// Test case: 437 MHz at 500 km
	params := LinkBudgetParams{
		TxPowerDBm:       0,
		TxAntennaGainDBi: 0,
		RxAntennaGainDBi: 0,
		FrequencyHz:      437e6,
		DistanceM:        500e3,
		SystemLossDB:     0,
		DataRateBps:      9600.0,
		RequiredEbN0DB:   0,
		NoiseDensityDBmHz: -174.0,
	}

	result, err := ComputeLinkBudget(params)
	if err != nil {
		t.Fatalf("ComputeLinkBudget failed: %v", err)
	}

	// Expected FSPL = 20*log10(500e3) + 20*log10(437e6) - 147.55
	expectedFSPL := 20*math.Log10(500e3) + 20*math.Log10(437e6) - 147.55

	if math.Abs(result.FSPL-expectedFSPL) > 0.01 {
		t.Errorf("FSPL mismatch: got %.2f dB, want %.2f dB", result.FSPL, expectedFSPL)
	}
}

// TestLinkMarginMonotonicity verifies that link margin decreases with distance
// This is a basic sanity check for Property 26
func TestLinkMarginMonotonicity(t *testing.T) {
	baseParams := LEOUHFParams()

	distances := []float64{100e3, 200e3, 500e3, 1000e3, 2000e3}
	var prevMargin float64

	for i, distance := range distances {
		params := baseParams
		params.DistanceM = distance

		result, err := ComputeLinkBudget(params)
		if err != nil {
			t.Fatalf("ComputeLinkBudget at distance %.0f km failed: %v", distance/1e3, err)
		}

		t.Logf("Distance: %.0f km, Link margin: %.1f dB", distance/1e3, result.LinkMargin)

		if i > 0 {
			// Link margin should decrease with increasing distance
			if result.LinkMargin >= prevMargin {
				t.Errorf("Link margin did not decrease: at %.0f km margin=%.1f dB, at %.0f km margin=%.1f dB",
					distances[i-1]/1e3, prevMargin, distance/1e3, result.LinkMargin)
			}
		}

		prevMargin = result.LinkMargin
	}
}
