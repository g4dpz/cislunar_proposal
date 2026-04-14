package iq

import (
	"math"
)

// Modulator generates IQ baseband samples from binary data
type Modulator struct {
	config       ModulationConfig
	phase        float64
	gaussianTaps []float64
}

// NewModulator creates a new modulator with the specified configuration
func NewModulator(config ModulationConfig) *Modulator {
	m := &Modulator{
		config: config,
		phase:  0,
	}

	// Generate Gaussian filter taps for GFSK/GMSK
	if config.Type == ModulationGFSK || config.Type == ModulationGMSK {
		m.gaussianTaps = generateGaussianTaps(config.FilterTaps, config.BTProduct)
	}

	return m
}

// Modulate converts binary data to IQ samples
func (m *Modulator) Modulate(data []byte) *IQBuffer {
	buffer := NewIQBuffer(len(data)*8*m.config.SamplesPerSym, m.config.SampleRate)

	switch m.config.Type {
	case ModulationGFSK, ModulationGMSK:
		m.modulateFSK(data, buffer)
	case ModulationBPSK:
		m.modulateBPSK(data, buffer)
	}

	return buffer
}

// modulateFSK performs GFSK/GMSK modulation
func (m *Modulator) modulateFSK(data []byte, buffer *IQBuffer) {
	// Convert bytes to bits
	bits := bytesToBits(data)

	// Apply Gaussian filtering to bit stream
	filtered := m.applyGaussianFilter(bits)

	// Generate IQ samples
	for _, symbol := range filtered {
		// Frequency deviation based on symbol (+1 or -1)
		freq := symbol * m.config.FrequencyDev

		// Generate samples for this symbol
		for i := 0; i < m.config.SamplesPerSym; i++ {
			// Phase accumulation
			m.phase += 2 * math.Pi * freq / m.config.SampleRate

			// Wrap phase to [-π, π]
			for m.phase > math.Pi {
				m.phase -= 2 * math.Pi
			}
			for m.phase < -math.Pi {
				m.phase += 2 * math.Pi
			}

			// Generate I/Q sample
			sample := IQSample{
				I: math.Cos(m.phase),
				Q: math.Sin(m.phase),
			}
			buffer.Append(sample)
		}
	}
}

// modulateBPSK performs BPSK modulation
func (m *Modulator) modulateBPSK(data []byte, buffer *IQBuffer) {
	// Convert bytes to bits
	bits := bytesToBits(data)

	// Generate IQ samples
	for _, bit := range bits {
		// BPSK: bit 0 → phase 0, bit 1 → phase π
		phase := 0.0
		if bit == 1 {
			phase = math.Pi
		}

		// Generate samples for this symbol
		for i := 0; i < m.config.SamplesPerSym; i++ {
			sample := IQSample{
				I: math.Cos(phase),
				Q: math.Sin(phase),
			}
			buffer.Append(sample)
		}
	}
}

// applyGaussianFilter applies Gaussian pulse shaping to the bit stream
func (m *Modulator) applyGaussianFilter(bits []int) []float64 {
	if len(m.gaussianTaps) == 0 {
		// No filtering, convert bits to symbols directly
		symbols := make([]float64, len(bits))
		for i, bit := range bits {
			if bit == 0 {
				symbols[i] = -1.0
			} else {
				symbols[i] = 1.0
			}
		}
		return symbols
	}

	// Convert bits to NRZ symbols (-1, +1)
	nrz := make([]float64, len(bits))
	for i, bit := range bits {
		if bit == 0 {
			nrz[i] = -1.0
		} else {
			nrz[i] = 1.0
		}
	}

	// Apply Gaussian filter via convolution
	filtered := make([]float64, len(nrz))
	halfTaps := len(m.gaussianTaps) / 2

	for i := range nrz {
		sum := 0.0
		for j, tap := range m.gaussianTaps {
			idx := i - halfTaps + j
			if idx >= 0 && idx < len(nrz) {
				sum += tap * nrz[idx]
			}
		}
		filtered[i] = sum
	}

	return filtered
}

// generateGaussianTaps generates Gaussian filter coefficients
func generateGaussianTaps(numTaps int, btProduct float64) []float64 {
	taps := make([]float64, numTaps)
	center := float64(numTaps-1) / 2.0
	alpha := math.Sqrt(math.Log(2) / 2) / btProduct

	sum := 0.0
	for i := 0; i < numTaps; i++ {
		t := float64(i) - center
		taps[i] = math.Exp(-alpha * alpha * t * t)
		sum += taps[i]
	}

	// Normalize
	for i := range taps {
		taps[i] /= sum
	}

	return taps
}

// bytesToBits converts a byte slice to a bit slice
func bytesToBits(data []byte) []int {
	bits := make([]int, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			if (b & (1 << uint(7-j))) != 0 {
				bits[i*8+j] = 1
			} else {
				bits[i*8+j] = 0
			}
		}
	}
	return bits
}

// bitsToBytes converts a bit slice to a byte slice
func bitsToBytes(bits []int) []byte {
	numBytes := (len(bits) + 7) / 8
	data := make([]byte, numBytes)

	for i := 0; i < len(bits); i++ {
		if bits[i] == 1 {
			byteIdx := i / 8
			bitIdx := 7 - (i % 8)
			data[byteIdx] |= 1 << uint(bitIdx)
		}
	}

	return data
}
