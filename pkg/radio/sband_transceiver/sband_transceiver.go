// Package sband_transceiver provides an interface to the S-band/X-band IQ transceiver
// for cislunar deep-space communication nodes.
//
// This transceiver supports S-band 2.2 GHz (or X-band 8.4 GHz) operation at 500 bps
// with BPSK modulation and strong forward error correction (LDPC or Turbo coding).
// It accounts for 1-2 second one-way light-time delay in cislunar operations.
//
// The transceiver interfaces directly with the STM32U585 (or higher-capability processor)
// via DAC/ADC or SPI, with no companion host required.
package sband_transceiver

import (
	"fmt"
	"sync"
	"time"

	"terrestrial-dtn/pkg/iq"
)

// SBandTransceiver represents an S-band/X-band IQ transceiver for cislunar operations
type SBandTransceiver struct {
	mu         sync.Mutex
	config     Config
	streaming  bool
	centerFreq float64
	sampleRate float64
	txPower    float64 // Higher power for deep-space link
	txGain     float64
	rxGain     float64
	txBuffer   chan *iq.IQBuffer
	rxBuffer   chan *iq.IQBuffer
	stopChan   chan struct{}
	wg         sync.WaitGroup
	
	// FEC state
	fecEnabled bool
	fecType    FECType
	
	// Delay compensation for cislunar operations
	lightTimeDelay time.Duration // 1-2 second one-way delay
}

// Config holds S-band/X-band transceiver configuration
type Config struct {
	// Interface type (DAC/ADC or SPI)
	InterfaceType InterfaceType

	// DAC/ADC configuration (if using analog interface)
	DACChannel int
	ADCChannel int

	// SPI configuration (if using digital interface)
	SPIDevice string
	SPISpeed  uint32

	// RF configuration
	Band       Band    // S-band or X-band
	CenterFreq float64 // Hz (e.g., 2.2e9 for S-band, 8.4e9 for X-band)
	SampleRate float64 // Samples/second (lower for 500 bps)
	TXPower    float64 // Watts (5-10W for deep-space)
	TXGain     float64 // dB
	RXGain     float64 // dB

	// FEC configuration
	FECEnabled bool
	FECType    FECType

	// DMA buffer size
	BufferSize int
	
	// Light-time delay compensation
	LightTimeDelay time.Duration // 1-2 seconds for Earth-Moon distance
}

// InterfaceType specifies how the transceiver connects to the processor
type InterfaceType int

const (
	// InterfaceDACADC uses analog DAC/ADC interface
	InterfaceDACADC InterfaceType = iota
	// InterfaceSPI uses digital SPI interface
	InterfaceSPI
)

// Band specifies the RF band
type Band int

const (
	// BandS is S-band (2.2 GHz)
	BandS Band = iota
	// BandX is X-band (8.4 GHz)
	BandX
)

func (b Band) String() string {
	switch b {
	case BandS:
		return "S-band"
	case BandX:
		return "X-band"
	default:
		return fmt.Sprintf("Unknown(%d)", b)
	}
}

// FECType specifies the forward error correction coding scheme
type FECType int

const (
	// FECLDPC is Low-Density Parity-Check coding
	FECLDPC FECType = iota
	// FECTurbo is Turbo coding
	FECTurbo
)

func (f FECType) String() string {
	switch f {
	case FECLDPC:
		return "LDPC"
	case FECTurbo:
		return "Turbo"
	default:
		return fmt.Sprintf("Unknown(%d)", f)
	}
}

// DefaultSBandConfig returns default configuration for S-band cislunar operation
func DefaultSBandConfig() Config {
	return Config{
		InterfaceType:  InterfaceDACADC,
		DACChannel:     1,
		ADCChannel:     1,
		Band:           BandS,
		CenterFreq:     2.2e9,  // 2.2 GHz S-band
		SampleRate:     8000.0, // 8 kHz sample rate for 500 bps
		TXPower:        5.0,    // 5W transmit power
		TXGain:         10.0,   // 10 dBi directional patch antenna
		RXGain:         35.0,   // 35 dBi ground dish
		FECEnabled:     true,
		FECType:        FECLDPC,
		BufferSize:     2048,
		LightTimeDelay: 1200 * time.Millisecond, // ~1.2 seconds nominal Earth-Moon delay
	}
}

// DefaultXBandConfig returns default configuration for X-band cislunar operation
func DefaultXBandConfig() Config {
	return Config{
		InterfaceType:  InterfaceDACADC,
		DACChannel:     1,
		ADCChannel:     1,
		Band:           BandX,
		CenterFreq:     8.4e9,  // 8.4 GHz X-band
		SampleRate:     8000.0, // 8 kHz sample rate for 500 bps
		TXPower:        5.0,    // 5W transmit power
		TXGain:         12.0,   // 12 dBi directional antenna
		RXGain:         40.0,   // 40 dBi ground dish
		FECEnabled:     true,
		FECType:        FECLDPC,
		BufferSize:     2048,
		LightTimeDelay: 1200 * time.Millisecond,
	}
}

// New creates a new S-band/X-band transceiver interface
func New(config Config) (*SBandTransceiver, error) {
	return &SBandTransceiver{
		config:         config,
		centerFreq:     config.CenterFreq,
		sampleRate:     config.SampleRate,
		txPower:        config.TXPower,
		txGain:         config.TXGain,
		rxGain:         config.RXGain,
		fecEnabled:     config.FECEnabled,
		fecType:        config.FECType,
		lightTimeDelay: config.LightTimeDelay,
		txBuffer:       make(chan *iq.IQBuffer, 8),
		rxBuffer:       make(chan *iq.IQBuffer, 8),
		stopChan:       make(chan struct{}),
	}, nil
}

// Open initializes the S-band/X-band transceiver
func (t *SBandTransceiver) Open() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// In real implementation, this would:
	// 1. Initialize DAC/ADC or SPI interface
	// 2. Configure transceiver IC registers
	// 3. Set center frequency, sample rate, gains, power
	// 4. Initialize FEC encoder/decoder
	// 5. Enable TX/RX paths
	// For simulation, we just log the configuration

	interfaceStr := "DAC/ADC"
	if t.config.InterfaceType == InterfaceSPI {
		interfaceStr = "SPI"
	}

	fmt.Printf("S-band/X-band Transceiver: Opened (%s interface)\n", interfaceStr)
	fmt.Printf("S-band/X-band Transceiver: Band = %s\n", t.config.Band)
	fmt.Printf("S-band/X-band Transceiver: Center freq = %.3f GHz\n", t.centerFreq/1e9)
	fmt.Printf("S-band/X-band Transceiver: Sample rate = %.3f kHz\n", t.sampleRate/1e3)
	fmt.Printf("S-band/X-band Transceiver: TX power = %.1f W, TX gain = %.1f dBi, RX gain = %.1f dBi\n",
		t.txPower, t.txGain, t.rxGain)
	if t.fecEnabled {
		fmt.Printf("S-band/X-band Transceiver: FEC = %s (enabled)\n", t.fecType)
	} else {
		fmt.Println("S-band/X-band Transceiver: FEC = disabled")
	}
	fmt.Printf("S-band/X-band Transceiver: Light-time delay = %.3f seconds\n",
		t.lightTimeDelay.Seconds())

	return nil
}

// Close shuts down the S-band/X-band transceiver
func (t *SBandTransceiver) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.streaming {
		return fmt.Errorf("cannot close while streaming")
	}

	// In real implementation, disable TX/RX paths and power down
	fmt.Println("S-band/X-band Transceiver: Closed")
	return nil
}

// StartStreaming begins IQ sample streaming
func (t *SBandTransceiver) StartStreaming() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.streaming {
		return fmt.Errorf("already streaming")
	}

	t.streaming = true
	t.stopChan = make(chan struct{})

	// Start TX worker
	t.wg.Add(1)
	go t.txWorker()

	// Start RX worker
	t.wg.Add(1)
	go t.rxWorker()

	fmt.Println("S-band/X-band Transceiver: Started streaming")
	return nil
}

// StopStreaming halts IQ sample streaming
func (t *SBandTransceiver) StopStreaming() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.streaming {
		return fmt.Errorf("not streaming")
	}

	close(t.stopChan)
	t.wg.Wait()
	t.streaming = false

	fmt.Println("S-band/X-band Transceiver: Stopped streaming")
	return nil
}

// Transmit queues an IQ buffer for transmission
func (t *SBandTransceiver) Transmit(buffer *iq.IQBuffer) error {
	select {
	case t.txBuffer <- buffer:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("TX buffer full")
	}
}

// Receive returns a received IQ buffer
// Note: Accounts for light-time delay in cislunar operations
func (t *SBandTransceiver) Receive() (*iq.IQBuffer, error) {
	// Add light-time delay to timeout to account for propagation
	timeout := 2*time.Second + t.lightTimeDelay
	
	select {
	case buffer := <-t.rxBuffer:
		return buffer, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("RX timeout (including %.3fs light-time delay)",
			t.lightTimeDelay.Seconds())
	}
}

// txWorker handles TX streaming via DAC/ADC or SPI
func (t *SBandTransceiver) txWorker() {
	defer t.wg.Done()

	for {
		select {
		case <-t.stopChan:
			return
		case buffer := <-t.txBuffer:
			// In real implementation:
			// 1. Apply FEC encoding (LDPC or Turbo)
			// 2. BPSK modulate to IQ samples
			// 3. Convert IQ samples to transceiver format
			// 4. Write to DAC (analog) or SPI (digital)
			// 5. Processor DMA streams samples to transceiver
			// 6. Transceiver up-converts to S-band/X-band RF and transmits
			// 7. Account for light-time delay in timing

			// For simulation, just log
			if err := t.transmitIQ(buffer); err != nil {
				fmt.Printf("S-band/X-band Transceiver TX error: %v\n", err)
			}
		}
	}
}

// rxWorker handles RX streaming via DAC/ADC or SPI
func (t *SBandTransceiver) rxWorker() {
	defer t.wg.Done()

	ticker := time.NewTicker(time.Duration(float64(t.config.BufferSize) / t.sampleRate * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case <-ticker.C:
			// In real implementation:
			// 1. Transceiver down-converts S-band/X-band RF to baseband IQ
			// 2. Read from ADC (analog) or SPI (digital)
			// 3. Processor DMA receives samples from transceiver
			// 4. BPSK demodulate IQ samples
			// 5. Apply FEC decoding (LDPC or Turbo)
			// 6. Convert to IQ buffer format
			// 7. Account for light-time delay in timing

			// For simulation, create buffer
			buffer, err := t.receiveIQ()
			if err != nil {
				fmt.Printf("S-band/X-band Transceiver RX error: %v\n", err)
				continue
			}

			select {
			case t.rxBuffer <- buffer:
			default:
				// Buffer full, drop samples
				fmt.Println("S-band/X-band Transceiver: RX buffer full, dropping samples")
			}
		}
	}
}

// transmitIQ transmits IQ samples via the transceiver with FEC encoding
func (t *SBandTransceiver) transmitIQ(buffer *iq.IQBuffer) error {
	// In real implementation:
	// 1. Apply FEC encoding (LDPC or Turbo) to improve link budget
	// 2. Write samples to DAC/ADC or SPI
	// 3. Account for higher transmit power (5-10W) for deep-space link
	
	// For simulation, just log
	fecStr := "no FEC"
	if t.fecEnabled {
		fecStr = fmt.Sprintf("%s FEC", t.fecType)
	}
	fmt.Printf("S-band/X-band Transceiver: TX %d samples (BPSK + %s, %.1fW)\n",
		len(buffer.Samples), fecStr, t.txPower)
	return nil
}

// receiveIQ receives IQ samples from the transceiver with FEC decoding
func (t *SBandTransceiver) receiveIQ() (*iq.IQBuffer, error) {
	// In real implementation:
	// 1. Read samples from ADC or SPI
	// 2. Apply FEC decoding (LDPC or Turbo) to correct errors
	// 3. Account for light-time delay in timing
	
	// For simulation, create empty buffer
	buffer := iq.NewIQBuffer(t.config.BufferSize, t.sampleRate)
	buffer.Timestamp = time.Now().UnixNano()
	return buffer, nil
}

// SetCenterFreq sets the center frequency
func (t *SBandTransceiver) SetCenterFreq(freq float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate frequency range based on band
	var minFreq, maxFreq float64
	if t.config.Band == BandS {
		// S-band: 2.0-2.3 GHz
		minFreq = 2.0e9
		maxFreq = 2.3e9
	} else {
		// X-band: 8.0-8.5 GHz
		minFreq = 8.0e9
		maxFreq = 8.5e9
	}

	if freq < minFreq || freq > maxFreq {
		return fmt.Errorf("frequency out of range for %s (%.1f-%.1f GHz)",
			t.config.Band, minFreq/1e9, maxFreq/1e9)
	}

	t.centerFreq = freq
	fmt.Printf("S-band/X-band Transceiver: Set center freq = %.3f GHz\n", freq/1e9)
	return nil
}

// SetSampleRate sets the sample rate
func (t *SBandTransceiver) SetSampleRate(rate float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range for 500 bps: 4-16 kHz
	if rate < 4e3 || rate > 16e3 {
		return fmt.Errorf("sample rate out of range (4-16 kHz)")
	}

	t.sampleRate = rate
	fmt.Printf("S-band/X-band Transceiver: Set sample rate = %.3f kHz\n", rate/1e3)
	return nil
}

// SetTXPower sets the transmit power
func (t *SBandTransceiver) SetTXPower(power float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 1-10 W for deep-space
	if power < 1.0 || power > 10.0 {
		return fmt.Errorf("TX power out of range (1-10 W)")
	}

	t.txPower = power
	fmt.Printf("S-band/X-band Transceiver: Set TX power = %.1f W\n", power)
	return nil
}

// SetTXGain sets the TX gain
func (t *SBandTransceiver) SetTXGain(gain float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 0-30 dB
	if gain < 0 || gain > 30 {
		return fmt.Errorf("TX gain out of range (0-30 dB)")
	}

	t.txGain = gain
	fmt.Printf("S-band/X-band Transceiver: Set TX gain = %.1f dB\n", gain)
	return nil
}

// SetRXGain sets the RX gain
func (t *SBandTransceiver) SetRXGain(gain float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 0-60 dB
	if gain < 0 || gain > 60 {
		return fmt.Errorf("RX gain out of range (0-60 dB)")
	}

	t.rxGain = gain
	fmt.Printf("S-band/X-band Transceiver: Set RX gain = %.1f dB\n", gain)
	return nil
}

// SetLightTimeDelay sets the light-time delay compensation
func (t *SBandTransceiver) SetLightTimeDelay(delay time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 1-2.5 seconds for Earth-Moon distance
	if delay < 1*time.Second || delay > 3*time.Second {
		return fmt.Errorf("light-time delay out of range (1-3 seconds)")
	}

	t.lightTimeDelay = delay
	fmt.Printf("S-band/X-band Transceiver: Set light-time delay = %.3f seconds\n", delay.Seconds())
	return nil
}

// EnableFEC enables forward error correction
func (t *SBandTransceiver) EnableFEC(fecType FECType) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.fecEnabled = true
	t.fecType = fecType
	fmt.Printf("S-band/X-band Transceiver: Enabled %s FEC\n", fecType)
	return nil
}

// DisableFEC disables forward error correction
func (t *SBandTransceiver) DisableFEC() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.fecEnabled = false
	fmt.Println("S-band/X-band Transceiver: Disabled FEC")
	return nil
}

// GetCenterFreq returns the current center frequency
func (t *SBandTransceiver) GetCenterFreq() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.centerFreq
}

// GetSampleRate returns the current sample rate
func (t *SBandTransceiver) GetSampleRate() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sampleRate
}

// GetTXPower returns the current transmit power
func (t *SBandTransceiver) GetTXPower() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.txPower
}

// GetLightTimeDelay returns the current light-time delay
func (t *SBandTransceiver) GetLightTimeDelay() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lightTimeDelay
}

// IsFECEnabled returns whether FEC is enabled
func (t *SBandTransceiver) IsFECEnabled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fecEnabled
}

// GetFECType returns the current FEC type
func (t *SBandTransceiver) GetFECType() FECType {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fecType
}

// IsStreaming returns whether the transceiver is currently streaming
func (t *SBandTransceiver) IsStreaming() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.streaming
}

// GetBand returns the RF band
func (t *SBandTransceiver) GetBand() Band {
	return t.config.Band
}

// GetInterfaceType returns the interface type
func (t *SBandTransceiver) GetInterfaceType() InterfaceType {
	return t.config.InterfaceType
}

// GetLinkMetrics returns current link quality metrics
func (t *SBandTransceiver) GetLinkMetrics() iq.LinkMetrics {
	t.mu.Lock()
	defer t.mu.Unlock()

	// In real implementation, read from transceiver hardware
	// For simulation, return placeholder values
	return iq.LinkMetrics{
		RSSI:      -120.0, // Weak signal for deep-space
		SNR:       8.0,    // Marginal SNR with FEC
		EVM:       15.0,   // Higher EVM due to long path
		FreqError: 50.0,   // Doppler shift
	}
}
