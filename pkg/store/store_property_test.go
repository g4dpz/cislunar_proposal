package store

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"terrestrial-dtn/pkg/bpa"
)

// Property 1: Bundle Store/Retrieve Round-Trip
// For any valid BPv7 bundle, storing it and then retrieving it by its ID
// SHALL produce a bundle identical to the original.
// Validates: Requirement 2.2
func TestProperty_BundleStoreRetrieveRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("store and retrieve returns identical bundle", prop.ForAll(
		func(payloadSize uint8, priority uint8, lifetime int64) bool {
			// Generate a valid bundle
			bundle := &bpa.Bundle{
				ID: bpa.BundleID{
					SourceEID: bpa.EndpointID{
						Scheme: "dtn",
						SSP:    "test-node",
					},
					CreationTimestamp: 1000,
					SequenceNumber:    1,
				},
				Destination: bpa.EndpointID{
					Scheme: "dtn",
					SSP:    "dest-node",
				},
				Payload:     make([]byte, payloadSize),
				Priority:    bpa.Priority(priority % 4), // 0-3
				Lifetime:    lifetime,
				CreatedAt:   1000,
				BundleType:  bpa.BundleTypeData,
			}

			// Create store with sufficient capacity
			store := NewBundleStore(1024 * 1024) // 1 MB

			// Store the bundle
			if err := store.Store(bundle); err != nil {
				return false
			}

			// Retrieve the bundle
			retrieved, err := store.Retrieve(bundle.ID)
			if err != nil {
				return false
			}

			// Verify identity
			if retrieved.ID.SourceEID.Scheme != bundle.ID.SourceEID.Scheme {
				return false
			}
			if retrieved.ID.SourceEID.SSP != bundle.ID.SourceEID.SSP {
				return false
			}
			if retrieved.ID.CreationTimestamp != bundle.ID.CreationTimestamp {
				return false
			}
			if retrieved.ID.SequenceNumber != bundle.ID.SequenceNumber {
				return false
			}
			if retrieved.Destination.Scheme != bundle.Destination.Scheme {
				return false
			}
			if retrieved.Destination.SSP != bundle.Destination.SSP {
				return false
			}
			if len(retrieved.Payload) != len(bundle.Payload) {
				return false
			}
			if retrieved.Priority != bundle.Priority {
				return false
			}
			if retrieved.Lifetime != bundle.Lifetime {
				return false
			}
			if retrieved.CreatedAt != bundle.CreatedAt {
				return false
			}
			if retrieved.BundleType != bundle.BundleType {
				return false
			}

			return true
		},
		gen.UInt8Range(1, 255),      // payloadSize
		gen.UInt8Range(0, 3),         // priority
		gen.Int64Range(1, 3600),      // lifetime (1 second to 1 hour)
	))

	properties.TestingRun(t)
}

// Property 5: Store Capacity Bound
// For any sequence of store and delete operations, the total stored bytes
// SHALL never exceed the configured maximum storage capacity.
// Validates: Requirement 2.6
func TestProperty_StoreCapacityBound(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("store capacity never exceeded", prop.ForAll(
		func(operations []bool, payloadSizes []uint8) bool {
			maxBytes := int64(10000) // 10 KB max
			store := NewBundleStore(maxBytes)

			bundleIDs := make([]bpa.BundleID, 0)
			seqNum := uint64(1)

			// Execute operations: true = store, false = delete
			for i, op := range operations {
				if i >= len(payloadSizes) {
					break
				}

				if op {
					// Store operation
					bundle := &bpa.Bundle{
						ID: bpa.BundleID{
							SourceEID: bpa.EndpointID{
								Scheme: "dtn",
								SSP:    "test",
							},
							CreationTimestamp: 1000,
							SequenceNumber:    seqNum,
						},
						Destination: bpa.EndpointID{
							Scheme: "dtn",
							SSP:    "dest",
						},
						Payload:    make([]byte, payloadSizes[i]),
						Priority:   bpa.PriorityNormal,
						Lifetime:   3600,
						CreatedAt:  1000,
						BundleType: bpa.BundleTypeData,
					}
					seqNum++

					err := store.Store(bundle)
					if err == nil {
						bundleIDs = append(bundleIDs, bundle.ID)
					}
					// If store fails due to capacity, that's expected
				} else {
					// Delete operation
					if len(bundleIDs) > 0 {
						// Delete a random bundle
						idx := i % len(bundleIDs)
						store.Delete(bundleIDs[idx])
						// Remove from tracking
						bundleIDs = append(bundleIDs[:idx], bundleIDs[idx+1:]...)
					}
				}

				// Check capacity invariant
				capacity := store.Capacity()
				if capacity.UsedBytes > capacity.TotalBytes {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.Bool()),
		gen.SliceOf(gen.UInt8Range(10, 200)),
	))

	properties.TestingRun(t)
}

// Property 3: Priority Ordering Invariant
// For any set of bundles stored, listing them by priority SHALL produce a sequence
// where each bundle's priority is greater than or equal to the next.
// Validates: Requirements 2.3, 5.3
func TestProperty_PriorityOrderingInvariant(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("bundles are ordered by priority", prop.ForAll(
		func(priorities []uint8) bool {
			if len(priorities) == 0 {
				return true
			}

			store := NewBundleStore(1024 * 1024) // 1 MB

			// Store bundles with various priorities
			for i, p := range priorities {
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID: bpa.EndpointID{
							Scheme: "dtn",
							SSP:    "test",
						},
						CreationTimestamp: 1000,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: bpa.EndpointID{
						Scheme: "dtn",
						SSP:    "dest",
					},
					Payload:    []byte("test"),
					Priority:   bpa.Priority(p % 4), // 0-3
					Lifetime:   3600,
					CreatedAt:  1000,
					BundleType: bpa.BundleTypeData,
				}

				if err := store.Store(bundle); err != nil {
					return false
				}
			}

			// List by priority
			bundles := store.ListByPriority()

			// Verify ordering: each priority >= next
			for i := 0; i < len(bundles)-1; i++ {
				if bundles[i].Priority < bundles[i+1].Priority {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(20, gen.UInt8Range(0, 3)),
	))

	properties.TestingRun(t)
}

// Property 4: Eviction Policy Ordering
// When eviction is triggered, expired bundles SHALL be evicted first,
// then lowest-priority bundles, and critical-priority bundles SHALL be
// preserved until all lower-priority bundles have been evicted.
// Validates: Requirements 2.4, 2.5
func TestProperty_EvictionPolicyOrdering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("expired bundles evicted first", prop.ForAll(
		func(numExpired uint8, numValid uint8) bool {
			if numExpired == 0 && numValid == 0 {
				return true
			}

			store := NewBundleStore(1024 * 1024)
			currentTime := int64(2000)

			// Add expired bundles
			for i := uint8(0); i < numExpired; i++ {
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID: bpa.EndpointID{
							Scheme: "dtn",
							SSP:    "test",
						},
						CreationTimestamp: 1000,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: bpa.EndpointID{
						Scheme: "dtn",
						SSP:    "dest",
					},
					Payload:    []byte("test"),
					Priority:   bpa.PriorityNormal,
					Lifetime:   500, // Expired at currentTime=2000
					CreatedAt:  1000,
					BundleType: bpa.BundleTypeData,
				}
				store.Store(bundle)
			}

			// Add valid bundles
			for i := uint8(0); i < numValid; i++ {
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID: bpa.EndpointID{
							Scheme: "dtn",
							SSP:    "test",
						},
						CreationTimestamp: 1000,
						SequenceNumber:    uint64(numExpired + i + 1),
					},
					Destination: bpa.EndpointID{
						Scheme: "dtn",
						SSP:    "dest",
					},
					Payload:    []byte("test"),
					Priority:   bpa.PriorityNormal,
					Lifetime:   3600, // Valid
					CreatedAt:  1000,
					BundleType: bpa.BundleTypeData,
				}
				store.Store(bundle)
			}

			initialCount := store.Count()

			// Evict expired bundles
			evicted := store.EvictExpired(currentTime)

			// Should have evicted all expired bundles
			if evicted != int(numExpired) {
				return false
			}

			// Should have only valid bundles remaining
			if store.Count() != int(numValid) {
				return false
			}

			// Verify initial count was correct
			_ = initialCount // Use the variable

			return true
		},
		gen.UInt8Range(0, 10),
		gen.UInt8Range(0, 10),
	))

	properties.Property("lowest priority bundles evicted first", prop.ForAll(
		func(priorities []uint8) bool {
			if len(priorities) == 0 {
				return true
			}

			store := NewBundleStore(1024 * 1024)

			// Store bundles with various priorities
			for i, p := range priorities {
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID: bpa.EndpointID{
							Scheme: "dtn",
							SSP:    "test",
						},
						CreationTimestamp: 1000,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: bpa.EndpointID{
						Scheme: "dtn",
						SSP:    "dest",
					},
					Payload:    []byte("test"),
					Priority:   bpa.Priority(p % 4),
					Lifetime:   3600,
					CreatedAt:  1000,
					BundleType: bpa.BundleTypeData,
				}
				store.Store(bundle)
			}

			// Evict lowest priority
			evicted, err := store.EvictLowestPriority()
			if err != nil {
				return false
			}

			// Verify all remaining bundles have priority >= evicted
			remaining := store.ListByPriority()
			for _, bundle := range remaining {
				if bundle.Priority < evicted.Priority {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(10, gen.UInt8Range(0, 3)),
	))

	properties.TestingRun(t)
}

// Property 6: Bundle Lifetime Enforcement
// For any set of bundles after a cleanup cycle, zero bundles SHALL have
// a creation timestamp plus lifetime less than or equal to the current time.
// Validates: Requirements 3.1, 3.2
func TestProperty_BundleLifetimeEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("no expired bundles after cleanup", prop.ForAll(
		func(lifetimes []int64, currentTime int64) bool {
			if len(lifetimes) == 0 {
				return true
			}

			store := NewBundleStore(1024 * 1024)
			createdAt := currentTime - 1000 // Bundles created 1000 seconds ago

			// Store bundles with various lifetimes
			for i, lifetime := range lifetimes {
				if lifetime <= 0 {
					continue // Skip invalid lifetimes
				}

				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID: bpa.EndpointID{
							Scheme: "dtn",
							SSP:    "test",
						},
						CreationTimestamp: createdAt,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: bpa.EndpointID{
						Scheme: "dtn",
						SSP:    "dest",
					},
					Payload:    []byte("test"),
					Priority:   bpa.PriorityNormal,
					Lifetime:   lifetime,
					CreatedAt:  createdAt,
					BundleType: bpa.BundleTypeData,
				}

				store.Store(bundle)
			}

			// Run cleanup cycle
			store.EvictExpired(currentTime)

			// Verify no expired bundles remain
			bundles := store.ListByPriority()
			for _, bundle := range bundles {
				if bundle.IsExpired(currentTime) {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(20, gen.Int64Range(1, 2000)),
		gen.Int64Range(1000, 3000),
	))

	properties.TestingRun(t)
}
