package integration

import (
	"fmt"
	"testing"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla/uhf_iq_b200"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/node"
	"terrestrial-dtn/pkg/nvm"
	"terrestrial-dtn/pkg/power"
	"terrestrial-dtn/pkg/sdr/b200mini"
	"terrestrial-dtn/pkg/store"
)

// TestEMSimulatedPass tests bundle upload/download during a simulated 8-minute orbital pass
func TestEMSimulatedPass(t *testing.T) {
	// Create EM node
	emNode := createEMNode(t, "em-test-1", 1, "KX0EM1")
	defer cleanupEMNode(emNode)

	// Create ground station node
	gsNode := createEMNode(t, "ground-1", 2, "KX0GS1")
	defer cleanupEMNode(gsNode)

	// Define simulated pass (8 minutes, 9.6 kbps UHF)
	passStart := time.Now().Add(5 * time.Second)
	passDuration := 8 * time.Minute
	passEnd := passStart.Add(passDuration)

	contactWindow := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "ground-1",
		StartTime:  uint64(passStart.Unix()),
		EndTime:    uint64(passEnd.Unix()),
		DataRate:   9600,
		LinkType:   contact.LinkTypeUHFIQB200,
	}

	// Load contact plan
	emNode.contactManager.LoadPlan([]contact.ContactWindow{contactWindow})
	gsNode.contactManager.LoadPlan([]contact.ContactWindow{contactWindow})

	// Create test bundles for upload
	testBundles := createTestBundles(t, 10, emNode.bpa)

	// Store bundles in ground station for upload
	for _, bundle := range testBundles {
		if err := gsNode.store.Store(bundle); err != nil {
			t.Fatalf("Failed to store bundle: %v", err)
		}
	}

	t.Logf("Simulated pass: %s to %s (8 min, 9.6 kbps UHF)", passStart.Format("15:04:05"), passEnd.Format("15:04:05"))

	// Wait for pass to start
	time.Sleep(time.Until(passStart))

	// Open CLAs for contact
	if err := gsNode.cla.Open(contactWindow); err != nil {
		t.Fatalf("Failed to open GS CLA: %v", err)
	}
	defer gsNode.cla.Close()

	if err := emNode.cla.Open(contactWindow); err != nil {
		t.Fatalf("Failed to open EM CLA: %v", err)
	}
	defer emNode.cla.Close()

	// Simulate pass: upload bundles from ground station to EM
	uploadedCount := 0
	for _, bundle := range testBundles {
		if time.Now().After(passEnd) {
			break
		}

		// Transmit from ground station
		metrics, err := gsNode.cla.SendBundle(bundle)
		if err != nil {
			t.Logf("Upload failed: %v", err)
			continue
		}

		t.Logf("Uploaded bundle %s (RSSI=%.1fdBm, SNR=%.1fdB)",
			bundle.ID.String(), metrics.RSSI, metrics.SNR)

		// Receive at EM node
		rxBundle, rxMetrics, err := emNode.cla.RecvBundle()
		if err != nil {
			t.Logf("Reception failed: %v", err)
			continue
		}

		// Store in EM node
		if err := emNode.store.Store(rxBundle); err != nil {
			t.Fatalf("Failed to store received bundle: %v", err)
		}

		uploadedCount++
		t.Logf("EM received bundle (RSSI=%.1fdBm, SNR=%.1fdB)", rxMetrics.RSSI, rxMetrics.SNR)

		// Simulate transmission time
		time.Sleep(100 * time.Millisecond)
	}

	// Close CLAs
	gsNode.cla.Close()
	emNode.cla.Close()

	t.Logf("Pass complete: uploaded %d/%d bundles", uploadedCount, len(testBundles))

	// Verify EM node stored bundles
	emBundles := emNode.store.ListByPriority()
	if len(emBundles) != uploadedCount {
		t.Errorf("Expected %d bundles in EM store, got %d", uploadedCount, len(emBundles))
	}

	// Validate power budget
	validatePowerBudget(t, emNode.powerMgr, passDuration)

	// Validate SRAM usage
	validateSRAMUsage(t, emNode)

	t.Log("EM simulated pass test passed")
}

// TestEMStoreAndForward tests store-and-forward through EM node
func TestEMStoreAndForward(t *testing.T) {
	// Create nodes
	gsA := createEMNode(t, "ground-a", 1, "KX0GSA")
	defer cleanupEMNode(gsA)

	emNode := createEMNode(t, "em-node", 2, "KX0EM1")
	defer cleanupEMNode(emNode)

	gsB := createEMNode(t, "ground-b", 3, "KX0GSB")
	defer cleanupEMNode(gsB)

	// Pass 1: Ground A → EM (upload)
	pass1Start := time.Now().Add(2 * time.Second)
	pass1Window := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "em-node",
		StartTime:  uint64(pass1Start.Unix()),
		EndTime:    uint64(pass1Start.Add(8 * time.Minute).Unix()),
		DataRate:   9600,
		LinkType:   contact.LinkTypeUHFIQB200,
	}

	// Pass 2: EM → Ground B (download) - 90 minutes later
	pass2Start := pass1Start.Add(90 * time.Minute)
	pass2Window := contact.ContactWindow{
		ContactID:  2,
		RemoteNode: "ground-b",
		StartTime:  uint64(pass2Start.Unix()),
		EndTime:    uint64(pass2Start.Add(8 * time.Minute).Unix()),
		DataRate:   9600,
		LinkType:   contact.LinkTypeUHFIQB200,
	}

	// Create test bundle
	bundle := bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: uint64(time.Now().Unix()),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "3.0"}, // Destination: Ground B
		Payload:     []byte("Test message from Ground A to Ground B via EM"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    7200, // 2 hours
		CreatedAt:   uint64(time.Now().Unix()),
		BundleType:  bpa.BundleTypeData,
	}

	// Store in Ground A
	if err := gsA.store.Store(bundle); err != nil {
		t.Fatalf("Failed to store bundle: %v", err)
	}

	t.Log("Pass 1: Ground A → EM (upload)")
	time.Sleep(time.Until(pass1Start))

	// Upload to EM
	gsA.cla.Open(pass1Window)
	emNode.cla.Open(pass1Window)

	metrics, err := gsA.cla.SendBundle(bundle)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	t.Logf("Uploaded (RSSI=%.1fdBm)", metrics.RSSI)

	rxBundle, _, err := emNode.cla.RecvBundle()
	if err != nil {
		t.Fatalf("Reception failed: %v", err)
	}

	emNode.store.Store(rxBundle)
	gsA.cla.Close()
	emNode.cla.Close()

	t.Log("EM stored bundle, entering Stop 2 mode for 90 minutes")

	// EM enters Stop 2 mode between passes
	emNode.powerMgr.EnterStop2(pass2Start.Add(-30 * time.Second))

	// Simulate 90-minute orbit gap (fast-forward for test)
	t.Log("Simulating 90-minute orbit gap...")

	// Wake up before Pass 2
	time.Sleep(100 * time.Millisecond) // Simulated wakeup
	emNode.powerMgr.ExitStop2()

	t.Log("Pass 2: EM → Ground B (download)")

	// Download to Ground B
	emNode.cla.Open(pass2Window)
	gsB.cla.Open(pass2Window)

	storedBundles := emNode.store.ListByDestination(bpa.EndpointID{Scheme: "ipn", SSP: "3.0"})
	if len(storedBundles) == 0 {
		t.Fatal("No bundles stored for Ground B")
	}

	metrics, err = emNode.cla.SendBundle(storedBundles[0])
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	t.Logf("Downloaded (RSSI=%.1fdBm)", metrics.RSSI)

	rxBundle, _, err = gsB.cla.RecvBundle()
	if err != nil {
		t.Fatalf("Reception failed: %v", err)
	}

	gsB.store.Store(rxBundle)
	emNode.cla.Close()
	gsB.cla.Close()

	// Verify delivery
	gsBBundles := gsB.store.ListByPriority()
	if len(gsBBundles) != 1 {
		t.Errorf("Expected 1 bundle at Ground B, got %d", len(gsBBundles))
	}

	t.Log("Store-and-forward test passed")
}

// Helper functions

type emNodeComponents struct {
	nodeID         string
	bpa            *bpa.BPA
	store          *store.Store
	contactManager *contact.Manager
	cla            *uhf_iq_b200.UHFIQB200CLA
	controller     *node.Controller
	powerMgr       *power.PowerManager
	nvm            *nvm.NVM
}

func createEMNode(t *testing.T, nodeID string, nodeNumber int, callsign string) *emNodeComponents {
	// NVM
	nvmConfig := nvm.DefaultConfig()
	nvmConfig.Capacity = 128 * 1024 * 1024
	nvmDevice, err := nvm.New(nvmConfig)
	if err != nil {
		t.Fatalf("Failed to create NVM: %v", err)
	}
	nvmDevice.Open()

	// Store
	storeConfig := store.Config{MaxBytes: 128 * 1024 * 1024}
	bundleStore := store.New(storeConfig)

	// BPA
	bpaConfig := bpa.Config{
		NodeEID: bpa.EndpointID{Scheme: "ipn", SSP: fmt.Sprintf("%d.0", nodeNumber)},
	}
	bundleAgent := bpa.New(bpaConfig)

	// Contact Manager
	contactManager := contact.NewManager()

	// CLA
	claConfig := uhf_iq_b200.DefaultConfig(callsign)
	cla, err := uhf_iq_b200.New(claConfig)
	if err != nil {
		t.Fatalf("Failed to create CLA: %v", err)
	}

	// Power Manager
	powerMgr := power.New(power.DefaultConfig())

	// Node Controller
	nodeConfig := node.NodeConfig{
		NodeID:          nodeID,
		NodeType:        node.NodeTypeEngineeringModel,
		Endpoints:       []bpa.EndpointID{{Scheme: "ipn", SSP: fmt.Sprintf("%d.0", nodeNumber)}},
		MaxStorageBytes: 128 * 1024 * 1024,
		SRAMBytes:       786 * 1024,
		DefaultPriority: bpa.PriorityNormal,
	}
	controller := node.NewController(nodeConfig, bundleAgent, bundleStore, contactManager, cla)

	return &emNodeComponents{
		nodeID:         nodeID,
		bpa:            bundleAgent,
		store:          bundleStore,
		contactManager: contactManager,
		cla:            cla,
		controller:     controller,
		powerMgr:       powerMgr,
		nvm:            nvmDevice,
	}
}

func cleanupEMNode(node *emNodeComponents) {
	if node.nvm != nil {
		node.nvm.Close()
	}
}

func createTestBundles(t *testing.T, count int, agent *bpa.BPA) []bpa.Bundle {
	bundles := make([]bpa.Bundle, count)
	for i := 0; i < count; i++ {
		bundles[i] = bpa.Bundle{
			ID: bpa.BundleID{
				SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
				CreationTimestamp: uint64(time.Now().Unix()),
				SequenceNumber:    uint64(i + 1),
			},
			Destination: bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			Payload:     []byte(fmt.Sprintf("Test bundle %d", i+1)),
			Priority:    bpa.PriorityNormal,
			Lifetime:    3600,
			CreatedAt:   uint64(time.Now().Unix()),
			BundleType:  bpa.BundleTypeData,
		}
	}
	return bundles
}

func validatePowerBudget(t *testing.T, powerMgr *power.PowerManager, passDuration time.Duration) {
	// Calculate power budget for 90-minute orbit (8 min active, 82 min sleep)
	runTime := passDuration
	sleepTime := 82 * time.Minute
	budget := powerMgr.GetPowerBudget(runTime, sleepTime)

	t.Logf("Power Budget: %s", budget.String())

	// Validate against 5-10 W average power budget for LEO
	maxPower := 10.0 // Watts
	if !budget.IsWithinBudget(maxPower) {
		t.Errorf("Power budget exceeded: %.3fW > %.3fW", budget.AveragePower, maxPower)
	} else {
		t.Logf("Power budget OK: %.3fW <= %.3fW", budget.AveragePower, maxPower)
	}
}

func validateSRAMUsage(t *testing.T, node *emNodeComponents) {
	// STM32U585 has 786 KB SRAM
	maxSRAM := uint64(786 * 1024)

	// Check bundle store memory usage (simplified)
	capacity := node.store.Capacity()
	if capacity.UsedBytes > maxSRAM {
		t.Errorf("SRAM usage exceeded: %d bytes > %d bytes", capacity.UsedBytes, maxSRAM)
	} else {
		t.Logf("SRAM usage OK: %d bytes <= %d bytes", capacity.UsedBytes, maxSRAM)
	}
}
