package security

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter enforces rate limiting on bundle acceptance
// Requirement 16.4: Enforce rate limiting to prevent store flooding
type RateLimiter struct {
	maxBundlesPerSecond int
	windowSize          time.Duration
	timestamps          []time.Time
	mu                  sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxBundlesPerSecond int) *RateLimiter {
	return &RateLimiter{
		maxBundlesPerSecond: maxBundlesPerSecond,
		windowSize:          time.Second,
		timestamps:          make([]time.Time, 0),
	}
}

// Allow checks if a bundle can be accepted based on rate limit
// Returns true if the bundle is within rate limit, false otherwise
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Remove timestamps outside the current window
	cutoff := now.Add(-rl.windowSize)
	validTimestamps := make([]time.Time, 0)
	for _, ts := range rl.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	rl.timestamps = validTimestamps

	// Check if we're at the rate limit
	if len(rl.timestamps) >= rl.maxBundlesPerSecond {
		return false
	}

	// Accept the bundle and record timestamp
	rl.timestamps = append(rl.timestamps, now)
	return true
}

// GetCurrentRate returns the current acceptance rate (bundles per second)
func (rl *RateLimiter) GetCurrentRate() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	count := 0
	for _, ts := range rl.timestamps {
		if ts.After(cutoff) {
			count++
		}
	}

	return count
}

// Reset clears the rate limiter state
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.timestamps = make([]time.Time, 0)
}

// CheckAndReject checks rate limit and returns error if exceeded
func (rl *RateLimiter) CheckAndReject() error {
	if !rl.Allow() {
		return fmt.Errorf("rate limit exceeded: max %d bundles per second", rl.maxBundlesPerSecond)
	}
	return nil
}
