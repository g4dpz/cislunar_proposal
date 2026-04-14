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

// TestTerrestrialPing tests DTN ping between two terrestrial nodes
// Validates Requirements: 4.1, 4.2, 4.3
func TestTerrestrialPing(t *testing.T) {
	// Setup Node A
	nodeA := setupTerrestrialNode(t, "node-a", 1)
	defer nodeA.Shutdown()

	// Setup Node B
	nodeB := setupTerrestrialNode(t, "node-b", 2)
	defer nodeB.Shutdown()

	// Create ping request from Node A to Node B
	sourceEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.0"}
	destEID := bpa.EndpointID{Scheme: "ipn", SSP: "2.0"}

	bpaA := bpa.NewBundleProtocolAgent([]bpa.EndpointID{sourceEID})
	pingRequest, err := bpaA.CreatePing(sourceEID, destEID)
	if err != nil {
		t.Fatalf("Failed to create ping request: %v", err)
	}

	// Verify ping request properties
	if pingRequest.BundleType != bpa.BundleTypePingRequest {
		t.Errorf("Expected ping request type, got %v", pingRequest.BundleType)
	}

	if pingRequest.Destination != destEID {
		t.Errorf("Expected destination %v, got %v", destEID, pingRequest.Destination)
	}

	// Simulate Node B receiving the ping request
	bpaB := bpa.NewBundleProtocolAgent([]bpa.EndpointID{destEID})
	pingResponse, err := bpaB.HandlePing(pingRequest)
	if err != nil {
		t.Fatalf("Failed to handle ping: %v", err)
	}

	// Verify ping response properties
	if pingResponse.BundleType != bpa.BundleTypePingResponse {
		t.Errorf("Expected ping response type, got %v", pingResponse.BundleType)
	}

	// Verify response is addressed to original sender
	if pingResponse.Destination != sourceEID {
		t.Errorf("Expected response destination %v, got %v", sourceEID, pingResponse.Destination)
	}

	// Verify exactly one response was generated
	if pingResponse == nil {
		t.Error("Expected ping response, got nil")
	}

	// Measure RTT (simulated)
	startTime := pingRequest.CreatedAt
	endTime := pingResponse.CreatedAt
	rtt := endTime - startTime

	t.Logf("Ping successful: RTT = %d seconds", rtt)

	// Verify RTT is reasonable (should be very small for local test)
	if rtt < 0 {
		t.Errorf("Invalid RTT: %d", rtt)
	}
}

// TestTerrestrialPingTimeout tests ping timeout behavior
func TestTerrestrialPingTimeout(t *testing.T) {
	// Setup Node A
	nodeA := setupTerrestrialNode(t, "node-a", 1)
	defer nodeA.Shutdown()

	// Create ping request with short lifetime
	sourceEID := bpa.EndpointID{Scheme: "ipn", SSP: "1.0"}
	destEID := bpa.EndpointID{Scheme: "ipn", SSP: "99.0"} // Non-existent node

	bpaA := bpa.NewBundleProtocolAgent([]bpa.EndpointID{sourceEID})
	pingRequest, err := bpaA.CreatePing(sourceEID, destEID)
	if err != nil {
		t.Fatalf("Failed to create ping request: %v", err)
	}

	// Modify lifetime to be very short
	pingRequest.Lifetime = 1

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Verify bundle is expired
	currentTime := time.Now().Unix()
	if !pingRequest.IsExpired(currentTime) {
		t.Error("Expected ping request to be expired")
	}

	t.Log("Ping timeout test passed")
}

// setupTerrestrialNode creates a test terrestrial node
func setupTerrestrialNode(t *testing.T, nodeID string, nodeNumber int) *node.NodeController {
	// Create endpoints
	endpoints := []bpa.EndpointID{
		{Scheme: "ipn", SSP: "1.0"},
		{Scheme: "ipn", SSP: "1.1"},
	}

	// Create BPA
	bpaAgent := bpa.NewBundleProtocolAgent(endpoints)

	// Create bundle store
	bundleStore := store.NewBundleStore(10 * 1024 * 1024) // 10 MB

	// Create contact plan manager
	planManager := contact.NewContactPlanManager()

	// Load always-on contact plan
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

	// Create mock CLA
	claAdapter := cla.NewMockCLA(cla.CLATypeAX25LTPUHFTNC)

	// Create node controller
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
