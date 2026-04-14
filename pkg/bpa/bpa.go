package bpa

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter interface for dependency injection
type RateLimiter interface {
	CheckAndReject() error
	Allow() bool
}

// BundleProtocolAgent wraps ION-DTN bundle operations
type BundleProtocolAgent struct {
	localEndpoints []EndpointID
	sequenceNum    atomic.Uint64
	rateLimiter    RateLimiter
	mu             sync.RWMutex
}

// NewBundleProtocolAgent creates a new BPA instance
func NewBundleProtocolAgent(endpoints []EndpointID) *BundleProtocolAgent {
	return &BundleProtocolAgent{
		localEndpoints: endpoints,
		rateLimiter:    nil, // Optional, can be set with SetRateLimiter
	}
}

// SetRateLimiter sets the rate limiter for bundle acceptance
// Requirement 16.4: Enforce rate limiting to prevent store flooding
func (bpa *BundleProtocolAgent) SetRateLimiter(limiter RateLimiter) {
	bpa.mu.Lock()
	defer bpa.mu.Unlock()
	bpa.rateLimiter = limiter
}

// CreateBundle creates a new data bundle
func (bpa *BundleProtocolAgent) CreateBundle(
	source EndpointID,
	destination EndpointID,
	payload []byte,
	priority Priority,
	lifetime int64,
) (*Bundle, error) {
	if destination.Scheme == "" || destination.SSP == "" {
		return nil, fmt.Errorf("invalid destination endpoint")
	}

	if lifetime <= 0 {
		return nil, fmt.Errorf("lifetime must be greater than zero")
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("payload cannot be empty")
	}

	now := time.Now().Unix()
	seqNum := bpa.sequenceNum.Add(1)

	bundle := &Bundle{
		ID: BundleID{
			SourceEID:         source,
			CreationTimestamp: now,
			SequenceNumber:    seqNum,
		},
		Destination: destination,
		Payload:     payload,
		Priority:    priority,
		Lifetime:    lifetime,
		CreatedAt:   now,
		BundleType:  BundleTypeData,
	}

	return bundle, nil
}

// CreatePing creates a ping request bundle
func (bpa *BundleProtocolAgent) CreatePing(
	source EndpointID,
	destination EndpointID,
) (*Bundle, error) {
	if destination.Scheme == "" || destination.SSP == "" {
		return nil, fmt.Errorf("invalid destination endpoint")
	}

	now := time.Now().Unix()
	seqNum := bpa.sequenceNum.Add(1)

	// Ping bundles have a default lifetime of 300 seconds (5 minutes)
	bundle := &Bundle{
		ID: BundleID{
			SourceEID:         source,
			CreationTimestamp: now,
			SequenceNumber:    seqNum,
		},
		Destination: destination,
		Payload:     []byte("PING"),
		Priority:    PriorityExpedited,
		Lifetime:    300,
		CreatedAt:   now,
		BundleType:  BundleTypePingRequest,
	}

	return bundle, nil
}

// ValidateBundle validates a received bundle
func (bpa *BundleProtocolAgent) ValidateBundle(bundle *Bundle, currentTime int64) error {
	// Requirement 1.2: Validate bundle fields
	
	// Check destination is valid
	if bundle.Destination.Scheme == "" || bundle.Destination.SSP == "" {
		return fmt.Errorf("invalid destination endpoint")
	}

	// Check lifetime is greater than zero
	if bundle.Lifetime <= 0 {
		return fmt.Errorf("lifetime must be greater than zero")
	}

	// Check creation timestamp does not exceed current time
	if bundle.CreatedAt > currentTime {
		return fmt.Errorf("creation timestamp %d exceeds current time %d", bundle.CreatedAt, currentTime)
	}

	// Check bundle is not expired
	if bundle.IsExpired(currentTime) {
		return fmt.Errorf("bundle expired at %d (current time %d)", bundle.CreatedAt+bundle.Lifetime, currentTime)
	}

	// Note: CRC validation would be done by ION-DTN at the protocol level
	// This is a higher-level validation

	return nil
}

// ReceiveBundle processes an incoming bundle
// Requirement 16.4: Enforce rate limiting on bundle acceptance
func (bpa *BundleProtocolAgent) ReceiveBundle(bundle *Bundle, currentTime int64) error {
	// Check rate limit if configured
	bpa.mu.RLock()
	limiter := bpa.rateLimiter
	bpa.mu.RUnlock()

	if limiter != nil {
		if err := limiter.CheckAndReject(); err != nil {
			return fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	// Validate the bundle
	if err := bpa.ValidateBundle(bundle, currentTime); err != nil {
		return fmt.Errorf("bundle validation failed: %w", err)
	}

	// Bundle is valid and can be processed
	// Actual storage/delivery is handled by the caller (Node Controller)
	return nil
}

// HandlePing processes a ping request and generates an echo response
func (bpa *BundleProtocolAgent) HandlePing(pingRequest *Bundle) (*Bundle, error) {
	if pingRequest.BundleType != BundleTypePingRequest {
		return nil, fmt.Errorf("bundle is not a ping request")
	}

	// Create echo response addressed to the original sender
	now := time.Now().Unix()
	seqNum := bpa.sequenceNum.Add(1)

	// Find a local endpoint to use as source
	if len(bpa.localEndpoints) == 0 {
		return nil, fmt.Errorf("no local endpoints configured")
	}
	source := bpa.localEndpoints[0]

	response := &Bundle{
		ID: BundleID{
			SourceEID:         source,
			CreationTimestamp: now,
			SequenceNumber:    seqNum,
		},
		Destination: pingRequest.ID.SourceEID, // Reply to sender
		Payload:     []byte("PONG"),
		Priority:    PriorityExpedited,
		Lifetime:    300,
		CreatedAt:   now,
		BundleType:  BundleTypePingResponse,
	}

	return response, nil
}

// IsLocalEndpoint checks if an endpoint is local to this node
func (bpa *BundleProtocolAgent) IsLocalEndpoint(endpoint EndpointID) bool {
	bpa.mu.RLock()
	defer bpa.mu.RUnlock()

	for _, local := range bpa.localEndpoints {
		if local.Scheme == endpoint.Scheme && local.SSP == endpoint.SSP {
			return true
		}
	}
	return false
}

// DeliverBundle delivers a bundle to the local application agent
func (bpa *BundleProtocolAgent) DeliverBundle(bundle *Bundle) error {
	if !bpa.IsLocalEndpoint(bundle.Destination) {
		return fmt.Errorf("bundle destination %s is not a local endpoint", bundle.Destination.String())
	}

	// In a real implementation, this would deliver to the application layer
	// For now, we just validate that delivery is possible
	return nil
}

// DeleteBundle is a placeholder for bundle deletion
// In practice, this would interact with ION-DTN's bundle storage
func (bpa *BundleProtocolAgent) DeleteBundle(bundleID BundleID) error {
	// This would call into ION-DTN to delete the bundle
	return nil
}

// QueryBundles is a placeholder for querying bundles
// In practice, this would interact with ION-DTN's bundle storage
func (bpa *BundleProtocolAgent) QueryBundles(filter func(*Bundle) bool) ([]*Bundle, error) {
	// This would call into ION-DTN to query bundles
	return nil, nil
}
