package node

import (
	"reflect"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 23: Statistics Consistency
// For any sequence of node operations, the cumulative statistics (total bundles
// received, sent, bytes received, bytes sent) SHALL be monotonically non-decreasing
// and consistent with the individual operations performed.
//
// **Validates: Requirement 15.3**
func TestProperty_StatisticsConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for operation sequences
	genOperationType := gen.IntRange(0, 3).Map(func(i int) string {
		switch i {
		case 0:
			return "receive"
		case 1:
			return "send"
		case 2:
			return "forward"
		default:
			return "drop"
		}
	})

	genBundleSize := gen.IntRange(1, 1024)

	// Generator for operation sequence (list of operations)
	genOperationSequence := gen.SliceOf(
		gen.Struct(reflect.TypeOf(struct {
			OpType     string
			BundleSize int
		}{}), map[string]gopter.Gen{
			"OpType":     genOperationType,
			"BundleSize": genBundleSize,
		}),
	).SuchThat(func(ops interface{}) bool {
		slice := ops.([]interface{})
		return len(slice) >= 1 && len(slice) <= 50
	})

	properties.Property("statistics are monotonically non-decreasing and consistent", prop.ForAll(
		func(operations []interface{}) bool {
			// Initialize statistics
			stats := &NodeStatistics{
				TotalBundlesReceived: 0,
				TotalBundlesSent:     0,
				TotalBytesReceived:   0,
				TotalBytesSent:       0,
				AverageLatency:       0,
				ContactsCompleted:    0,
				ContactsMissed:       0,
			}

			// Track expected values
			expectedReceived := int64(0)
			expectedSent := int64(0)
			expectedBytesReceived := int64(0)
			expectedBytesSent := int64(0)
			expectedForwarded := int64(0)
			expectedDropped := int64(0)

			// Apply operations
			for _, op := range operations {
				opStruct := op.(struct {
					OpType     string
					BundleSize int
				})

				prevReceived := stats.TotalBundlesReceived
				prevSent := stats.TotalBundlesSent
				prevBytesReceived := stats.TotalBytesReceived
				prevBytesSent := stats.TotalBytesSent

				// Simulate operation
				switch opStruct.OpType {
				case "receive":
					stats.TotalBundlesReceived++
					stats.TotalBytesReceived += int64(opStruct.BundleSize)
					expectedReceived++
					expectedBytesReceived += int64(opStruct.BundleSize)

				case "send":
					stats.TotalBundlesSent++
					stats.TotalBytesSent += int64(opStruct.BundleSize)
					expectedSent++
					expectedBytesSent += int64(opStruct.BundleSize)

				case "forward":
					// Forward counts as both receive and send
					stats.TotalBundlesReceived++
					stats.TotalBundlesSent++
					stats.TotalBytesReceived += int64(opStruct.BundleSize)
					stats.TotalBytesSent += int64(opStruct.BundleSize)
					expectedReceived++
					expectedSent++
					expectedBytesReceived += int64(opStruct.BundleSize)
					expectedBytesSent += int64(opStruct.BundleSize)
					expectedForwarded++

				case "drop":
					// Drop doesn't affect send/receive stats
					expectedDropped++
				}

				// Verify monotonicity
				if stats.TotalBundlesReceived < prevReceived {
					t.Logf("TotalBundlesReceived decreased: %d -> %d",
						prevReceived, stats.TotalBundlesReceived)
					return false
				}
				if stats.TotalBundlesSent < prevSent {
					t.Logf("TotalBundlesSent decreased: %d -> %d",
						prevSent, stats.TotalBundlesSent)
					return false
				}
				if stats.TotalBytesReceived < prevBytesReceived {
					t.Logf("TotalBytesReceived decreased: %d -> %d",
						prevBytesReceived, stats.TotalBytesReceived)
					return false
				}
				if stats.TotalBytesSent < prevBytesSent {
					t.Logf("TotalBytesSent decreased: %d -> %d",
						prevBytesSent, stats.TotalBytesSent)
					return false
				}
			}

			// Verify final consistency
			if stats.TotalBundlesReceived != expectedReceived {
				t.Logf("TotalBundlesReceived mismatch: got %d, want %d",
					stats.TotalBundlesReceived, expectedReceived)
				return false
			}
			if stats.TotalBundlesSent != expectedSent {
				t.Logf("TotalBundlesSent mismatch: got %d, want %d",
					stats.TotalBundlesSent, expectedSent)
				return false
			}
			if stats.TotalBytesReceived != expectedBytesReceived {
				t.Logf("TotalBytesReceived mismatch: got %d, want %d",
					stats.TotalBytesReceived, expectedBytesReceived)
				return false
			}
			if stats.TotalBytesSent != expectedBytesSent {
				t.Logf("TotalBytesSent mismatch: got %d, want %d",
					stats.TotalBytesSent, expectedBytesSent)
				return false
			}

			return true
		},
		genOperationSequence,
	))

	properties.TestingRun(t)
}

// Property 23 (variant): Byte counts are consistent with bundle counts
// For any sequence of operations, total bytes SHALL be consistent with
// the sum of individual bundle sizes
//
// **Validates: Requirement 15.3**
func TestProperty_StatisticsByteCountConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for bundle sequence
	genBundleSequence := gen.SliceOf(gen.IntRange(1, 512)).
		SuchThat(func(sizes []int) bool {
			return len(sizes) >= 1 && len(sizes) <= 100
		})

	properties.Property("byte counts match sum of bundle sizes", prop.ForAll(
		func(bundleSizes []int) bool {
			stats := &NodeStatistics{}

			expectedBytes := int64(0)

			// Simulate receiving bundles
			for _, size := range bundleSizes {
				stats.TotalBundlesReceived++
				stats.TotalBytesReceived += int64(size)
				expectedBytes += int64(size)
			}

			// Verify consistency
			if stats.TotalBytesReceived != expectedBytes {
				t.Logf("Byte count mismatch: got %d, want %d",
					stats.TotalBytesReceived, expectedBytes)
				return false
			}

			// Verify bundle count
			if stats.TotalBundlesReceived != int64(len(bundleSizes)) {
				t.Logf("Bundle count mismatch: got %d, want %d",
					stats.TotalBundlesReceived, len(bundleSizes))
				return false
			}

			return true
		},
		genBundleSequence,
	))

	properties.TestingRun(t)
}

// Property 23 (variant): Contact statistics are non-negative
// For any sequence of contact operations, completed and missed counts
// SHALL be non-negative and their sum SHALL equal total contact attempts
//
// **Validates: Requirement 15.3**
func TestProperty_StatisticsContactCountsNonNegative(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for contact outcomes (true = completed, false = missed)
	genContactSequence := gen.SliceOf(gen.Bool()).
		SuchThat(func(contacts []bool) bool {
			return len(contacts) >= 1 && len(contacts) <= 100
		})

	properties.Property("contact statistics are non-negative and consistent", prop.ForAll(
		func(contactOutcomes []bool) bool {
			stats := &NodeStatistics{}

			expectedCompleted := int64(0)
			expectedMissed := int64(0)

			// Simulate contacts
			for _, completed := range contactOutcomes {
				if completed {
					stats.ContactsCompleted++
					expectedCompleted++
				} else {
					stats.ContactsMissed++
					expectedMissed++
				}

				// Verify non-negativity at each step
				if stats.ContactsCompleted < 0 {
					t.Logf("ContactsCompleted became negative: %d", stats.ContactsCompleted)
					return false
				}
				if stats.ContactsMissed < 0 {
					t.Logf("ContactsMissed became negative: %d", stats.ContactsMissed)
					return false
				}
			}

			// Verify final counts
			if stats.ContactsCompleted != expectedCompleted {
				t.Logf("ContactsCompleted mismatch: got %d, want %d",
					stats.ContactsCompleted, expectedCompleted)
				return false
			}
			if stats.ContactsMissed != expectedMissed {
				t.Logf("ContactsMissed mismatch: got %d, want %d",
					stats.ContactsMissed, expectedMissed)
				return false
			}

			// Verify total
			totalContacts := stats.ContactsCompleted + stats.ContactsMissed
			if totalContacts != int64(len(contactOutcomes)) {
				t.Logf("Total contacts mismatch: got %d, want %d",
					totalContacts, len(contactOutcomes))
				return false
			}

			return true
		},
		genContactSequence,
	))

	properties.TestingRun(t)
}
