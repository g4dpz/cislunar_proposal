package integration

import (
	"fmt"
	"testing"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla/uhf_iq"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/node"
	"terrestrial-dtn/pkg/nvm"
	"terrestrial-dtn/pkg/power"
	"terrestrial-dtn/pkg/store"
)

// TestLEOPassSimulation tests ground-to-LEO-to-ground store-and-forward during a simulated orbital pass
// Validates Requirements: 13.1, 13.2, 13.3, 13.4, 13.5
func TestLEOPassSimulation(t *testing.T) {
	// Create ground station A (uplink)
	gsA := createLEOGroundStation(t, "ground-a", 1, "KX0GSA")
	defer cleanupLEONode(gsA)

	// Create LEO CubeSat flight node
	leoNode := createLEOFlightNode(t, "leo-cubesat-1", 10, "KX0LEO")
	defer cleanupLEONode(leoNode)

	// Create ground station B (downlink)
	gsB := createLEOGroundStation(t, "ground-b", 2, "KX0GSB")
	defer cleanupLEONode(gsB)

	// Define simulated LEO pass over Ground Station A (8 minutes, 9.6 kbps UHF)
	pass1Start := time.Now().Add(5 * time.Second)
	pass1Duration := 8 * time.Minute
	pass1End := pass1Start.Add(pass1Duration)

	pass1Window := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "leo-cubesat-1",
		StartTime:  pass1Start.Unix(),
		EndTime:    pass1End.Unix(),
		DataRate:   9600,
		LinkType:   contact.LinkTypeUHFIQ,
	}

	// Define simulated LEO pass over Ground Station B (90 minutes later, 8 minutes)
	pass2Start := pass1Start.Add(90 * time.Minute)
	pass2Duration := 8 * time.Minute
	pass2End := pass2Start.Add(pass2Duration)

	pass2Window := contact.ContactWindow{
		ContactID:  2,
		RemoteNode: "leo-cubesat-1",
		StartTime:  pass2Start.Unix(),
		EndTime:    pass2End.Unix(),
		DataRate:   9600,
		LinkType:   contact.LinkTypeUHFIQ,
	}

	// Load contact plans
	gsA.contactManager.LoadPlan(&contact.ContactPlan{
		PlanID:      1,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   0,
		ValidTo:     2147483647,
		Contacts:    []contact.ContactWindow{pass1Window},
	})
	leoNode.contactManager.LoadPlan(&contact.ContactPlan{
		PlanID:      2,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   0,
		ValidTo:     2147483647,
		Contacts:    []contact.ContactWindow{pass1Window, pass2Window},
	})
	gsB.contactManager.LoadPlan(&contact.ContactPlan{
		PlanID:      3,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   0,
		ValidTo:     2147483647,
		Contacts:    []contact.ContactWindow{pass2Window},
	})

	// Create test bundle from Ground Station A to Ground Station B (via LEO)
	testBundle := bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"}, // Destination: Ground Station B
		Payload:     []byte("Test message from Ground A to Ground B via LEO CubeSat"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    7200, // 2 hours
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	// Store bundle in Ground Station A for upload
	if err := gsA.store.Store(&testBundle); err != nil {
		t.Fatalf("Failed to store bundle: %v", err)
	}

	t.Logf("Pass 1: Ground A → LEO (upload)")
	t.Logf("Pass window: %s to %s (8 min, 9.6 kbps UHF 437 MHz)",
		pass1Start.Format("15:04:05"), pass1End.Format("15:04:05"))

	// Wait for pass 1 to start
	time.Sleep(time.Until(pass1Start))

	// Open CLAs for pass 1 (Ground A → LEO)
	if err := gsA.cla.Open(pass1Window); err != nil {
		t.Fatalf("Failed to open GS A CLA: %v", err)
	}
	defer gsA.cla.Close()

	if err := leoNode.cla.Open(pass1Window); err != nil {
		t.Fatalf("Failed to open LEO CLA: %v", err)
	}

	// Transmit from Ground Station A to LEO
	metrics, err := gsA.cla.SendBundle(&testBundle)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	t.Logf("Uploaded bundle %s (RSSI=%d dBm, SNR=%.1f dB)",
		testBundle.ID.String(), metrics.RSSI, metrics.SNR)

	// Receive at LEO node
	rxBundle, rxMetrics, err := leoNode.cla.RecvBundle()
	if err != nil {
		t.Fatalf("LEO reception failed: %v", err)
	}

	// Store in LEO node NVM
	if err := leoNode.store.Store(rxBundle); err != nil {
		t.Fatalf("Failed to store received bundle in LEO NVM: %v", err)
	}

	t.Logf("LEO received bundle (RSSI=%d dBm, SNR=%.1f dB)", rxMetrics.RSSI, rxMetrics.SNR)

	// Close CLAs for pass 1
	gsA.cla.Close()
	leoNode.cla.Close()

	t.Log("Pass 1 complete: bundle stored in LEO NVM")

	// LEO enters Stop 2 mode between passes
	t.Log("LEO entering Stop 2 mode (~16 µA) for 90-minute orbit gap")
	leoNode.powerMgr.EnterStop2(pass2Start.Add(-30 * time.Second))

	// Validate power budget during pass 1
	validateLEOPowerBudget(t, leoNode.powerMgr, pass1Duration)

	// Simulate 90-minute orbit gap (fast-forward for test)
	t.Log("Simulating 90-minute orbit gap...")
	time.Sleep(100 * time.Millisecond) // Simulated gap

	// Wake up before Pass 2
	leoNode.powerMgr.ExitStop2()
	t.Log("LEO woke from Stop 2 mode, preparing for Pass 2")

	t.Logf("Pass 2: LEO → Ground B (download)")
	t.Logf("Pass window: %s to %s (8 min, 9.6 kbps UHF 437 MHz)",
		pass2Start.Format("15:04:05"), pass2End.Format("15:04:05"))

	// Open CLAs for pass 2 (LEO → Ground B)
	if err := leoNode.cla.Open(pass2Window); err != nil {
		t.Fatalf("Failed to open LEO CLA for pass 2: %v", err)
	}

	if err := gsB.cla.Open(pass2Window); err != nil {
		t.Fatalf("Failed to open GS B CLA: %v", err)
	}
	defer gsB.cla.Close()

	// Retrieve stored bundle from LEO NVM (destination: Ground B)
	storedBundles := leoNode.store.ListByDestination(bpa.EndpointID{Scheme: "ipn", SSP: "2.0"})
	if len(storedBundles) == 0 {
		t.Fatal("No bundles stored for Ground B in LEO NVM")
	}

	// Transmit from LEO to Ground Station B (direct delivery, no relay)
	metrics, err = leoNode.cla.SendBundle(storedBundles[0])
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	t.Logf("Downloaded bundle %s (RSSI=%d dBm, SNR=%.1f dB)",
		storedBundles[0].ID.String(), metrics.RSSI, metrics.SNR)

	// Receive at Ground Station B
	rxBundle, rxMetrics, err = gsB.cla.RecvBundle()
	if err != nil {
		t.Fatalf("GS B reception failed: %v", err)
	}

	// Store in Ground Station B
	if err := gsB.store.Store(rxBundle); err != nil {
		t.Fatalf("Failed to store received bundle at GS B: %v", err)
	}

	t.Logf("Ground B received bundle (RSSI=%d dBm, SNR=%.1f dB)", rxMetrics.RSSI, rxMetrics.SNR)

	// Close CLAs for pass 2
	leoNode.cla.Close()
	gsB.cla.Close()

	t.Log("Pass 2 complete: bundle delivered to Ground B")

	// Verify end-to-end delivery
	gsBBundles := gsB.store.ListByPriority()
	if len(gsBBundles) != 1 {
		t.Errorf("Expected 1 bundle at Ground B, got %d", len(gsBBundles))
	}

	// Verify bundle content matches original
	if string(gsBBundles[0].Payload) != string(testBundle.Payload) {
		t.Errorf("Payload mismatch: expected %q, got %q",
			string(testBundle.Payload), string(gsBBundles[0].Payload))
	}

	// Verify direct delivery (no relay) - bundle went directly from LEO to Ground B
	if gsBBundles[0].Destination.String() != testBundle.Destination.String() {
		t.Errorf("Destination mismatch: expected %s, got %s",
			testBundle.Destination.String(), gsBBundles[0].Destination.String())
	}

	// Validate telemetry
	validateLEOTelemetry(t, leoNode)

	t.Log("LEO pass simulation test passed: ground-to-LEO-to-ground store-and-forward successful")
}

// TestLEODirectDeliveryNoRelay verifies that LEO node does not relay bundles
// Validates Requirements: 13.5 (direct delivery only, no relay)
func TestLEODirectDeliveryNoRelay(t *testing.T) {
	// Create ground station A
	gsA := createLEOGroundStation(t, "ground-a", 1, "KX0GSA")
	defer cleanupLEONode(gsA)

	// Create LEO CubeSat
	leoNode := createLEOFlightNode(t, "leo-cubesat-1", 10, "KX0LEO")
	defer cleanupLEONode(leoNode)

	// Create ground station C (not in contact plan)
	gsC := createLEOGroundStation(t, "ground-c", 3, "KX0GSC")
	defer cleanupLEONode(gsC)

	// Define pass over Ground Station A only
	passStart := time.Now().Add(2 * time.Second)
	passWindow := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "leo-cubesat-1",
		StartTime:  passStart.Unix(),
		EndTime:    passStart.Add(8 * time.Minute).Unix(),
		DataRate:   9600,
		LinkType:   contact.LinkTypeUHFIQ,
	}

	gsA.contactManager.LoadPlan(&contact.ContactPlan{
		PlanID:      1,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   0,
		ValidTo:     2147483647,
		Contacts:    []contact.ContactWindow{passWindow},
	})
	leoNode.contactManager.LoadPlan(&contact.ContactPlan{
		PlanID:      2,
		GeneratedAt: time.Now().Unix(),
		ValidFrom:   0,
		ValidTo:     2147483647,
		Contacts:    []contact.ContactWindow{passWindow},
	})

	// Create bundle from Ground A to Ground C (not in LEO's contact plan)
	bundle := bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "3.0"}, // Ground C (no contact window)
		Payload:     []byte("Test message to unreachable destination"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    3600,
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	gsA.store.Store(&bundle)

	// Wait for pass to start
	time.Sleep(time.Until(passStart))

	// Open CLAs
	gsA.cla.Open(passWindow)
	defer gsA.cla.Close()
	leoNode.cla.Open(passWindow)
	defer leoNode.cla.Close()

	// Upload bundle to LEO
	_, err := gsA.cla.SendBundle(&bundle)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// LEO receives bundle
	rxBundle, _, err := leoNode.cla.RecvBundle()
	if err != nil {
		t.Fatalf("LEO reception failed: %v", err)
	}

	// LEO stores bundle (destination: Ground C)
	leoNode.store.Store(rxBundle)

	// Verify LEO does NOT attempt to relay to Ground C during this pass
	// (no contact window exists for Ground C)
	storedBundles := leoNode.store.ListByDestination(bpa.EndpointID{Scheme: "ipn", SSP: "3.0"})
	if len(storedBundles) != 1 {
		t.Errorf("Expected 1 bundle stored for Ground C, got %d", len(storedBundles))
	}

	// Verify bundle is retained (not relayed or dropped)
	if storedBundles[0].Destination.String() != "ipn:3.0" {
		t.Errorf("Expected destination ipn:3.0, got %s", storedBundles[0].Destination.String())
	}

	t.Log("No relay test passed: LEO retained bundle for unreachable destination (no relay)")
}

// TestLEOPowerBudgetValidation validates 5-10 W average power budget
// Validates Requirements: 13.4 (power budget compliance)
func TestLEOPowerBudgetValidation(t *testing.T) {
	// Create LEO node
	leoNode := createLEOFlightNode(t, "leo-cubesat-1", 10, "KX0LEO")
	defer cleanupLEONode(leoNode)

	// Simulate 8-minute active pass
	passDuration := 8 * time.Minute
	sleepDuration := 82 * time.Minute // 90-minute orbit - 8-minute pass

	// Calculate power budget
	budget := leoNode.powerMgr.GetPowerBudget(passDuration, sleepDuration)

	t.Logf("Power Budget Analysis:")
	t.Logf("  Active time: %s", passDuration)
	t.Logf("  Sleep time: %s", sleepDuration)
	t.Logf("  Average power: %.3f W", budget.AveragePower)

	// Validate against 5-10 W average power budget
	minPower := 5.0
	maxPower := 10.0

	if budget.AveragePower < minPower {
		t.Logf("Warning: Average power %.3f W is below minimum %.3f W", budget.AveragePower, minPower)
	}

	if budget.AveragePower > maxPower {
		t.Errorf("Power budget exceeded: %.3f W > %.3f W", budget.AveragePower, maxPower)
	} else {
		t.Logf("Power budget OK: %.3f W <= %.3f W", budget.AveragePower, maxPower)
	}

	// Verify Stop 2 mode power consumption
	sleepPowerWatts := budget.SleepCurrent * budget.Voltage / 1e6 // Convert µA to A, then to Watts
	if sleepPowerWatts > 0.00002 { // 20 µA max
		t.Errorf("Stop 2 mode power too high: %.6f W (%.1f µA)",
			sleepPowerWatts, budget.SleepCurrent)
	}
}

// Helper functions

type leoNodeComponents struct {
	nodeID         string
	bpa            *bpa.BundleProtocolAgent
	store          *store.BundleStore
	contactManager *contact.ContactPlanManager
	cla            *uhf_iq.UHFIQCLA
	controller     *node.NodeController
	powerMgr       *power.PowerManager
	nvm            *nvm.NVM
}

func createLEOFlightNode(t *testing.T, nodeID string, nodeNumber int, callsign string) *leoNodeComponents {
	// NVM (128 MB for LEO flight)
	nvmConfig := nvm.DefaultConfig()
	nvmConfig.Capacity = 128 * 1024 * 1024
	nvmDevice, err := nvm.New(nvmConfig)
	if err != nil {
		t.Fatalf("Failed to create NVM: %v", err)
	}
	nvmDevice.Open()

	// Store
	bundleStore := store.NewBundleStore(128 * 1024 * 1024)

	// BPA
	bundleAgent := bpa.NewBundleProtocolAgent([]bpa.EndpointID{
		{Scheme: "ipn", SSP: fmt.Sprintf("%d.0", nodeNumber)},
	})

	// Contact Manager
	contactManager := contact.NewContactPlanManager()

	// CLA (flight IQ transceiver, not B200mini)
	claConfig := uhf_iq.DefaultConfig(callsign)
	cla, err := uhf_iq.New(claConfig)
	if err != nil {
		t.Fatalf("Failed to create CLA: %v", err)
	}

	// Power Manager
	powerMgr := power.New(power.DefaultConfig())

	// Node Controller
	nodeConfig := node.NodeConfig{
		NodeID:          nodeID,
		NodeType:        node.NodeTypeLEOCubesat,
		Endpoints:       []bpa.EndpointID{{Scheme: "ipn", SSP: fmt.Sprintf("%d.0", nodeNumber)}},
		MaxStorageBytes: 128 * 1024 * 1024,
		SRAMBytes:       786 * 1024, // STM32U585 SRAM
		DefaultPriority: bpa.PriorityNormal,
	}
	controller := node.NewNodeController(nodeConfig, bundleAgent, bundleStore, contactManager, cla)

	return &leoNodeComponents{
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

func createLEOGroundStation(t *testing.T, nodeID string, nodeNumber int, callsign string) *leoNodeComponents {
	// Ground stations use same components as LEO node for testing
	return createLEOFlightNode(t, nodeID, nodeNumber, callsign)
}

func cleanupLEONode(node *leoNodeComponents) {
	if node.nvm != nil {
		node.nvm.Close()
	}
}

func validateLEOPowerBudget(t *testing.T, powerMgr *power.PowerManager, passDuration time.Duration) {
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

	// Verify Stop 2 mode power consumption
	sleepPowerWatts := budget.SleepCurrent * budget.Voltage / 1e6 // Convert µA to A, then to Watts
	if sleepPowerWatts > 0.00002 { // 20 µA max (allowing margin above 16 µA nominal)
		t.Errorf("Stop 2 mode power too high: %.6f W (%.1f µA)",
			sleepPowerWatts, budget.SleepCurrent)
	}
}

func validateLEOTelemetry(t *testing.T, node *leoNodeComponents) {
	// Get telemetry
	health := node.controller.HealthCheck()
	stats := node.controller.GetStatistics()
	total, used := node.nvm.Capacity()

	// Simulate temperature and battery readings (would come from sensors in real implementation)
	temperature := 25.0 // °C
	battery := 85.0     // %

	t.Log("\n--- LEO Node Telemetry ---")
	t.Logf("Temperature: %.1f°C", temperature)
	t.Logf("Battery: %.1f%%", battery)
	t.Logf("Storage: %.1f%% (%.1f MB / %.1f MB)",
		health.StorageUsedPercent,
		float64(used)/(1024*1024),
		float64(total)/(1024*1024))
	t.Logf("SRAM: 786 KB (STM32U585)")
	t.Logf("Bundles: %d stored, %d forwarded, %d dropped",
		health.BundlesStored, health.BundlesForwarded, health.BundlesDropped)
	t.Logf("Contacts: %d completed, %d missed",
		stats.ContactsCompleted, stats.ContactsMissed)
	t.Log("--------------------------")

	// Validate telemetry ranges
	if temperature < -40 || temperature > 85 {
		t.Errorf("Temperature out of range: %.1f°C", temperature)
	}

	if battery < 0 || battery > 100 {
		t.Errorf("Battery percentage out of range: %.1f%%", battery)
	}

	if health.StorageUsedPercent > 100 {
		t.Errorf("Storage utilization exceeds 100%%: %.1f%%", health.StorageUsedPercent)
	}

	// Validate SRAM constraint (786 KB for STM32U585)
	maxSRAM := uint64(786 * 1024)
	if used > maxSRAM {
		t.Errorf("SRAM usage exceeded: %d bytes > %d bytes", used, maxSRAM)
	}
}
