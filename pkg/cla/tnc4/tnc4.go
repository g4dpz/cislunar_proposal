package tnc4

import (
	"fmt"
	"sync"
	"time"

	"go.bug.st/serial"
	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
)

// TNC4CLA implements the Convergence Layer Adapter for Mobilinkd TNC4
// Supports KISS framing over USB serial at 9600 baud for VHF/UHF
type TNC4CLA struct {
	devicePath    string
	port          serial.Port
	claType       cla.CLAType
	status        cla.CLAStatus
	metrics       cla.LinkMetrics
	isOpen        bool
	currentWindow *contact.ContactWindow
	mu            sync.RWMutex
}

// KISS protocol constants
const (
	KISS_FEND  = 0xC0 // Frame End
	KISS_FESC  = 0xDB // Frame Escape
	KISS_TFEND = 0xDC // Transposed Frame End
	KISS_TFESC = 0xDD // Transposed Frame Escape
	
	KISS_CMD_DATA = 0x00 // Data frame
)

// NewTNC4CLA creates a new TNC4 CLA instance
func NewTNC4CLA(devicePath string, claType cla.CLAType) *TNC4CLA {
	return &TNC4CLA{
		devicePath: devicePath,
		claType:    claType,
		status:     cla.CLAStatusIdle,
		metrics: cla.LinkMetrics{
			RSSI:             -90,
			SNR:              10.0,
			BitErrorRate:     0.01,
			BytesTransferred: 0,
		},
	}
}

// Type returns the CLA type
func (t *TNC4CLA) Type() cla.CLAType {
	return t.claType
}

// Open establishes the link for a contact window
func (t *TNC4CLA) Open(window contact.ContactWindow) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isOpen {
		return fmt.Errorf("link already open")
	}

	// Open serial port
	mode := &serial.Mode{
		BaudRate: 9600, // 9600 baud for G3RUH-compatible GFSK
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(t.devicePath, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", t.devicePath, err)
	}

	t.port = port
	t.isOpen = true
	t.currentWindow = &window
	t.status = cla.CLAStatusIdle

	// Initialize KISS mode
	if err := t.initKISS(); err != nil {
		t.port.Close()
		t.isOpen = false
		return fmt.Errorf("failed to initialize KISS mode: %w", err)
	}

	return nil
}

// Close closes the link
func (t *TNC4CLA) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isOpen {
		return fmt.Errorf("link not open")
	}

	if err := t.port.Close(); err != nil {
		return fmt.Errorf("failed to close serial port: %w", err)
	}

	t.isOpen = false
	t.currentWindow = nil
	t.status = cla.CLAStatusIdle
	return nil
}

// SendBundle transmits a bundle over the link
func (t *TNC4CLA) SendBundle(bundle *bpa.Bundle) (*cla.LinkMetrics, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isOpen {
		return nil, fmt.Errorf("link not open")
	}

	t.status = cla.CLAStatusTransmitting

	// Serialize bundle to bytes (simplified - real implementation would use ION-DTN)
	bundleData := t.serializeBundle(bundle)

	// Encapsulate in KISS frame
	kissFrame := t.encodeKISS(bundleData)

	// Write to serial port
	n, err := t.port.Write(kissFrame)
	if err != nil {
		t.status = cla.CLAStatusError
		return nil, fmt.Errorf("failed to write to serial port: %w", err)
	}

	// Update metrics
	t.metrics.BytesTransferred += int64(n)
	t.status = cla.CLAStatusIdle

	// Return a copy of the metrics
	metricsCopy := t.metrics
	return &metricsCopy, nil
}

// RecvBundle receives a bundle from the link
func (t *TNC4CLA) RecvBundle() (*bpa.Bundle, *cla.LinkMetrics, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isOpen {
		return nil, nil, fmt.Errorf("link not open")
	}

	t.status = cla.CLAStatusReceiving

	// Set read timeout
	t.port.SetReadTimeout(100 * time.Millisecond)

	// Read KISS frame
	kissFrame, err := t.readKISSFrame()
	if err != nil {
		t.status = cla.CLAStatusIdle
		return nil, nil, fmt.Errorf("failed to read KISS frame: %w", err)
	}

	// Decode KISS frame
	bundleData := t.decodeKISS(kissFrame)

	// Deserialize bundle (simplified - real implementation would use ION-DTN)
	bundle, err := t.deserializeBundle(bundleData)
	if err != nil {
		t.status = cla.CLAStatusIdle
		return nil, nil, fmt.Errorf("failed to deserialize bundle: %w", err)
	}

	// Update metrics
	t.metrics.BytesTransferred += int64(len(kissFrame))
	t.status = cla.CLAStatusIdle

	// Return a copy of the metrics
	metricsCopy := t.metrics
	return bundle, &metricsCopy, nil
}

// Status returns the current CLA status
func (t *TNC4CLA) Status() cla.CLAStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// LinkMetrics returns the current link metrics
func (t *TNC4CLA) LinkMetrics() *cla.LinkMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy
	metricsCopy := t.metrics
	return &metricsCopy
}

// initKISS initializes the TNC4 in KISS mode
func (t *TNC4CLA) initKISS() error {
	// Send KISS initialization sequence
	// Exit TNC2 mode and enter KISS mode
	initSeq := []byte{KISS_FEND, KISS_CMD_DATA, KISS_FEND}
	_, err := t.port.Write(initSeq)
	return err
}

// encodeKISS encapsulates data in a KISS frame
func (t *TNC4CLA) encodeKISS(data []byte) []byte {
	frame := make([]byte, 0, len(data)*2+4)
	
	// Start frame
	frame = append(frame, KISS_FEND)
	frame = append(frame, KISS_CMD_DATA)
	
	// Escape data
	for _, b := range data {
		switch b {
		case KISS_FEND:
			frame = append(frame, KISS_FESC, KISS_TFEND)
		case KISS_FESC:
			frame = append(frame, KISS_FESC, KISS_TFESC)
		default:
			frame = append(frame, b)
		}
	}
	
	// End frame
	frame = append(frame, KISS_FEND)
	
	return frame
}

// decodeKISS extracts data from a KISS frame
func (t *TNC4CLA) decodeKISS(frame []byte) []byte {
	if len(frame) < 3 {
		return nil
	}
	
	// Skip FEND and command byte
	data := frame[2 : len(frame)-1]
	
	// Unescape data
	result := make([]byte, 0, len(data))
	escaped := false
	
	for _, b := range data {
		if escaped {
			switch b {
			case KISS_TFEND:
				result = append(result, KISS_FEND)
			case KISS_TFESC:
				result = append(result, KISS_FESC)
			default:
				result = append(result, b)
			}
			escaped = false
		} else if b == KISS_FESC {
			escaped = true
		} else {
			result = append(result, b)
		}
	}
	
	return result
}

// readKISSFrame reads a complete KISS frame from the serial port
func (t *TNC4CLA) readKISSFrame() ([]byte, error) {
	frame := make([]byte, 0, 1024)
	buf := make([]byte, 1)
	
	// Wait for start FEND
	for {
		n, err := t.port.Read(buf)
		if err != nil {
			return nil, err
		}
		if n > 0 && buf[0] == KISS_FEND {
			frame = append(frame, buf[0])
			break
		}
	}
	
	// Read until end FEND
	for {
		n, err := t.port.Read(buf)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			frame = append(frame, buf[0])
			if buf[0] == KISS_FEND && len(frame) > 2 {
				break
			}
		}
	}
	
	return frame, nil
}

// serializeBundle converts a bundle to bytes (simplified)
func (t *TNC4CLA) serializeBundle(bundle *bpa.Bundle) []byte {
	// In a real implementation, this would use ION-DTN's BPv7 serialization
	// For now, use a simple format
	data := make([]byte, 0, len(bundle.Payload)+256)
	
	// Add bundle metadata (simplified)
	data = append(data, []byte(bundle.ID.SourceEID.String())...)
	data = append(data, 0x00) // Separator
	data = append(data, []byte(bundle.Destination.String())...)
	data = append(data, 0x00) // Separator
	data = append(data, bundle.Payload...)
	
	return data
}

// deserializeBundle converts bytes to a bundle (simplified)
func (t *TNC4CLA) deserializeBundle(data []byte) (*bpa.Bundle, error) {
	// In a real implementation, this would use ION-DTN's BPv7 deserialization
	// For now, use a simple format
	
	if len(data) < 10 {
		return nil, fmt.Errorf("data too short")
	}
	
	// Parse simplified format
	// This is a placeholder - real implementation would parse BPv7
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID: bpa.EndpointID{
				Scheme: "dtn",
				SSP:    "remote-node",
			},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{
			Scheme: "dtn",
			SSP:    "local-node",
		},
		Payload:    data,
		Priority:   bpa.PriorityNormal,
		Lifetime:   3600,
		CreatedAt:  time.Now().Unix(),
		BundleType: bpa.BundleTypeData,
	}
	
	return bundle, nil
}
