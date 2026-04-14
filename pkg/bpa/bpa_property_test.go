package bpa

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 2: Bundle Validation Correctness
// For any bundle, the BPA validation function SHALL accept the bundle if and only if
// its destination is valid, its lifetime is greater than zero, its creation timestamp
// does not exceed the current time, and it is not expired.
// Validates: Requirements 1.1, 1.2, 1.3
func TestProperty_BundleValidationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	bpaAgent := NewBundleProtocolAgent([]EndpointID{
		{Scheme: "dtn", SSP: "local-node"},
	})

	properties.Property("valid bundles are accepted", prop.ForAll(
		func(lifetime int64, createdAt int64) bool {
			currentTime := createdAt + 100 // Current time is after creation

			bundle := &Bundle{
				ID: BundleID{
					SourceEID: EndpointID{
						Scheme: "dtn",
						SSP:    "source",
					},
					CreationTimestamp: createdAt,
					SequenceNumber:    1,
				},
				Destination: EndpointID{
					Scheme: "dtn",
					SSP:    "dest",
				},
				Payload:    []byte("test"),
				Priority:   PriorityNormal,
				Lifetime:   lifetime,
				CreatedAt:  createdAt,
				BundleType: BundleTypeData,
			}

			err := bpaAgent.ValidateBundle(bundle, currentTime)

			// Should be valid if lifetime > 0 and not expired
			shouldBeValid := lifetime > 0 && (createdAt+lifetime > currentTime)

			if shouldBeValid {
				return err == nil
			}
			return err != nil
		},
		gen.Int64Range(1, 3600),    // lifetime (positive)
		gen.Int64Range(1000, 2000), // createdAt
	))

	properties.Property("bundles with invalid destination are rejected", prop.ForAll(
		func(scheme string, ssp string) bool {
			bundle := &Bundle{
				ID: BundleID{
					SourceEID: EndpointID{
						Scheme: "dtn",
						SSP:    "source",
					},
					CreationTimestamp: 1000,
					SequenceNumber:    1,
				},
				Destination: EndpointID{
					Scheme: scheme,
					SSP:    ssp,
				},
				Payload:    []byte("test"),
				Priority:   PriorityNormal,
				Lifetime:   3600,
				CreatedAt:  1000,
				BundleType: BundleTypeData,
			}

			err := bpaAgent.ValidateBundle(bundle, 1100)

			// Should be invalid if scheme or SSP is empty
			if scheme == "" || ssp == "" {
				return err != nil
			}
			return err == nil
		},
		gen.OneConstOf("", "dtn", "ipn"),
		gen.OneConstOf("", "node1", "node2"),
	))

	properties.Property("bundles with zero or negative lifetime are rejected", prop.ForAll(
		func(lifetime int64) bool {
			bundle := &Bundle{
				ID: BundleID{
					SourceEID: EndpointID{
						Scheme: "dtn",
						SSP:    "source",
					},
					CreationTimestamp: 1000,
					SequenceNumber:    1,
				},
				Destination: EndpointID{
					Scheme: "dtn",
					SSP:    "dest",
				},
				Payload:    []byte("test"),
				Priority:   PriorityNormal,
				Lifetime:   lifetime,
				CreatedAt:  1000,
				BundleType: BundleTypeData,
			}

			err := bpaAgent.ValidateBundle(bundle, 1100)

			// Should be invalid if lifetime <= 0
			if lifetime <= 0 {
				return err != nil
			}
			// Should be valid if lifetime > 0 and not expired
			if lifetime > 0 && (1000+lifetime > 1100) {
				return err == nil
			}
			// Should be invalid if expired
			return err != nil
		},
		gen.Int64Range(-100, 100),
	))

	properties.Property("bundles with future creation timestamp are rejected", prop.ForAll(
		func(createdAt int64, currentTime int64) bool {
			bundle := &Bundle{
				ID: BundleID{
					SourceEID: EndpointID{
						Scheme: "dtn",
						SSP:    "source",
					},
					CreationTimestamp: createdAt,
					SequenceNumber:    1,
				},
				Destination: EndpointID{
					Scheme: "dtn",
					SSP:    "dest",
				},
				Payload:    []byte("test"),
				Priority:   PriorityNormal,
				Lifetime:   3600,
				CreatedAt:  createdAt,
				BundleType: BundleTypeData,
			}

			err := bpaAgent.ValidateBundle(bundle, currentTime)

			// Should be invalid if createdAt > currentTime
			if createdAt > currentTime {
				return err != nil
			}
			// Also check if expired
			if createdAt+3600 <= currentTime {
				return err != nil
			}
			return err == nil
		},
		gen.Int64Range(1000, 3000),
		gen.Int64Range(1000, 3000),
	))

	properties.TestingRun(t)
}

// Property 7: Ping Echo Correctness
// For any ping request bundle, exactly one ping response bundle SHALL be generated
// with its destination set to the original sender's endpoint.
// Validates: Requirements 4.1, 4.2
func TestProperty_PingEchoCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	bpaAgent := NewBundleProtocolAgent([]EndpointID{
		{Scheme: "dtn", SSP: "local-node"},
	})

	properties.Property("ping request generates exactly one response", prop.ForAll(
		func(sourceSSP string, destSSP string) bool {
			if sourceSSP == "" || destSSP == "" {
				return true // Skip invalid inputs
			}

			pingRequest := &Bundle{
				ID: BundleID{
					SourceEID: EndpointID{
						Scheme: "dtn",
						SSP:    sourceSSP,
					},
					CreationTimestamp: 1000,
					SequenceNumber:    1,
				},
				Destination: EndpointID{
					Scheme: "dtn",
					SSP:    destSSP,
				},
				Payload:    []byte("PING"),
				Priority:   PriorityExpedited,
				Lifetime:   300,
				CreatedAt:  1000,
				BundleType: BundleTypePingRequest,
			}

			response, err := bpaAgent.HandlePing(pingRequest)
			if err != nil {
				return false
			}

			// Verify response is addressed to original sender
			if response.Destination.Scheme != pingRequest.ID.SourceEID.Scheme {
				return false
			}
			if response.Destination.SSP != pingRequest.ID.SourceEID.SSP {
				return false
			}

			// Verify response is a ping response
			if response.BundleType != BundleTypePingResponse {
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}
