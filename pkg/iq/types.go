// Package iq provides IQ (In-phase/Quadrature) baseband sample generation and processing
// for software-defined radio operations on the STM32U585 microcontroller.
//
// This package supports GFSK, GMSK, and BPSK modulation/demodulation for:
// - Engineering Model: UHF 437 MHz via Ettus B200mini SDR
// - LEO Flight: UHF 437 MHz via flight IQ transceiver
// - Cislunar: S-band 2.2 GHz via flight IQ transceiver
//
// The STM32U585 generates TX IQ samples and processes RX IQ samples via DMA,
// providing full software control over modulation/demodulation.
package iq

import (
	"fmt"
	"math"
)

// ModulationType represents the modulation scheme
type ModulationType int

const (
	// ModulationGFSK is Gaussian Frequency Shift Keying (9600 baud, terrestrial/EM)
	ModulationGFSK ModulationType = iota
	// ModulationGMSK is Gaussian Minimum Shift Keying (9.6 kbps, LEO UHF)
	ModulationGMSK
	// ModulationBPSK is Binary Phase Shift Keying (500 bps, cislunar S-band)
	ModulationBPSK
)

func (m ModulationType) String() string {
	switch m {
	case ModulationGFSK:
		return "GFSK"
	case ModulationGMSK:
		return "GMSK"
	case ModulationBPSK:
		return "BPSK"
	default:
		return fmt.Sprintf("Unknown(%d)", m)
	}
}

// IQSample represents a single I/Q baseband sample
type IQSample struct {
	I float64 // In-phase component
	Q float64 // Quadrature component
}

// Complex returns the complex representation of the IQ sample
func (s IQSample) Complex() complex128 {
	return complex(s.I, s.Q)
}

// Magnitude returns the magnitude of the IQ sample
func (s IQSample) Magnitude() float64 {
	return math.Sqrt(s.I*s.I + s.Q*s.Q)
}

// Phase returns the phase angle of the IQ sample in radians
func (s IQSample) Phase() float64 {
	return math.Atan2(s.Q, s.I)
}

// IQBuffer represents a buffer of IQ samples for DMA streaming
type IQBuffer struct {
	Samples    []IQSample
	SampleRate float64 // Samples per second
	Timestamp  int64   // Unix timestamp in nanoseconds
}

// NewIQBuffer creates a new IQ buffer with the specified capacity
func NewIQBuffer(capacity int, sampleRate float64) *IQBuffer {
	return &IQBuffer{
		Samples:    make([]IQSample, 0, capacity),
		SampleRate: sampleRate,
	}
}

// Append adds an IQ sample to the buffer
func (b *IQBuffer) Append(sample IQSample) {
	b.Samples = append(b.Samples, sample)
}

// Clear resets the buffer
func (b *IQBuffer) Clear() {
	b.Samples = b.Samples[:0]
}

// Len returns the number of samples in the buffer
func (b *IQBuffer) Len() int {
	return len(b.Samples)
}

// Cap returns the capacity of the buffer
func (b *IQBuffer) Cap() int {
	return cap(b.Samples)
}

// ModulationConfig holds configuration for modulation/demodulation
type ModulationConfig struct {
	Type           ModulationType
	SampleRate     float64 // Samples per second
	SymbolRate     float64 // Symbols per second (baud rate)
	CarrierFreq    float64 // Carrier frequency in Hz (0 for baseband)
	FrequencyDev   float64 // Frequency deviation for FSK (Hz)
	BTProduct      float64 // Gaussian filter BT product (for GFSK/GMSK)
	SamplesPerSym  int     // Samples per symbol
	FilterTaps     int     // Number of Gaussian filter taps
}

// DefaultGFSKConfig returns default configuration for GFSK modulation (9600 baud)
func DefaultGFSKConfig() ModulationConfig {
	sampleRate := 48000.0 // 48 kHz sample rate
	symbolRate := 9600.0  // 9600 baud
	return ModulationConfig{
		Type:          ModulationGFSK,
		SampleRate:    sampleRate,
		SymbolRate:    symbolRate,
		CarrierFreq:   0, // Baseband
		FrequencyDev:  3000.0,
		BTProduct:     0.5,
		SamplesPerSym: int(sampleRate / symbolRate),
		FilterTaps:    8,
	}
}

// DefaultGMSKConfig returns default configuration for GMSK modulation (9.6 kbps)
func DefaultGMSKConfig() ModulationConfig {
	sampleRate := 48000.0 // 48 kHz sample rate
	symbolRate := 9600.0  // 9.6 kbps
	return ModulationConfig{
		Type:          ModulationGMSK,
		SampleRate:    sampleRate,
		SymbolRate:    symbolRate,
		CarrierFreq:   0, // Baseband
		FrequencyDev:  2400.0, // MSK: deviation = symbol_rate / 4
		BTProduct:     0.3,
		SamplesPerSym: int(sampleRate / symbolRate),
		FilterTaps:    8,
	}
}

// DefaultBPSKConfig returns default configuration for BPSK modulation (500 bps)
func DefaultBPSKConfig() ModulationConfig {
	sampleRate := 8000.0 // 8 kHz sample rate (lower for cislunar)
	symbolRate := 500.0  // 500 bps
	return ModulationConfig{
		Type:          ModulationBPSK,
		SampleRate:    sampleRate,
		SymbolRate:    symbolRate,
		CarrierFreq:   0, // Baseband
		FrequencyDev:  0, // Not used for BPSK
		BTProduct:     0,
		SamplesPerSym: int(sampleRate / symbolRate),
		FilterTaps:    0,
	}
}

// DemodulationState holds state for demodulation
type DemodulationState struct {
	Config         ModulationConfig
	PrevPhase      float64
	PrevSample     IQSample
	SymbolBuffer   []float64
	SymbolIndex    int
	BitBuffer      []byte
	SyncLocked     bool
	SyncThreshold  float64
}

// NewDemodulationState creates a new demodulation state
func NewDemodulationState(config ModulationConfig) *DemodulationState {
	return &DemodulationState{
		Config:        config,
		SymbolBuffer:  make([]float64, config.SamplesPerSym),
		BitBuffer:     make([]byte, 0, 1024),
		SyncThreshold: 0.7,
	}
}

// LinkMetrics holds RF link quality metrics
type LinkMetrics struct {
	RSSI      float64 // Received Signal Strength Indicator (dBm)
	SNR       float64 // Signal-to-Noise Ratio (dB)
	EVM       float64 // Error Vector Magnitude (%)
	FreqError float64 // Frequency error (Hz)
}

// String returns a string representation of link metrics
func (m LinkMetrics) String() string {
	return fmt.Sprintf("RSSI=%.1fdBm SNR=%.1fdB EVM=%.1f%% FreqErr=%.1fHz",
		m.RSSI, m.SNR, m.EVM, m.FreqError)
}
