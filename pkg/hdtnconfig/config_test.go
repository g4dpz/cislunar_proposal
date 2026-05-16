package hdtnconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig() *HDTNConfig {
	return &HDTNConfig{
		HDTNConfigName: "test-node",
		MyNodeID:       1,
		MySchemeStr:    "ipn",
		MyDtnEidStr:    "ipn:1.0",
		StoragePath:    "/tmp/hdtn-storage",
		InductsConfig: InductsConfig{
			InductVector: []Induct{
				{ConvergenceLayer: "ltp_over_udp", Name: "udp-induct", BoundPort: 4556},
			},
		},
		OutductsConfig: OutductsConfig{
			OutductVector: []Outduct{
				{ConvergenceLayer: "ltp_over_udp", Name: "udp-outduct", NextHopNodeID: 2, RemoteHostname: "127.0.0.1", RemotePort: 4557},
			},
		},
		ContactPlanJSON: ContactPlanJSON{
			Contacts: []ContactEntry{
				{Source: 1, Dest: 2, StartTime: 0, EndTime: 86400, RateBitsPerSec: 9600},
			},
		},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error for valid config, got: %v", err)
	}
}

func TestValidate_InvalidNodeID(t *testing.T) {
	cfg := validConfig()
	cfg.MyNodeID = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for node ID <= 0")
	}
	if !strings.Contains(err.Error(), "myNodeId") {
		t.Fatalf("error should mention myNodeId, got: %v", err)
	}
}

func TestValidate_EmptyStoragePath(t *testing.T) {
	cfg := validConfig()
	cfg.StoragePath = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty storage path")
	}
	if !strings.Contains(err.Error(), "storagePath") {
		t.Fatalf("error should mention storagePath, got: %v", err)
	}
}

func TestValidate_EmptyInductVector(t *testing.T) {
	cfg := validConfig()
	cfg.InductsConfig.InductVector = nil
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty induct vector")
	}
	if !strings.Contains(err.Error(), "inductVector") {
		t.Fatalf("error should mention inductVector, got: %v", err)
	}
}

func TestValidate_EmptyOutductVector(t *testing.T) {
	cfg := validConfig()
	cfg.OutductsConfig.OutductVector = nil
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty outduct vector")
	}
	if !strings.Contains(err.Error(), "outductVector") {
		t.Fatalf("error should mention outductVector, got: %v", err)
	}
}

func TestValidate_EmptyContactPlan(t *testing.T) {
	cfg := validConfig()
	cfg.ContactPlanJSON.Contacts = nil
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty contact plan")
	}
	if !strings.Contains(err.Error(), "contacts") {
		t.Fatalf("error should mention contacts, got: %v", err)
	}
}

func TestWriteToFile(t *testing.T) {
	cfg := validConfig()

	dir := t.TempDir()
	path := filepath.Join(dir, "hdtn-config.json")

	if err := cfg.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	var loaded HDTNConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse written JSON: %v", err)
	}

	if loaded.MyNodeID != cfg.MyNodeID {
		t.Errorf("expected MyNodeID %d, got %d", cfg.MyNodeID, loaded.MyNodeID)
	}
	if loaded.StoragePath != cfg.StoragePath {
		t.Errorf("expected StoragePath %q, got %q", cfg.StoragePath, loaded.StoragePath)
	}
	if len(loaded.InductsConfig.InductVector) != 1 {
		t.Errorf("expected 1 induct, got %d", len(loaded.InductsConfig.InductVector))
	}
	if len(loaded.OutductsConfig.OutductVector) != 1 {
		t.Errorf("expected 1 outduct, got %d", len(loaded.OutductsConfig.OutductVector))
	}
	if len(loaded.ContactPlanJSON.Contacts) != 1 {
		t.Errorf("expected 1 contact, got %d", len(loaded.ContactPlanJSON.Contacts))
	}
}

func TestWriteToFile_CreatesDirectory(t *testing.T) {
	cfg := validConfig()

	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "hdtn-config.json")

	if err := cfg.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile failed to create nested directory: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist after WriteToFile")
	}
}
