// Package uhf_iq implements the Convergence Layer Adapter for LEO CubeSat flight nodes
// using a flight-qualified IQ transceiver IC.
//
// This CLA interfaces directly with the STM32U585 OBC via DAC/ADC or SPI, with no
// companion host required (unlike the B200mini used in EM phase).
//
// Supports GMSK/BPSK modulation at 9.6 kbps on UHF 437 MHz for LEO operations.
// Implements AX.25/LTP protocol stack over IQ baseband.
package uhf_iq

import (
	"fmt"
	"sync"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/iq"
	"terrestrial-dtn/pkg/radio/iq_transceiver"
)

// UHFIQCLA implements the CLA interface for LEO flight IQ transceiver
type UHFIQCLA struct {
	mu          sync.Mutex
	config      Config
	transceiver *iq_transceiver.IQTransceiver
	modulator   *iq.Modulator
	demodulator *iq.Demodulator
	status      cla.CLAStatus
	metrics     cla.LinkMetrics
	isOpen      bool
}

// Config holds UHF IQ CLA configuration
type Config struct {
	Callsign   string
	CenterFreq float64 // Hz (default: 437e6 for UHF 437 MHz)
	SampleRate float64 // Samples/second (default: 1e6)
	DataRate   int     // bps (default: 9600)
	TXGain     float64 // dB
	RXGain     float64 // dB
	Modulation iq.ModulationType
}

// DefaultConfig returns default configuration for LEO UHF operation
func DefaultConfig(callsign string) Config {
	return Config{
		Callsign:   callsign,
		CenterFreq: 437e6,  // 437 MHz UHF
		SampleRate: 1e6,    // 1 MHz sample rate
		DataRate:   9600,   // 9.6 kbps
		TXGain:     20.0,   // 20 dB TX gain
		RXGain:     30.0,   // 30 dB RX gain
		Modulation: iq.ModulationGMSK, // GMSK for LEO
	}
}

// New creates a new UHF IQ CLA for LEO flight
func New(config Config) (*UHFIQCLA, error) {
	// Create IQ transceiver
	transceiverConfig := iq_transceiver.DefaultConfig()
	transceiverConfig.CenterFreq = config.CenterFreq
	transceiverConfig.SampleRate = config.SampleRate
	transceiverConfig.TXGain = config.TXGain
	transceiverConfig.RXGain = config.RXGain

	transceiver, err := iq_transceiver.New(transceiverConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create IQ transceiver: %w", err)
	}

	// Create modulator
	modulatorConfig := iq.ModulationConfig{
		Type:          config.Modulation,
		SampleRate:    config.SampleRate,
		SymbolRate:    float64(config.DataRate),
		CarrierFreq:   0, // Baseband
		FrequencyDev:  2400.0, // For GMSK
		BTProduct:     0.3,
		SamplesPerSym: int(config.SampleRate / float64(config.DataRate)),
		FilterTaps:    8,
	}
	modulator := iq.NewModulator(modulatorConfig)

	// Create demodulator
	demodulatorConfig := iq.ModulationConfig{
		Type:          config.Modulation,
		SampleRate:    config.SampleRate,
		SymbolRate:    float64(config.DataRate),
		CarrierFreq:   0, // Baseband
		FrequencyDev:  2400.0, // For GMSK
		BTProduct:     0.3,
		SamplesPerSym: int(config.SampleRate / float64(config.DataRate)),
		FilterTaps:    8,
	}
	demodulator := iq.NewDemodulator(demodulatorConfig)

	return &UHFIQCLA{
		config:      config,
		transceiver: transceiver,
		modulator:   modulator,
		demodulator: demodulator,
		status:      cla.CLAStatusIdle,
	}, nil
}

// Type returns the CLA type
func (c *UHFIQCLA) Type() cla.CLAType {
	return cla.CLATypeAX25LTPUHFIQ
}

// Open establishes the link for a contact window
func (c *UHFIQCLA) Open(window contact.ContactWindow) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isOpen {
		return fmt.Errorf("link already open")
	}

	// Open IQ transceiver
	if err := c.transceiver.Open(); err != nil {
		return fmt.Errorf("failed to open transceiver: %w", err)
	}

	// Start IQ streaming
	if err := c.transceiver.StartStreaming(); err != nil {
		c.transceiver.Close()
		return fmt.Errorf("failed to start streaming: %w", err)
	}

	c.isOpen = true
	c.status = cla.CLAStatusIdle

	fmt.Printf("UHF IQ CLA: Link opened for contact with %s\n", window.RemoteNode)
	fmt.Printf("UHF IQ CLA: %s modulation at %d bps on %.3f MHz\n",
		c.config.Modulation, c.config.DataRate, c.config.CenterFreq/1e6)

	return nil
}

// Close closes the link
func (c *UHFIQCLA) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isOpen {
		return nil
	}

	// Stop IQ streaming
	if err := c.transceiver.StopStreaming(); err != nil {
		return fmt.Errorf("failed to stop streaming: %w", err)
	}

	// Close transceiver
	if err := c.transceiver.Close(); err != nil {
		return fmt.Errorf("failed to close transceiver: %w", err)
	}

	c.isOpen = false
	c.status = cla.CLAStatusIdle

	fmt.Println("UHF IQ CLA: Link closed")
	return nil
}

// SendBundle transmits a bundle over the link
func (c *UHFIQCLA) SendBundle(bundle *bpa.Bundle) (*cla.LinkMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isOpen {
		return nil, fmt.Errorf("link not open")
	}

	c.status = cla.CLAStatusTransmitting

	// Serialize bundle to bytes (using a simple encoding)
	data := c.serializeBundle(bundle)

	// Wrap in AX.25 frame with callsign addressing
	ax25Frame := c.createAX25Frame(data)

	// Modulate to IQ samples
	iqBuffer := c.modulator.Modulate(ax25Frame)

	// Transmit IQ samples via transceiver
	if err := c.transceiver.Transmit(iqBuffer); err != nil {
		c.status = cla.CLAStatusError
		return nil, fmt.Errorf("transmission failed: %w", err)
	}

	// Update metrics
	c.metrics.BytesTransferred += int64(len(data))

	c.status = cla.CLAStatusIdle

	fmt.Printf("UHF IQ CLA: Transmitted bundle %s (%d bytes)\n", bundle.ID, len(data))

	return &c.metrics, nil
}

// RecvBundle receives a bundle from the link
func (c *UHFIQCLA) RecvBundle() (*bpa.Bundle, *cla.LinkMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isOpen {
		return nil, nil, fmt.Errorf("link not open")
	}

	c.status = cla.CLAStatusReceiving

	// Receive IQ samples from transceiver
	iqBuffer, err := c.transceiver.Receive()
	if err != nil {
		c.status = cla.CLAStatusIdle
		return nil, nil, fmt.Errorf("receive failed: %w", err)
	}

	// Demodulate IQ samples to bytes
	data, linkMetrics := c.demodulator.Demodulate(iqBuffer)

	// Extract AX.25 frame and validate callsign
	payload, err := c.extractAX25Frame(data)
	if err != nil {
		c.status = cla.CLAStatusError
		return nil, nil, fmt.Errorf("AX.25 frame extraction failed: %w", err)
	}

	// Deserialize bundle
	bundle, err := c.deserializeBundle(payload)
	if err != nil {
		c.status = cla.CLAStatusError
		return nil, nil, fmt.Errorf("bundle deserialization failed: %w", err)
	}

	// Update metrics from demodulator
	c.metrics.RSSI = int(linkMetrics.RSSI)
	c.metrics.SNR = linkMetrics.SNR
	c.metrics.BitErrorRate = linkMetrics.EVM / 100.0 // Convert EVM to BER approximation
	c.metrics.BytesTransferred += int64(len(data))

	c.status = cla.CLAStatusIdle

	fmt.Printf("UHF IQ CLA: Received bundle %s (%d bytes, RSSI=%d dBm, SNR=%.1f dB)\n",
		bundle.ID, len(data), c.metrics.RSSI, c.metrics.SNR)

	return bundle, &c.metrics, nil
}

// Status returns the current CLA status
func (c *UHFIQCLA) Status() cla.CLAStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

// LinkMetrics returns the current link metrics
func (c *UHFIQCLA) LinkMetrics() *cla.LinkMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &c.metrics
}

// createAX25Frame wraps payload in AX.25 frame with callsign addressing
func (c *UHFIQCLA) createAX25Frame(payload []byte) []byte {
	// AX.25 frame format (simplified):
	// - Destination callsign (7 bytes)
	// - Source callsign (7 bytes)
	// - Control field (1 byte)
	// - PID (1 byte)
	// - Payload
	// - FCS (2 bytes)
	//
	// In real implementation, would use proper AX.25 encoding
	// For simulation, just prepend callsign header

	header := []byte(fmt.Sprintf("AX25:%s>", c.config.Callsign))
	frame := append(header, payload...)
	
	return frame
}

// extractAX25Frame extracts payload from AX.25 frame and validates callsign
func (c *UHFIQCLA) extractAX25Frame(frame []byte) ([]byte, error) {
	// In real implementation, would parse AX.25 frame structure
	// and validate source/destination callsigns
	// For simulation, just strip header

	headerLen := len(fmt.Sprintf("AX25:%s>", c.config.Callsign))
	if len(frame) < headerLen {
		return nil, fmt.Errorf("frame too short")
	}

	return frame[headerLen:], nil
}

// serializeBundle serializes a bundle to bytes (simplified encoding)
func (c *UHFIQCLA) serializeBundle(bundle *bpa.Bundle) []byte {
	// In real implementation, would use proper BPv7 CBOR encoding
	// For simulation, use a simple format:
	// [bundle_type][priority][lifetime][destination][payload]
	
	data := []byte{byte(bundle.BundleType), byte(bundle.Priority)}
	
	// Add lifetime (8 bytes)
	lifetimeBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		lifetimeBytes[i] = byte(bundle.Lifetime >> (8 * i))
	}
	data = append(data, lifetimeBytes...)
	
	// Add destination (simplified)
	destBytes := []byte(bundle.Destination.String())
	data = append(data, byte(len(destBytes)))
	data = append(data, destBytes...)
	
	// Add payload
	data = append(data, bundle.Payload...)
	
	return data
}

// deserializeBundle deserializes bytes to a bundle (simplified decoding)
func (c *UHFIQCLA) deserializeBundle(data []byte) (*bpa.Bundle, error) {
	if len(data) < 11 { // Minimum: type(1) + priority(1) + lifetime(8) + dest_len(1)
		return nil, fmt.Errorf("data too short")
	}
	
	// Parse bundle type and priority
	bundleType := bpa.BundleType(data[0])
	priority := bpa.Priority(data[1])
	
	// Parse lifetime
	var lifetime int64
	for i := 0; i < 8; i++ {
		lifetime |= int64(data[2+i]) << (8 * i)
	}
	
	// Parse destination
	destLen := int(data[10])
	if len(data) < 11+destLen {
		return nil, fmt.Errorf("data too short for destination")
	}
	destStr := string(data[11 : 11+destLen])
	
	// Parse destination endpoint (simplified)
	dest := bpa.EndpointID{Scheme: "ipn", SSP: destStr}
	
	// Parse payload
	payload := data[11+destLen:]
	
	// Create bundle
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "0.0"},
			CreationTimestamp: 0,
			SequenceNumber:    0,
		},
		Destination: dest,
		Payload:     payload,
		Priority:    priority,
		Lifetime:    lifetime,
		CreatedAt:   0,
		BundleType:  bundleType,
	}
	
	return bundle, nil
}
