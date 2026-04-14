package cla

import (
	"bytes"
	"testing"
	"terrestrial-dtn/ax25"
)

// TestLTPSegmentationRoundTrip tests that large bundles can be segmented
// into multiple AX.25 frames and reassembled correctly.
//
// This simulates LTP segmentation/reassembly for bundles exceeding a single frame.
// In the real implementation, ION-DTN handles LTP segmentation/reassembly.
// This test validates the concept using AX.25 frame size limits.
//
// Validates: Requirement 10.3
func TestLTPSegmentationRoundTrip(t *testing.T) {
	// AX.25 frame has a practical payload limit (typically 256 bytes for UI frames)
	// For this test, we'll use a smaller limit to force segmentation
	const maxFramePayload = 128

	tests := []struct {
		name        string
		payloadSize int
	}{
		{
			name:        "small payload (no segmentation needed)",
			payloadSize: 64,
		},
		{
			name:        "exactly one frame",
			payloadSize: maxFramePayload,
		},
		{
			name:        "requires 2 segments",
			payloadSize: maxFramePayload + 1,
		},
		{
			name:        "requires 3 segments",
			payloadSize: maxFramePayload*2 + 50,
		},
		{
			name:        "large bundle (10 segments)",
			payloadSize: maxFramePayload * 10,
		},
	}

	src := ax25.Callsign{Call: "W1AW", SSID: 0}
	dst := ax25.Callsign{Call: "N0CALL", SSID: 1}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate test payload
			originalPayload := make([]byte, tt.payloadSize)
			for i := range originalPayload {
				originalPayload[i] = byte(i % 256)
			}

			// Simulate LTP segmentation: split payload into chunks
			segments := segmentPayload(originalPayload, maxFramePayload)
			t.Logf("Payload size: %d bytes, segments: %d", tt.payloadSize, len(segments))

			// Transmit each segment as an AX.25 frame
			var frames [][]byte
			for i, segment := range segments {
				frame, err := ax25.BuildUIFrame(dst, src, segment)
				if err != nil {
					t.Fatalf("BuildUIFrame for segment %d failed: %v", i, err)
				}

				// Validate the frame
				if err := ValidateAX25Frame(frame); err != nil {
					t.Fatalf("Frame validation failed for segment %d: %v", i, err)
				}

				frames = append(frames, frame)
			}

			// Simulate LTP reassembly: parse frames and concatenate payloads
			var reassembled []byte
			for i, frame := range frames {
				parsed, err := ax25.ParseFrame(frame)
				if err != nil {
					t.Fatalf("ParseFrame for segment %d failed: %v", i, err)
				}

				// Verify callsigns match
				if parsed.Src.Call != src.Call || parsed.Src.SSID != src.SSID {
					t.Errorf("Segment %d: source callsign mismatch", i)
				}
				if parsed.Dst.Call != dst.Call || parsed.Dst.SSID != dst.SSID {
					t.Errorf("Segment %d: destination callsign mismatch", i)
				}

				reassembled = append(reassembled, parsed.Info...)
			}

			// Verify reassembled payload matches original
			if len(reassembled) != len(originalPayload) {
				t.Errorf("Reassembled payload length mismatch: got %d, want %d",
					len(reassembled), len(originalPayload))
			}

			if !bytes.Equal(reassembled, originalPayload) {
				t.Errorf("Reassembled payload does not match original")
				// Find first mismatch
				for i := 0; i < len(originalPayload) && i < len(reassembled); i++ {
					if originalPayload[i] != reassembled[i] {
						t.Errorf("First mismatch at byte %d: got 0x%02X, want 0x%02X",
							i, reassembled[i], originalPayload[i])
						break
					}
				}
			} else {
				t.Logf("Segmentation/reassembly round-trip successful: %d bytes", len(reassembled))
			}
		})
	}
}

// segmentPayload splits a payload into chunks of maximum size
func segmentPayload(payload []byte, maxSize int) [][]byte {
	if len(payload) == 0 {
		return [][]byte{{}}
	}

	var segments [][]byte
	for i := 0; i < len(payload); i += maxSize {
		end := i + maxSize
		if end > len(payload) {
			end = len(payload)
		}
		segment := make([]byte, end-i)
		copy(segment, payload[i:end])
		segments = append(segments, segment)
	}

	return segments
}

// TestLTPSegmentationEdgeCases tests edge cases for LTP segmentation
// Validates: Requirement 10.3
func TestLTPSegmentationEdgeCases(t *testing.T) {
	const maxFramePayload = 128

	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "empty payload",
			payload: []byte{},
		},
		{
			name:    "single byte",
			payload: []byte{0x42},
		},
		{
			name:    "all zeros",
			payload: make([]byte, 256),
		},
		{
			name:    "all ones",
			payload: bytes.Repeat([]byte{0xFF}, 256),
		},
		{
			name:    "alternating pattern",
			payload: bytes.Repeat([]byte{0xAA, 0x55}, 128),
		},
	}

	src := ax25.Callsign{Call: "TEST", SSID: 0}
	dst := ax25.Callsign{Call: "DEST", SSID: 0}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Segment
			segments := segmentPayload(tt.payload, maxFramePayload)

			// Build frames
			var frames [][]byte
			for _, segment := range segments {
				frame, err := ax25.BuildUIFrame(dst, src, segment)
				if err != nil {
					t.Fatalf("BuildUIFrame failed: %v", err)
				}
				frames = append(frames, frame)
			}

			// Reassemble
			var reassembled []byte
			for _, frame := range frames {
				parsed, err := ax25.ParseFrame(frame)
				if err != nil {
					t.Fatalf("ParseFrame failed: %v", err)
				}
				reassembled = append(reassembled, parsed.Info...)
			}

			// Verify
			if !bytes.Equal(reassembled, tt.payload) {
				t.Errorf("Reassembled payload mismatch: got %d bytes, want %d bytes",
					len(reassembled), len(tt.payload))
			} else {
				t.Logf("Edge case passed: %d bytes in %d segments", len(reassembled), len(segments))
			}
		})
	}
}

// TestLTPSegmentationBinaryData tests segmentation with binary data
// Validates: Requirement 10.3
func TestLTPSegmentationBinaryData(t *testing.T) {
	const maxFramePayload = 100

	// Generate binary data with all possible byte values
	payload := make([]byte, 512)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	src := ax25.Callsign{Call: "BIN", SSID: 1}
	dst := ax25.Callsign{Call: "DATA", SSID: 2}

	// Segment
	segments := segmentPayload(payload, maxFramePayload)
	t.Logf("Binary payload: %d bytes, %d segments", len(payload), len(segments))

	// Build and parse frames
	var reassembled []byte
	for i, segment := range segments {
		frame, err := ax25.BuildUIFrame(dst, src, segment)
		if err != nil {
			t.Fatalf("BuildUIFrame for segment %d failed: %v", i, err)
		}

		parsed, err := ax25.ParseFrame(frame)
		if err != nil {
			t.Fatalf("ParseFrame for segment %d failed: %v", i, err)
		}

		reassembled = append(reassembled, parsed.Info...)
	}

	// Verify every byte
	if len(reassembled) != len(payload) {
		t.Fatalf("Length mismatch: got %d, want %d", len(reassembled), len(payload))
	}

	for i := range payload {
		if reassembled[i] != payload[i] {
			t.Errorf("Byte %d mismatch: got 0x%02X, want 0x%02X", i, reassembled[i], payload[i])
		}
	}

	t.Logf("Binary data segmentation/reassembly successful: %d bytes", len(reassembled))
}
