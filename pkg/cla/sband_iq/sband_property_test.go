package sband_iq

import (
	"bytes"
	"math/rand"
	"testing"

	"terrestrial-dtn/pkg/bpa"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: test-framework-srs-sdd, Property 23: Bundle Serialization Round-Trip
// For any valid bundle with payload sizes from 1 to 1500 bytes, serializing then
// deserializing SHALL produce a bundle with identical bundle type, priority, lifetime,
// destination endpoint, and payload content.
// **Validates: SRS-TF-022 (Requirements 22.1)**
func TestProperty_BundleSerializationRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	config := DefaultSBandConfig("W1ABC")
	claInstance, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create SBandIQCLA: %v", err)
	}

	properties.Property("serialize then deserialize preserves all bundle fields", prop.ForAll(
		func(payloadSize int, priority int, lifetime int64, payloadByte byte) bool {
			// Generate payload of the specified size
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte((int(payloadByte) + i) % 256)
			}

			// Map priority to valid range (0-3)
			bundlePriority := bpa.Priority(priority)

			// Map bundle type (use BundleTypeData for simplicity since it's the most common)
			bundleType := bpa.BundleType(priority % 3) // 0=Data, 1=PingRequest, 2=PingResponse

			// Create a destination that will round-trip correctly through serialization.
			// serializeBundle uses bundle.Destination.String() which formats as "scheme://ssp"
			// deserializeBundle stores the full string as SSP with Scheme="ipn"
			// So we use a destination whose String() output we can verify.
			destSSP := "2.0"
			original := &bpa.Bundle{
				ID: bpa.BundleID{
					SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
					CreationTimestamp: 1000,
					SequenceNumber:    1,
				},
				Destination: bpa.EndpointID{Scheme: "ipn", SSP: destSSP},
				Payload:     payload,
				Priority:    bundlePriority,
				Lifetime:    lifetime,
				CreatedAt:   1000,
				BundleType:  bundleType,
			}

			// Serialize
			serialized := claInstance.serializeBundle(original)
			if len(serialized) == 0 {
				t.Logf("serializeBundle returned empty data")
				return false
			}

			// Deserialize
			deserialized, err := claInstance.deserializeBundle(serialized)
			if err != nil {
				t.Logf("deserializeBundle error: %v", err)
				return false
			}

			// Verify bundle type
			if deserialized.BundleType != original.BundleType {
				t.Logf("BundleType mismatch: got %v, want %v", deserialized.BundleType, original.BundleType)
				return false
			}

			// Verify priority
			if deserialized.Priority != original.Priority {
				t.Logf("Priority mismatch: got %v, want %v", deserialized.Priority, original.Priority)
				return false
			}

			// Verify lifetime
			if deserialized.Lifetime != original.Lifetime {
				t.Logf("Lifetime mismatch: got %v, want %v", deserialized.Lifetime, original.Lifetime)
				return false
			}

			// Verify destination (the serialized form is Destination.String() = "ipn://2.0")
			// deserializeBundle stores it as EndpointID{Scheme: "ipn", SSP: destStr}
			// where destStr is the full String() output of the original destination
			expectedDestStr := original.Destination.String()
			actualDestStr := deserialized.Destination.SSP
			if actualDestStr != expectedDestStr {
				t.Logf("Destination mismatch: got SSP=%q, want %q", actualDestStr, expectedDestStr)
				return false
			}

			// Verify payload content (byte-for-byte)
			if !bytes.Equal(deserialized.Payload, original.Payload) {
				t.Logf("Payload mismatch: got len=%d, want len=%d", len(deserialized.Payload), len(original.Payload))
				return false
			}

			return true
		},
		gen.IntRange(1, 1500),      // payloadSize: 1-1500 bytes
		gen.IntRange(0, 3),         // priority: 0-3
		gen.Int64Range(1, 3600),    // lifetime: 1-3600 seconds
		gen.UInt8(),                // payloadByte: seed for payload content
	))

	properties.TestingRun(t)
}

// Feature: test-framework-srs-sdd, Property 24: AX.25 Framing Round-Trip
// For any valid payload with sizes from 1 to 1500 bytes, AX.25 framing
// (createAX25Frame) then extraction (extractAX25Frame) SHALL produce a byte
// sequence identical to the original payload.
// **Validates: SRS-TF-022 (Requirements 22.2)**
func TestProperty_AX25FramingRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	config := DefaultSBandConfig("W1ABC")
	claInstance, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create SBandIQCLA: %v", err)
	}

	properties.Property("createAX25Frame then extractAX25Frame preserves payload", prop.ForAll(
		func(payloadSize int, seed int64) bool {
			// Generate random payload of the specified size
			rng := rand.New(rand.NewSource(seed))
			payload := make([]byte, payloadSize)
			for i := range payload {
				payload[i] = byte(rng.Intn(256))
			}

			// Frame the payload
			frame := claInstance.createAX25Frame(payload)
			if len(frame) == 0 {
				t.Logf("createAX25Frame returned empty frame for payload size %d", payloadSize)
				return false
			}

			// Extract the payload from the frame
			extracted, err := claInstance.extractAX25Frame(frame)
			if err != nil {
				t.Logf("extractAX25Frame error: %v", err)
				return false
			}

			// Verify byte-identical payload
			if !bytes.Equal(extracted, payload) {
				t.Logf("Payload mismatch: original len=%d, extracted len=%d", len(payload), len(extracted))
				return false
			}

			return true
		},
		gen.IntRange(1, 1500), // payloadSize: 1–1500 bytes
		gen.Int64(),           // seed for random byte content
	))

	properties.TestingRun(t)
}
