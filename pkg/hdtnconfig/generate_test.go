package hdtnconfig

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func validTerrestrialOpts() TerrestrialOpts {
	return TerrestrialOpts{
		NodeNumber:       1,
		NodeName:         "node-a",
		Callsign:         "W1AW",
		StoragePath:      "/var/hdtn/storage",
		TNCDevice:        "/dev/ttyUSB0",
		TNCBaudRate:      9600,
		UDPLocalPort:     4556,
		UDPRemoteHost:    "192.168.1.2",
		UDPRemotePort:    4556,
		RemoteNodeNumber: 2,
		ContactDataRate:  9600,
	}
}

func TestGenerateTerrestrialConfig_ValidOpts(t *testing.T) {
	opts := validTerrestrialOpts()
	cfg, err := GenerateTerrestrialConfig(opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check EID format
	expectedEID := fmt.Sprintf("ipn:%d.0", opts.NodeNumber)
	if cfg.MyDtnEidStr != expectedEID {
		t.Errorf("expected EID %q, got %q", expectedEID, cfg.MyDtnEidStr)
	}

	// Check node ID
	if cfg.MyNodeID != opts.NodeNumber {
		t.Errorf("expected MyNodeID %d, got %d", opts.NodeNumber, cfg.MyNodeID)
	}

	// Check storage path
	if cfg.StoragePath != opts.StoragePath {
		t.Errorf("expected StoragePath %q, got %q", opts.StoragePath, cfg.StoragePath)
	}

	// Check inducts: should have LTP-over-UDP and KISS
	if len(cfg.InductsConfig.InductVector) != 2 {
		t.Fatalf("expected 2 inducts, got %d", len(cfg.InductsConfig.InductVector))
	}
	udpInduct := cfg.InductsConfig.InductVector[0]
	if udpInduct.ConvergenceLayer != "ltp_over_udp" {
		t.Errorf("expected first induct to be ltp_over_udp, got %q", udpInduct.ConvergenceLayer)
	}
	if udpInduct.BoundPort != opts.UDPLocalPort {
		t.Errorf("expected BoundPort %d, got %d", opts.UDPLocalPort, udpInduct.BoundPort)
	}
	kissInduct := cfg.InductsConfig.InductVector[1]
	if kissInduct.ConvergenceLayer != "kiss" {
		t.Errorf("expected second induct to be kiss, got %q", kissInduct.ConvergenceLayer)
	}
	if kissInduct.KissTncDevice != opts.TNCDevice {
		t.Errorf("expected KissTncDevice %q, got %q", opts.TNCDevice, kissInduct.KissTncDevice)
	}
	if kissInduct.KissBaudRate != opts.TNCBaudRate {
		t.Errorf("expected KissBaudRate %d, got %d", opts.TNCBaudRate, kissInduct.KissBaudRate)
	}

	// Check outducts: should have LTP-over-UDP and KISS
	if len(cfg.OutductsConfig.OutductVector) != 2 {
		t.Fatalf("expected 2 outducts, got %d", len(cfg.OutductsConfig.OutductVector))
	}
	udpOutduct := cfg.OutductsConfig.OutductVector[0]
	if udpOutduct.ConvergenceLayer != "ltp_over_udp" {
		t.Errorf("expected first outduct to be ltp_over_udp, got %q", udpOutduct.ConvergenceLayer)
	}
	if udpOutduct.RemoteHostname != opts.UDPRemoteHost {
		t.Errorf("expected RemoteHostname %q, got %q", opts.UDPRemoteHost, udpOutduct.RemoteHostname)
	}
	if udpOutduct.RemotePort != opts.UDPRemotePort {
		t.Errorf("expected RemotePort %d, got %d", opts.UDPRemotePort, udpOutduct.RemotePort)
	}
	kissOutduct := cfg.OutductsConfig.OutductVector[1]
	if kissOutduct.ConvergenceLayer != "kiss" {
		t.Errorf("expected second outduct to be kiss, got %q", kissOutduct.ConvergenceLayer)
	}
	if kissOutduct.KissTncDevice != opts.TNCDevice {
		t.Errorf("expected KissTncDevice %q, got %q", opts.TNCDevice, kissOutduct.KissTncDevice)
	}

	// Check contact plan
	if len(cfg.ContactPlanJSON.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(cfg.ContactPlanJSON.Contacts))
	}
	contact := cfg.ContactPlanJSON.Contacts[0]
	if contact.Source != opts.NodeNumber {
		t.Errorf("expected contact Source %d, got %d", opts.NodeNumber, contact.Source)
	}
	if contact.Dest != opts.RemoteNodeNumber {
		t.Errorf("expected contact Dest %d, got %d", opts.RemoteNodeNumber, contact.Dest)
	}
	if contact.RateBitsPerSec != opts.ContactDataRate {
		t.Errorf("expected contact RateBitsPerSec %d, got %d", opts.ContactDataRate, contact.RateBitsPerSec)
	}

	// Check LTP engine IDs
	if udpInduct.ThisLtpEngineID != uint64(opts.NodeNumber) {
		t.Errorf("expected UDP induct ThisLtpEngineID %d, got %d", opts.NodeNumber, udpInduct.ThisLtpEngineID)
	}
	if udpInduct.RemoteLtpEngineID != uint64(opts.RemoteNodeNumber) {
		t.Errorf("expected UDP induct RemoteLtpEngineID %d, got %d", opts.RemoteNodeNumber, udpInduct.RemoteLtpEngineID)
	}
}

func TestGenerateTerrestrialConfig_NodeB(t *testing.T) {
	opts := TerrestrialOpts{
		NodeNumber:       2,
		NodeName:         "node-b",
		Callsign:         "W2AW",
		StoragePath:      "/var/hdtn/storage-b",
		TNCDevice:        "/dev/ttyUSB1",
		TNCBaudRate:      9600,
		UDPLocalPort:     4557,
		UDPRemoteHost:    "192.168.1.1",
		UDPRemotePort:    4556,
		RemoteNodeNumber: 1,
		ContactDataRate:  9600,
	}

	cfg, err := GenerateTerrestrialConfig(opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedEID := "ipn:2.0"
	if cfg.MyDtnEidStr != expectedEID {
		t.Errorf("expected EID %q, got %q", expectedEID, cfg.MyDtnEidStr)
	}
	if cfg.MyNodeID != 2 {
		t.Errorf("expected MyNodeID 2, got %d", cfg.MyNodeID)
	}
}

func TestGenerateTerrestrialConfig_InvalidNodeNumber(t *testing.T) {
	opts := validTerrestrialOpts()
	opts.NodeNumber = 0

	_, err := GenerateTerrestrialConfig(opts)
	if err == nil {
		t.Fatal("expected error for node number 0")
	}
}

func TestGenerateTerrestrialConfig_EmptyStoragePath(t *testing.T) {
	opts := validTerrestrialOpts()
	opts.StoragePath = ""

	_, err := GenerateTerrestrialConfig(opts)
	if err == nil {
		t.Fatal("expected error for empty storage path")
	}
}

func TestGenerateTerrestrialConfig_EIDFormat(t *testing.T) {
	testCases := []struct {
		nodeNumber int
		expected   string
	}{
		{1, "ipn:1.0"},
		{2, "ipn:2.0"},
		{42, "ipn:42.0"},
		{100, "ipn:100.0"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("node_%d", tc.nodeNumber), func(t *testing.T) {
			opts := validTerrestrialOpts()
			opts.NodeNumber = tc.nodeNumber
			cfg, err := GenerateTerrestrialConfig(opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.MyDtnEidStr != tc.expected {
				t.Errorf("expected EID %q, got %q", tc.expected, cfg.MyDtnEidStr)
			}
		})
	}
}

// --- Task 1.7: Unit tests for config generation ---

// TestGenerateTerrestrialConfig_NodeA_JSONStructure verifies that the generated
// config for node-a produces valid JSON with the expected structure.
func TestGenerateTerrestrialConfig_NodeA_JSONStructure(t *testing.T) {
	opts := TerrestrialOpts{
		NodeNumber:       1,
		NodeName:         "node-a",
		Callsign:         "W1AW",
		StoragePath:      "/var/hdtn/storage",
		TNCDevice:        "/dev/ttyUSB0",
		TNCBaudRate:      9600,
		UDPLocalPort:     4556,
		UDPRemoteHost:    "192.168.1.2",
		UDPRemotePort:    4556,
		RemoteNodeNumber: 2,
		ContactDataRate:  9600,
	}

	cfg, err := GenerateTerrestrialConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	// Parse as generic map to verify JSON structure
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Verify top-level fields
	if jsonMap["hdtnConfigName"] != "node-a" {
		t.Errorf("expected hdtnConfigName 'node-a', got %v", jsonMap["hdtnConfigName"])
	}
	if jsonMap["myNodeId"].(float64) != 1 {
		t.Errorf("expected myNodeId 1, got %v", jsonMap["myNodeId"])
	}
	if jsonMap["mySchemeStr"] != "ipn" {
		t.Errorf("expected mySchemeStr 'ipn', got %v", jsonMap["mySchemeStr"])
	}
	if jsonMap["myDtnEidStr"] != "ipn:1.0" {
		t.Errorf("expected myDtnEidStr 'ipn:1.0', got %v", jsonMap["myDtnEidStr"])
	}
	if jsonMap["storagePath"] != "/var/hdtn/storage" {
		t.Errorf("expected storagePath '/var/hdtn/storage', got %v", jsonMap["storagePath"])
	}

	// Verify inductsConfig structure
	inductsConfig, ok := jsonMap["inductsConfig"].(map[string]interface{})
	if !ok {
		t.Fatal("inductsConfig is not an object")
	}
	inductVector, ok := inductsConfig["inductVector"].([]interface{})
	if !ok {
		t.Fatal("inductVector is not an array")
	}
	if len(inductVector) != 2 {
		t.Fatalf("expected 2 inducts, got %d", len(inductVector))
	}

	// First induct should be ltp_over_udp
	udpInduct := inductVector[0].(map[string]interface{})
	if udpInduct["convergenceLayer"] != "ltp_over_udp" {
		t.Errorf("expected first induct convergenceLayer 'ltp_over_udp', got %v", udpInduct["convergenceLayer"])
	}
	if udpInduct["boundPort"].(float64) != 4556 {
		t.Errorf("expected boundPort 4556, got %v", udpInduct["boundPort"])
	}

	// Second induct should be kiss
	kissInduct := inductVector[1].(map[string]interface{})
	if kissInduct["convergenceLayer"] != "kiss" {
		t.Errorf("expected second induct convergenceLayer 'kiss', got %v", kissInduct["convergenceLayer"])
	}
	if kissInduct["kissTncDevice"] != "/dev/ttyUSB0" {
		t.Errorf("expected kissTncDevice '/dev/ttyUSB0', got %v", kissInduct["kissTncDevice"])
	}

	// Verify outductsConfig structure
	outductsConfig, ok := jsonMap["outductsConfig"].(map[string]interface{})
	if !ok {
		t.Fatal("outductsConfig is not an object")
	}
	outductVector, ok := outductsConfig["outductVector"].([]interface{})
	if !ok {
		t.Fatal("outductVector is not an array")
	}
	if len(outductVector) != 2 {
		t.Fatalf("expected 2 outducts, got %d", len(outductVector))
	}

	// Verify contactPlanJson structure
	contactPlan, ok := jsonMap["contactPlanJson"].(map[string]interface{})
	if !ok {
		t.Fatal("contactPlanJson is not an object")
	}
	contacts, ok := contactPlan["contacts"].([]interface{})
	if !ok {
		t.Fatal("contacts is not an array")
	}
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}
	contact := contacts[0].(map[string]interface{})
	if contact["source"].(float64) != 1 {
		t.Errorf("expected contact source 1, got %v", contact["source"])
	}
	if contact["dest"].(float64) != 2 {
		t.Errorf("expected contact dest 2, got %v", contact["dest"])
	}
	if contact["rateBitsPerSec"].(float64) != 9600 {
		t.Errorf("expected contact rateBitsPerSec 9600, got %v", contact["rateBitsPerSec"])
	}
}

// TestGenerateTerrestrialConfig_NodeB_JSONStructure verifies that the generated
// config for node-b produces valid JSON with the expected structure.
func TestGenerateTerrestrialConfig_NodeB_JSONStructure(t *testing.T) {
	opts := TerrestrialOpts{
		NodeNumber:       2,
		NodeName:         "node-b",
		Callsign:         "W2AW",
		StoragePath:      "/var/hdtn/storage-b",
		TNCDevice:        "/dev/ttyUSB1",
		TNCBaudRate:      9600,
		UDPLocalPort:     4557,
		UDPRemoteHost:    "192.168.1.1",
		UDPRemotePort:    4556,
		RemoteNodeNumber: 1,
		ContactDataRate:  9600,
	}

	cfg, err := GenerateTerrestrialConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	// Parse as generic map to verify JSON structure
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Verify top-level fields for node-b
	if jsonMap["hdtnConfigName"] != "node-b" {
		t.Errorf("expected hdtnConfigName 'node-b', got %v", jsonMap["hdtnConfigName"])
	}
	if jsonMap["myNodeId"].(float64) != 2 {
		t.Errorf("expected myNodeId 2, got %v", jsonMap["myNodeId"])
	}
	if jsonMap["myDtnEidStr"] != "ipn:2.0" {
		t.Errorf("expected myDtnEidStr 'ipn:2.0', got %v", jsonMap["myDtnEidStr"])
	}
	if jsonMap["storagePath"] != "/var/hdtn/storage-b" {
		t.Errorf("expected storagePath '/var/hdtn/storage-b', got %v", jsonMap["storagePath"])
	}

	// Verify outduct points back to node 1
	outductsConfig := jsonMap["outductsConfig"].(map[string]interface{})
	outductVector := outductsConfig["outductVector"].([]interface{})
	udpOutduct := outductVector[0].(map[string]interface{})
	if udpOutduct["remoteHostname"] != "192.168.1.1" {
		t.Errorf("expected remoteHostname '192.168.1.1', got %v", udpOutduct["remoteHostname"])
	}
	if udpOutduct["nextHopNodeId"].(float64) != 1 {
		t.Errorf("expected nextHopNodeId 1, got %v", udpOutduct["nextHopNodeId"])
	}

	// Verify contact plan points from node 2 to node 1
	contactPlan := jsonMap["contactPlanJson"].(map[string]interface{})
	contacts := contactPlan["contacts"].([]interface{})
	contact := contacts[0].(map[string]interface{})
	if contact["source"].(float64) != 2 {
		t.Errorf("expected contact source 2, got %v", contact["source"])
	}
	if contact["dest"].(float64) != 1 {
		t.Errorf("expected contact dest 1, got %v", contact["dest"])
	}

	// Verify KISS induct has correct TNC device for node-b
	inductsConfig := jsonMap["inductsConfig"].(map[string]interface{})
	inductVector := inductsConfig["inductVector"].([]interface{})
	kissInduct := inductVector[1].(map[string]interface{})
	if kissInduct["kissTncDevice"] != "/dev/ttyUSB1" {
		t.Errorf("expected kissTncDevice '/dev/ttyUSB1', got %v", kissInduct["kissTncDevice"])
	}
}

// TestValidationErrorMessages verifies that each type of invalid field
// produces a descriptive error message containing the field name.
func TestValidationErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(*HDTNConfig)
		expectedField string
	}{
		{
			name:          "node ID zero",
			mutate:        func(c *HDTNConfig) { c.MyNodeID = 0 },
			expectedField: "myNodeId",
		},
		{
			name:          "node ID negative",
			mutate:        func(c *HDTNConfig) { c.MyNodeID = -5 },
			expectedField: "myNodeId",
		},
		{
			name:          "empty storage path",
			mutate:        func(c *HDTNConfig) { c.StoragePath = "" },
			expectedField: "storagePath",
		},
		{
			name:          "nil induct vector",
			mutate:        func(c *HDTNConfig) { c.InductsConfig.InductVector = nil },
			expectedField: "inductVector",
		},
		{
			name:          "empty induct vector",
			mutate:        func(c *HDTNConfig) { c.InductsConfig.InductVector = []Induct{} },
			expectedField: "inductVector",
		},
		{
			name:          "nil outduct vector",
			mutate:        func(c *HDTNConfig) { c.OutductsConfig.OutductVector = nil },
			expectedField: "outductVector",
		},
		{
			name:          "empty outduct vector",
			mutate:        func(c *HDTNConfig) { c.OutductsConfig.OutductVector = []Outduct{} },
			expectedField: "outductVector",
		},
		{
			name:          "nil contacts",
			mutate:        func(c *HDTNConfig) { c.ContactPlanJSON.Contacts = nil },
			expectedField: "contacts",
		},
		{
			name:          "empty contacts",
			mutate:        func(c *HDTNConfig) { c.ContactPlanJSON.Contacts = []ContactEntry{} },
			expectedField: "contacts",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfig()
			tc.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected validation error for %s, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.expectedField) {
				t.Fatalf("expected error to contain %q, got: %v", tc.expectedField, err)
			}
		})
	}
}

// TestGenerateTerrestrialConfig_JSONParseable verifies that generated configs
// produce JSON that is parseable without error and contains no null required fields.
func TestGenerateTerrestrialConfig_JSONParseable(t *testing.T) {
	testCases := []struct {
		name string
		opts TerrestrialOpts
	}{
		{
			name: "node-a",
			opts: TerrestrialOpts{
				NodeNumber:       1,
				NodeName:         "node-a",
				Callsign:         "W1AW",
				StoragePath:      "/var/hdtn/storage",
				TNCDevice:        "/dev/ttyUSB0",
				TNCBaudRate:      9600,
				UDPLocalPort:     4556,
				UDPRemoteHost:    "192.168.1.2",
				UDPRemotePort:    4556,
				RemoteNodeNumber: 2,
				ContactDataRate:  9600,
			},
		},
		{
			name: "node-b",
			opts: TerrestrialOpts{
				NodeNumber:       2,
				NodeName:         "node-b",
				Callsign:         "W2AW",
				StoragePath:      "/var/hdtn/storage-b",
				TNCDevice:        "/dev/ttyUSB1",
				TNCBaudRate:      9600,
				UDPLocalPort:     4557,
				UDPRemoteHost:    "192.168.1.1",
				UDPRemotePort:    4556,
				RemoteNodeNumber: 1,
				ContactDataRate:  9600,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GenerateTerrestrialConfig(tc.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			data, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Verify JSON is parseable
			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("generated JSON is not parseable: %v", err)
			}

			// Verify required fields are not null
			requiredFields := []string{"hdtnConfigName", "myNodeId", "mySchemeStr", "myDtnEidStr", "storagePath", "inductsConfig", "outductsConfig", "contactPlanJson"}
			for _, field := range requiredFields {
				val, exists := parsed[field]
				if !exists {
					t.Errorf("required field %q missing from JSON", field)
				}
				if val == nil {
					t.Errorf("required field %q is null in JSON", field)
				}
			}

			// Verify EID format
			expectedEID := fmt.Sprintf("ipn:%d.0", tc.opts.NodeNumber)
			if parsed["myDtnEidStr"] != expectedEID {
				t.Errorf("expected EID %q, got %v", expectedEID, parsed["myDtnEidStr"])
			}
		})
	}
}
