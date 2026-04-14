package security

import (
	"fmt"
	"sync"
)

// TrustZoneKeyStore manages cryptographic keys in STM32U585 TrustZone secure world
// Requirement 16.3: Store keys in TrustZone secure world
type TrustZoneKeyStore struct {
	// In a real STM32U585 implementation, this would interface with
	// TrustZone secure world APIs for isolated key storage
	keys map[string][]byte
	mu   sync.RWMutex
}

// NewTrustZoneKeyStore creates a new TrustZone key store
func NewTrustZoneKeyStore() *TrustZoneKeyStore {
	return &TrustZoneKeyStore{
		keys: make(map[string][]byte),
	}
}

// StoreKey stores a cryptographic key in TrustZone secure world
func (tzks *TrustZoneKeyStore) StoreKey(keyID string, key []byte) error {
	tzks.mu.Lock()
	defer tzks.mu.Unlock()

	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}

	// In real implementation, would call TrustZone secure world API
	// to store key in isolated secure memory
	tzks.keys[keyID] = make([]byte, len(key))
	copy(tzks.keys[keyID], key)

	return nil
}

// GetKey retrieves a cryptographic key from TrustZone secure world
func (tzks *TrustZoneKeyStore) GetKey(keyID string) ([]byte, error) {
	tzks.mu.RLock()
	defer tzks.mu.RUnlock()

	key, exists := tzks.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	// Return a copy to prevent modification
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	return keyCopy, nil
}

// DeleteKey removes a cryptographic key from TrustZone secure world
func (tzks *TrustZoneKeyStore) DeleteKey(keyID string) error {
	tzks.mu.Lock()
	defer tzks.mu.Unlock()

	if _, exists := tzks.keys[keyID]; !exists {
		return fmt.Errorf("key not found: %s", keyID)
	}

	// In real implementation, would call TrustZone secure world API
	// to securely erase the key
	delete(tzks.keys, keyID)

	return nil
}

// ListKeys returns all key IDs stored in TrustZone
func (tzks *TrustZoneKeyStore) ListKeys() []string {
	tzks.mu.RLock()
	defer tzks.mu.RUnlock()

	keyIDs := make([]string, 0, len(tzks.keys))
	for keyID := range tzks.keys {
		keyIDs = append(keyIDs, keyID)
	}

	return keyIDs
}

// KeyExists checks if a key exists in TrustZone
func (tzks *TrustZoneKeyStore) KeyExists(keyID string) bool {
	tzks.mu.RLock()
	defer tzks.mu.RUnlock()

	_, exists := tzks.keys[keyID]
	return exists
}
