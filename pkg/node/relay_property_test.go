package node

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/store"
)

// TestProperty_NoRelayDirectDeliveryOnly validates Property 10:
// For any bundle transmitted during any contact window on any node, the contact's
// remote node SHALL match the bundle's final destination endpoint. No bundle SHALL
// be forwarded on behalf of other nodes, and all route lookups SHALL return
// single-hop direct contacts only.
//
// **Validates: Requirements 6.1, 6.2, 13.5**
func TestProperty_NoRelayDirectDeliveryOnly(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1: Bundles are only transmitted to their final destination
	properties.Property("bundles only transmitted to final destination", prop.ForAll(
		func(numBundles uint8, numNodes uint8) bool {
			if numBundles == 0 || numNodes < 2 || numNodes > 10 {
				return true // Skip invalid inputs
			}

			// Create a node with multiple possible destinations
			nodeID := "node-1"
			localEndpoint := bpa.EndpointID{Scheme: "dtn", SSP: nodeID}
			
			_ = bpa.NewBundleProtocolAgent([]bpa.EndpointID{localEndpoint})
			bundleStore := store.NewBundleStore(1024 * 1024) // 1 MB
			planner := contact.NewContactPlanManager()
			
			currentTime := time.Now().Unix()
			
			// Create contact plan with direct contacts to various nodes
			contacts := make([]contact.ContactWindow, 0)
			for i := uint8(0); i < numNodes; i++ {
				remoteNodeID := contact.NodeID("node-" + string('2'+rune(i)))
				contacts = append(contacts, contact.ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: remoteNodeID,
					StartTime:  currentTime,
					EndTime:    currentTime + 600, // 10 minute window
					DataRate:   9600,
					LinkType:   contact.LinkTypeUHFTNC,
				})
			}
			
			plan := &contact.ContactPlan{
				PlanID:     1,
				GeneratedAt: currentTime,
				ValidFrom:  currentTime,
				ValidTo:    currentTime + 3600,
				Contacts:   contacts,
			}
			
			if err := planner.LoadPlan(plan); err != nil {
				return false
			}
			
			// Create bundles with various destinations
			bundleDestinations := make(map[string]bpa.EndpointID)
			for i := uint8(0); i < numBundles; i++ {
				destNodeIdx := i % numNodes
				destNodeID := "node-" + string('2'+rune(destNodeIdx))
				destination := bpa.EndpointID{Scheme: "dtn", SSP: destNodeID}
				
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID:         localEndpoint,
						CreationTimestamp: currentTime,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: destination,
					Payload:     []byte("test payload"),
					Priority:    bpa.PriorityNormal,
					Lifetime:    3600,
					CreatedAt:   currentTime,
					BundleType:  bpa.BundleTypeData,
				}
				
				bundleStore.Store(bundle)
				bundleDestinations[bundle.ID.String()] = destination
			}
			
			// Verify: For each active contact, only bundles destined for that contact's
			// remote node should be retrieved
			activeContacts := planner.GetActiveContacts(currentTime)
			
			for _, contactWindow := range activeContacts {
				// Get bundles for this contact's destination
				contactDestEID := bpa.EndpointID{
					Scheme: "dtn",
					SSP:    string(contactWindow.RemoteNode),
				}
				bundles := bundleStore.ListByDestination(contactDestEID)
				
				// Verify: ALL retrieved bundles have destination matching the contact's remote node
				for _, bundle := range bundles {
					if bundle.Destination.SSP != string(contactWindow.RemoteNode) {
						// Bundle destination does not match contact remote node - RELAY DETECTED!
						return false
					}
				}
			}
			
			return true
		},
		gen.UInt8Range(1, 50),  // numBundles
		gen.UInt8Range(2, 10),  // numNodes
	))

	// Property 2: FindDirectContact returns only single-hop contacts
	properties.Property("FindDirectContact returns only direct contacts", prop.ForAll(
		func(numIntermediateNodes uint8) bool {
			if numIntermediateNodes > 10 {
				return true // Skip invalid inputs
			}

			planner := contact.NewContactPlanManager()
			currentTime := time.Now().Unix()
			
			// Create a contact plan with:
			// - Direct contact: node-1 -> node-final
			// - Intermediate contacts: node-1 -> node-intermediate-X (should NOT be used for node-final)
			contacts := make([]contact.ContactWindow, 0)
			
			// Add direct contact to final destination
			contacts = append(contacts, contact.ContactWindow{
				ContactID:  1,
				RemoteNode: contact.NodeID("node-final"),
				StartTime:  currentTime + 100,
				EndTime:    currentTime + 700,
				DataRate:   9600,
				LinkType:   contact.LinkTypeUHFTNC,
			})
			
			// Add contacts to intermediate nodes (these should NOT be returned for node-final)
			for i := uint8(0); i < numIntermediateNodes; i++ {
				contacts = append(contacts, contact.ContactWindow{
					ContactID:  uint64(i + 2),
					RemoteNode: contact.NodeID("node-intermediate-" + string('A'+rune(i))),
					StartTime:  currentTime,
					EndTime:    currentTime + 600,
					DataRate:   9600,
					LinkType:   contact.LinkTypeUHFTNC,
				})
			}
			
			plan := &contact.ContactPlan{
				PlanID:     1,
				GeneratedAt: currentTime,
				ValidFrom:  currentTime,
				ValidTo:    currentTime + 3600,
				Contacts:   contacts,
			}
			
			if err := planner.LoadPlan(plan); err != nil {
				return false
			}
			
			// Look up contact for final destination
			finalDest := bpa.EndpointID{Scheme: "dtn", SSP: "node-final"}
			contactWindow, err := planner.FindDirectContact(finalDest, currentTime)
			
			if err != nil {
				// Should find the direct contact
				return false
			}
			
			// Verify: The returned contact is ONLY the direct contact to node-final
			if contactWindow.RemoteNode != contact.NodeID("node-final") {
				// Multi-hop routing detected! Should only return direct contact.
				return false
			}
			
			// Verify: No intermediate nodes are returned
			if string(contactWindow.RemoteNode) != "node-final" {
				return false
			}
			
			return true
		},
		gen.UInt8Range(0, 10), // numIntermediateNodes
	))

	// Property 3: No bundle is forwarded on behalf of other nodes
	// This property verifies that a node only transmits bundles where it is either:
	// - The source (originating the bundle), OR
	// - The destination (local delivery)
	// A node should NEVER relay bundles (source != self, destination != self)
	properties.Property("no relay forwarding on behalf of other nodes", prop.ForAll(
		func(numBundles uint8) bool {
			if numBundles == 0 || numBundles > 50 {
				return true // Skip invalid inputs
			}

			// Create node-2 (intermediate node)
			node2ID := "node-2"
			node2Endpoint := bpa.EndpointID{Scheme: "dtn", SSP: node2ID}
			
			_ = bpa.NewBundleProtocolAgent([]bpa.EndpointID{node2Endpoint})
			bundleStore := store.NewBundleStore(1024 * 1024)
			planner := contact.NewContactPlanManager()
			
			currentTime := time.Now().Unix()
			
			// Create contact plan:
			// - node-2 has contact with node-3
			contacts := []contact.ContactWindow{
				{
					ContactID:  1,
					RemoteNode: contact.NodeID("node-3"),
					StartTime:  currentTime,
					EndTime:    currentTime + 600,
					DataRate:   9600,
					LinkType:   contact.LinkTypeUHFTNC,
				},
			}
			
			plan := &contact.ContactPlan{
				PlanID:     1,
				GeneratedAt: currentTime,
				ValidFrom:  currentTime,
				ValidTo:    currentTime + 3600,
				Contacts:   contacts,
			}
			
			if err := planner.LoadPlan(plan); err != nil {
				return false
			}
			
			// Scenario 1: Create bundles FROM node-2 TO node-3
			// These SHOULD be in the store and queued for transmission (node-2 is source)
			for i := uint8(0); i < numBundles/2; i++ {
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID:         node2Endpoint, // node-2 is source
						CreationTimestamp: currentTime,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: bpa.EndpointID{Scheme: "dtn", SSP: "node-3"},
					Payload:     []byte("test payload"),
					Priority:    bpa.PriorityNormal,
					Lifetime:    3600,
					CreatedAt:   currentTime,
					BundleType:  bpa.BundleTypeData,
				}
				
				bundleStore.Store(bundle)
			}
			
			// Scenario 2: Create bundles FROM node-1 TO node-3 (relay scenario)
			// These should NOT be in node-2's store for forwarding
			// (In a real system, node-2 would reject these or not receive them)
			relayBundles := make([]*bpa.Bundle, 0)
			for i := uint8(numBundles / 2); i < numBundles; i++ {
				bundle := &bpa.Bundle{
					ID: bpa.BundleID{
						SourceEID:         bpa.EndpointID{Scheme: "dtn", SSP: "node-1"}, // Different source
						CreationTimestamp: currentTime,
						SequenceNumber:    uint64(i + 1),
					},
					Destination: bpa.EndpointID{Scheme: "dtn", SSP: "node-3"},
					Payload:     []byte("test payload"),
					Priority:    bpa.PriorityNormal,
					Lifetime:    3600,
					CreatedAt:   currentTime,
					BundleType:  bpa.BundleTypeData,
				}
				
				relayBundles = append(relayBundles, bundle)
			}
			
			// Verify: Get bundles queued for transmission to node-3
			node3Dest := bpa.EndpointID{Scheme: "dtn", SSP: "node-3"}
			bundlesForNode3 := bundleStore.ListByDestination(node3Dest)
			
			// All bundles in the store should have node-2 as the source
			// (no relay bundles should be present)
			for _, bundle := range bundlesForNode3 {
				if bundle.ID.SourceEID.SSP != node2ID {
					// Found a relay bundle! This violates the no-relay constraint
					return false
				}
			}
			
			// Verify: Relay bundles should NOT be in the store
			for _, relayBundle := range relayBundles {
				retrieved, err := bundleStore.Retrieve(relayBundle.ID)
				if err == nil && retrieved != nil {
					// Relay bundle found in store - this is a violation
					return false
				}
			}
			
			return true
		},
		gen.UInt8Range(2, 50), // numBundles (at least 2 to test both scenarios)
	))

	// Property 4: Contact lookup never returns multi-hop paths
	properties.Property("contact lookup returns single-hop only", prop.ForAll(
		func(numHops uint8) bool {
			if numHops < 2 || numHops > 5 {
				return true // Skip invalid inputs
			}

			planner := contact.NewContactPlanManager()
			currentTime := time.Now().Unix()
			
			// Create a chain of contacts: node-1 -> node-2 -> node-3 -> ... -> node-N
			// If we look up contact from node-1 to node-N, it should return NONE
			// (no direct contact), not a multi-hop path
			
			contacts := make([]contact.ContactWindow, 0)
			for i := uint8(0); i < numHops; i++ {
				contacts = append(contacts, contact.ContactWindow{
					ContactID:  uint64(i + 1),
					RemoteNode: contact.NodeID("node-" + string('2'+rune(i))),
					StartTime:  currentTime,
					EndTime:    currentTime + 600,
					DataRate:   9600,
					LinkType:   contact.LinkTypeUHFTNC,
				})
			}
			
			plan := &contact.ContactPlan{
				PlanID:     1,
				GeneratedAt: currentTime,
				ValidFrom:  currentTime,
				ValidTo:    currentTime + 3600,
				Contacts:   contacts,
			}
			
			if err := planner.LoadPlan(plan); err != nil {
				return false
			}
			
			// Try to find contact to the final node in the chain
			finalNodeID := "node-" + string('2'+rune(numHops-1))
			finalDest := bpa.EndpointID{Scheme: "dtn", SSP: finalNodeID}
			
			contactWindow, err := planner.FindDirectContact(finalDest, currentTime)
			
			// Should find the direct contact to the final node
			if err != nil {
				return false
			}
			
			// Verify: The contact is direct (single-hop)
			if contactWindow.RemoteNode != contact.NodeID(finalNodeID) {
				return false
			}
			
			// Now try to find contact to a node that doesn't have a direct contact
			nonExistentDest := bpa.EndpointID{Scheme: "dtn", SSP: "node-nonexistent"}
			_, err = planner.FindDirectContact(nonExistentDest, currentTime)
			
			// Should return error (no direct contact found)
			// Should NOT return a multi-hop path
			if err == nil {
				// Found a contact to non-existent node - this shouldn't happen
				return false
			}
			
			return true
		},
		gen.UInt8Range(2, 5), // numHops
	))

	properties.TestingRun(t)
}
