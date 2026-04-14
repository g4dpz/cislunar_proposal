package cla

import (
	"bytes"
	"testing"
	"terrestrial-dtn/ax25"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 21: LTP Segmentation/Reassembly Round-Trip
// For any bundle whose size exceeds a single AX.25 frame, LTP segmentation
// followed by reassembly SHALL produce a bundle identical to the original.
//
// **Validates: Requirement 10.3**
func TestProperty_LTPSegmentationReassemblyRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for payload size (0-2048 bytes to test various segmentation scenarios)
	genPayloadSize := gen.IntRange(0, 2048)

	// Generator for max frame size (32-256 bytes to force segmentation)
	genMaxFrameSize := gen.IntRange(32, 256)

	// Generator for callsigns
	genCallsign := gen.RegexMatch("[A-Z0-9]{1,6}")

	genSSID := gen.IntRange(0, ax25.MaxSSID).Map(func(i int) uint8 { return uint8(i) })

	properties.Property("segmentation and reassembly preserves payload", prop.ForAll(
		func(payloadSize int, maxFrameSize int, srcCall string, srcSSID uint8, dstCall string, dstSSID uint8) bool {
			// Generate payload with deterministic pattern
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			src := ax25.Callsign{Call: srcCall, SSID: srcSSID}
			dst := ax25.Callsign{Call: dstCall, SSID: dstSSID}

			// Segment the payload
			segments := segmentPayload(payload, maxFrameSize)

			// Build AX.25 frames for each segment
			var frames [][]byte
			for _, segment := range segments {
				frame, err := ax25.BuildUIFrame(dst, src, segment)
				if err != nil {
					t.Logf("BuildUIFrame failed: %v", err)
					return false
				}

				// Validate each frame
				if err := ValidateAX25Frame(frame); err != nil {
					t.Logf("Frame validation failed: %v", err)
					return false
				}

				frames = append(frames, frame)
			}

			// Reassemble: parse frames and concatenate payloads
			var reassembled []byte
			for i, frame := range frames {
				parsed, err := ax25.ParseFrame(frame)
				if err != nil {
					t.Logf("ParseFrame for segment %d failed: %v", i, err)
					return false
				}

				// Verify callsigns are preserved in each segment
				if parsed.Src.Call != src.Call || parsed.Src.SSID != src.SSID {
					t.Logf("Segment %d: source callsign mismatch", i)
					return false
				}
				if parsed.Dst.Call != dst.Call || parsed.Dst.SSID != dst.SSID {
					t.Logf("Segment %d: destination callsign mismatch", i)
					return false
				}

				reassembled = append(reassembled, parsed.Info...)
			}

			// Verify reassembled payload matches original
			if !bytes.Equal(reassembled, payload) {
				t.Logf("Reassembled payload mismatch: got %d bytes, want %d bytes",
					len(reassembled), len(payload))
				return false
			}

			return true
		},
		genPayloadSize,
		genMaxFrameSize,
		genCallsign,
		genSSID,
		genCallsign,
		genSSID,
	))

	properties.TestingRun(t)
}

// Property 21 (variant): Segmentation count is correct
// For any payload and max frame size, the number of segments SHALL be
// ceil(payload_size / max_frame_size)
//
// **Validates: Requirement 10.3**
func TestProperty_LTPSegmentationCount(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genPayloadSize := gen.IntRange(0, 1024)
	genMaxFrameSize := gen.IntRange(1, 256)

	properties.Property("segment count is correct", prop.ForAll(
		func(payloadSize int, maxFrameSize int) bool {
			payload := make([]byte, payloadSize)
			segments := segmentPayload(payload, maxFrameSize)

			// Expected segment count
			expectedCount := payloadSize / maxFrameSize
			if payloadSize%maxFrameSize != 0 {
				expectedCount++
			}
			if payloadSize == 0 {
				expectedCount = 1 // Empty payload still produces one segment
			}

			if len(segments) != expectedCount {
				t.Logf("Segment count mismatch: got %d, want %d (payload=%d, maxFrame=%d)",
					len(segments), expectedCount, payloadSize, maxFrameSize)
				return false
			}

			// Verify total reassembled size matches original
			totalSize := 0
			for _, seg := range segments {
				totalSize += len(seg)
			}

			if totalSize != payloadSize {
				t.Logf("Total reassembled size mismatch: got %d, want %d", totalSize, payloadSize)
				return false
			}

			return true
		},
		genPayloadSize,
		genMaxFrameSize,
	))

	properties.TestingRun(t)
}

// Property 21 (variant): Each segment is within max frame size
// For any segmentation, each segment SHALL be at most max_frame_size bytes
// (except possibly the last segment which may be smaller)
//
// **Validates: Requirement 10.3**
func TestProperty_LTPSegmentSizeConstraint(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genPayloadSize := gen.IntRange(0, 2048)
	genMaxFrameSize := gen.IntRange(1, 256)

	properties.Property("each segment respects max frame size", prop.ForAll(
		func(payloadSize int, maxFrameSize int) bool {
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			segments := segmentPayload(payload, maxFrameSize)

			for i, segment := range segments {
				if len(segment) > maxFrameSize {
					t.Logf("Segment %d exceeds max frame size: %d > %d", i, len(segment), maxFrameSize)
					return false
				}

				// All segments except the last should be exactly maxFrameSize
				// (unless payload is smaller than maxFrameSize)
				if i < len(segments)-1 && len(segment) != maxFrameSize {
					t.Logf("Non-final segment %d is not max size: %d != %d", i, len(segment), maxFrameSize)
					return false
				}
			}

			return true
		},
		genPayloadSize,
		genMaxFrameSize,
	))

	properties.TestingRun(t)
}

// Property 21 (variant): Binary data preservation
// For any binary payload (including all byte values 0x00-0xFF), segmentation
// and reassembly SHALL preserve the exact byte sequence
//
// **Validates: Requirement 10.3**
func TestProperty_LTPSegmentationBinaryDataPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for arbitrary binary data
	genBinaryPayload := gen.SliceOfN(512, gen.UInt8()).Map(func(b []uint8) []byte {
		result := make([]byte, len(b))
		for i, v := range b {
			result[i] = byte(v)
		}
		return result
	})

	genMaxFrameSize := gen.IntRange(16, 128)

	genCallsign := gen.RegexMatch("[A-Z0-9]{1,6}")

	genSSID := gen.IntRange(0, ax25.MaxSSID).Map(func(i int) uint8 { return uint8(i) })

	properties.Property("binary data is preserved through segmentation", prop.ForAll(
		func(payload []byte, maxFrameSize int, srcCall string, srcSSID uint8, dstCall string, dstSSID uint8) bool {
			src := ax25.Callsign{Call: srcCall, SSID: srcSSID}
			dst := ax25.Callsign{Call: dstCall, SSID: dstSSID}

			// Segment
			segments := segmentPayload(payload, maxFrameSize)

			// Build frames and reassemble
			var reassembled []byte
			for _, segment := range segments {
				frame, err := ax25.BuildUIFrame(dst, src, segment)
				if err != nil {
					return false
				}

				parsed, err := ax25.ParseFrame(frame)
				if err != nil {
					return false
				}

				reassembled = append(reassembled, parsed.Info...)
			}

			// Verify exact byte-for-byte match
			return bytes.Equal(reassembled, payload)
		},
		genBinaryPayload,
		genMaxFrameSize,
		genCallsign,
		genSSID,
		genCallsign,
		genSSID,
	))

	properties.TestingRun(t)
}
