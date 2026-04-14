package iq

import (
	"math"
)

// Demodulator recovers binary data from IQ baseband samples
type Demodulator struct {
	state *DemodulationState
}

// NewDemodulator creates a new demodulator with the specified configuration
func NewDemodulator(config ModulationConfig) *Demodulator {
	return &Demodulator{
		state: NewDemodulationState(config),
	}
}

// Demodulate converts IQ samples to binary data
func (d *Demodulator) Demodulate(buffer *IQBuffer) ([]byte, LinkMetrics) {
	switch d.state.Config.Type {
	case ModulationGFSK, ModulationGMSK:
		return d.demodulateFSK(buffer)
	case ModulationBPSK:
		return d.demodulateBPSK(buffer)
	default:
		return nil, LinkMetrics{}
	}
}

// demodulateFSK performs GFSK/GMSK demodulation using frequency discrimination
func (d *Demodulator) demodulateFSK(buffer *IQBuffer) ([]byte, LinkMetrics) {
	if len(buffer.Samples) == 0 {
		return nil, LinkMetrics{}
	}

	// Calculate link metrics
	metrics := d.calculateMetrics(buffer)

	// Frequency discrimination: compute instantaneous frequency from phase changes
	phases := make([]float64, len(buffer.Samples))
	for i, sample := range buffer.Samples {
		phases[i] = sample.Phase()
	}

	// Compute phase differences (unwrapped)
	freqs := make([]float64, len(phases)-1)
	for i := 0; i < len(phases)-1; i++ {
		diff := phases[i+1] - phases[i]
		// Unwrap phase
		for diff > math.Pi {
			diff -= 2 * math.Pi
		}
		for diff < -math.Pi {
			diff += 2 * math.Pi
		}
		freqs[i] = diff * d.state.Config.SampleRate / (2 * math.Pi)
	}

	// Symbol decision: sample at symbol rate
	bits := make([]int, 0, len(freqs)/d.state.Config.SamplesPerSym)
	for i := 0; i < len(freqs); i += d.state.Config.SamplesPerSym {
		// Average frequency over symbol period
		sum := 0.0
		count := 0
		for j := 0; j < d.state.Config.SamplesPerSym && i+j < len(freqs); j++ {
			sum += freqs[i+j]
			count++
		}
		avgFreq := sum / float64(count)

		// Decision: positive frequency → bit 1, negative → bit 0
		if avgFreq > 0 {
			bits = append(bits, 1)
		} else {
			bits = append(bits, 0)
		}
	}

	// Convert bits to bytes
	data := bitsToBytes(bits)
	return data, metrics
}

// demodulateBPSK performs BPSK demodulation using phase detection
func (d *Demodulator) demodulateBPSK(buffer *IQBuffer) ([]byte, LinkMetrics) {
	if len(buffer.Samples) == 0 {
		return nil, LinkMetrics{}
	}

	// Calculate link metrics
	metrics := d.calculateMetrics(buffer)

	// Symbol decision: sample at symbol rate
	bits := make([]int, 0, len(buffer.Samples)/d.state.Config.SamplesPerSym)
	for i := 0; i < len(buffer.Samples); i += d.state.Config.SamplesPerSym {
		// Average I component over symbol period (Q should be near zero for BPSK)
		sumI := 0.0
		count := 0
		for j := 0; j < d.state.Config.SamplesPerSym && i+j < len(buffer.Samples); j++ {
			sumI += buffer.Samples[i+j].I
			count++
		}
		avgI := sumI / float64(count)

		// Decision: positive I → bit 0 (phase 0), negative I → bit 1 (phase π)
		if avgI > 0 {
			bits = append(bits, 0)
		} else {
			bits = append(bits, 1)
		}
	}

	// Convert bits to bytes
	data := bitsToBytes(bits)
	return data, metrics
}

// calculateMetrics computes link quality metrics from IQ samples
func (d *Demodulator) calculateMetrics(buffer *IQBuffer) LinkMetrics {
	if len(buffer.Samples) == 0 {
		return LinkMetrics{}
	}

	// Calculate average signal power
	sumPower := 0.0
	for _, sample := range buffer.Samples {
		power := sample.I*sample.I + sample.Q*sample.Q
		sumPower += power
	}
	avgPower := sumPower / float64(len(buffer.Samples))

	// RSSI: convert power to dBm (assuming normalized samples)
	// This is a simplified calculation; real implementation would calibrate
	rssi := 10 * math.Log10(avgPower) - 30 // Arbitrary offset for simulation

	// SNR estimation using signal variance
	// Simplified: assume noise floor at -100 dBm
	noiseFloor := -100.0
	snr := rssi - noiseFloor

	// EVM: simplified calculation
	// Real implementation would compare to ideal constellation points
	evm := 5.0 // Placeholder: 5% EVM

	// Frequency error: simplified
	freqError := 0.0 // Placeholder

	return LinkMetrics{
		RSSI:      rssi,
		SNR:       snr,
		EVM:       evm,
		FreqError: freqError,
	}
}

// Reset resets the demodulator state
func (d *Demodulator) Reset() {
	d.state = NewDemodulationState(d.state.Config)
}
