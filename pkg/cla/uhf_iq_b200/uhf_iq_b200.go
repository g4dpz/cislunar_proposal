// Package uhf_iq_b200 implements the Convergence Layer Adapter for the Engineering Model
// using UHF IQ baseband via STM32U585 + Ettus B200mini SDR.
//
// This CLA integrates:
// - IQ baseband modulation/demodulation (GFSK/GMSK at 9.6 kbps)
// - B200mini SDR as RF front-end (UHF 437 MHz)
// - AX.25/LTP framing (provided by ION-DTN)
//
// This is EM-only; flight unit uses dedicated IQ transceiver IC.
package uhf_iq_b200

import (
	"fmt"
	"sync"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/iq"
	"terrestrial-dtn/pkg/sdr/b200mini"
)

// UHFIQB200CLA implements the CLA interface for EM UHF IQ operation
type UHFIQB200CLA struct {
	mu         sync.Mutex
	callsign   string
	b200       *b200mini.B200mini
	modulator  *iq.Modulator
	demodulator *iq.Demodulator
	status     cla.CLAStatus
	metrics    cla.LinkMetrics
	active     bool
}

// Config holds configuration for the UHF IQ B200mini CLA
type Config struct {
	Callsign     string
	B200Config   b200mini.Config
	ModConfig    iq.ModulationConfig
}

// DefaultConfig returns default configuration for EM UHF operation
func DefaultConfig(callsign string) Config {
	return Config{
		Callsign:   callsign,
		B200Config: b200mini.DefaultB200miniConfig(),
		ModConfig:  iq.DefaultGMSKConfig(), // GMSK for EM
	}
}

// New creates a new UHF IQ B200mini CLA
func New(config Config) (*UHFIQB200CLA, error) {
	// Create B200mini interface
	b200, err := b200mini.New(config.B200Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create B200mini: %w", err)
	}

	// Create modulator and demodulator
	modulator := iq.NewModulator(config.ModConfig)
	demodulator := iq.NewDemodulator(config.ModConfig)

	return &UHFIQB200CLA{
		callsign:    config.Callsign,
		b200:        b200,
		modulator:   modulator,
		demodulator: demodulator,
		status:      cla.CLAStatusIdle,
	}, nil
}

// CLAType returns the CLA type
func (c *UHFIQB200CLA) CLAType() cla.CLAType {
	return cla.CLATypeAX25LTP_UHF_IQ_B200
}

// Open establishes the link for a contact window
func (c *UHFIQB200CLA) Open(window contact.ContactWindow) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.active {
		return fmt.Errorf("CLA already active")
	}

	// Open B200mini
	if err := c.b200.Open(); err != nil {
		return fmt.Errorf("failed to open B200mini: %w", err)
	}

	// Start streaming
	if err := c.b200.StartStreaming(); err != nil {
		c.b200.Close()
		return fmt.Errorf("failed to start streaming: %w", err)
	}

	c.active = true
	c.status = cla.CLAStatusIdle
	fmt.Printf("UHF IQ B200 CLA: Opened for contact with %s (UHF 437 MHz, 9.6 kbps)\n", window.RemoteNode)
	return nil
}

// Close terminates the link
func (c *UHFIQB200CLA) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return fmt.Errorf("CLA not active")
	}

	// Stop streaming
	if err := c.b200.StopStreaming(); err != nil {
		return fmt.Errorf("failed to stop streaming: %w", err)
	}

	// Close B200mini
	if err := c.b200.Close(); err != nil {
		return fmt.Errorf("failed to close B200mini: %w", err)
	}

	c.active = false
	c.status = cla.CLAStatusIdle
	fmt.Println("UHF IQ B200 CLA: Closed")
	return nil
}

// SendBundle transmits a bundle over the UHF IQ link
func (c *UHFIQB200CLA) SendBundle(bundle bpa.Bundle) (cla.LinkMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return cla.LinkMetrics{}, fmt.Errorf("CLA not active")
	}

	c.status = cla.CLAStatusTransmitting

	// In real implementation, this would:
	// 1. Serialize bundle to bytes (ION-DTN handles BPv7 encoding)
	// 2. Encapsulate in AX.25 frame with callsigns (ION-DTN handles this)
	// 3. Wrap in LTP segments (ION-DTN handles this)
	// 4. Modulate to IQ samples
	// 5. Transmit via B200mini

	// For simulation, we modulate the payload
	iqBuffer := c.modulator.Modulate(bundle.Payload)

	// Transmit via B200mini
	if err := c.b200.Transmit(iqBuffer); err != nil {
		c.status = cla.CLAStatusError
		return cla.LinkMetrics{}, fmt.Errorf("transmission failed: %w", err)
	}

	// Update metrics (simulated)
	c.metrics = cla.LinkMetrics{
		RSSI:             -70.0, // Simulated
		SNR:              15.0,
		BitErrorRate:     1e-5,
		BytesTransferred: len(bundle.Payload),
	}

	c.status = cla.CLAStatusIdle
	fmt.Printf("UHF IQ B200 CLA: Sent bundle %s (%d bytes)\n", bundle.ID.String(), len(bundle.Payload))
	return c.metrics, nil
}

// RecvBundle receives a bundle from the UHF IQ link
func (c *UHFIQB200CLA) RecvBundle() (bpa.Bundle, cla.LinkMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return bpa.Bundle{}, cla.LinkMetrics{}, fmt.Errorf("CLA not active")
	}

	c.status = cla.CLAStatusReceiving

	// Receive IQ samples from B200mini
	iqBuffer, err := c.b200.Receive()
	if err != nil {
		c.status = cla.CLAStatusError
		return bpa.Bundle{}, cla.LinkMetrics{}, fmt.Errorf("reception failed: %w", err)
	}

	// Demodulate IQ samples
	data, iqMetrics := c.demodulator.Demodulate(iqBuffer)

	// In real implementation, this would:
	// 1. Demodulate IQ to bits
	// 2. Decode LTP segments (ION-DTN handles this)
	// 3. Decode AX.25 frames (ION-DTN handles this)
	// 4. Deserialize BPv7 bundle (ION-DTN handles this)

	// For simulation, create a bundle from demodulated data
	bundle := bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "dtn", SSP: "//remote/data"},
			CreationTimestamp: uint64(time.Now().Unix()),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "dtn", SSP: "//local/data"},
		Payload:     data,
		Priority:    bpa.PriorityNormal,
		Lifetime:    3600,
		CreatedAt:   uint64(time.Now().Unix()),
		BundleType:  bpa.BundleTypeData,
	}

	// Convert IQ metrics to CLA metrics
	c.metrics = cla.LinkMetrics{
		RSSI:             iqMetrics.RSSI,
		SNR:              iqMetrics.SNR,
		BitErrorRate:     iqMetrics.EVM / 100.0, // Convert EVM to BER approximation
		BytesTransferred: len(data),
	}

	c.status = cla.CLAStatusIdle
	fmt.Printf("UHF IQ B200 CLA: Received bundle (%d bytes)\n", len(data))
	return bundle, c.metrics, nil
}

// Status returns the current CLA status
func (c *UHFIQB200CLA) Status() cla.CLAStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

// LinkMetrics returns the current link metrics
func (c *UHFIQB200CLA) LinkMetrics() cla.LinkMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.metrics
}

// IsActive returns whether the CLA is active
func (c *UHFIQB200CLA) IsActive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active
}
