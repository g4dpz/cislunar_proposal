package security

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sync"

	"terrestrial-dtn/pkg/contact"
)

// ContactPlanVerifier verifies signed contact plans for space nodes
// Requirement 16.5: Verify contact plan integrity using signed plans
type ContactPlanVerifier struct {
	trustedKeys   map[string]ed25519.PublicKey
	hardwareCrypto *HardwareCrypto
	mu            sync.RWMutex
}

// NewContactPlanVerifier creates a new contact plan verifier
func NewContactPlanVerifier() *ContactPlanVerifier {
	return &ContactPlanVerifier{
		trustedKeys:   make(map[string]ed25519.PublicKey),
		hardwareCrypto: NewHardwareCrypto(),
	}
}

// AddTrustedKey adds a trusted public key for contact plan verification
func (cpv *ContactPlanVerifier) AddTrustedKey(keyID string, publicKey ed25519.PublicKey) {
	cpv.mu.Lock()
	defer cpv.mu.Unlock()

	cpv.trustedKeys[keyID] = publicKey
}

// SignContactPlan signs a contact plan using a private key
func (cpv *ContactPlanVerifier) SignContactPlan(plan *contact.ContactPlan, privateKey ed25519.PrivateKey) (*ContactPlanSignature, error) {
	// Serialize the contact plan
	planData, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize contact plan: %w", err)
	}

	// Compute hash using hardware accelerator
	planHash := cpv.hardwareCrypto.ComputeSHA256(planData)

	// Sign the hash
	signature := ed25519.Sign(privateKey, planHash[:])

	return &ContactPlanSignature{
		PlanHash:  planHash,
		Signature: signature,
		PublicKey: privateKey.Public().(ed25519.PublicKey),
		Timestamp: plan.GeneratedAt,
	}, nil
}

// VerifyContactPlan verifies a signed contact plan
func (cpv *ContactPlanVerifier) VerifyContactPlan(plan *contact.ContactPlan, sig *ContactPlanSignature, keyID string) error {
	cpv.mu.RLock()
	defer cpv.mu.RUnlock()

	// Get the trusted public key
	publicKey, exists := cpv.trustedKeys[keyID]
	if !exists {
		return fmt.Errorf("unknown key ID: %s", keyID)
	}

	// Serialize the contact plan
	planData, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("failed to serialize contact plan: %w", err)
	}

	// Compute hash using hardware accelerator
	planHash := cpv.hardwareCrypto.ComputeSHA256(planData)

	// Verify the hash matches
	if planHash != sig.PlanHash {
		return fmt.Errorf("contact plan hash mismatch")
	}

	// Verify the signature
	if !ed25519.Verify(publicKey, planHash[:], sig.Signature) {
		return fmt.Errorf("contact plan signature verification failed")
	}

	return nil
}

// RemoveTrustedKey removes a trusted public key
func (cpv *ContactPlanVerifier) RemoveTrustedKey(keyID string) {
	cpv.mu.Lock()
	defer cpv.mu.Unlock()

	delete(cpv.trustedKeys, keyID)
}

// ListTrustedKeys returns all trusted key IDs
func (cpv *ContactPlanVerifier) ListTrustedKeys() []string {
	cpv.mu.RLock()
	defer cpv.mu.RUnlock()

	keys := make([]string, 0, len(cpv.trustedKeys))
	for keyID := range cpv.trustedKeys {
		keys = append(keys, keyID)
	}
	return keys
}
