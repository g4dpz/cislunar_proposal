package kiss

import (
	"bytes"
	"testing"

	"pgregory.net/rapid"
)

// Feature: hdtn-migration, Property 11: KISS encode/decode round-trip
// For any byte sequence of length 1 to 1500, KISS-encode then KISS-decode
// produces the original byte sequence. This holds regardless of whether the
// input contains KISS special bytes (0xC0, 0xDB).
//
// **Validates: Requirements 5.2, 5.3, 5.7, 10.6**

func TestPropertyKissEncodeDecodeRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a byte slice of length 1 to 1500
		length := rapid.IntRange(1, 1500).Draw(t, "length")
		data := make([]byte, length)
		for i := range data {
			data[i] = byte(rapid.IntRange(0, 255).Draw(t, "byte"))
		}

		// Encode
		encoded, err := Encode(data)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Decode
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		// Assert round-trip produces original data
		if !bytes.Equal(decoded, data) {
			t.Fatalf("round-trip mismatch: original length=%d, decoded length=%d", len(data), len(decoded))
		}
	})
}
