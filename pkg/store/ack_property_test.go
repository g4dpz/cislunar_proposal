package store

import (
	"testing"
	"time"

	"terrestrial-dtn/pkg/bpa"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_ACKDeletesNoACKRetains validates Property 9:
// For any bundle transmitted during a contact window, if the remote node acknowledges
// receipt, the bundle SHALL be deleted from the Bundle_Store; if the transmission is
// not acknowledged, the bundle SHALL remain in the Bundle_Store for retry.
//
// Validates Requirements 5.4, 5.5
func TestProperty_ACKDeletesNoACKRetains(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: ACK deletes bundle, no-ACK retains bundle
	properties.Property("ACK deletes, no-ACK retains", prop.ForAll(
		func(ackReceived bool, seqNum uint64) bool {
			// Create store
			config := Config{
				MaxBytes: 1024 * 1024,
			}
			store := New(config)

			// Create and store bundle
			bundle := bpa.Bundle{
				ID: bpa.BundleID{
					SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
					CreationTimestamp: uint64(time.Now().Unix()),
					SequenceNumber:    seqNum,
				},
				Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
				Payload:     []byte("test payload"),
				Priority:    bpa.PriorityNormal,
				Lifetime:    3600,
				CreatedAt:   uint64(time.Now().Unix()),
				BundleType:  bpa.BundleTypeData,
			}

			if err := store.Store(bundle); err != nil {
				return false
			}

			// Simulate transmission
			if ackReceived {
				// ACK received: delete bundle
				if err := store.Delete(bundle.ID); err != nil {
					return false
				}
			}
			// No ACK: bundle remains in store

			// Verify bundle presence
			retrieved, err := store.Retrieve(bundle.ID)

			if ackReceived {
				// Bundle should be deleted (not found)
				return err != nil && retrieved == nil
			} else {
				// Bundle should be retained
				return err == nil && retrieved != nil
			}
		},
		gen.Bool(),
		gen.UInt64(),
	))

	properties.TestingRun(t)
}

// TestProperty_RetryAfterNoACK validates that bundles are available for retry
func TestProperty_RetryAfterNoACK(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Bundles without ACK remain available for retry
	properties.Property("bundles available for retry after no-ACK", prop.ForAll(
		func(numBundles int) bool {
			if numBundles < 0 || numBundles > 100 {
				return true // Skip invalid inputs
			}

			config := Config{
				MaxBytes: 10 * 1024 * 1024,
			}
			store := New(config)

			// Store bundles
			bundleIDs := make([]bpa.BundleID, numBundles)
			for i := 0; i < numBundles; i++ {
				bundle := bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
						CreationTimestamp: uint64(time.Now().Unix()),
						SequenceNumber:    uint64(i),
					},
					Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
					Payload:     []byte("test"),
					Priority:    bpa.PriorityNormal,
					Lifetime:    3600,
					CreatedAt:   uint64(time.Now().Unix()),
					BundleType:  bpa.BundleTypeData,
				}

				if err := store.Store(bundle); err != nil {
					return false
				}
				bundleIDs[i] = bundle.ID
			}

			// Simulate transmission without ACK (no deletion)
			// All bundles should still be in store

			// Verify all bundles are available for retry
			for _, id := range bundleIDs {
				if _, err := store.Retrieve(id); err != nil {
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}

// TestProperty_ACKSequence validates ACK/no-ACK sequences
func TestProperty_ACKSequence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: ACK sequence correctly updates store
	properties.Property("ACK sequence updates store correctly", prop.ForAll(
		func(ackPattern []bool) bool {
			if len(ackPattern) == 0 || len(ackPattern) > 50 {
				return true // Skip invalid inputs
			}

			config := Config{
				MaxBytes: 10 * 1024 * 1024,
			}
			store := New(config)

			// Store bundles
			bundles := make([]bpa.Bundle, len(ackPattern))
			for i := range ackPattern {
				bundles[i] = bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
						CreationTimestamp: uint64(time.Now().Unix()),
						SequenceNumber:    uint64(i),
					},
					Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
					Payload:     []byte("test"),
					Priority:    bpa.PriorityNormal,
					Lifetime:    3600,
					CreatedAt:   uint64(time.Now().Unix()),
					BundleType:  bpa.BundleTypeData,
				}

				if err := store.Store(bundles[i]); err != nil {
					return false
				}
			}

			// Process ACK pattern
			expectedRetained := 0
			for i, ack := range ackPattern {
				if ack {
					// ACK received: delete
					store.Delete(bundles[i].ID)
				} else {
					// No ACK: retain
					expectedRetained++
				}
			}

			// Verify retained count
			allBundles := store.ListByPriority()
			return len(allBundles) == expectedRetained
		},
		gen.SliceOf(gen.Bool()),
	))

	properties.TestingRun(t)
}

// TestProperty_ACKIdempotence validates that multiple ACKs don't cause errors
func TestProperty_ACKIdempotence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Multiple ACKs for same bundle are idempotent
	properties.Property("multiple ACKs are idempotent", prop.ForAll(
		func(numACKs int, seqNum uint64) bool {
			if numACKs < 1 || numACKs > 10 {
				return true // Skip invalid inputs
			}

			config := Config{
				MaxBytes: 1024 * 1024,
			}
			store := New(config)

			// Store bundle
			bundle := bpa.Bundle{
				ID: bpa.BundleID{
					SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
					CreationTimestamp: uint64(time.Now().Unix()),
					SequenceNumber:    seqNum,
				},
				Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
				Payload:     []byte("test"),
				Priority:    bpa.PriorityNormal,
				Lifetime:    3600,
				CreatedAt:   uint64(time.Now().Unix()),
				BundleType:  bpa.BundleTypeData,
			}

			if err := store.Store(bundle); err != nil {
				return false
			}

			// Process multiple ACKs
			for i := 0; i < numACKs; i++ {
				store.Delete(bundle.ID) // Ignore errors (bundle may already be deleted)
			}

			// Verify bundle is deleted
			_, err := store.Retrieve(bundle.ID)
			return err != nil // Should not be found
		},
		gen.IntRange(1, 10),
		gen.UInt64(),
	))

	properties.TestingRun(t)
}
