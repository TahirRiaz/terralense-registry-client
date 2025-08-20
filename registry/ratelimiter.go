package registry

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	tokens       int
	maxTokens    int
	refillRate   int
	refillPeriod time.Duration
	lastRefill   time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, period time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:       maxRequests,
		maxTokens:    maxRequests,
		refillRate:   maxRequests,
		refillPeriod: period,
		lastRefill:   time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.TryAcquire() {
			return nil
		}

		// Calculate wait time until next token
		waitTime := r.timeUntilNextToken()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// TryAcquire attempts to acquire a token without blocking
func (r *RateLimiter) TryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// refill adds tokens based on elapsed time
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	if elapsed >= r.refillPeriod {
		// Full refill
		r.tokens = r.maxTokens
		r.lastRefill = now
	} else {
		// Partial refill based on elapsed time
		tokensToAdd := int(float64(r.refillRate) * (float64(elapsed) / float64(r.refillPeriod)))
		if tokensToAdd > 0 {
			r.tokens = min(r.tokens+tokensToAdd, r.maxTokens)
			r.lastRefill = now
		}
	}
}

// timeUntilNextToken calculates the time until the next token is available
func (r *RateLimiter) timeUntilNextToken() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tokens > 0 {
		return 0
	}

	timeSinceLastRefill := time.Since(r.lastRefill)
	timePerToken := r.refillPeriod / time.Duration(r.refillRate)

	if timeSinceLastRefill >= r.refillPeriod {
		return 0
	}

	return timePerToken - (timeSinceLastRefill % timePerToken)
}

// Reset resets the rate limiter to full capacity
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens = r.maxTokens
	r.lastRefill = time.Now()
}

// TokensRemaining returns the number of tokens currently available
func (r *RateLimiter) TokensRemaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()
	return r.tokens
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
