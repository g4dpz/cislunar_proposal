package ionconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateCislunarConfig(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	// Define cislunar orbital parameters
	epoch := time.Now()
	orbitalParams := &OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  384400000.0, // ~384,400 km (Earth-Moon distance)
		Eccentricity:    0.05,
		InclinationDeg:  5.0,
		RAANDeg:         45.0,
		ArgPeriapsisDeg: 90.0,
		TrueAnomalyDeg:  0.0,
	}

	// Define cislunar contact plan
	contacts := []CislunarContact{
		{
			RemoteNodeNumber: 100,
			RemoteCallsign:   "W1ABC",
			StartTime:        epoch.Add(1 * time.Hour),
			Duration:         2 * time.Hour, // 2-hour contact window
			DataRate:         500,           // 500 bps S-band
			MaxElevationDeg:  45.0,
			Confidence:       0.85,
			LightTimeDelay:   1.28, // ~1.28 seconds for Earth-Moon
		},
		{
			RemoteNodeNumber: 101,
			RemoteCallsign:   "K2XYZ",
			StartTime:        epoch.Add(12 * time.Hour),
			Duration:         3 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  60.0,
			Confidence:       0.75,
			LightTimeDelay:   1.30,
		},
	}

	config := CislunarNodeConfig{
		NodeID:           "cislunar-01",
		NodeNumber:       10,
		Callsign:         "N0CALL",
		StorageBytes:     512 * 1024 * 1024, // 512 MB
		SRAMBytes:        786 * 1024,         // 786 KB
		Band:             SBand,
		ContactPlan:      contacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	err := GenerateCislunarConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateCislunarConfig failed: %v", err)
	}

	// Verify all config files were created
	expectedFiles := []string{
		"node.ionrc",
		"node.ltprc",
		"node.bprc",
		"node.ipnrc",
		"sband.ionconfig",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", filename)
		}
	}

	// Verify ionrc content
	ionrcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ionrc"))
	if err != nil {
		t.Fatalf("Failed to read node.ionrc: %v", err)
	}

	ionrcStr := string(ionrcContent)
	if !strings.Contains(ionrcStr, "cislunar node cislunar-01") {
		t.Errorf("ionrc missing node ID")
	}
	if !strings.Contains(ionrcStr, "1 10 ''") {
		t.Errorf("ionrc missing node initialization")
	}
	if !strings.Contains(ionrcStr, "1-2 second light-time delay") {
		t.Errorf("ionrc missing light-time delay comment")
	}

	// Verify ltprc content
	ltprcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ltprc"))
	if err != nil {
		t.Fatalf("Failed to read node.ltprc: %v", err)
	}

	ltprcStr := string(ltprcContent)
	if !strings.Contains(ltprcStr, "Long-delay session management") {
		t.Errorf("ltprc missing long-delay session management comment")
	}
	if !strings.Contains(ltprcStr, "Light-time delay: 1.28") {
		t.Errorf("ltprc missing light-time delay for first contact")
	}
	if !strings.Contains(ltprcStr, "RTT: 2.56") {
		t.Errorf("ltprc missing round-trip time calculation")
	}
	if !strings.Contains(ltprcStr, "a span 100") {
		t.Errorf("ltprc missing LTP span for first contact")
	}
	if !strings.Contains(ltprcStr, "a span 101") {
		t.Errorf("ltprc missing LTP span for second contact")
	}

	// Verify bprc content
	bprcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.bprc"))
	if err != nil {
		t.Fatalf("Failed to read node.bprc: %v", err)
	}

	bprcStr := string(bprcContent)
	if !strings.Contains(bprcStr, "Long-duration message storage") {
		t.Errorf("bprc missing long-duration storage comment")
	}
	if !strings.Contains(bprcStr, "a endpoint ipn:10.0 q") {
		t.Errorf("bprc missing endpoint configuration")
	}
	if !strings.Contains(bprcStr, "a endpoint ipn:10.10 q") {
		t.Errorf("bprc missing telemetry endpoint")
	}
	if !strings.Contains(bprcStr, "a outduct ltp/100") {
		t.Errorf("bprc missing outduct for first contact")
	}

	// Verify ipnrc content
	ipnrcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ipnrc"))
	if err != nil {
		t.Fatalf("Failed to read node.ipnrc: %v", err)
	}

	ipnrcStr := string(ipnrcContent)
	if !strings.Contains(ipnrcStr, "Confidence: 0.85") {
		t.Errorf("ipnrc missing confidence for first contact")
	}
	if !strings.Contains(ipnrcStr, "a plan 100") {
		t.Errorf("ipnrc missing plan for first contact")
	}

	// Verify ionconfig content
	ionconfigContent, err := os.ReadFile(filepath.Join(tmpDir, "sband.ionconfig"))
	if err != nil {
		t.Fatalf("Failed to read sband.ionconfig: %v", err)
	}

	ionconfigStr := string(ionconfigContent)
	if !strings.Contains(ionconfigStr, "NODE_NUMBER=10") {
		t.Errorf("ionconfig missing node number")
	}
	if !strings.Contains(ionconfigStr, "CALLSIGN=N0CALL") {
		t.Errorf("ionconfig missing callsign")
	}
	if !strings.Contains(ionconfigStr, "CENTER_FREQ=2200000000") {
		t.Errorf("ionconfig missing S-band center frequency")
	}
	if !strings.Contains(ionconfigStr, "MODULATION=BPSK") {
		t.Errorf("ionconfig missing BPSK modulation")
	}
	if !strings.Contains(ionconfigStr, "FEC_TYPE=LDPC") {
		t.Errorf("ionconfig missing LDPC FEC type")
	}
	if !strings.Contains(ionconfigStr, "FEC_ENABLED=true") {
		t.Errorf("ionconfig missing FEC enabled flag")
	}
	if !strings.Contains(ionconfigStr, "SEMI_MAJOR_AXIS_M=3.844e+08") {
		t.Errorf("ionconfig missing orbital parameters")
	}
	if !strings.Contains(ionconfigStr, "LIGHT_TIME_DELAY=1.29") {
		t.Errorf("ionconfig missing light-time delay")
	}
	if !strings.Contains(ionconfigStr, "LONG_DURATION_STORAGE=true") {
		t.Errorf("ionconfig missing long-duration storage flag")
	}
}

func TestGenerateCislunarConfig_XBand(t *testing.T) {
	tmpDir := t.TempDir()

	epoch := time.Now()
	contacts := []CislunarContact{
		{
			RemoteNodeNumber: 200,
			RemoteCallsign:   "W5DEF",
			StartTime:        epoch.Add(2 * time.Hour),
			Duration:         1 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  50.0,
			Confidence:       0.80,
			LightTimeDelay:   1.25,
		},
	}

	config := CislunarNodeConfig{
		NodeID:           "cislunar-xband",
		NodeNumber:       20,
		Callsign:         "K9TEST",
		StorageBytes:     1024 * 1024 * 1024, // 1 GB
		SRAMBytes:        2 * 1024 * 1024,    // 2 MB (higher-capability processor)
		Band:             XBand,
		ContactPlan:      contacts,
		OrbitalParams:    nil, // No orbital params
		TelemetryEnabled: false,
	}

	err := GenerateCislunarConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateCislunarConfig (X-band) failed: %v", err)
	}

	// Verify X-band ionconfig was created
	ionconfigPath := filepath.Join(tmpDir, "xband.ionconfig")
	if _, err := os.Stat(ionconfigPath); os.IsNotExist(err) {
		t.Errorf("Expected xband.ionconfig was not created")
	}

	// Verify X-band specific parameters
	ionconfigContent, err := os.ReadFile(ionconfigPath)
	if err != nil {
		t.Fatalf("Failed to read xband.ionconfig: %v", err)
	}

	ionconfigStr := string(ionconfigContent)
	if !strings.Contains(ionconfigStr, "CENTER_FREQ=8400000000") {
		t.Errorf("ionconfig missing X-band center frequency (8.4 GHz)")
	}
	if !strings.Contains(ionconfigStr, "FEC_TYPE=Turbo") {
		t.Errorf("ionconfig missing Turbo FEC type for X-band")
	}
	if !strings.Contains(ionconfigStr, "X-band 8.4 GHz") {
		t.Errorf("ionconfig missing X-band description")
	}
	if strings.Contains(ionconfigStr, "TELEMETRY_ENABLED=true") {
		t.Errorf("ionconfig should not have telemetry enabled")
	}
}

func TestGenerateCislunarWithCGRPrediction(t *testing.T) {
	tmpDir := t.TempDir()

	epoch := time.Now()
	orbitalParams := &OrbitalParameters{
		Epoch:           epoch,
		SemiMajorAxisM:  384400000.0,
		Eccentricity:    0.05,
		InclinationDeg:  5.0,
		RAANDeg:         0.0,
		ArgPeriapsisDeg: 0.0,
		TrueAnomalyDeg:  0.0,
	}

	predictedContacts := []CislunarContact{
		{
			RemoteNodeNumber: 300,
			RemoteCallsign:   "N1GS",
			StartTime:        epoch.Add(6 * time.Hour),
			Duration:         4 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  70.0,
			Confidence:       0.90,
			LightTimeDelay:   1.28,
		},
	}

	err := GenerateCislunarWithCGRPrediction(
		"cislunar-cgr",
		30,
		"W0CGR",
		SBand,
		orbitalParams,
		predictedContacts,
		tmpDir,
	)
	if err != nil {
		t.Fatalf("GenerateCislunarWithCGRPrediction failed: %v", err)
	}

	// Verify config was created
	ionconfigPath := filepath.Join(tmpDir, "sband.ionconfig")
	if _, err := os.Stat(ionconfigPath); os.IsNotExist(err) {
		t.Errorf("Expected sband.ionconfig was not created")
	}

	// Verify CGR-predicted contact is in config
	ltprcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ltprc"))
	if err != nil {
		t.Fatalf("Failed to read node.ltprc: %v", err)
	}

	ltprcStr := string(ltprcContent)
	if !strings.Contains(ltprcStr, "a span 300") {
		t.Errorf("ltprc missing CGR-predicted contact span")
	}
	if !strings.Contains(ltprcStr, "Confidence: 0.9") {
		t.Errorf("ltprc missing confidence value")
	}
}

func TestCislunarConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		config      CislunarNodeConfig
		expectError bool
	}{
		{
			name: "valid S-band config",
			config: CislunarNodeConfig{
				NodeID:       "valid-sband",
				NodeNumber:   1,
				Callsign:     "W1ABC",
				StorageBytes: 512 * 1024 * 1024,
				SRAMBytes:    786 * 1024,
				Band:         SBand,
				ContactPlan:  []CislunarContact{},
			},
			expectError: false,
		},
		{
			name: "valid X-band config",
			config: CislunarNodeConfig{
				NodeID:       "valid-xband",
				NodeNumber:   2,
				Callsign:     "K2XYZ",
				StorageBytes: 1024 * 1024 * 1024,
				SRAMBytes:    2 * 1024 * 1024,
				Band:         XBand,
				ContactPlan:  []CislunarContact{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := GenerateCislunarConfig(tt.config, testDir)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCislunarLightTimeDelayConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	epoch := time.Now()
	contacts := []CislunarContact{
		{
			RemoteNodeNumber: 400,
			RemoteCallsign:   "W4LTD",
			StartTime:        epoch,
			Duration:         1 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  45.0,
			Confidence:       0.85,
			LightTimeDelay:   1.5, // 1.5 seconds
		},
		{
			RemoteNodeNumber: 401,
			RemoteCallsign:   "K4LTD",
			StartTime:        epoch.Add(6 * time.Hour),
			Duration:         2 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  50.0,
			Confidence:       0.80,
			LightTimeDelay:   2.0, // 2.0 seconds (max for cislunar)
		},
	}

	config := CislunarNodeConfig{
		NodeID:       "cislunar-ltd",
		NodeNumber:   40,
		Callsign:     "N4LTD",
		StorageBytes: 512 * 1024 * 1024,
		SRAMBytes:    786 * 1024,
		Band:         SBand,
		ContactPlan:  contacts,
	}

	err := GenerateCislunarConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateCislunarConfig failed: %v", err)
	}

	// Verify light-time delay is properly configured in ltprc
	ltprcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ltprc"))
	if err != nil {
		t.Fatalf("Failed to read node.ltprc: %v", err)
	}

	ltprcStr := string(ltprcContent)

	// Check first contact: 1.5s delay, 3.0s RTT, 13.0s timeout
	if !strings.Contains(ltprcStr, "Light-time delay: 1.5s") {
		t.Errorf("ltprc missing 1.5s light-time delay for first contact")
	}
	if !strings.Contains(ltprcStr, "RTT: 3s") {
		t.Errorf("ltprc missing 3s RTT for first contact")
	}
	if !strings.Contains(ltprcStr, "LTP timeout configured for deep-space delay (13s)") {
		t.Errorf("ltprc missing 13s LTP timeout for first contact")
	}

	// Check second contact: 2.0s delay, 4.0s RTT, 14.0s timeout
	if !strings.Contains(ltprcStr, "Light-time delay: 2s") {
		t.Errorf("ltprc missing 2s light-time delay for second contact")
	}
	if !strings.Contains(ltprcStr, "RTT: 4s") {
		t.Errorf("ltprc missing 4s RTT for second contact")
	}
	if !strings.Contains(ltprcStr, "LTP timeout configured for deep-space delay (14s)") {
		t.Errorf("ltprc missing 14s LTP timeout for second contact")
	}

	// Verify average light-time delay in ionconfig
	ionconfigContent, err := os.ReadFile(filepath.Join(tmpDir, "sband.ionconfig"))
	if err != nil {
		t.Fatalf("Failed to read sband.ionconfig: %v", err)
	}

	ionconfigStr := string(ionconfigContent)
	// Average of 1.5 and 2.0 is 1.75
	if !strings.Contains(ionconfigStr, "LIGHT_TIME_DELAY=1.75") {
		t.Errorf("ionconfig missing average light-time delay (1.75s)")
	}
}

func TestCislunarLongDurationStorage(t *testing.T) {
	tmpDir := t.TempDir()

	epoch := time.Now()
	contacts := []CislunarContact{
		{
			RemoteNodeNumber: 500,
			RemoteCallsign:   "W5LDS",
			StartTime:        epoch,
			Duration:         6 * time.Hour, // Long 6-hour contact window
			DataRate:         500,
			MaxElevationDeg:  80.0,
			Confidence:       0.95,
			LightTimeDelay:   1.28,
		},
	}

	config := CislunarNodeConfig{
		NodeID:       "cislunar-lds",
		NodeNumber:   50,
		Callsign:     "N5LDS",
		StorageBytes: 1024 * 1024 * 1024, // 1 GB for long-duration storage
		SRAMBytes:    786 * 1024,
		Band:         SBand,
		ContactPlan:  contacts,
	}

	err := GenerateCislunarConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateCislunarConfig failed: %v", err)
	}

	// Verify long-duration storage configuration
	ionconfigContent, err := os.ReadFile(filepath.Join(tmpDir, "sband.ionconfig"))
	if err != nil {
		t.Fatalf("Failed to read sband.ionconfig: %v", err)
	}

	ionconfigStr := string(ionconfigContent)
	if !strings.Contains(ionconfigStr, "STORAGE_BYTES=1073741824") {
		t.Errorf("ionconfig missing 1 GB storage configuration")
	}
	if !strings.Contains(ionconfigStr, "LONG_DURATION_STORAGE=true") {
		t.Errorf("ionconfig missing long-duration storage flag")
	}
	if !strings.Contains(ionconfigStr, "EXTENDED_CONTACT_GAPS=true") {
		t.Errorf("ionconfig missing extended contact gaps flag")
	}

	// Verify LTP configuration for long-duration contacts
	ltprcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ltprc"))
	if err != nil {
		t.Fatalf("Failed to read node.ltprc: %v", err)
	}

	ltprcStr := string(ltprcContent)
	// Check for larger block size (50000 vs 10000 for LEO)
	if !strings.Contains(ltprcStr, "a span 500 50000") {
		t.Errorf("ltprc missing larger block size for long-duration contacts")
	}
}

func TestCislunarConfidenceDegradation(t *testing.T) {
	tmpDir := t.TempDir()

	epoch := time.Now()
	contacts := []CislunarContact{
		{
			RemoteNodeNumber: 600,
			RemoteCallsign:   "W6CD1",
			StartTime:        epoch.Add(1 * time.Hour),
			Duration:         1 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  45.0,
			Confidence:       0.95, // High confidence (near epoch)
			LightTimeDelay:   1.28,
		},
		{
			RemoteNodeNumber: 601,
			RemoteCallsign:   "W6CD2",
			StartTime:        epoch.Add(3 * 24 * time.Hour), // 3 days later
			Duration:         1 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  45.0,
			Confidence:       0.55, // Lower confidence (far from epoch)
			LightTimeDelay:   1.28,
		},
		{
			RemoteNodeNumber: 602,
			RemoteCallsign:   "W6CD3",
			StartTime:        epoch.Add(7 * 24 * time.Hour), // 7 days later
			Duration:         1 * time.Hour,
			DataRate:         500,
			MaxElevationDeg:  45.0,
			Confidence:       0.25, // Very low confidence (very far from epoch)
			LightTimeDelay:   1.28,
		},
	}

	config := CislunarNodeConfig{
		NodeID:       "cislunar-cd",
		NodeNumber:   60,
		Callsign:     "N6CD",
		StorageBytes: 512 * 1024 * 1024,
		SRAMBytes:    786 * 1024,
		Band:         SBand,
		ContactPlan:  contacts,
	}

	err := GenerateCislunarConfig(config, tmpDir)
	if err != nil {
		t.Fatalf("GenerateCislunarConfig failed: %v", err)
	}

	// Verify confidence values are properly documented
	ipnrcContent, err := os.ReadFile(filepath.Join(tmpDir, "node.ipnrc"))
	if err != nil {
		t.Fatalf("Failed to read node.ipnrc: %v", err)
	}

	ipnrcStr := string(ipnrcContent)
	if !strings.Contains(ipnrcStr, "Confidence: 0.95") {
		t.Errorf("ipnrc missing high confidence for near-epoch contact")
	}
	if !strings.Contains(ipnrcStr, "Confidence: 0.55") {
		t.Errorf("ipnrc missing medium confidence for 3-day contact")
	}
	if !strings.Contains(ipnrcStr, "Confidence: 0.25") {
		t.Errorf("ipnrc missing low confidence for 7-day contact")
	}
	if !strings.Contains(ipnrcStr, "degrades faster for cislunar") {
		t.Errorf("ipnrc missing confidence degradation comment")
	}
}
