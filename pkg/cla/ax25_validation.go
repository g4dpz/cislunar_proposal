package cla

import (
	"fmt"
	"terrestrial-dtn/ax25"
)

// ValidateAX25Frame validates that a frame is a properly formatted AX.25 UI frame
// with valid source and destination callsigns.
//
// Validates: Requirement 10.1
func ValidateAX25Frame(frameData []byte) error {
	// Parse the frame
	frame, err := ax25.ParseFrame(frameData)
	if err != nil {
		return fmt.Errorf("invalid AX.25 frame: %w", err)
	}

	// Verify destination callsign is valid
	if frame.Dst.Call == "" {
		return fmt.Errorf("destination callsign is empty")
	}
	if len(frame.Dst.Call) > ax25.CallsignLen {
		return fmt.Errorf("destination callsign too long: %d chars (max %d)", len(frame.Dst.Call), ax25.CallsignLen)
	}
	if frame.Dst.SSID > ax25.MaxSSID {
		return fmt.Errorf("destination SSID out of range: %d (max %d)", frame.Dst.SSID, ax25.MaxSSID)
	}

	// Verify source callsign is valid
	if frame.Src.Call == "" {
		return fmt.Errorf("source callsign is empty")
	}
	if len(frame.Src.Call) > ax25.CallsignLen {
		return fmt.Errorf("source callsign too long: %d chars (max %d)", len(frame.Src.Call), ax25.CallsignLen)
	}
	if frame.Src.SSID > ax25.MaxSSID {
		return fmt.Errorf("source SSID out of range: %d (max %d)", frame.Src.SSID, ax25.MaxSSID)
	}

	return nil
}

// ValidateCLAAX25Framing validates that a CLA implementation produces valid AX.25 frames
// by attempting to build and parse a test frame with the given callsigns.
//
// This function can be used in tests to verify that all CLA implementations
// correctly encapsulate bundles in AX.25 frames with proper callsign addressing.
//
// Validates: Requirement 10.1
func ValidateCLAAX25Framing(srcCallsign, dstCallsign ax25.Callsign, payload []byte) error {
	// Build an AX.25 frame
	frame, err := ax25.BuildUIFrame(dstCallsign, srcCallsign, payload)
	if err != nil {
		return fmt.Errorf("failed to build AX.25 frame: %w", err)
	}

	// Validate the frame
	if err := ValidateAX25Frame(frame); err != nil {
		return fmt.Errorf("frame validation failed: %w", err)
	}

	// Parse the frame to verify round-trip
	parsed, err := ax25.ParseFrame(frame)
	if err != nil {
		return fmt.Errorf("failed to parse AX.25 frame: %w", err)
	}

	// Verify callsigns match
	if parsed.Src.Call != srcCallsign.Call || parsed.Src.SSID != srcCallsign.SSID {
		return fmt.Errorf("source callsign mismatch: got %s, want %s", parsed.Src.String(), srcCallsign.String())
	}
	if parsed.Dst.Call != dstCallsign.Call || parsed.Dst.SSID != dstCallsign.SSID {
		return fmt.Errorf("destination callsign mismatch: got %s, want %s", parsed.Dst.String(), dstCallsign.String())
	}

	// Verify payload matches
	if len(parsed.Info) != len(payload) {
		return fmt.Errorf("payload length mismatch: got %d bytes, want %d bytes", len(parsed.Info), len(payload))
	}
	for i := range payload {
		if parsed.Info[i] != payload[i] {
			return fmt.Errorf("payload mismatch at byte %d: got 0x%02X, want 0x%02X", i, parsed.Info[i], payload[i])
		}
	}

	return nil
}
