package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConfigFromYAML(t *testing.T) {
	// Create a temporary YAML config file with HDTN fields
	content := `
node_id: "ground-station-1"
node_number: 1
callsign: "W1AW"
hdtn_binary: "/usr/local/bin/hdtn-one-process"
hdtn_config: "/etc/hdtn/hdtn-config.json"
contact_plan_file: "/etc/hdtn/contacts.yaml"
telemetry_port: 9090
telemetry_file: "/var/log/hdtn-telemetry.json"
health_interval: 15
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set the flag to point to our test config
	*configFile = configPath
	// Clear other flags to avoid interference
	*nodeID = ""
	*nodeNumber = 0
	*hdtnBinary = ""
	*hdtnConfig = ""
	*telemetryPort = 0

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() returned error: %v", err)
	}

	if config.NodeID != "ground-station-1" {
		t.Errorf("NodeID = %q, want %q", config.NodeID, "ground-station-1")
	}
	if config.NodeNumber != 1 {
		t.Errorf("NodeNumber = %d, want %d", config.NodeNumber, 1)
	}
	if config.Callsign != "W1AW" {
		t.Errorf("Callsign = %q, want %q", config.Callsign, "W1AW")
	}
	if config.HDTNBinary != "/usr/local/bin/hdtn-one-process" {
		t.Errorf("HDTNBinary = %q, want %q", config.HDTNBinary, "/usr/local/bin/hdtn-one-process")
	}
	if config.HDTNConfig != "/etc/hdtn/hdtn-config.json" {
		t.Errorf("HDTNConfig = %q, want %q", config.HDTNConfig, "/etc/hdtn/hdtn-config.json")
	}
	if config.ContactPlanFile != "/etc/hdtn/contacts.yaml" {
		t.Errorf("ContactPlanFile = %q, want %q", config.ContactPlanFile, "/etc/hdtn/contacts.yaml")
	}
	if config.TelemetryPort != 9090 {
		t.Errorf("TelemetryPort = %d, want %d", config.TelemetryPort, 9090)
	}
	if config.TelemetryFile != "/var/log/hdtn-telemetry.json" {
		t.Errorf("TelemetryFile = %q, want %q", config.TelemetryFile, "/var/log/hdtn-telemetry.json")
	}
	if config.HealthInterval != 15 {
		t.Errorf("HealthInterval = %d, want %d", config.HealthInterval, 15)
	}
}

func TestLoadConfigFromJSON(t *testing.T) {
	// Create a temporary JSON config file with HDTN fields
	cfg := Config{
		NodeID:          "relay-node-2",
		NodeNumber:      2,
		Callsign:        "KD2ABC",
		HDTNBinary:      "/opt/hdtn/bin/hdtn-one-process",
		HDTNConfig:      "/opt/hdtn/config/hdtn-config.json",
		ContactPlanFile: "/opt/hdtn/contacts.json",
		TelemetryPort:   8081,
		TelemetryFile:   "/tmp/telemetry.json",
		HealthInterval:  30,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set the flag to point to our test config
	*configFile = configPath
	*nodeID = ""
	*nodeNumber = 0
	*hdtnBinary = ""
	*hdtnConfig = ""
	*telemetryPort = 0

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() returned error: %v", err)
	}

	if config.NodeID != "relay-node-2" {
		t.Errorf("NodeID = %q, want %q", config.NodeID, "relay-node-2")
	}
	if config.NodeNumber != 2 {
		t.Errorf("NodeNumber = %d, want %d", config.NodeNumber, 2)
	}
	if config.Callsign != "KD2ABC" {
		t.Errorf("Callsign = %q, want %q", config.Callsign, "KD2ABC")
	}
	if config.HDTNBinary != "/opt/hdtn/bin/hdtn-one-process" {
		t.Errorf("HDTNBinary = %q, want %q", config.HDTNBinary, "/opt/hdtn/bin/hdtn-one-process")
	}
	if config.HDTNConfig != "/opt/hdtn/config/hdtn-config.json" {
		t.Errorf("HDTNConfig = %q, want %q", config.HDTNConfig, "/opt/hdtn/config/hdtn-config.json")
	}
	if config.ContactPlanFile != "/opt/hdtn/contacts.json" {
		t.Errorf("ContactPlanFile = %q, want %q", config.ContactPlanFile, "/opt/hdtn/contacts.json")
	}
	if config.TelemetryPort != 8081 {
		t.Errorf("TelemetryPort = %d, want %d", config.TelemetryPort, 8081)
	}
	if config.HealthInterval != 30 {
		t.Errorf("HealthInterval = %d, want %d", config.HealthInterval, 30)
	}
}

func TestLoadConfigValidationRejectsMissingNodeID(t *testing.T) {
	content := `
node_number: 1
hdtn_binary: "/usr/local/bin/hdtn-one-process"
hdtn_config: "/etc/hdtn/hdtn-config.json"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	*configFile = configPath
	*nodeID = ""
	*nodeNumber = 0
	*hdtnBinary = ""
	*hdtnConfig = ""
	*telemetryPort = 0

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() should return error for missing node_id")
	}
	if got := err.Error(); got != "node_id is required" {
		t.Errorf("error = %q, want %q", got, "node_id is required")
	}
}

func TestLoadConfigValidationRejectsMissingNodeNumber(t *testing.T) {
	content := `
node_id: "test-node"
hdtn_binary: "/usr/local/bin/hdtn-one-process"
hdtn_config: "/etc/hdtn/hdtn-config.json"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	*configFile = configPath
	*nodeID = ""
	*nodeNumber = 0
	*hdtnBinary = ""
	*hdtnConfig = ""
	*telemetryPort = 0

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() should return error for missing node_number")
	}
	if got := err.Error(); got != "node_number is required and must be positive" {
		t.Errorf("error = %q, want %q", got, "node_number is required and must be positive")
	}
}

func TestLoadConfigValidationRejectsMissingHDTNBinary(t *testing.T) {
	content := `
node_id: "test-node"
node_number: 1
hdtn_config: "/etc/hdtn/hdtn-config.json"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	*configFile = configPath
	*nodeID = ""
	*nodeNumber = 0
	*hdtnBinary = ""
	*hdtnConfig = ""
	*telemetryPort = 0

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() should return error for missing hdtn_binary")
	}
	if got := err.Error(); got != "hdtn_binary is required" {
		t.Errorf("error = %q, want %q", got, "hdtn_binary is required")
	}
}

func TestLoadConfigValidationRejectsMissingHDTNConfig(t *testing.T) {
	content := `
node_id: "test-node"
node_number: 1
hdtn_binary: "/usr/local/bin/hdtn-one-process"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	*configFile = configPath
	*nodeID = ""
	*nodeNumber = 0
	*hdtnBinary = ""
	*hdtnConfig = ""
	*telemetryPort = 0

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() should return error for missing hdtn_config")
	}
	if got := err.Error(); got != "hdtn_config is required" {
		t.Errorf("error = %q, want %q", got, "hdtn_config is required")
	}
}

func TestLoadConfigYAMLFieldMapping(t *testing.T) {
	// Verify that the YAML struct tags correctly map to the Config fields
	content := `
node_id: "yaml-test"
node_number: 42
callsign: "N0CALL"
hdtn_binary: "/bin/hdtn"
hdtn_config: "/etc/hdtn.json"
contact_plan_file: "/contacts.yaml"
telemetry_port: 7070
telemetry_file: "/tmp/tel.json"
health_interval: 5
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if cfg.NodeID != "yaml-test" {
		t.Errorf("NodeID = %q, want %q", cfg.NodeID, "yaml-test")
	}
	if cfg.NodeNumber != 42 {
		t.Errorf("NodeNumber = %d, want %d", cfg.NodeNumber, 42)
	}
	if cfg.Callsign != "N0CALL" {
		t.Errorf("Callsign = %q, want %q", cfg.Callsign, "N0CALL")
	}
	if cfg.HDTNBinary != "/bin/hdtn" {
		t.Errorf("HDTNBinary = %q, want %q", cfg.HDTNBinary, "/bin/hdtn")
	}
	if cfg.HDTNConfig != "/etc/hdtn.json" {
		t.Errorf("HDTNConfig = %q, want %q", cfg.HDTNConfig, "/etc/hdtn.json")
	}
	if cfg.ContactPlanFile != "/contacts.yaml" {
		t.Errorf("ContactPlanFile = %q, want %q", cfg.ContactPlanFile, "/contacts.yaml")
	}
	if cfg.TelemetryPort != 7070 {
		t.Errorf("TelemetryPort = %d, want %d", cfg.TelemetryPort, 7070)
	}
	if cfg.TelemetryFile != "/tmp/tel.json" {
		t.Errorf("TelemetryFile = %q, want %q", cfg.TelemetryFile, "/tmp/tel.json")
	}
	if cfg.HealthInterval != 5 {
		t.Errorf("HealthInterval = %d, want %d", cfg.HealthInterval, 5)
	}
}

func TestLoadConfigJSONFieldMapping(t *testing.T) {
	// Verify that the JSON struct tags correctly map to the Config fields
	content := `{
		"node_id": "json-test",
		"node_number": 99,
		"callsign": "KB1ABC",
		"hdtn_binary": "/usr/bin/hdtn",
		"hdtn_config": "/config/hdtn.json",
		"contact_plan_file": "/plan.json",
		"telemetry_port": 6060,
		"telemetry_file": "/var/tel.json",
		"health_interval": 20
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if cfg.NodeID != "json-test" {
		t.Errorf("NodeID = %q, want %q", cfg.NodeID, "json-test")
	}
	if cfg.NodeNumber != 99 {
		t.Errorf("NodeNumber = %d, want %d", cfg.NodeNumber, 99)
	}
	if cfg.Callsign != "KB1ABC" {
		t.Errorf("Callsign = %q, want %q", cfg.Callsign, "KB1ABC")
	}
	if cfg.HDTNBinary != "/usr/bin/hdtn" {
		t.Errorf("HDTNBinary = %q, want %q", cfg.HDTNBinary, "/usr/bin/hdtn")
	}
	if cfg.HDTNConfig != "/config/hdtn.json" {
		t.Errorf("HDTNConfig = %q, want %q", cfg.HDTNConfig, "/config/hdtn.json")
	}
	if cfg.ContactPlanFile != "/plan.json" {
		t.Errorf("ContactPlanFile = %q, want %q", cfg.ContactPlanFile, "/plan.json")
	}
	if cfg.TelemetryPort != 6060 {
		t.Errorf("TelemetryPort = %d, want %d", cfg.TelemetryPort, 6060)
	}
	if cfg.TelemetryFile != "/var/tel.json" {
		t.Errorf("TelemetryFile = %q, want %q", cfg.TelemetryFile, "/var/tel.json")
	}
	if cfg.HealthInterval != 20 {
		t.Errorf("HealthInterval = %d, want %d", cfg.HealthInterval, 20)
	}
}
