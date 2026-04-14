package contact

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_ActiveContactsQueryCorrectness validates Property 11:
// For any contact plan and query time T, GetActiveContacts(T) SHALL return
// ALL and ONLY contacts where StartTime ≤ T < EndTime. No contacts outside
// this time range should be returned, and all contacts within this time range
// should be returned.
//
// **Validates: Requirement 7.2**
func TestProperty_ActiveContactsQueryCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: GetActiveContacts returns exactly the contacts that are active at query time
	properties.Property("active contacts query returns all and only active contacts", prop.ForAll(
		func(numContacts uint8, queryTimeOffset int64) bool {
			if numContacts == 0 || numContacts > 50 {
				return true // Skip invalid inputs
			}

			// Create a contact plan manager
			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Generate arbitrary contact windows with various time ranges
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numContacts; i++ {
				// Create contacts with varying start/end times relative to currentTime
				// Some before, some during, some after the query time
				startOffset := int64(i) * 100 // Spread contacts across time
				duration := int64(50 + i*10)  // Varying durations

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID("node-" + string('A'+rune(i%10))),
					StartTime:  currentTime + startOffset,
					EndTime:    currentTime + startOffset + duration,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			// Create and load the contact plan
			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   currentTime - 1000,
				ValidTo:     currentTime + 10000,
				Contacts:    contacts,
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at a specific time (bounded to be within reasonable range)
			queryTime := currentTime + (queryTimeOffset % 5000)

			// Get active contacts from the manager
			activeContacts := planner.GetActiveContacts(queryTime)

			// Manually compute which contacts should be active
			expectedActive := make(map[uint64]bool)
			for _, contact := range contacts {
				if contact.IsActive(queryTime) {
					expectedActive[contact.ContactID] = true
				}
			}

			// Verify: All returned contacts are actually active
			for _, activeContact := range activeContacts {
				if !activeContact.IsActive(queryTime) {
					// False positive: returned contact is not actually active
					return false
				}
				// Mark as found
				delete(expectedActive, activeContact.ContactID)
			}

			// Verify: No active contacts were missed (all expected contacts were returned)
			if len(expectedActive) > 0 {
				// False negative: some active contacts were not returned
				return false
			}

			return true
		},
		gen.UInt8Range(1, 50),      // numContacts
		gen.Int64Range(-500, 5000), // queryTimeOffset
	))

	// Property: Active contacts satisfy the time range constraint
	properties.Property("all active contacts satisfy StartTime <= T < EndTime", prop.ForAll(
		func(numContacts uint8) bool {
			if numContacts == 0 || numContacts > 50 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Generate contacts
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numContacts; i++ {
				startOffset := int64(i) * 100
				duration := int64(50 + i*10)

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID("node-" + string('A'+rune(i%10))),
					StartTime:  currentTime + startOffset,
					EndTime:    currentTime + startOffset + duration,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   currentTime - 1000,
				ValidTo:     currentTime + 10000,
				Contacts:    contacts,
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Test at multiple query times
			for offset := int64(-100); offset < 5000; offset += 100 {
				queryTime := currentTime + offset
				activeContacts := planner.GetActiveContacts(queryTime)

				// Verify each returned contact satisfies the time constraint
				for _, contact := range activeContacts {
					if !(contact.StartTime <= queryTime && queryTime < contact.EndTime) {
						// Contact does not satisfy the active time constraint
						return false
					}
				}
			}

			return true
		},
		gen.UInt8Range(1, 50), // numContacts
	))

	// Property: No false positives - contacts outside time range are never returned
	properties.Property("no false positives - inactive contacts never returned", prop.ForAll(
		func(numContacts uint8) bool {
			if numContacts == 0 || numContacts > 50 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Generate contacts that are clearly not active at currentTime
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numContacts; i++ {
				// Create contacts in the past (ended before currentTime)
				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID("node-past-" + string('A'+rune(i%10))),
					StartTime:  currentTime - 1000 - int64(i)*100,
					EndTime:    currentTime - 500 - int64(i)*100,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})

				// Create contacts in the future (start after currentTime)
				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(numContacts + i + 1),
					RemoteNode: NodeID("node-future-" + string('A'+rune(i%10))),
					StartTime:  currentTime + 1000 + int64(i)*100,
					EndTime:    currentTime + 2000 + int64(i)*100,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   currentTime - 5000,
				ValidTo:     currentTime + 10000,
				Contacts:    contacts,
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at currentTime - no contacts should be active
			activeContacts := planner.GetActiveContacts(currentTime)

			// Verify: No contacts are returned (all are inactive)
			if len(activeContacts) > 0 {
				// False positive detected
				return false
			}

			return true
		},
		gen.UInt8Range(1, 25), // numContacts (doubled in test, so max 50)
	))

	// Property: No false negatives - all active contacts are returned
	properties.Property("no false negatives - all active contacts returned", prop.ForAll(
		func(numContacts uint8) bool {
			if numContacts == 0 || numContacts > 50 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Generate contacts that are ALL active at currentTime
			// Each contact must have a unique remote node to avoid overlap validation errors
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numContacts; i++ {
				// All contacts span currentTime, each with unique remote node
				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID(fmt.Sprintf("node-%d", i)), // Unique node per contact
					StartTime:  currentTime - 100, // Started before currentTime
					EndTime:    currentTime + 100, // Ends after currentTime
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime,
				ValidFrom:   currentTime - 1000,
				ValidTo:     currentTime + 1000,
				Contacts:    contacts,
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at currentTime - all contacts should be active
			activeContacts := planner.GetActiveContacts(currentTime)

			// Verify: All contacts are returned
			if len(activeContacts) != int(numContacts) {
				// Some active contacts were not returned (false negative)
				return false
			}

			// Verify: Each contact is present
			foundContacts := make(map[uint64]bool)
			for _, contact := range activeContacts {
				foundContacts[contact.ContactID] = true
			}

			for _, contact := range contacts {
				if !foundContacts[contact.ContactID] {
					// Contact not found (false negative)
					return false
				}
			}

			return true
		},
		gen.UInt8Range(1, 50), // numContacts
	))

	// Property: Boundary condition - contact ending exactly at query time is not active
	properties.Property("boundary: contact ending at T is not active", prop.ForAll(
		func(duration uint16) bool {
			if duration == 0 {
				return true // Skip invalid duration
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Create a contact that ends exactly at currentTime
			contact := ContactWindow{
				ContactID:  1,
				RemoteNode: NodeID("node-boundary"),
				StartTime:  currentTime - int64(duration),
				EndTime:    currentTime, // Ends exactly at query time
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime - int64(duration) - 100,
				ValidFrom:   currentTime - int64(duration) - 100,
				ValidTo:     currentTime + 1000,
				Contacts:    []ContactWindow{contact},
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at currentTime (contact.EndTime)
			activeContacts := planner.GetActiveContacts(currentTime)

			// Verify: Contact is NOT active (EndTime is exclusive)
			if len(activeContacts) > 0 {
				// Contact should not be active at its EndTime
				return false
			}

			return true
		},
		gen.UInt16Range(1, 1000), // duration
	))

	// Property: Boundary condition - contact starting exactly at query time is active
	properties.Property("boundary: contact starting at T is active", prop.ForAll(
		func(duration uint16) bool {
			if duration == 0 {
				return true // Skip invalid duration
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Create a contact that starts exactly at currentTime
			contact := ContactWindow{
				ContactID:  1,
				RemoteNode: NodeID("node-boundary"),
				StartTime:  currentTime, // Starts exactly at query time
				EndTime:    currentTime + int64(duration),
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime - 100,
				ValidFrom:   currentTime - 100,
				ValidTo:     currentTime + int64(duration) + 100,
				Contacts:    []ContactWindow{contact},
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at currentTime (contact.StartTime)
			activeContacts := planner.GetActiveContacts(currentTime)

			// Verify: Contact IS active (StartTime is inclusive)
			if len(activeContacts) != 1 {
				// Contact should be active at its StartTime
				return false
			}

			if activeContacts[0].ContactID != contact.ContactID {
				return false
			}

			return true
		},
		gen.UInt16Range(1, 1000), // duration
	))

	properties.TestingRun(t)
}
