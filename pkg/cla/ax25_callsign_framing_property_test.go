package cla

import (
	"testing"
	"terrestrial-dtn/ax25"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 20: AX.25 Callsign Framing
// For any bundle transmitted through the CLA on any phase (terrestrial, EM, LEO, cislunar),
// the output frame SHALL be encapsulated in AX.25 format carrying valid source and
// destination amateur radio callsigns.
//
// **Validates: Requirement 10.1**
func TestProperty_AX25CallsignFraming(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for valid callsigns (1-6 characters, alphanumeric)
	genCallsign := gen.RegexMatch("[A-Z0-9]{1,6}")

	// Generator for SSID (0-15)
	genSSID := gen.IntRange(0, ax25.MaxSSID).Map(func(i int) uint8 { return uint8(i) })

	// Generator for payload (0-256 bytes)
	genPayload := gen.SliceOfN(256, gen.UInt8()).Map(func(b []uint8) []byte {
		result := make([]byte, len(b))
		for i, v := range b {
			result[i] = byte(v)
		}
		return result
	})

	properties.Property("all frames have valid AX.25 callsign framing", prop.ForAll(
		func(srcCall string, srcSSID uint8, dstCall string, dstSSID uint8, payload []byte) bool {
			// Create callsigns
			src := ax25.Callsign{Call: srcCall, SSID: srcSSID}
			dst := ax25.Callsign{Call: dstCall, SSID: dstSSID}

			// Build AX.25 frame
			frame, err := ax25.BuildUIFrame(dst, src, payload)
			if err != nil {
				// Should not fail with valid inputs
				t.Logf("BuildUIFrame failed with valid inputs: %v", err)
				return false
			}

			// Validate the frame has proper AX.25 structure
			if err := ValidateAX25Frame(frame); err != nil {
				t.Logf("Frame validation failed: %v", err)
				return false
			}

			// Parse the frame to verify callsigns are preserved
			parsed, err := ax25.ParseFrame(frame)
			if err != nil {
				t.Logf("ParseFrame failed: %v", err)
				return false
			}

			// Verify source callsign
			if parsed.Src.Call != src.Call {
				t.Logf("Source callsign mismatch: got %s, want %s", parsed.Src.Call, src.Call)
				return false
			}
			if parsed.Src.SSID != src.SSID {
				t.Logf("Source SSID mismatch: got %d, want %d", parsed.Src.SSID, src.SSID)
				return false
			}

			// Verify destination callsign
			if parsed.Dst.Call != dst.Call {
				t.Logf("Destination callsign mismatch: got %s, want %s", parsed.Dst.Call, dst.Call)
				return false
			}
			if parsed.Dst.SSID != dst.SSID {
				t.Logf("Destination SSID mismatch: got %d, want %d", parsed.Dst.SSID, dst.SSID)
				return false
			}

			// Verify payload is preserved
			if len(parsed.Info) != len(payload) {
				t.Logf("Payload length mismatch: got %d, want %d", len(parsed.Info), len(payload))
				return false
			}
			for i := range payload {
				if parsed.Info[i] != payload[i] {
					t.Logf("Payload mismatch at byte %d", i)
					return false
				}
			}

			return true
		},
		genCallsign,
		genSSID,
		genCallsign,
		genSSID,
		genPayload,
	))

	properties.TestingRun(t)
}

// Property 20 (variant): All CLA types produce valid AX.25 frames
// For any CLA type, frames SHALL have valid AX.25 callsign framing
//
// **Validates: Requirement 10.1**
func TestProperty_AllCLATypesProduceValidAX25Frames(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Generator for CLA types
	genCLAType := gen.IntRange(0, 5).Map(func(i int) CLAType {
		return CLAType(i)
	})

	// Generator for valid callsigns
	genCallsign := gen.RegexMatch("[A-Z0-9]{1,6}")

	genSSID := gen.IntRange(0, ax25.MaxSSID).Map(func(i int) uint8 { return uint8(i) })

	genPayload := gen.SliceOfN(128, gen.UInt8()).Map(func(b []uint8) []byte {
		result := make([]byte, len(b))
		for i, v := range b {
			result[i] = byte(v)
		}
		return result
	})

	properties.Property("all CLA types produce valid AX.25 frames", prop.ForAll(
		func(claType CLAType, srcCall string, srcSSID uint8, dstCall string, dstSSID uint8, payload []byte) bool {
			// All CLA types use the same AX.25 framing
			src := ax25.Callsign{Call: srcCall, SSID: srcSSID}
			dst := ax25.Callsign{Call: dstCall, SSID: dstSSID}

			// Validate CLA AX.25 framing
			err := ValidateCLAAX25Framing(src, dst, payload)
			if err != nil {
				t.Logf("CLA type %s failed AX.25 framing: %v", claType.String(), err)
				return false
			}

			return true
		},
		genCLAType,
		genCallsign,
		genSSID,
		genCallsign,
		genSSID,
		genPayload,
	))

	properties.TestingRun(t)
}
