package security

import (
	"testing"

	"terrestrial-dtn/pkg/bpa"
)

func TestBPSecManager_AddAndVerifyIntegrityBlock(t *testing.T) {
	// Create BPSec manager
	manager := NewBPSecManager()

	// Store a test key
	keyID := "test-key"
	key := []byte("test-secret-key-12345678901234567890")
	err := manager.GetKeyStore().StoreKey(keyID, key)
	if err != nil {
		t.Fatalf("failed to store key: %v", err)
	}

	// Create a test bundle
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "dtn", SSP: "//source"},
			CreationTimestamp: 1000000,
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "dtn", SSP: "//dest"},
		Payload:     []byte("test payload data"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    300,
		CreatedAt:   1000000,
		BundleType:  bpa.BundleTypeData,
	}

	// Add integrity block
	integrityBlock, err := manager.AddIntegrityBlock(bundle, keyID)
	if err != nil {
		t.Fatalf("failed to add integrity block: %v", err)
	}

	if integrityBlock == nil {
		t.Fatal("integrity block is nil")
	}

	if len(integrityBlock.SecurityResults) == 0 {
		t.Fatal("integrity block has no security results")
	}

	// Verify integrity block
	err = manager.VerifyIntegrityBlock(bundle, integrityBlock, keyID)
	if err != nil {
		t.Fatalf("failed to verify integrity block: %v", err)
	}
}

func TestBPSecManager_VerifyIntegrityBlock_Tampered(t *testing.T) {
	// Create BPSec manager
	manager := NewBPSecManager()

	// Store a test key
	keyID := "test-key"
	key := []byte("test-secret-key-12345678901234567890")
	err := manager.GetKeyStore().StoreKey(keyID, key)
	if err != nil {
		t.Fatalf("failed to store key: %v", err)
	}

	// Create a test bundle
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "dtn", SSP: "//source"},
			CreationTimestamp: 1000000,
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "dtn", SSP: "//dest"},
		Payload:     []byte("test payload data"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    300,
		CreatedAt:   1000000,
		BundleType:  bpa.BundleTypeData,
	}

	// Add integrity block
	integrityBlock, err := manager.AddIntegrityBlock(bundle, keyID)
	if err != nil {
		t.Fatalf("failed to add integrity block: %v", err)
	}

	// Tamper with the bundle payload
	bundle.Payload = []byte("tampered payload data")

	// Verify should fail
	err = manager.VerifyIntegrityBlock(bundle, integrityBlock, keyID)
	if err == nil {
		t.Fatal("expected verification to fail for tampered bundle")
	}
}

func TestBPSecManager_VerifyIntegrityBlock_WrongKey(t *testing.T) {
	// Create BPSec manager
	manager := NewBPSecManager()

	// Store two different keys
	keyID1 := "key1"
	key1 := []byte("test-secret-key-1-1234567890123456")
	err := manager.GetKeyStore().StoreKey(keyID1, key1)
	if err != nil {
		t.Fatalf("failed to store key1: %v", err)
	}

	keyID2 := "key2"
	key2 := []byte("test-secret-key-2-1234567890123456")
	err = manager.GetKeyStore().StoreKey(keyID2, key2)
	if err != nil {
		t.Fatalf("failed to store key2: %v", err)
	}

	// Create a test bundle
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "dtn", SSP: "//source"},
			CreationTimestamp: 1000000,
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "dtn", SSP: "//dest"},
		Payload:     []byte("test payload data"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    300,
		CreatedAt:   1000000,
		BundleType:  bpa.BundleTypeData,
	}

	// Add integrity block with key1
	integrityBlock, err := manager.AddIntegrityBlock(bundle, keyID1)
	if err != nil {
		t.Fatalf("failed to add integrity block: %v", err)
	}

	// Verify with key2 should fail
	err = manager.VerifyIntegrityBlock(bundle, integrityBlock, keyID2)
	if err == nil {
		t.Fatal("expected verification to fail with wrong key")
	}
}
