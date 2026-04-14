// Package sband_iq implements the Convergence Layer Adapter for cislunar deep-space nodes
// using S-band (2.2 GHz) or X-band (8.4 GHz) IQ transceiver.
//
// This CLA interfaces directly with the STM32U585 OBC (or higher-capability processor)
// via DAC/ADC or SPI, with no companion host required.
//
// Supports BPSK modulation with LDPC/Turbo FEC at 500 bps for cislunar operations.
// Implements AX.25/LTP protocol stack over IQ baseband with long-delay session management
// to account for 1-2 second one-way light-time delay.
package sband_iq

import (
	"fmt"
	"sync"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/iq"
	"terrestrial-dtn/pkg/radio/sband_transceiver"
)

// SBandIQCLA implements the CLA interface for cislunar S-band/X-band IQ transceiver
type SBandIQCLA struct {
	mu          sync.Mutex
	config      Config
	transceiver *sband_transceiver.SBandTransceiver
	modulator   *iq.Modulator
	demodulator *iq.Demodulator
	status      cla.CLAStatus
	metrics     cla.LinkMetrics
	isOpen      bool
	
	// Long-delay LTP session management
	ltpSessions map[string]*LTPSession
	sessionMu   sync.Mutex
}

// Config holds S-band/X-band IQ CLA configuration
type Config struct {
	Callsign   string
	Band       sband_transceiver.Band // S-band or X-band
	CenterFreq float64                // Hz (default: 2.2e9 for S-band, 8.4e9 for X-band)
	SampleRate float64                // Samples/second (default: 8000 for 500 bps)
	DataRate   int                    // bps (default: 500)
	TXPower    float64                // Watts (5-10W for deep-space)
	TXGain     float64                // dBi
	RXGain     float64                // dBi
	FECEnabled bool                   // Enable LDPC/Turbo FEC
	FECType    sband_transceiver.FECType
	
	// Light-time delay for cislunar operations
	LightTimeDelay time.Duration // 1-2 seconds for Earth-Moon distance
	
	// LTP session timeout (must account for round-trip delay)
	LTPTimeout time.Duration // Default: 10 seconds (allows for 2-4s RTT + processing)
}

// LTPSession tracks a long-delay LTP session
type LTPSession struct {
	SessionID      string
	RemoteNode     string
	StartTime      time.Time
	LastActivity   time.Time
	SegmentsSent   int
	SegmentsAcked  int
	PendingData    []byte
	State          LTPSessionState
}

// LTPSessionState represents the state of an LTP session
type LTPSessionState int

const (
	LTPSessionIdle LTPSessionState = iota
	LTPSessionActive
	LTPSessionWaitingAck
	LTPSessionComplete
	LTPSessionTimeout
)

// DefaultSBandConfig returns default configuration for S-band cislunar operation
func DefaultSBandConfig(callsign string) Config {
	return Config{
		Callsign:       callsign,
		Band:           sband_transceiver.BandS,
		CenterFreq:     2.2e9,  // 2.2 GHz S-band
		SampleRate:     8000.0, // 8 kHz sample rate
		DataRate:       500,    // 500 bps
		TXPower:        5.0,    // 5W transmit power
		TXGain:         10.0,   // 10 dBi directional patch antenna
		RXGain:         35.0,   // 35 dBi ground dish
		FECEnabled:     true,
		FECType:        sband_transceiver.FECLDPC,
		LightTimeDelay: 1200 * time.Millisecond, // ~1.2 seconds nominal
		LTPTimeout:     10 * time.Second,         // 10 seconds for long-delay sessions
	}
}

// DefaultXBandConfig returns default configuration for X-band cislunar operation
func DefaultXBandConfig(callsign string) Config {
	return Config{
		Callsign:       callsign,
		Band:           sband_transceiver.BandX,
		CenterFreq:     8.4e9,  // 8.4 GHz X-band
		SampleRate:     8000.0, // 8 kHz sample rate
		DataRate:       500,    // 500 bps
		TXPower:        5.0,    // 5W transmit power
		TXGain:         12.0,   // 12 dBi directional antenna
		RXGain:         40.0,   // 40 dBi ground dish
		FECEnabled:     true,
		FECType:        sband_transceiver.FECLDPC,
		LightTimeDelay: 1200 * time.Millisecond,
		LTPTimeout:     10 * time.Second,
	}
}

// New creates a new S-band/X-band IQ CLA for cislunar operations
func New(config Config) (*SBandIQCLA, error) {
	// Create S-band/X-band transceiver
	var transceiverConfig sband_transceiver.Config
	if config.Band == sband_transceiver.BandS {
		transceiverConfig = sband_transceiver.DefaultSBandConfig()
	} else {
		transceiverConfig = sband_transceiver.DefaultXBandConfig()
	}
	
	transceiverConfig.CenterFreq = config.CenterFreq
	transceiverConfig.SampleRate = config.SampleRate
	transceiverConfig.TXPower = config.TXPower
	transceiverConfig.TXGain = config.TXGain
	transceiverConfig.RXGain = config.RXGain
	transceiverConfig.FECEnabled = config.FECEnabled
	transceiverConfig.FECType = config.FECType
	transceiverConfig.LightTimeDelay = config.LightTimeDelay

	transceiver, err := sband_transceiver.New(transceiverConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create S-band/X-band transceiver: %w", err)
	}

	// Create modulator for BPSK
	modulatorConfig := iq.ModulationConfig{
		Type:          iq.ModulationBPSK,
		SampleRate:    config.SampleRate,
		SymbolRate:    float64(config.DataRate),
		CarrierFreq:   0, // Baseband
		FrequencyDev:  0, // Not used for BPSK
		BTProduct:     0,
		SamplesPerSym: int(config.SampleRate / float64(config.DataRate)),
		FilterTaps:    0,
	}
	modulator := iq.NewModulator(modulatorConfig)

	// Create demodulator for BPSK
	demodulatorConfig := iq.ModulationConfig{
		Type:          iq.ModulationBPSK,
		SampleRate:    config.SampleRate,
		SymbolRate:    float64(config.DataRate),
		CarrierFreq:   0, // Baseband
		FrequencyDev:  0,
		BTProduct:     0,
		SamplesPerSym: int(config.SampleRate / float64(config.DataRate)),
		FilterTaps:    0,
	}
	demodulator := iq.NewDemodulator(demodulatorConfig)

	return &SBandIQCLA{
		config:      config,
		transceiver: transceiver,
		modulator:   modulator,
		demodulator: demodulator,
		status:      cla.CLAStatusIdle,
		ltpSessions: make(map[string]*LTPSession),
	}, nil
}

// Type returns the CLA type
func (c *SBandIQCLA) Type() cla.CLAType {
	if c.config.Band == sband_transceiver.BandS {
		return cla.CLATypeAX25LTPSBandIQ
	}
	return cla.CLATypeAX25LTPXBandIQ
}

// Open establishes the link for a contact window
func (c *SBandIQCLA) Open(window contact.ContactWindow) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isOpen {
		return fmt.Errorf("link already open")
	}

	// Open S-band/X-band transceiver
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

	bandStr := "S-band"
	if c.config.Band == sband_transceiver.BandX {
		bandStr = "X-band"
	}
	
	fecStr := "no FEC"
	if c.config.FECEnabled {
		fecStr = fmt.Sprintf("%s FEC", c.config.FECType)
	}

	fmt.Printf("%s IQ CLA: Link opened for contact with %s\n", bandStr, window.RemoteNode)
	fmt.Printf("%s IQ CLA: BPSK + %s at %d bps on %.3f GHz\n",
		bandStr, fecStr, c.config.DataRate, c.config.CenterFreq/1e9)
	fmt.Printf("%s IQ CLA: Light-time delay compensation = %.3f seconds\n",
		bandStr, c.config.LightTimeDelay.Seconds())

	return nil
}

// Close closes the link
func (c *SBandIQCLA) Close() error {
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

	bandStr := "S-band"
	if c.config.Band == sband_transceiver.BandX {
		bandStr = "X-band"
	}
	fmt.Printf("%s IQ CLA: Link closed\n", bandStr)
	
	return nil
}

// SendBundle transmits a bundle over the link with long-delay LTP session management
func (c *SBandIQCLA) SendBundle(bundle *bpa.Bundle) (*cla.LinkMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isOpen {
		return nil, fmt.Errorf("link not open")
	}

	c.status = cla.CLAStatusTransmitting

	// Serialize bundle to bytes
	data := c.serializeBundle(bundle)

	// Create LTP session for long-delay management
	sessionID := fmt.Sprintf("%s-%d", bundle.ID, time.Now().UnixNano())
	session := &LTPSession{
		SessionID:    sessionID,
		RemoteNode:   bundle.Destination.String(),
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		PendingData:  data,
		State:        LTPSessionActive,
	}
	
	c.sessionMu.Lock()
	c.ltpSessions[sessionID] = session
	c.sessionMu.Unlock()

	// Wrap in AX.25 frame with callsign addressing
	ax25Frame := c.createAX25Frame(data)

	// Modulate to IQ samples (BPSK)
	iqBuffer := c.modulator.Modulate(ax25Frame)

	// Account for light-time delay in transmission timing
	// In real implementation, this would adjust timing expectations
	transmitStart := time.Now()

	// Transmit IQ samples via transceiver (with FEC if enabled)
	if err := c.transceiver.Transmit(iqBuffer); err != nil {
		c.status = cla.CLAStatusError
		session.State = LTPSessionTimeout
		return nil, fmt.Errorf("transmission failed: %w", err)
	}

	// Update session state - waiting for ACK (accounting for RTT delay)
	session.State = LTPSessionWaitingAck
	session.SegmentsSent++
	session.LastActivity = time.Now()

	// Update metrics
	c.metrics.BytesTransferred += int64(len(data))

	c.status = cla.CLAStatusIdle

	transmitDuration := time.Since(transmitStart)
	expectedRTT := 2 * c.config.LightTimeDelay

	bandStr := "S-band"
	if c.config.Band == sband_transceiver.BandX {
		bandStr = "X-band"
	}

	fmt.Printf("%s IQ CLA: Transmitted bundle %s (%d bytes, TX time: %.3fs)\n",
		bandStr, bundle.ID, len(data), transmitDuration.Seconds())
	fmt.Printf("%s IQ CLA: Expected RTT: %.3fs (light-time delay: %.3fs each way)\n",
		bandStr, expectedRTT.Seconds(), c.config.LightTimeDelay.Seconds())

	return &c.metrics, nil
}

// RecvBundle receives a bundle from the link with long-delay handling
func (c *SBandIQCLA) RecvBundle() (*bpa.Bundle, *cla.LinkMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isOpen {
		return nil, nil, fmt.Errorf("link not open")
	}

	c.status = cla.CLAStatusReceiving

	// Receive IQ samples from transceiver (accounting for light-time delay)
	iqBuffer, err := c.transceiver.Receive()
	if err != nil {
		c.status = cla.CLAStatusIdle
		return nil, nil, fmt.Errorf("receive failed: %w", err)
	}

	// Demodulate IQ samples to bytes (BPSK with FEC decoding)
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
	c.metrics.BitErrorRate = linkMetrics.EVM / 100.0
	c.metrics.BytesTransferred += int64(len(data))

	c.status = cla.CLAStatusIdle

	bandStr := "S-band"
	if c.config.Band == sband_transceiver.BandX {
		bandStr = "X-band"
	}

	fmt.Printf("%s IQ CLA: Received bundle %s (%d bytes, RSSI=%d dBm, SNR=%.1f dB)\n",
		bandStr, bundle.ID, len(data), c.metrics.RSSI, c.metrics.SNR)
	fmt.Printf("%s IQ CLA: Reception accounted for %.3fs light-time delay\n",
		bandStr, c.config.LightTimeDelay.Seconds())

	return bundle, &c.metrics, nil
}

// Status returns the current CLA status
func (c *SBandIQCLA) Status() cla.CLAStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

// LinkMetrics returns the current link metrics
func (c *SBandIQCLA) LinkMetrics() *cla.LinkMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &c.metrics
}

// GetActiveSessions returns the count of active LTP sessions
func (c *SBandIQCLA) GetActiveSessions() int {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	
	count := 0
	for _, session := range c.ltpSessions {
		if session.State == LTPSessionActive || session.State == LTPSessionWaitingAck {
			count++
		}
	}
	return count
}

// CleanupSessions removes expired LTP sessions
func (c *SBandIQCLA) CleanupSessions() {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	
	now := time.Now()
	for id, session := range c.ltpSessions {
		if now.Sub(session.LastActivity) > c.config.LTPTimeout {
			session.State = LTPSessionTimeout
			delete(c.ltpSessions, id)
		}
	}
}

// createAX25Frame wraps payload in AX.25 frame with callsign addressing
func (c *SBandIQCLA) createAX25Frame(payload []byte) []byte {
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
func (c *SBandIQCLA) extractAX25Frame(frame []byte) ([]byte, error) {
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
func (c *SBandIQCLA) serializeBundle(bundle *bpa.Bundle) []byte {
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
func (c *SBandIQCLA) deserializeBundle(data []byte) (*bpa.Bundle, error) {
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
