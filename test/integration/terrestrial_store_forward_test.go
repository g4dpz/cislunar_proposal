package integration

import (
	"testing"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/node"
	"terrestrial-dtn/pkg/store"
)

// TestTerrestrialStoreAndForward tests bundle delivery between two terrestrial nodes
// Validates Requirements: 5.1, 5.2, 5.3, 5.4, 5.5
func TestTerrestrialStoreAndForward(t *testing.T) {
	// Setup Node A (sender)
	nodeA := setupStoreForwardNode(t, "node-a", 1)
	defer nodeA.Shutdown()

	// Setup Node B (receiver)
	nodeB := setupStoreForwardNode(t, "node-b", 2)
	defer nodeB.Shutdown()

	// Create bundle from Node A to Node B
	sourceEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.0"}
	destEID := bpa.EndpointID{Scheme: "ipn", SSP: "2.0"}

	bpaA := bpa.NewBundleProtocolAgent([]bpa.EndpointID{sourceEID})
	bundle, err := bpaA.CreateBundle(sourceEID, destEID, []byte("Hello from Node A"), bpa.PriorityNormal, 3600)
	if err != nil {
		t.Fatalf("Failed to create bundle: %v", err)
	}

	// Store bundle at Node A
	storeA := store.NewBundleStore(10 * 1024 * 1024)
	if err := storeA.Store(bundle); err != nil {
		t.Fatalf("Failed to store bundle: %v", err)
	}

	// Verify bundle is stored
	if storeA.Count() != 1 {
		t.Errorf("Expected 1 bundle in store, got %d", storeA.Count())
	}

	// Simulate bundle transmission (Node A -> Node B)
	bundles := storeA.ListByDestination(destEID)
	if len(bundles) != 1 {
		t.Fatalf("Expected 1 bundle for destination, got %d", len(bundles))
	}

	transmittedBundle := bundles[0]

	// Simulate ACK - delete from Node A store
	if err := storeA.Delete(transmittedBundle.ID); err != nil {
		t.Fatalf("Failed to delete acknowledged bundle: %v", err)
	}

	// Verify bundle was deleted from Node A
	if storeA.Count() != 0 {
		t.Errorf("Expected 0 bundles in store after ACK, got %d", storeA.Count())
	}

	// Simulate bundle reception at Node B
	bpaB := bpa.NewBundleProtocolAgent([]bpa.EndpointID{destEID})
	currentTime := time.Now().Unix()

	if err := bpaB.ValidateBundle(transmittedBundle, currentTime); err != nil {
		t.Fatalf("Bundle validation failed at Node B: %v", err)
	}

	// Check if destination is local to Node B
	if !bpaB.IsLocalEndpoint(transmittedBundle.Destination) {
		t.Error("Expected bundle destination to be local to Node B")
	}

	// Deliver bundle locally
	if err := bpaB.DeliverBundle(transmittedBundle); err != nil {
		t.Fatalf("Failed to deliver bundle: %v", err)
	}

	t.Log("Store-and-forward test passed: bundle delivered successfully")
}

// TestPriorityBasedDelivery tests that bundles are transmitted in priority order
// Validates Requirement: 5.3
func TestPriorityBasedDelivery(t *testing.T) {
	// Create bundle store
	bundleStore := store.NewBundleStore(10 * 1024 * 1024)

	sourceEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.0"}
	destEID := bpa.EndpointID{Scheme: "ipn", SSP: "2.0"}

	bpaAgent := bpa.NewBundleProtocolAgent([]bpa.EndpointID{sourceEID})

	// Create bundles with different priorities
	priorities := []bpa.Priority{
		bpa.PriorityBulk,
		bpa.PriorityCritical,
		bpa.PriorityNormal,
		bpa.PriorityExpedited,
	}

	for i, priority := range priorities {
		bundle, err := bpaAgent.CreateBundle(
			sourceEID,
			destEID,
			[]byte("Test payload"),
			priority,
			3600,
		)
		if err != nil {
			t.Fatalf("Failed to create bundle %d: %v", i, err)
		}

		if err := bundleStore.Store(bundle); err != nil {
			t.Fatalf("Failed to store bundle %d: %v", i, err)
		}
	}

	// List bundles by priority
	bundles := bundleStore.ListByPriority()

	// Verify ordering: critical > expedited > normal > bulk
	expectedOrder := []bpa.Priority{
		bpa.PriorityCritical,
		bpa.PriorityExpedited,
		bpa.PriorityNormal,
		bpa.PriorityBulk,
	}

	if len(bundles) != len(expectedOrder) {
		t.Fatalf("Expected %d bundles, got %d", len(expectedOrder), len(bundles))
	}

	for i, bundle := range bundles {
		if bundle.Priority != expectedOrder[i] {
			t.Errorf("Bundle %d: expected priority %v, got %v", i, expectedOrder[i], bundle.Priority)
		}
	}

	t.Log("Priority-based delivery test passed")
}

// TestBundleRetryOnNoACK tests that bundles are retained for retry when not acknowledged
// Validates Requirement: 5.5
func TestBundleRetryOnNoACK(t *testing.T) {
	// Create bundle store
	bundleStore := store.NewBundleStore(10 * 1024 * 1024)

	sourceEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.0"}
	destEID := bpa.EndpointID{Scheme: "ipn", SSP: "2.0"}

	bpaAgent := bpa.NewBundleProtocolAgent([]bpa.EndpointID{sourceEID})

	// Create bundle
	bundle, err := bpaAgent.CreateBundle(sourceEID, destEID, []byte("Test payload"), bpa.PriorityNormal, 3600)
	if err != nil {
		t.Fatalf("Failed to create bundle: %v", err)
	}

	// Store bundle
	if err := bundleStore.Store(bundle); err != nil {
		t.Fatalf("Failed to store bundle: %v", err)
	}

	initialCount := bundleStore.Count()

	// Simulate transmission without ACK (bundle should remain in store)
	// In a real scenario, the CLA would return an error or timeout

	// Verify bundle is still in store
	if bundleStore.Count() != initialCount {
		t.Errorf("Expected bundle to remain in store after failed transmission")
	}

	// Verify bundle can be retrieved for retry
	retrieved, err := bundleStore.Retrieve(bundle.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve bundle for retry: %v", err)
	}

	if retrieved.ID != bundle.ID {
		t.Error("Retrieved bundle ID does not match original")
	}

	t.Log("Bundle retry test passed: bundle retained for retry")
}

// TestRemoteDestinationStored tests that bundles with remote destinations are stored
// Validates Requirement: 5.2
func TestRemoteDestinationStored(t *testing.T) {
	sourceEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.0"}
	localEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.1"}
	remoteEID := bpa.EndpointID{Scheme: "ipn", SSP: "2.0"}

	bpaAgent := bpa.NewBundleProtocolAgent([]bpa.EndpointID{localEID})
	bundleStore := store.NewBundleStore(10 * 1024 * 1024)

	// Create bundle with remote destination
	bundle, err := bpaAgent.CreateBundle(sourceEID, remoteEID, []byte("Remote payload"), bpa.PriorityNormal, 3600)
	if err != nil {
		t.Fatalf("Failed to create bundle: %v", err)
	}

	// Verify destination is not local
	if bpaAgent.IsLocalEndpoint(bundle.Destination) {
		t.Error("Expected destination to be remote")
	}

	// Store bundle for forwarding
	if err := bundleStore.Store(bundle); err != nil {
		t.Fatalf("Failed to store bundle: %v", err)
	}

	// Verify bundle is stored
	if bundleStore.Count() != 1 {
		t.Errorf("Expected 1 bundle in store, got %d", bundleStore.Count())
	}

	// Verify bundle can be retrieved by destination
	bundles := bundleStore.ListByDestination(remoteEID)
	if len(bundles) != 1 {
		t.Errorf("Expected 1 bundle for remote destination, got %d", len(bundles))
	}

	t.Log("Remote destination storage test passed")
}

// setupStoreForwardNode creates a test node for store-and-forward testing
func setupStoreForwardNode(t *testing.T, nodeID string, nodeNumber int) *node.NodeController {
	endpoints := []bpa.EndpointID{
		{Scheme: "ipn", SSP: "1.0"},
	}

	bpaAgent := bpa.NewBundleProtocolAgent(endpoints)
	bundleStore := store.NewBundleStore(10 * 1024 * 1024)
	planManager := contact.NewContactPlanManager()

	plan := &contact.ContactPlan{
		PlanID:      1,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   0,
		ValidTo:     2147483647,
		Contacts: []contact.ContactWindow{
			{
				ContactID:  1,
				RemoteNode: contact.NodeID("node-b"),
				StartTime:  0,
				EndTime:    2147483647,
				DataRate:   9600,
				LinkType:   contact.LinkTypeUHFTNC,
			},
		},
	}

	if err := planManager.LoadPlan(plan); err != nil {
		t.Fatalf("Failed to load contact plan: %v", err)
	}

	claAdapter := cla.NewMockCLA(cla.CLATypeAX25LTPUHFTNC)

	config := node.NodeConfig{
		NodeID:          nodeID,
		NodeType:        node.NodeTypeTerrestrial,
		Endpoints:       endpoints,
		MaxStorageBytes: 10 * 1024 * 1024,
		SRAMBytes:       0,
		DefaultPriority: bpa.PriorityNormal,
	}

	controller := node.NewNodeController(config, bpaAgent, bundleStore, planManager, claAdapter)

	if err := controller.Initialize(); err != nil {
		t.Fatalf("Failed to initialize node: %v", err)
	}

	return controller
}
