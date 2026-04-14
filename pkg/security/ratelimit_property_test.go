package security

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// Property 24: Rate Limiting
// **Validates: Requirement 16.4**
// For any sequence of rapid bundle submissions exceeding the configured acceptance rate,
// the BPA SHALL reject bundles beyond the rate limit while accepting bundles within the limit.

func TestProperty_RateLimiting(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a rate limit (1-100 bundles per second)
		maxRate := rapid.IntRange(1, 100).Draw(t, "maxRate")

		// Create rate limiter
		limiter := NewRateLimiter(maxRate)

		// Generate a sequence of bundle submission attempts
		// Try to submit more than the rate limit
		numAttempts := rapid.IntRange(maxRate+1, maxRate*2).Draw(t, "numAttempts")

		accepted := 0
		rejected := 0

		// Submit bundles rapidly
		for i := 0; i < numAttempts; i++ {
			if limiter.Allow() {
				accepted++
			} else {
				rejected++
			}
		}

		// Property: accepted should be <= maxRate
		if accepted > maxRate {
			t.Fatalf("rate limit violated: accepted %d bundles, max rate %d", accepted, maxRate)
		}

		// Property: rejected should be > 0 (since we submitted more than maxRate)
		if rejected == 0 {
			t.Fatalf("expected some rejections when submitting %d bundles with max rate %d", numAttempts, maxRate)
		}

		// Property: accepted + rejected should equal total attempts
		if accepted+rejected != numAttempts {
			t.Fatalf("accounting error: accepted %d + rejected %d != attempts %d", accepted, rejected, numAttempts)
		}
	})
}

func TestProperty_RateLimitingWithinWindow(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a rate limit
		maxRate := rapid.IntRange(5, 50).Draw(t, "maxRate")

		// Create rate limiter
		limiter := NewRateLimiter(maxRate)

		// Submit bundles at exactly the rate limit
		for i := 0; i < maxRate; i++ {
			if !limiter.Allow() {
				t.Fatalf("bundle %d rejected when within rate limit %d", i, maxRate)
			}
		}

		// The next bundle should be rejected
		if limiter.Allow() {
			t.Fatalf("bundle accepted when rate limit %d exceeded", maxRate)
		}
	})
}

func TestProperty_RateLimitingWindowReset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a rate limit
		maxRate := rapid.IntRange(5, 20).Draw(t, "maxRate")

		// Create rate limiter
		limiter := NewRateLimiter(maxRate)

		// Fill the rate limit
		for i := 0; i < maxRate; i++ {
			if !limiter.Allow() {
				t.Fatalf("bundle %d rejected when within rate limit %d", i, maxRate)
			}
		}

		// Next bundle should be rejected
		if limiter.Allow() {
			t.Fatalf("bundle accepted when rate limit exceeded")
		}

		// Wait for window to expire
		time.Sleep(1100 * time.Millisecond)

		// Should be able to accept bundles again
		if !limiter.Allow() {
			t.Fatalf("bundle rejected after window reset")
		}
	})
}

func TestProperty_RateLimitingMonotonicity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a rate limit
		maxRate := rapid.IntRange(10, 50).Draw(t, "maxRate")

		// Create rate limiter
		limiter := NewRateLimiter(maxRate)

		// Property: current rate should never exceed max rate
		for i := 0; i < maxRate*2; i++ {
			limiter.Allow()
			currentRate := limiter.GetCurrentRate()
			if currentRate > maxRate {
				t.Fatalf("current rate %d exceeds max rate %d", currentRate, maxRate)
			}
		}
	})
}
