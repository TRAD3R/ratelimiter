// Package ratelimiter provides a simple and efficient rate limiter for Go applications.
// It allows limiting the number of requests per time interval with thread-safe operations.
package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter that allows a maximum number
// of requests per time interval. It is thread-safe and can be used concurrently
// from multiple goroutines.
type RateLimiter struct {
	rps        int
	interval   time.Duration
	requests   int
	retryAfter time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the specified requests per second
// limit and reset interval. The rate limiter will allow up to 'rps' requests
// within each 'interval' time window.
func NewRateLimiter(rps int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		rps:        rps,
		interval:   interval,
		requests:   0,
		retryAfter: time.Now().Add(interval),
	}
}

// AllowRequest checks if a request is allowed according to the rate limit.
// It returns true if the request is allowed and increments the internal counter,
// or false if the rate limit has been exceeded.
func (rl *RateLimiter) AllowRequest() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if time.Now().After(rl.retryAfter) {
		rl.requests = 0
		rl.retryAfter = time.Now().Add(rl.interval)
	}

	if rl.requests < rl.rps {
		rl.requests++
		return true
	}

	return false
}

// CurrentState returns the current state of the rate limiter.
// It returns the number of requests made in the current interval and
// the time remaining until the next reset.
func (rl *RateLimiter) CurrentState() (int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.requests, time.Until(rl.retryAfter)
}

// GetRPS returns the configured requests per second limit.
// This method is primarily for testing purposes.
func (rl *RateLimiter) GetRPS() int {
	return rl.rps
}

// GetInterval returns the configured reset interval.
// This method is primarily for testing purposes.
func (rl *RateLimiter) GetInterval() time.Duration {
	return rl.interval
}

// GetRetryAfter returns the time when the rate limiter will reset.
// This method is primarily for testing purposes.
func (rl *RateLimiter) GetRetryAfter() time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.retryAfter
}
