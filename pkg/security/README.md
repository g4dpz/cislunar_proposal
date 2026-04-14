# Security Package

This package implements security features for the Cislunar Amateur DTN Payload, including BPSec integration, rate limiting, and contact plan verification.

## Components

### BPSec Integration (Requirement 16.1, 16.2, 16.3)

The `BPSecManager` provides Bundle Protocol Security (RFC 9172) integrity blocks for bundle origin authentication:

- **Hardware Crypto**: Uses STM32U585 hardware crypto accelerator (AES-256, SHA-256, PKA) for cryptographic operations
- **TrustZone Key Storage**: Stores cryptographic keys in TrustZone secure world, isolated from non-secure application code
- **Integrity Blocks**: Adds and verifies HMAC-SHA256 integrity blocks to bundles

```go
// Create BPSec manager
bpsec := security.NewBPSecManager()

// Store a key in TrustZone
keyStore := bpsec.GetKeyStore()
keyStore.StoreKey("my-key", []byte("secret-key-data"))

// Add integrity block to a bundle
integrityBlock, err := bpsec.AddIntegrityBlock(bundle, "my-key")

// Verify integrity block
err = bpsec.VerifyIntegrityBlock(bundle, integrityBlock, "my-key")
```

### Rate Limiting (Requirement 16.4)

The `RateLimiter` enforces rate limiting on bundle acceptance to prevent store flooding attacks:

```go
// Create rate limiter (max 10 bundles per second)
limiter := security.NewRateLimiter(10)

// Check if bundle can be accepted
if !limiter.Allow() {
    return fmt.Errorf("rate limit exceeded")
}

// Or use CheckAndReject for error handling
if err := limiter.CheckAndReject(); err != nil {
    return err
}
```

### Contact Plan Verification (Requirement 16.5)

The `ContactPlanVerifier` verifies signed contact plans for space nodes:

```go
// Create verifier
verifier := security.NewContactPlanVerifier()

// Add trusted public key
verifier.AddTrustedKey("ground-station-1", publicKey)

// Sign a contact plan (ground station)
signature, err := verifier.SignContactPlan(plan, privateKey)

// Verify a contact plan (space node)
err = verifier.VerifyContactPlan(plan, signature, "ground-station-1")
```

## Hardware Integration

### STM32U585 Hardware Crypto Accelerator

The `HardwareCrypto` type provides an interface to the STM32U585 hardware crypto accelerator:

- **SHA-256**: Hardware-accelerated hash computation
- **HMAC-SHA256**: Hardware-accelerated HMAC for integrity blocks
- **AES-256-GCM**: Hardware-accelerated encryption/decryption

In the current implementation, these use software fallbacks. On actual STM32U585 hardware, these would interface with the hardware crypto peripherals.

### TrustZone Secure World

The `TrustZoneKeyStore` provides secure key storage in the STM32U585 TrustZone secure world:

- Keys are isolated from non-secure application code
- Keys cannot be read or modified by compromised application code
- Secure erase on key deletion

In the current implementation, this uses in-memory storage. On actual STM32U585 hardware, this would interface with TrustZone secure world APIs.

## Testing

The package includes comprehensive property-based tests:

### Property 24: Rate Limiting (Requirement 16.4)

Validates that the rate limiter correctly rejects bundles beyond the configured rate while accepting bundles within the limit.

```bash
go test -v ./pkg/security/... -run TestProperty_RateLimiting
```

## Integration with BPA

The BPA can be configured with a rate limiter:

```go
// Create BPA
bpa := bpa.NewBundleProtocolAgent(endpoints)

// Create and set rate limiter
limiter := security.NewRateLimiter(10)
bpa.SetRateLimiter(limiter)

// Now ReceiveBundle will enforce rate limiting
err := bpa.ReceiveBundle(bundle, currentTime)
```

## Error Handling

All security operations return errors that should be handled appropriately:

- **Rate limit exceeded**: Bundle should be rejected and sender notified
- **Integrity verification failed**: Bundle should be discarded and event logged
- **Key not found**: Operation should fail and alert operator
- **Signature verification failed**: Contact plan should be rejected

## Future Enhancements

- Integration with actual STM32U585 hardware crypto peripherals
- Integration with TrustZone secure world APIs
- Support for BPSec confidentiality blocks (encryption)
- Support for additional cipher suites
- Key rotation and management
- Certificate-based authentication
