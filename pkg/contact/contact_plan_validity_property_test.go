package contact

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_ContactPlanValidityInvariants validates Property 13:
// For any valid contact plan, all contact windows SHALL fall within the plan's
// valid-from and valid-to time range, and no two contacts on the same link for
// a given node SHALL overlap in time.
//
// **Validates: Requirements 7.4, 7.5**
func TestProperty_ContactPlanValidityInvariants(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: All contacts fall within valid time range
	properties.Property("all contacts fall within plan valid time range", prop.ForAll(
		func(numContacts uint8, validDuration int64) bool {
			if numContacts == 0 || numContacts > 50 {
				return true // Skip invalid inputs
			}
			if validDuration <= 0 || validDuration > 86400 {
				return true // Skip invalid durations
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime
			validTo := currentTime + validDuration

			// Generate contacts within the valid range
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numContacts; i++ {
				// Ensure contact falls within valid range
				startOffset := int64(i) * (validDuration / int64(numContacts+1))
				duration := int64(100)

				// Ensure end time doesn't exceed validTo
				startTime := validFrom + startOffset
				endTime := startTime + duration
				if endTime > validTo {
					endTime = validTo
				}
				if startTime >= endTime {
					continue // Skip invalid contact
				}

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID("node-" + string(rune('A'+i%10))),
					StartTime:  startTime,
					EndTime:    endTime,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			// Create contact plan
			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    contacts,
			}

			// Validate the plan
			err := plan.Validate()
			if err != nil {
				// Plan should be valid
				return false
			}

			// Verify all contacts fall within valid range
			for _, contact := range plan.Contacts {
				if contact.StartTime < plan.ValidFrom || contact.EndTime > plan.ValidTo {
					return false
				}
			}

			return true
		},
		gen.UInt8Range(1, 50),
		gen.Int64Range(1000, 86400),
	))

	// Property: No overlapping contacts for the same node
	properties.Property("no overlapping contacts for same node", prop.ForAll(
		func(numContacts uint8) bool {
			if numContacts < 2 || numContacts > 20 {
				return true // Skip invalid inputs
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime
			validTo := currentTime + 10000

			// Generate non-overlapping contacts for the same node
			targetNode := NodeID("target-node")
			contacts := make([]ContactWindow, 0)

			for i := uint8(0); i < numContacts; i++ {
				// Space contacts apart to ensure no overlap
				startTime := validFrom + int64(i)*200
				endTime := startTime + 100

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: targetNode,
					StartTime:  startTime,
					EndTime:    endTime,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			// Create contact plan
			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    contacts,
			}

			// Validate the plan - should succeed (no overlaps)
			err := plan.Validate()
			if err != nil {
				return false
			}

			// Manually verify no overlaps
			for i := 0; i < len(contacts); i++ {
				for j := i + 1; j < len(contacts); j++ {
					c1, c2 := contacts[i], contacts[j]
					if c1.RemoteNode == c2.RemoteNode {
						// Check for overlap: c1.start < c2.end && c2.start < c1.end
						if c1.StartTime < c2.EndTime && c2.StartTime < c1.EndTime {
							return false // Overlap detected
						}
					}
				}
			}

			return true
		},
		gen.UInt8Range(2, 20),
	))

	// Property: Overlapping contacts for same node are rejected
	properties.Property("overlapping contacts for same node are rejected", prop.ForAll(
		func(offset int64) bool {
			if offset <= 0 || offset >= 100 {
				return true // Skip invalid offsets
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime
			validTo := currentTime + 10000

			targetNode := NodeID("target-node")

			// Create two overlapping contacts
			contact1 := ContactWindow{
				ContactID:  1,
				RemoteNode: targetNode,
				StartTime:  validFrom + 100,
				EndTime:    validFrom + 300,
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			// Second contact overlaps with first
			contact2 := ContactWindow{
				ContactID:  2,
				RemoteNode: targetNode,
				StartTime:  validFrom + 100 + offset, // Starts during contact1
				EndTime:    validFrom + 400,
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			// Create contact plan with overlapping contacts
			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    []ContactWindow{contact1, contact2},
			}

			// Validate the plan - should fail due to overlap
			err := plan.Validate()
			if err == nil {
				// Should have detected overlap
				return false
			}

			return true
		},
		gen.Int64Range(1, 99),
	))

	// Property: Contacts outside valid range are rejected
	properties.Property("contacts outside valid range are rejected", prop.ForAll(
		func(outsideOffset int64) bool {
			if outsideOffset <= 0 || outsideOffset > 1000 {
				return true // Skip invalid offsets
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime + 1000
			validTo := currentTime + 5000

			// Create contact that starts before validFrom
			contactBefore := ContactWindow{
				ContactID:  1,
				RemoteNode: NodeID("node-a"),
				StartTime:  validFrom - outsideOffset, // Before validFrom
				EndTime:    validFrom + 100,
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			plan1 := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    []ContactWindow{contactBefore},
			}

			// Should fail validation
			if plan1.Validate() == nil {
				return false
			}

			// Create contact that ends after validTo
			contactAfter := ContactWindow{
				ContactID:  2,
				RemoteNode: NodeID("node-b"),
				StartTime:  validTo - 100,
				EndTime:    validTo + outsideOffset, // After validTo
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			plan2 := &ContactPlan{
				PlanID:      2,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    []ContactWindow{contactAfter},
			}

			// Should fail validation
			if plan2.Validate() == nil {
				return false
			}

			return true
		},
		gen.Int64Range(1, 1000),
	))

	// Property: ValidFrom must be less than ValidTo
	properties.Property("validFrom must be less than validTo", prop.ForAll(
		func(duration int64) bool {
			if duration <= 0 {
				return true // Skip invalid durations
			}

			currentTime := time.Now().Unix()

			// Create plan with validFrom >= validTo
			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   currentTime + 1000,
				ValidTo:     currentTime + 1000, // Equal to validFrom
				Contacts:    []ContactWindow{},
			}

			// Should fail validation
			if plan.Validate() == nil {
				return false
			}

			// Create plan with validFrom > validTo
			plan2 := &ContactPlan{
				PlanID:      2,
				GeneratedAt: currentTime,
				ValidFrom:   currentTime + 2000,
				ValidTo:     currentTime + 1000, // Less than validFrom
				Contacts:    []ContactWindow{},
			}

			// Should fail validation
			if plan2.Validate() == nil {
				return false
			}

			return true
		},
		gen.Int64Range(1, 10000),
	))

	// Property: Contacts with different nodes can overlap
	properties.Property("contacts with different nodes can overlap", prop.ForAll(
		func(numNodes uint8) bool {
			if numNodes < 2 || numNodes > 10 {
				return true // Skip invalid inputs
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime
			validTo := currentTime + 10000

			// Create overlapping contacts for different nodes
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numNodes; i++ {
				// All contacts overlap in time but have different remote nodes
				contact := ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID("node-" + string(rune('A'+i))),
					StartTime:  validFrom + 100, // Same start time
					EndTime:    validFrom + 300, // Same end time
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				}
				contacts = append(contacts, contact)
			}

			// Create contact plan
			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    contacts,
			}

			// Validate the plan - should succeed (different nodes can overlap)
			err := plan.Validate()
			if err != nil {
				return false
			}

			return true
		},
		gen.UInt8Range(2, 10),
	))

	// Property: Contact with invalid time range is rejected
	properties.Property("contact with startTime >= endTime is rejected", prop.ForAll(
		func(duration int64) bool {
			if duration < 0 {
				return true // Skip invalid durations
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime
			validTo := currentTime + 10000

			// Create contact with startTime >= endTime
			contact := ContactWindow{
				ContactID:  1,
				RemoteNode: NodeID("node-a"),
				StartTime:  validFrom + 1000,
				EndTime:    validFrom + 1000 - duration, // endTime <= startTime
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    []ContactWindow{contact},
			}

			// Should fail validation
			if plan.Validate() == nil {
				return false
			}

			return true
		},
		gen.Int64Range(0, 1000),
	))

	// Property: Contact with invalid data rate is rejected
	properties.Property("contact with dataRate <= 0 is rejected", prop.ForAll(
		func(dataRate int64) bool {
			if dataRate > 0 {
				return true // Skip valid data rates
			}

			currentTime := time.Now().Unix()
			validFrom := currentTime
			validTo := currentTime + 10000

			// Create contact with invalid data rate
			contact := ContactWindow{
				ContactID:  1,
				RemoteNode: NodeID("node-a"),
				StartTime:  validFrom + 100,
				EndTime:    validFrom + 300,
				DataRate:   dataRate, // Invalid: <= 0
				LinkType:   LinkTypeUHFTNC,
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   validFrom,
				ValidTo:     validTo,
				Contacts:    []ContactWindow{contact},
			}

			// Should fail validation
			if plan.Validate() == nil {
				return false
			}

			return true
		},
		gen.Int64Range(-1000, 0),
	))

	properties.TestingRun(t)
}
