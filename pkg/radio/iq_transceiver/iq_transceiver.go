// Package iq_transceiver provides an interface to the flight-qualified IQ transceiver IC
// for LEO CubeSat and cislunar flight nodes.
//
// This replaces the B200mini SDR used in the EM phase. The flight transceiver interfaces
// directly with the STM32U585 via DAC/ADC or SPI, with no companion host required.
//
// Supports GMSK/BPSK at 9.6 kbps on UHF 437 MHz for LEO operations.
package iq_transceiver

import (
	"fmt"
	"sync"
	"time"

	"terrestrial-dtn/pkg/iq"
)

// IQTransceiver represents a flight-qualified IQ transceiver IC
type IQTransceiver struct {
	mu         sync.Mutex
	config     Config
	streaming  bool
	centerFreq float64
	sampleRate float64
	txGain     float64
	rxGain     float64
	txBuffer   chan *iq.IQBuffer
	rxBuffer   chan *iq.IQBuffer
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// Config holds IQ transceiver configuration
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
	CenterFreq float64 // Hz (e.g., 437e6 for UHF 437 MHz)
	SampleRate float64 // Samples/second
	TXGain     float64 // dB
	RXGain     float64 // dB

	// DMA buffer size
	BufferSize int
}

// InterfaceType specifies how the transceiver connects to STM32U585
type InterfaceType int

const (
	// InterfaceDACADC uses analog DAC/ADC interface
	InterfaceDACADC InterfaceType = iota
	// InterfaceSPI uses digital SPI interface
	InterfaceSPI
)

// DefaultConfig returns default configuration for LEO UHF operation
func DefaultConfig() Config {
	return Config{
		InterfaceType: InterfaceDACADC,
		DACChannel:    1,
		ADCChannel:    1,
		CenterFreq:    437e6,  // 437 MHz UHF
		SampleRate:    1e6,    // 1 MHz sample rate
		TXGain:        20.0,   // 20 dB TX gain
		RXGain:        30.0,   // 30 dB RX gain
		BufferSize:    4096,
	}
}

// New creates a new IQ transceiver interface
func New(config Config) (*IQTransceiver, error) {
	return &IQTransceiver{
		config:     config,
		centerFreq: config.CenterFreq,
		sampleRate: config.SampleRate,
		txGain:     config.TXGain,
		rxGain:     config.RXGain,
		txBuffer:   make(chan *iq.IQBuffer, 8),
		rxBuffer:   make(chan *iq.IQBuffer, 8),
		stopChan:   make(chan struct{}),
	}, nil
}

// Open initializes the IQ transceiver
func (t *IQTransceiver) Open() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// In real implementation, this would:
	// 1. Initialize DAC/ADC or SPI interface
	// 2. Configure transceiver IC registers
	// 3. Set center frequency, sample rate, gains
	// 4. Enable TX/RX paths
	// For simulation, we just log the configuration

	interfaceStr := "DAC/ADC"
	if t.config.InterfaceType == InterfaceSPI {
		interfaceStr = "SPI"
	}

	fmt.Printf("IQ Transceiver: Opened (%s interface)\n", interfaceStr)
	fmt.Printf("IQ Transceiver: Center freq = %.3f MHz\n", t.centerFreq/1e6)
	fmt.Printf("IQ Transceiver: Sample rate = %.3f MHz\n", t.sampleRate/1e6)
	fmt.Printf("IQ Transceiver: TX gain = %.1f dB, RX gain = %.1f dB\n", t.txGain, t.rxGain)

	return nil
}

// Close shuts down the IQ transceiver
func (t *IQTransceiver) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.streaming {
		return fmt.Errorf("cannot close while streaming")
	}

	// In real implementation, disable TX/RX paths and power down
	fmt.Println("IQ Transceiver: Closed")
	return nil
}

// StartStreaming begins IQ sample streaming
func (t *IQTransceiver) StartStreaming() error {
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

	fmt.Println("IQ Transceiver: Started streaming")
	return nil
}

// StopStreaming halts IQ sample streaming
func (t *IQTransceiver) StopStreaming() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.streaming {
		return fmt.Errorf("not streaming")
	}

	close(t.stopChan)
	t.wg.Wait()
	t.streaming = false

	fmt.Println("IQ Transceiver: Stopped streaming")
	return nil
}

// Transmit queues an IQ buffer for transmission
func (t *IQTransceiver) Transmit(buffer *iq.IQBuffer) error {
	select {
	case t.txBuffer <- buffer:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("TX buffer full")
	}
}

// Receive returns a received IQ buffer
func (t *IQTransceiver) Receive() (*iq.IQBuffer, error) {
	select {
	case buffer := <-t.rxBuffer:
		return buffer, nil
	case <-time.After(1 * time.Second):
		return nil, fmt.Errorf("RX timeout")
	}
}

// txWorker handles TX streaming via DAC/ADC or SPI
func (t *IQTransceiver) txWorker() {
	defer t.wg.Done()

	for {
		select {
		case <-t.stopChan:
			return
		case buffer := <-t.txBuffer:
			// In real implementation:
			// 1. Convert IQ samples to transceiver format
			// 2. Write to DAC (analog) or SPI (digital)
			// 3. STM32U585 DMA streams samples to transceiver
			// 4. Transceiver up-converts to RF and transmits

			// For simulation, just log
			if err := t.transmitIQ(buffer); err != nil {
				fmt.Printf("IQ Transceiver TX error: %v\n", err)
			}
		}
	}
}

// rxWorker handles RX streaming via DAC/ADC or SPI
func (t *IQTransceiver) rxWorker() {
	defer t.wg.Done()

	ticker := time.NewTicker(time.Duration(float64(t.config.BufferSize) / t.sampleRate * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case <-ticker.C:
			// In real implementation:
			// 1. Transceiver down-converts RF to baseband IQ
			// 2. Read from ADC (analog) or SPI (digital)
			// 3. STM32U585 DMA receives samples from transceiver
			// 4. Convert to IQ buffer format

			// For simulation, create buffer
			buffer, err := t.receiveIQ()
			if err != nil {
				fmt.Printf("IQ Transceiver RX error: %v\n", err)
				continue
			}

			select {
			case t.rxBuffer <- buffer:
			default:
				// Buffer full, drop samples
				fmt.Println("IQ Transceiver: RX buffer full, dropping samples")
			}
		}
	}
}

// transmitIQ transmits IQ samples via the transceiver
func (t *IQTransceiver) transmitIQ(buffer *iq.IQBuffer) error {
	// In real implementation, write samples to DAC/ADC or SPI
	// For simulation, just log
	fmt.Printf("IQ Transceiver: TX %d samples\n", len(buffer.Samples))
	return nil
}

// receiveIQ receives IQ samples from the transceiver
func (t *IQTransceiver) receiveIQ() (*iq.IQBuffer, error) {
	// In real implementation, read samples from ADC or SPI
	// For simulation, create empty buffer
	buffer := iq.NewIQBuffer(t.config.BufferSize, t.sampleRate)
	buffer.Timestamp = time.Now().UnixNano()
	return buffer, nil
}

// SetCenterFreq sets the center frequency
func (t *IQTransceiver) SetCenterFreq(freq float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range for UHF transceiver: 400-470 MHz
	if freq < 400e6 || freq > 470e6 {
		return fmt.Errorf("frequency out of range (400-470 MHz)")
	}

	t.centerFreq = freq
	fmt.Printf("IQ Transceiver: Set center freq = %.3f MHz\n", freq/1e6)
	return nil
}

// SetSampleRate sets the sample rate
func (t *IQTransceiver) SetSampleRate(rate float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 100 kHz - 10 MHz
	if rate < 100e3 || rate > 10e6 {
		return fmt.Errorf("sample rate out of range (100 kHz - 10 MHz)")
	}

	t.sampleRate = rate
	fmt.Printf("IQ Transceiver: Set sample rate = %.3f MHz\n", rate/1e6)
	return nil
}

// SetTXGain sets the TX gain
func (t *IQTransceiver) SetTXGain(gain float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 0-30 dB
	if gain < 0 || gain > 30 {
		return fmt.Errorf("TX gain out of range (0-30 dB)")
	}

	t.txGain = gain
	fmt.Printf("IQ Transceiver: Set TX gain = %.1f dB\n", gain)
	return nil
}

// SetRXGain sets the RX gain
func (t *IQTransceiver) SetRXGain(gain float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Typical range: 0-60 dB
	if gain < 0 || gain > 60 {
		return fmt.Errorf("RX gain out of range (0-60 dB)")
	}

	t.rxGain = gain
	fmt.Printf("IQ Transceiver: Set RX gain = %.1f dB\n", gain)
	return nil
}

// GetCenterFreq returns the current center frequency
func (t *IQTransceiver) GetCenterFreq() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.centerFreq
}

// GetSampleRate returns the current sample rate
func (t *IQTransceiver) GetSampleRate() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sampleRate
}

// IsStreaming returns whether the transceiver is currently streaming
func (t *IQTransceiver) IsStreaming() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.streaming
}

// GetInterfaceType returns the interface type
func (t *IQTransceiver) GetInterfaceType() InterfaceType {
	return t.config.InterfaceType
}
