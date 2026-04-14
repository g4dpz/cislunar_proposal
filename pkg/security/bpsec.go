package security

import (
	"fmt"
	"sync"

	"terrestrial-dtn/pkg/bpa"
)

// BPSecManager manages BPSec operations for bundle integrity
type BPSecManager struct {
	keyStore      *TrustZoneKeyStore
	hardwareCrypto *HardwareCrypto
	mu            sync.RWMutex
}

// NewBPSecManager creates a new BPSec manager
func NewBPSecManager() *BPSecManager {
	return &BPSecManager{
		keyStore:      NewTrustZoneKeyStore(),
		hardwareCrypto: NewHardwareCrypto(),
	}
}

// AddIntegrityBlock adds a BPSec integrity block to a bundle
// Requirement 16.1: Support BPSec (RFC 9172) integrity blocks
func (bm *BPSecManager) AddIntegrityBlock(bundle *bpa.Bundle, keyID string) (*IntegrityBlock, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Get the key from TrustZone secure storage
	key, err := bm.keyStore.GetKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key from TrustZone: %w", err)
	}

	// Serialize bundle payload for integrity computation
	data := bundle.Payload

	// Compute HMAC-SHA256 using hardware accelerator
	// Requirement 16.2: Use STM32U585 hardware crypto (SHA-256)
	hmac := bm.hardwareCrypto.ComputeHMAC(key, data)

	integrityBlock := &IntegrityBlock{
		SecurityTargets: []int{1}, // Protect payload block
		SecurityContext: BPSecContext{
			SecuritySource:  bundle.ID.SourceEID.String(),
			SecurityTarget:  bundle.Destination.String(),
			SecurityService: SecurityServiceIntegrity,
			CipherSuite:     CipherSuiteSHA256HMAC,
		},
		SecurityResults: hmac,
	}

	return integrityBlock, nil
}

// VerifyIntegrityBlock verifies a BPSec integrity block
// Requirement 16.1: Support BPSec (RFC 9172) integrity blocks
func (bm *BPSecManager) VerifyIntegrityBlock(bundle *bpa.Bundle, block *IntegrityBlock, keyID string) error {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Get the key from TrustZone secure storage
	key, err := bm.keyStore.GetKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to get key from TrustZone: %w", err)
	}

	// Serialize bundle payload for integrity verification
	data := bundle.Payload

	// Compute expected HMAC using hardware accelerator
	// Requirement 16.2: Use STM32U585 hardware crypto (SHA-256)
	expectedHMAC := bm.hardwareCrypto.ComputeHMAC(key, data)

	// Compare HMACs
	if len(expectedHMAC) != len(block.SecurityResults) {
		return fmt.Errorf("HMAC length mismatch")
	}

	for i := range expectedHMAC {
		if expectedHMAC[i] != block.SecurityResults[i] {
			return fmt.Errorf("HMAC verification failed")
		}
	}

	return nil
}

// GetHardwareCrypto returns the hardware crypto accelerator
func (bm *BPSecManager) GetHardwareCrypto() *HardwareCrypto {
	return bm.hardwareCrypto
}

// GetKeyStore returns the TrustZone key store
func (bm *BPSecManager) GetKeyStore() *TrustZoneKeyStore {
	return bm.keyStore
}
