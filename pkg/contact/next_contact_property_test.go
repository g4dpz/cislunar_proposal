package contact

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_NextContactLookupCorrectness validates Property 12:
// For any contact plan, destination node D, and query time T:
// - GetNextContact(D, T) should return the earliest contact with D where StartTime ≥ T
// - If multiple contacts exist with D, return the one with the earliest StartTime
// - If no future contacts exist with D, return an error
// - The returned contact must have StartTime ≥ T
//
// **Validates: Requirement 7.3**
func TestProperty_NextContactLookupCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: GetNextContact returns the earliest future contact with the destination
	properties.Property("next contact returns earliest future contact", prop.ForAll(
		func(numContacts uint8, queryTimeOffset int64, targetNodeIndex uint8) bool {
			if numContacts == 0 || numContacts > 50 {
				return true // Skip invalid inputs
			}

			// Create a contact plan manager
			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Generate arbitrary contact windows with various time ranges
			// Create contacts with different destination nodes
			contacts := make([]ContactWindow, 0)
			nodeNames := []string{"node-A", "node-B", "node-C", "node-D", "node-E"}

			for i := uint8(0); i < numContacts; i++ {
				// Create contacts with varying start/end times relative to currentTime
				startOffset := int64(i) * 100 // Spread contacts across time
				duration := int64(50 + i*10)  // Varying durations

				// Assign to different nodes
				nodeIndex := i % uint8(len(nodeNames))

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID(nodeNames[nodeIndex]),
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

			// Pick a target node to query
			targetNode := NodeID(nodeNames[targetNodeIndex%uint8(len(nodeNames))])

			// Query at a specific time (bounded to be within reasonable range)
			queryTime := currentTime + (queryTimeOffset % 5000)

			// Get next contact from the manager
			nextContact, err := planner.GetNextContact(targetNode, queryTime)

			// Manually compute which contact should be returned
			var expectedContact *ContactWindow
			for i := range contacts {
				contact := &contacts[i]
				if contact.RemoteNode == targetNode && contact.StartTime >= queryTime {
					if expectedContact == nil || contact.StartTime < expectedContact.StartTime {
						expectedContact = contact
					}
				}
			}

			// Verify: If no future contact exists, error should be returned
			if expectedContact == nil {
				if err == nil {
					// Should have returned an error
					return false
				}
				// Correct: no contact available, error returned
				return true
			}

			// Verify: If a future contact exists, it should be returned
			if err != nil {
				// Should not have returned an error
				return false
			}

			// Verify: The returned contact matches the expected contact
			if nextContact.ContactID != expectedContact.ContactID {
				return false
			}

			// Verify: The returned contact has StartTime >= queryTime
			if nextContact.StartTime < queryTime {
				return false
			}

			// Verify: The returned contact is for the correct destination
			if nextContact.RemoteNode != targetNode {
				return false
			}

			return true
		},
		gen.UInt8Range(1, 50),       // numContacts
		gen.Int64Range(-500, 5000),  // queryTimeOffset
		gen.UInt8Range(0, 4),        // targetNodeIndex (0-4 for 5 nodes)
	))

	// Property: GetNextContact returns earliest when multiple future contacts exist
	properties.Property("next contact returns earliest of multiple future contacts", prop.ForAll(
		func(numFutureContacts uint8) bool {
			if numFutureContacts < 2 || numFutureContacts > 20 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()
			targetNode := NodeID("target-node")

			// Create multiple future contacts with the same destination
			// Ensure unique start times by spacing them out
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numFutureContacts; i++ {
				// All contacts are in the future, with varying start times
				// Use larger spacing to ensure uniqueness
				startOffset := int64(100 + int64(i)*200) // Spread out in time
				duration := int64(50)

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: targetNode,
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

			// Query at current time
			nextContact, err := planner.GetNextContact(targetNode, currentTime)

			if err != nil {
				// Should have found a contact
				return false
			}

			// Verify: The returned contact is the earliest one
			// (first contact has the earliest start time by construction)
			if nextContact.ContactID != contacts[0].ContactID {
				return false
			}

			// Verify: All other contacts have later start times
			for i := 1; i < len(contacts); i++ {
				if contacts[i].StartTime < nextContact.StartTime {
					// Found a contact with earlier start time - should have been returned
					return false
				}
			}

			return true
		},
		gen.UInt8Range(2, 20), // numFutureContacts
	))

	// Property: GetNextContact returns error when no future contacts exist
	properties.Property("next contact returns error when no future contacts", prop.ForAll(
		func(numPastContacts uint8) bool {
			if numPastContacts == 0 || numPastContacts > 30 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()
			targetNode := NodeID("target-node")

			// Create contacts that are all in the past
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numPastContacts; i++ {
				// All contacts ended before currentTime
				startOffset := int64(-1000 - int64(i)*100)
				duration := int64(50)

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: targetNode,
					StartTime:  currentTime + startOffset,
					EndTime:    currentTime + startOffset + duration,
					DataRate:   9600,
					LinkType:   LinkTypeUHFTNC,
				})
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime - 5000,
				ValidFrom:   currentTime - 5000,
				ValidTo:     currentTime + 10000,
				Contacts:    contacts,
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at current time - no future contacts should exist
			_, err := planner.GetNextContact(targetNode, currentTime)

			// Verify: Error should be returned
			if err == nil {
				// Should have returned an error (no future contacts)
				return false
			}

			return true
		},
		gen.UInt8Range(1, 30), // numPastContacts
	))

	// Property: GetNextContact returns error for non-existent destination
	properties.Property("next contact returns error for non-existent destination", prop.ForAll(
		func(numContacts uint8) bool {
			if numContacts == 0 || numContacts > 30 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Create contacts with various destinations (but not "non-existent-node")
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numContacts; i++ {
				startOffset := int64(i) * 100
				duration := int64(50)

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: NodeID(fmt.Sprintf("node-%d", i)),
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

			// Query for a node that doesn't exist in the contact plan
			nonExistentNode := NodeID("non-existent-node")
			_, err := planner.GetNextContact(nonExistentNode, currentTime)

			// Verify: Error should be returned
			if err == nil {
				// Should have returned an error (destination not in plan)
				return false
			}

			return true
		},
		gen.UInt8Range(1, 30), // numContacts
	))

	// Property: Boundary condition - contact starting exactly at query time is returned
	properties.Property("boundary: contact starting at query time is returned", prop.ForAll(
		func(duration uint16) bool {
			if duration == 0 {
				return true // Skip invalid duration
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()
			targetNode := NodeID("target-node")

			// Create a contact that starts exactly at currentTime
			contact := ContactWindow{
				ContactID:  1,
				RemoteNode: targetNode,
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
			nextContact, err := planner.GetNextContact(targetNode, currentTime)

			// Verify: Contact should be returned (StartTime >= queryTime, inclusive)
			if err != nil {
				// Should have found the contact
				return false
			}

			if nextContact.ContactID != contact.ContactID {
				return false
			}

			return true
		},
		gen.UInt16Range(1, 1000), // duration
	))

	// Property: Query time just after contact start excludes that contact
	properties.Property("query time after contact start excludes past contact", prop.ForAll(
		func(duration uint16, offset uint16) bool {
			if duration == 0 || offset == 0 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()
			targetNode := NodeID("target-node")

			// Create a contact that starts before the query time
			contact := ContactWindow{
				ContactID:  1,
				RemoteNode: targetNode,
				StartTime:  currentTime - int64(offset), // Started in the past
				EndTime:    currentTime + int64(duration),
				DataRate:   9600,
				LinkType:   LinkTypeUHFTNC,
			}

			plan := &ContactPlan{
				PlanID:      1,
				GeneratedAt: currentTime - 1000,
				ValidFrom:   currentTime - 1000,
				ValidTo:     currentTime + int64(duration) + 100,
				Contacts:    []ContactWindow{contact},
			}

			if err := planner.LoadPlan(plan); err != nil {
				return false
			}

			// Query at currentTime (after contact.StartTime)
			_, err := planner.GetNextContact(targetNode, currentTime)

			// Verify: Contact should NOT be returned (StartTime < queryTime)
			if err == nil {
				// Should have returned an error (contact already started)
				return false
			}

			return true
		},
		gen.UInt16Range(1, 1000), // duration
		gen.UInt16Range(1, 100),  // offset
	))

	// Property: Multiple destinations - correct destination returned
	properties.Property("multiple destinations - correct destination returned", prop.ForAll(
		func(numNodes uint8) bool {
			if numNodes < 2 || numNodes > 10 {
				return true // Skip invalid inputs
			}

			planner := NewContactPlanManager()
			currentTime := time.Now().Unix()

			// Create contacts for multiple different nodes
			contacts := make([]ContactWindow, 0)
			for i := uint8(0); i < numNodes; i++ {
				nodeID := NodeID(fmt.Sprintf("node-%d", i))
				startOffset := int64(i) * 100

				contacts = append(contacts, ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: nodeID,
					StartTime:  currentTime + startOffset,
					EndTime:    currentTime + startOffset + 50,
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

			// Query for each node and verify correct contact is returned
			for i := uint8(0); i < numNodes; i++ {
				targetNode := NodeID(fmt.Sprintf("node-%d", i))
				nextContact, err := planner.GetNextContact(targetNode, currentTime)

				if err != nil {
					// Should have found a contact
					return false
				}

				// Verify: The returned contact is for the correct destination
				if nextContact.RemoteNode != targetNode {
					return false
				}

				// Verify: The returned contact matches the expected contact
				if nextContact.ContactID != uint64(i+1) {
					return false
				}
			}

			return true
		},
		gen.UInt8Range(2, 10), // numNodes
	))

	properties.TestingRun(t)
}
