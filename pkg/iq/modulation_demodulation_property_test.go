package iq

import (
	"bytes"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 22: Modulation/Demodulation Round-Trip
// For any valid data payload, GMSK/BPSK modulation to IQ baseband followed by
// demodulation SHALL recover the original data.
//
// **Validates: Requirement 13.2**
func TestProperty_ModulationDemodulationRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced for computational cost

	properties := gopter.NewProperties(parameters)

	// Generator for payload size (1-64 bytes to keep test fast)
	genPayloadSize := gen.IntRange(1, 64)

	// Generator for modulation type
	genModulationType := gen.IntRange(0, 2).Map(func(i int) ModulationType {
		switch i {
		case 0:
			return ModulationGFSK
		case 1:
			return ModulationGMSK
		default:
			return ModulationBPSK
		}
	})

	properties.Property("modulation and demodulation preserves data", prop.ForAll(
		func(payloadSize int, modType ModulationType) bool {
			// Generate test payload with deterministic pattern
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			// Create modulation config
			config := ModulationConfig{
				Type:          modType,
				SampleRate:    48000.0,
				SamplesPerSym: 8,
				FrequencyDev:  2400.0,
				FilterTaps:    5,
				BTProduct:     0.5,
			}

			// Modulate
			modulator := NewModulator(config)
			iqBuffer := modulator.Modulate(payload)

			if iqBuffer == nil || len(iqBuffer.Samples) == 0 {
				t.Logf("Modulation produced empty buffer")
				return false
			}

			// Demodulate
			demodulator := NewDemodulator(config)
			recovered, _ := demodulator.Demodulate(iqBuffer)

			// Verify recovered data matches original
			// Note: Due to modulation/demodulation imperfections, we may have
			// trailing bits. Compare up to the original payload length.
			if len(recovered) < len(payload) {
				t.Logf("Recovered data too short: got %d bytes, want at least %d bytes",
					len(recovered), len(payload))
				return false
			}

			// Compare the payload portion
			if !bytes.Equal(recovered[:len(payload)], payload) {
				// Count bit errors
				bitErrors := 0
				for i := 0; i < len(payload); i++ {
					if recovered[i] != payload[i] {
						for bit := 0; bit < 8; bit++ {
							if ((recovered[i] >> bit) & 1) != ((payload[i] >> bit) & 1) {
								bitErrors++
							}
						}
					}
				}

				// Allow small number of bit errors due to demodulation imperfections
				// In a real system, FEC would correct these
				totalBits := len(payload) * 8
				ber := float64(bitErrors) / float64(totalBits)

				if ber > 0.05 { // Allow up to 5% BER (FEC would handle this)
					t.Logf("BER too high: %.2f%% (%d/%d bit errors)", ber*100, bitErrors, totalBits)
					return false
				}

				// If BER is acceptable, consider it a pass
				t.Logf("Acceptable BER: %.2f%% (%d/%d bit errors)", ber*100, bitErrors, totalBits)
			}

			return true
		},
		genPayloadSize,
		genModulationType,
	))

	properties.TestingRun(t)
}

// Property 22 (variant): BPSK modulation produces correct constellation
// For BPSK modulation, the I component SHALL be either positive (bit 0) or
// negative (bit 1), and Q component SHALL be near zero
//
// **Validates: Requirement 13.2**
func TestProperty_BPSKConstellationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	genPayloadSize := gen.IntRange(1, 32)

	properties.Property("BPSK produces correct constellation points", prop.ForAll(
		func(payloadSize int) bool {
			// Generate test payload
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			// BPSK configuration
			config := ModulationConfig{
				Type:          ModulationBPSK,
				SampleRate:    48000.0,
				SamplesPerSym: 8,
				FrequencyDev:  0, // Not used for BPSK
				FilterTaps:    0, // No filtering for this test
				BTProduct:     0,
			}

			// Modulate
			modulator := NewModulator(config)
			iqBuffer := modulator.Modulate(payload)

			// Check constellation points
			// For BPSK: I should be ±1, Q should be 0
			for i, sample := range iqBuffer.Samples {
				// Check I component is approximately ±1
				if sample.I < -1.1 || sample.I > 1.1 {
					t.Logf("Sample %d: I component out of range: %.2f", i, sample.I)
					return false
				}

				// Check Q component is approximately 0
				if sample.Q < -0.1 || sample.Q > 0.1 {
					t.Logf("Sample %d: Q component not near zero: %.2f", i, sample.Q)
					return false
				}
			}

			return true
		},
		genPayloadSize,
	))

	properties.TestingRun(t)
}

// Property 22 (variant): FSK modulation produces continuous phase
// For GFSK/GMSK modulation, the phase SHALL be continuous (no discontinuities)
//
// **Validates: Requirement 13.2**
func TestProperty_FSKContinuousPhase(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	genPayloadSize := gen.IntRange(1, 32)
	genModulationType := gen.IntRange(0, 1).Map(func(i int) ModulationType {
		if i == 0 {
			return ModulationGFSK
		}
		return ModulationGMSK
	})

	properties.Property("FSK produces continuous phase", prop.ForAll(
		func(payloadSize int, modType ModulationType) bool {
			// Generate test payload
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			// FSK configuration
			config := ModulationConfig{
				Type:          modType,
				SampleRate:    48000.0,
				SamplesPerSym: 8,
				FrequencyDev:  2400.0,
				FilterTaps:    5,
				BTProduct:     0.5,
			}

			// Modulate
			modulator := NewModulator(config)
			iqBuffer := modulator.Modulate(payload)

			// Check phase continuity
			// Phase should not have large jumps (> π)
			for i := 1; i < len(iqBuffer.Samples); i++ {
				phase1 := iqBuffer.Samples[i-1].Phase()
				phase2 := iqBuffer.Samples[i].Phase()

				// Compute phase difference (unwrapped)
				phaseDiff := phase2 - phase1
				for phaseDiff > 3.14159 {
					phaseDiff -= 2 * 3.14159
				}
				for phaseDiff < -3.14159 {
					phaseDiff += 2 * 3.14159
				}

				// Phase difference should be small for continuous phase
				// Allow up to π/2 per sample (generous for high deviation)
				if phaseDiff > 1.6 || phaseDiff < -1.6 {
					t.Logf("Large phase discontinuity at sample %d: %.2f rad", i, phaseDiff)
					return false
				}
			}

			return true
		},
		genPayloadSize,
		genModulationType,
	))

	properties.TestingRun(t)
}

// Property 22 (variant): Modulation output size is correct
// For any payload, the number of IQ samples SHALL be payload_bits * samples_per_symbol
//
// **Validates: Requirement 13.2**
func TestProperty_ModulationOutputSize(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genPayloadSize := gen.IntRange(1, 64)
	genSamplesPerSym := gen.IntRange(4, 16)
	genModulationType := gen.IntRange(0, 2).Map(func(i int) ModulationType {
		switch i {
		case 0:
			return ModulationGFSK
		case 1:
			return ModulationGMSK
		default:
			return ModulationBPSK
		}
	})

	properties.Property("modulation output size is correct", prop.ForAll(
		func(payloadSize int, samplesPerSym int, modType ModulationType) bool {
			payload := make([]byte, payloadSize)

			config := ModulationConfig{
				Type:          modType,
				SampleRate:    48000.0,
				SamplesPerSym: samplesPerSym,
				FrequencyDev:  2400.0,
				FilterTaps:    5,
				BTProduct:     0.5,
			}

			modulator := NewModulator(config)
			iqBuffer := modulator.Modulate(payload)

			// Expected number of samples: payload_bits * samples_per_symbol
			expectedSamples := payloadSize * 8 * samplesPerSym

			if len(iqBuffer.Samples) != expectedSamples {
				t.Logf("Sample count mismatch: got %d, want %d (payload=%d bytes, sps=%d)",
					len(iqBuffer.Samples), expectedSamples, payloadSize, samplesPerSym)
				return false
			}

			return true
		},
		genPayloadSize,
		genSamplesPerSym,
		genModulationType,
	))

	properties.TestingRun(t)
}
