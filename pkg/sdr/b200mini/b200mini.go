// Package b200mini provides an interface to the Ettus Research USRP B200mini SDR
// for the Engineering Model phase. The B200mini connects via USB 3.0 to a companion
// Raspberry Pi or PC running UHD (USRP Hardware Driver), which bridges IQ samples
// to/from the STM32U585 OBC over SPI or UART/DMA.
//
// This is EM-only hardware; the flight unit replaces the B200mini with a dedicated
// flight-qualified IQ transceiver IC.
package b200mini

import (
	"fmt"
	"sync"
	"time"

	"terrestrial-dtn/pkg/iq"
)

// B200mini represents the Ettus B200mini SDR interface
type B200mini struct {
	mu            sync.Mutex
	config        Config
	bridge        *SPIBridge
	streaming     bool
	centerFreq    float64
	sampleRate    float64
	txGain        float64
	rxGain        float64
	txBuffer      chan *iq.IQBuffer
	rxBuffer      chan *iq.IQBuffer
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// Config holds B200mini configuration
type Config struct {
	// UHD device arguments (e.g., "type=b200")
	DeviceArgs string

	// Center frequency in Hz (e.g., 437e6 for UHF 437 MHz)
	CenterFreq float64

	// Sample rate in samples/second (e.g., 1e6 for 1 MHz)
	SampleRate float64

	// TX gain in dB (0-89.8 dB for B200mini)
	TXGain float64

	// RX gain in dB (0-76 dB for B200mini)
	RXGain float64

	// SPI/UART bridge configuration
	BridgeConfig SPIBridgeConfig
}

// DefaultB200miniConfig returns default configuration for UHF 437 MHz operation
func DefaultB200miniConfig() Config {
	return Config{
		DeviceArgs: "type=b200",
		CenterFreq: 437e6,  // 437 MHz UHF
		SampleRate: 1e6,    // 1 MHz sample rate
		TXGain:     50.0,   // 50 dB TX gain
		RXGain:     40.0,   // 40 dB RX gain
		BridgeConfig: SPIBridgeConfig{
			Device:     "/dev/spidev0.0",
			Speed:      1000000, // 1 MHz SPI clock
			BufferSize: 4096,
		},
	}
}

// New creates a new B200mini SDR interface
func New(config Config) (*B200mini, error) {
	// Create SPI/UART bridge to STM32U585
	bridge, err := NewSPIBridge(config.BridgeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SPI bridge: %w", err)
	}

	b := &B200mini{
		config:     config,
		bridge:     bridge,
		centerFreq: config.CenterFreq,
		sampleRate: config.SampleRate,
		txGain:     config.TXGain,
		rxGain:     config.RXGain,
		txBuffer:   make(chan *iq.IQBuffer, 8),
		rxBuffer:   make(chan *iq.IQBuffer, 8),
		stopChan:   make(chan struct{}),
	}

	return b, nil
}

// Open initializes the B200mini and establishes connection
func (b *B200mini) Open() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Open SPI bridge
	if err := b.bridge.Open(); err != nil {
		return fmt.Errorf("failed to open SPI bridge: %w", err)
	}

	// In a real implementation, this would:
	// 1. Initialize UHD library
	// 2. Create USRP device handle
	// 3. Set center frequency, sample rate, gains
	// 4. Configure TX/RX channels
	// For simulation, we just log the configuration

	fmt.Printf("B200mini: Opened device %s\n", b.config.DeviceArgs)
	fmt.Printf("B200mini: Center freq = %.3f MHz\n", b.centerFreq/1e6)
	fmt.Printf("B200mini: Sample rate = %.3f MHz\n", b.sampleRate/1e6)
	fmt.Printf("B200mini: TX gain = %.1f dB, RX gain = %.1f dB\n", b.txGain, b.rxGain)

	return nil
}

// Close shuts down the B200mini
func (b *B200mini) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.streaming {
		return fmt.Errorf("cannot close while streaming")
	}

	// Close SPI bridge
	if err := b.bridge.Close(); err != nil {
		return fmt.Errorf("failed to close SPI bridge: %w", err)
	}

	fmt.Println("B200mini: Closed")
	return nil
}

// StartStreaming begins IQ sample streaming
func (b *B200mini) StartStreaming() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.streaming {
		return fmt.Errorf("already streaming")
	}

	b.streaming = true
	b.stopChan = make(chan struct{})

	// Start TX worker
	b.wg.Add(1)
	go b.txWorker()

	// Start RX worker
	b.wg.Add(1)
	go b.rxWorker()

	fmt.Println("B200mini: Started streaming")
	return nil
}

// StopStreaming halts IQ sample streaming
func (b *B200mini) StopStreaming() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.streaming {
		return fmt.Errorf("not streaming")
	}

	close(b.stopChan)
	b.wg.Wait()
	b.streaming = false

	fmt.Println("B200mini: Stopped streaming")
	return nil
}

// Transmit queues an IQ buffer for transmission
func (b *B200mini) Transmit(buffer *iq.IQBuffer) error {
	select {
	case b.txBuffer <- buffer:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("TX buffer full")
	}
}

// Receive returns a received IQ buffer
func (b *B200mini) Receive() (*iq.IQBuffer, error) {
	select {
	case buffer := <-b.rxBuffer:
		return buffer, nil
	case <-time.After(1 * time.Second):
		return nil, fmt.Errorf("RX timeout")
	}
}

// txWorker handles TX streaming via USB 3.0 and SPI bridge
func (b *B200mini) txWorker() {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopChan:
			return
		case buffer := <-b.txBuffer:
			// In real implementation:
			// 1. Convert IQ samples to UHD format
			// 2. Send to B200mini via UHD TX stream
			// 3. B200mini up-converts to RF and transmits

			// For simulation, send IQ samples to STM32U585 via SPI bridge
			if err := b.bridge.SendIQ(buffer); err != nil {
				fmt.Printf("B200mini TX error: %v\n", err)
			}
		}
	}
}

// rxWorker handles RX streaming via USB 3.0 and SPI bridge
func (b *B200mini) rxWorker() {
	defer b.wg.Done()

	ticker := time.NewTicker(time.Duration(float64(b.bridge.config.BufferSize) / b.sampleRate * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-b.stopChan:
			return
		case <-ticker.C:
			// In real implementation:
			// 1. Receive RF from B200mini
			// 2. B200mini down-converts to baseband IQ
			// 3. Stream IQ samples via USB 3.0 (UHD)
			// 4. Bridge to STM32U585 via SPI/UART

			// For simulation, receive IQ samples from STM32U585 via SPI bridge
			buffer, err := b.bridge.ReceiveIQ()
			if err != nil {
				fmt.Printf("B200mini RX error: %v\n", err)
				continue
			}

			select {
			case b.rxBuffer <- buffer:
			default:
				// Buffer full, drop samples
				fmt.Println("B200mini: RX buffer full, dropping samples")
			}
		}
	}
}

// SetCenterFreq sets the center frequency
func (b *B200mini) SetCenterFreq(freq float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if freq < 70e6 || freq > 6e9 {
		return fmt.Errorf("frequency out of range (70 MHz - 6 GHz)")
	}

	b.centerFreq = freq
	fmt.Printf("B200mini: Set center freq = %.3f MHz\n", freq/1e6)
	return nil
}

// SetSampleRate sets the sample rate
func (b *B200mini) SetSampleRate(rate float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if rate < 200e3 || rate > 56e6 {
		return fmt.Errorf("sample rate out of range (200 kHz - 56 MHz)")
	}

	b.sampleRate = rate
	fmt.Printf("B200mini: Set sample rate = %.3f MHz\n", rate/1e6)
	return nil
}

// SetTXGain sets the TX gain
func (b *B200mini) SetTXGain(gain float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gain < 0 || gain > 89.8 {
		return fmt.Errorf("TX gain out of range (0 - 89.8 dB)")
	}

	b.txGain = gain
	fmt.Printf("B200mini: Set TX gain = %.1f dB\n", gain)
	return nil
}

// SetRXGain sets the RX gain
func (b *B200mini) SetRXGain(gain float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gain < 0 || gain > 76 {
		return fmt.Errorf("RX gain out of range (0 - 76 dB)")
	}

	b.rxGain = gain
	fmt.Printf("B200mini: Set RX gain = %.1f dB\n", gain)
	return nil
}

// GetCenterFreq returns the current center frequency
func (b *B200mini) GetCenterFreq() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.centerFreq
}

// GetSampleRate returns the current sample rate
func (b *B200mini) GetSampleRate() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sampleRate
}

// IsStreaming returns whether the B200mini is currently streaming
func (b *B200mini) IsStreaming() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.streaming
}
