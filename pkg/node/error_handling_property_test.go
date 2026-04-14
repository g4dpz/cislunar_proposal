package node

import (
	"testing"
	"time"

	"pgregory.net/rapid"
	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/store"
)

// Property 18: No Transmission After Window End
// **Validates: Requirement 9.2**
// For any contact window and transmission attempt, no transmission SHALL occur
// after the contact window's end time has been reached.

func TestProperty_NoTransmissionAfterWindowEnd(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contact window
		startTime := rapid.Int64Range(1000000, 2000000).Draw(t, "startTime")
		duration := rapid.Int64Range(60, 600).Draw(t, "duration")
		endTime := startTime + duration

		contactWindow := contact.ContactWindow{
			ContactID:  1,
			RemoteNode: "test-node",
			StartTime:  startTime,
			EndTime:    endTime,
			DataRate:   9600,
			LinkType:   contact.LinkTypeUHFIQ,
		}

		// Create test components
		endpoints := []bpa.EndpointID{{Scheme: "dtn", SSP: "//test-node"}}
		bpaAgent := bpa.NewBundleProtocolAgent(endpoints)
		bundleStore := store.NewBundleStore(1024 * 1024)
		planManager := contact.NewContactPlanManager()
		mockCLA := cla.NewMockCLA()

		config := NodeConfig{
			NodeID:          "test-controller",
			NodeType:        NodeTypeTerrestrial,
			Endpoints:       endpoints,
			MaxStorageBytes: 1024 * 1024,
		}

		controller := NewNodeController(config, bpaAgent, bundleStore, planManager, mockCLA)

		// Create bundles destined for the remote node
		numBundles := rapid.IntRange(5, 20).Draw(t, "numBundles")
		for i := 0; i < numBundles; i++ {
			bundle, _ := bpaAgent.CreateBundle(
				endpoints[0],
				bpa.EndpointID{Scheme: "dtn", SSP: "test-node"},
				[]byte("test payload"),
				bpa.PriorityNormal,
				300,
			)
			bundleStore.Store(bundle)
		}

		// Execute contact window at a time AFTER the window has ended
		currentTime := endTime + 10

		// Property: executeContactWindow should not transmit any bundles
		// because currentTime >= endTime
		err := controller.executeContactWindow(contactWindow, currentTime)

		// The function should complete without error
		if err != nil {
			t.Fatalf("executeContactWindow failed: %v", err)
		}

		// Property: no bundles should have been sent (all should remain in store)
		remainingBundles := bundleStore.Count()
		if remainingBundles != numBundles {
			t.Fatalf("bundles transmitted after window end: expected %d remaining, got %d",
				numBundles, remainingBundles)
		}
	})
}

// Property 19: Missed Contact Retains Bundles
// **Validates: Requirement 9.4**
// For any scheduled contact window where the CLA fails to establish a link,
// all bundles queued for that contact's destination SHALL remain in the Bundle_Store,
// and the contacts-missed counter SHALL be incremented.

func TestProperty_MissedContactRetainsBundles(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contact window
		startTime := rapid.Int64Range(1000000, 2000000).Draw(t, "startTime")
		duration := rapid.Int64Range(60, 600).Draw(t, "duration")

		contactWindow := contact.ContactWindow{
			ContactID:  1,
			RemoteNode: "test-node",
			StartTime:  startTime,
			EndTime:    startTime + duration,
			DataRate:   9600,
			LinkType:   contact.LinkTypeUHFIQ,
		}

		// Create test components
		endpoints := []bpa.EndpointID{{Scheme: "dtn", SSP: "//test-node"}}
		bpaAgent := bpa.NewBundleProtocolAgent(endpoints)
		bundleStore := store.NewBundleStore(1024 * 1024)
		planManager := contact.NewContactPlanManager()
		
		// Create a mock CLA that fails to open
		mockCLA := cla.NewMockCLA()
		mockCLA.SetOpenError("link establishment failed")

		config := NodeConfig{
			NodeID:          "test-controller",
			NodeType:        NodeTypeTerrestrial,
			Endpoints:       endpoints,
			MaxStorageBytes: 1024 * 1024,
		}

		controller := NewNodeController(config, bpaAgent, bundleStore, planManager, mockCLA)

		// Create bundles destined for the remote node
		numBundles := rapid.IntRange(1, 10).Draw(t, "numBundles")
		for i := 0; i < numBundles; i++ {
			bundle, _ := bpaAgent.CreateBundle(
				endpoints[0],
				bpa.EndpointID{Scheme: "dtn", SSP: "test-node"},
				[]byte("test payload"),
				bpa.PriorityNormal,
				300,
			)
			bundleStore.Store(bundle)
		}

		initialMissed := controller.GetStatistics().ContactsMissed

		// Execute contact window (should fail to open link)
		currentTime := startTime + 10
		err := controller.executeContactWindow(contactWindow, currentTime)

		// Property: executeContactWindow should return an error
		if err == nil {
			t.Fatalf("expected error when link fails to open")
		}

		// Property: all bundles should remain in store
		remainingBundles := bundleStore.Count()
		if remainingBundles != numBundles {
			t.Fatalf("bundles lost on missed contact: expected %d, got %d", numBundles, remainingBundles)
		}

		// Property: contacts-missed counter should be incremented
		finalMissed := controller.GetStatistics().ContactsMissed
		if finalMissed != initialMissed+1 {
			t.Fatalf("contacts-missed not incremented: expected %d, got %d", initialMissed+1, finalMissed)
		}
	})
}

// Property 25: Bundles Retained When No Contact Available
// **Validates: Requirements 17.5, 5.5**
// For any bundle whose destination has no direct contact window in the current contact plan,
// the Bundle_Store SHALL retain the bundle until the contact plan is updated or the bundle's lifetime expires.

func TestProperty_BundlesRetainedWhenNoContactAvailable(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Create test components
		endpoints := []bpa.EndpointID{{Scheme: "dtn", SSP: "//local-node"}}
		bpaAgent := bpa.NewBundleProtocolAgent(endpoints)
		bundleStore := store.NewBundleStore(1024 * 1024)
		planManager := contact.NewContactPlanManager()
		mockCLA := cla.NewMockCLA()

		// Load an empty contact plan (no contacts available)
		emptyPlan := &contact.ContactPlan{
			PlanID:      1,
			GeneratedAt: time.Now().Unix(),
			ValidFrom:   time.Now().Unix(),
			ValidTo:     time.Now().Unix() + 86400,
			Contacts:    []contact.ContactWindow{},
		}
		planManager.LoadPlan(emptyPlan)

		config := NodeConfig{
			NodeID:          "test-controller",
			NodeType:        NodeTypeTerrestrial,
			Endpoints:       endpoints,
			MaxStorageBytes: 1024 * 1024,
		}

		controller := NewNodeController(config, bpaAgent, bundleStore, planManager, mockCLA)

		// Create a bundle destined for a remote node (no contact available)
		remoteEndpoint := bpa.EndpointID{Scheme: "dtn", SSP: "remote-node"}
		bundle, _ := bpaAgent.CreateBundle(
			endpoints[0],
			remoteEndpoint,
			[]byte("test payload"),
			bpa.PriorityNormal,
			300,
		)

		currentTime := time.Now().Unix()

		// Process the incoming bundle
		err := controller.processIncomingBundle(bundle, currentTime)

		// Property: bundle should be stored even though no contact is available
		if err != nil {
			t.Fatalf("failed to process bundle with no contact available: %v", err)
		}

		// Property: bundle should be in the store
		storedBundle, err := bundleStore.Retrieve(bundle.ID)
		if err != nil {
			t.Fatalf("bundle not retained when no contact available: %v", err)
		}

		if storedBundle.ID != bundle.ID {
			t.Fatalf("wrong bundle retrieved: expected %v, got %v", bundle.ID, storedBundle.ID)
		}

		// Property: bundle should remain in store until contact plan is updated or lifetime expires
		// Simulate time passing but not exceeding lifetime
		futureTime := currentTime + 100 // Still within 300s lifetime

		// Run a cycle - bundle should still be there
		controller.RunCycle(futureTime)

		_, err = bundleStore.Retrieve(bundle.ID)
		if err != nil {
			t.Fatalf("bundle evicted before lifetime expired: %v", err)
		}
	})
}
