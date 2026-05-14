# Security Package

This package implements security features for the Cislunar Amateur DTN Payload, including rate limiting and contact plan verification.

Note: No cryptographic operations are used in this project. Amateur radio regulations prohibit encryption and cryptography on transmitted signals.

## Components

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

Security operations return errors that should be handled appropriately:

- **Rate limit exceeded**: Bundle should be rejected and sender notified
