package ionconfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateLEOConfig(t *testing.T) {
	// Create temporary output directory
	tmpDir := t.TempDir()

	// Create test orbital parameters
	epoch := time.Now()
	orbitalParams := &OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  6871000.0, // ~500 km altitude LEO
		Eccentricity:    0.001,     // Nearly circular
		InclinationDeg:  51.6,      // ISS-like inclination
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	// Create test contact plan with CGR-predicted passes
	contacts := []LEOContact{
		{
			RemoteNodeNumber: 1,
			RemoteCallsign:   "KA1ABC",
			StartTime:        epoch.Add(10 * time.Minute),
			Duration:         8 * time.Minute,
			DataRate:         9600,
			MaxElevationDeg:  45.0,
			Confidence:       0.95,
		},
		{
			RemoteNodeNumber: 2,
			RemoteCallsign:   "KB2XYZ",
			StartTime:        epoch.Add(100 * time.Minute),
			Duration:         6 * time.Minute,
			DataRate:         9600,
			MaxElevationDeg:  30.0,
			Confidence:       0.90,
		},
	}

	// Create LEO node config
	config := LEONodeConfig{
		NodeID:           "leo-cubesat-01",
		NodeNumber:       10,
		Callsign:         "KL0SAT",
		StorageBytes:     128 * 1024 * 1024, // 128 MB
		SRAMBytes:        786 * 1024,         // 786 KB
		ContactPlan:      contacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	// Generate configuration
	err := GenerateLEOConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateLEOConfig failed: %v", err)
	}

	// Verify all config files were created
	expectedFiles := []string{
		"node.ionrc",
		"node.ltprc",
		"node.bprc",
		"node.ipnrc",
		"leo.ionconfig",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", filename)
		}
	}

	// Verify ionrc content
	ionrcPath := filepath.Join(tmpDir, "node.ionrc")
	ionrcContent, err := os.ReadFile(ionrcPath)
	if err != nil {
		t.Fatalf("Failed to read ionrc: %v", err)
	}

	ionrcStr := string(ionrcContent)
	if !contains(ionrcStr, "1 10 ''") {
		t.Errorf("ionrc missing node initialization")
	}

	// Verify ltprc content
	ltprcPath := filepath.Join(tmpDir, "node.ltprc")
	ltprcContent, err := os.ReadFile(ltprcPath)
	if err != nil {
		t.Fatalf("Failed to read ltprc: %v", err)
	}

	ltprcStr := string(ltprcContent)
	if !contains(ltprcStr, "a span 1") {
		t.Errorf("ltprc missing span for contact 1")
	}
	if !contains(ltprcStr, "a span 2") {
		t.Errorf("ltprc missing span for contact 2")
	}

	// Verify bprc content
	bprcPath := filepath.Join(tmpDir, "node.bprc")
	bprcContent, err := os.ReadFile(bprcPath)
	if err != nil {
		t.Fatalf("Failed to read bprc: %v", err)
	}

	bprcStr := string(bprcContent)
	if !contains(bprcStr, "a endpoint ipn:10.0 q") {
		t.Errorf("bprc missing endpoint configuration")
	}
	if !contains(bprcStr, "a endpoint ipn:10.10 q") {
		t.Errorf("bprc missing telemetry endpoint")
	}
	if !contains(bprcStr, "a outduct ltp/1") {
		t.Errorf("bprc missing outduct for contact 1")
	}

	// Verify ipnrc content
	ipnrcPath := filepath.Join(tmpDir, "node.ipnrc")
	ipnrcContent, err := os.ReadFile(ipnrcPath)
	if err != nil {
		t.Fatalf("Failed to read ipnrc: %v", err)
	}

	ipnrcStr := string(ipnrcContent)
	if !contains(ipnrcStr, "a plan 1 ltp/1") {
		t.Errorf("ipnrc missing plan for contact 1")
	}

	// Verify leo.ionconfig content
	ionconfigPath := filepath.Join(tmpDir, "leo.ionconfig")
	ionconfigContent, err := os.ReadFile(ionconfigPath)
	if err != nil {
		t.Fatalf("Failed to read leo.ionconfig: %v", err)
	}

	ionconfigStr := string(ionconfigContent)
	if !contains(ionconfigStr, "NODE_NUMBER=10") {
		t.Errorf("leo.ionconfig missing node number")
	}
	if !contains(ionconfigStr, "CALLSIGN=KL0SAT") {
		t.Errorf("leo.ionconfig missing callsign")
	}
	if !contains(ionconfigStr, "RADIO_TYPE=flight_iq") {
		t.Errorf("leo.ionconfig missing radio type")
	}
	if !contains(ionconfigStr, "ORBITAL_EPOCH=") {
		t.Errorf("leo.ionconfig missing orbital parameters")
	}
	if !contains(ionconfigStr, "TELEMETRY_ENABLED=true") {
		t.Errorf("leo.ionconfig missing telemetry configuration")
	}
}

func TestGenerateLEOWithCGRPrediction(t *testing.T) {
	tmpDir := t.TempDir()

	epoch := time.Now()
	orbitalParams := &OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  6871000.0,
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	predictedContacts := []LEOContact{
		{
			RemoteNodeNumber: 1,
			RemoteCallsign:   "KA1ABC",
			StartTime:        epoch.Add(15 * time.Minute),
			Duration:         7 * time.Minute,
			DataRate:         9600,
			MaxElevationDeg:  50.0,
			Confidence:       0.98,
		},
	}

	err := GenerateLEOWithCGRPrediction(
		"leo-test",
		20,
		"KL0TST",
		orbitalParams,
		predictedContacts,
		tmpDir,
	)

	if err != nil {
		t.Fatalf("GenerateLEOWithCGRPrediction failed: %v", err)
	}

	// Verify config was created
	ionconfigPath := filepath.Join(tmpDir, "leo.ionconfig")
	if _, err := os.Stat(ionconfigPath); os.IsNotExist(err) {
		t.Errorf("leo.ionconfig was not created")
	}
}

func TestLEOConfigWithoutOrbitalParams(t *testing.T) {
	tmpDir := t.TempDir()

	config := LEONodeConfig{
		NodeID:           "leo-simple",
		NodeNumber:       30,
		Callsign:         "KL0SMP",
		StorageBytes:     64 * 1024 * 1024,
		SRAMBytes:        786 * 1024,
		ContactPlan:      []LEOContact{},
		OrbitalParams:    nil, // No orbital params
		TelemetryEnabled: false,
	}

	err := GenerateLEOConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateLEOConfig failed: %v", err)
	}

	// Verify config was created without orbital parameters
	ionconfigPath := filepath.Join(tmpDir, "leo.ionconfig")
	ionconfigContent, err := os.ReadFile(ionconfigPath)
	if err != nil {
		t.Fatalf("Failed to read leo.ionconfig: %v", err)
	}

	ionconfigStr := string(ionconfigContent)
	if contains(ionconfigStr, "ORBITAL_EPOCH=") {
		t.Errorf("leo.ionconfig should not contain orbital parameters")
	}
	if contains(ionconfigStr, "TELEMETRY_ENABLED=true") {
		t.Errorf("leo.ionconfig should have telemetry disabled")
	}
}

func TestUpdateLEOContactPlan(t *testing.T) {
	tmpDir := t.TempDir()

	// First create an initial config
	epoch := time.Now()
	orbitalParams := &OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  6871000.0,
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	initialContacts := []LEOContact{
		{
			RemoteNodeNumber: 1,
			RemoteCallsign:   "KA1ABC",
			StartTime:        epoch.Add(10 * time.Minute),
			Duration:         8 * time.Minute,
			DataRate:         9600,
			MaxElevationDeg:  45.0,
			Confidence:       0.95,
		},
	}

	config := LEONodeConfig{
		NodeID:           "leo-update-test",
		NodeNumber:       40,
		Callsign:         "KL0UPD",
		StorageBytes:     128 * 1024 * 1024,
		SRAMBytes:        786 * 1024,
		ContactPlan:      initialContacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	err := GenerateLEOConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateLEOConfig failed: %v", err)
	}

	// Now update with new contact plan
	newEpoch := epoch.Add(24 * time.Hour)
	newOrbitalParams := &OrbitalParameters{
		Epoch:           newEpoch,
		SemiMajorAxisM:  6871000.0,
		Eccentricity:    0.001,
		InclinationDeg:  51.6,
		RAANDeg:         46.0, // Updated RAAN
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	newContacts := []LEOContact{
		{
			RemoteNodeNumber: 1,
			RemoteCallsign:   "KA1ABC",
			StartTime:        newEpoch.Add(15 * time.Minute),
			Duration:         7 * time.Minute,
			DataRate:         9600,
			MaxElevationDeg:  40.0,
			Confidence:       0.93,
		},
		{
			RemoteNodeNumber: 2,
			RemoteCallsign:   "KB2XYZ",
			StartTime:        newEpoch.Add(105 * time.Minute),
			Duration:         6 * time.Minute,
			DataRate:         9600,
			MaxElevationDeg:  35.0,
			Confidence:       0.88,
		},
	}

	err = UpdateLEOContactPlan(tmpDir, newContacts, newOrbitalParams)
	if err != nil {
		t.Fatalf("UpdateLEOContactPlan failed: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
