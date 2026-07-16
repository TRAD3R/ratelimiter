// Package ratelimiter provides a simple and efficient rate limiter for Go applications.
// It allows limiting the number of requests per time interval with thread-safe operations.
package ratelimiter

import (
	"math"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter that allows a maximum number
// of requests per time interval. Tokens are refilled continuously (not in bursts
// at fixed boundaries), so it never allows more than 'rps' requests within any
// sliding window of length 'interval'. It is thread-safe and can be used
// concurrently from multiple goroutines.
type RateLimiter struct {
	rps        int
	interval   time.Duration
	capacity   float64
	rate       float64 // tokens per nanosecond
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the specified requests per second
// limit and reset interval. The rate limiter will allow up to 'rps' requests
// within any sliding window of length 'interval'. The bucket starts full, so an
// initial burst of up to 'rps' requests is allowed immediately.
func NewRateLimiter(rps int, interval time.Duration) *RateLimiter {
	capacity := float64(rps)

	return &RateLimiter{
		rps:        rps,
		interval:   interval,
		capacity:   capacity,
		rate:       capacity / float64(interval),
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// AllowRequest checks if a request is allowed according to the rate limit.
// It returns true if the request is allowed and consumes one token,
// or false if the rate limit has been exceeded.
func (rl *RateLimiter) AllowRequest() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill(time.Now())

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

// refill adds tokens accumulated since the last refill, capped at capacity.
func (rl *RateLimiter) refill(now time.Time) {
	elapsed := now.Sub(rl.lastRefill)
	if elapsed <= 0 {
		return
	}

	rl.tokens = min(rl.capacity, rl.tokens+float64(elapsed)*rl.rate)
	rl.lastRefill = now
}

// CurrentState returns the current state of the rate limiter.
// It returns the number of tokens currently available (rounded down) and
// the time remaining until the next token becomes available (0 if a token
// is already available).
func (rl *RateLimiter) CurrentState() (int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill(time.Now())

	if rl.tokens >= 1 {
		return int(rl.tokens), 0
	}

	if rl.rate <= 0 {
		return int(rl.tokens), time.Duration(math.MaxInt64)
	}

	missing := 1 - rl.tokens
	wait := time.Duration(missing / rl.rate)

	return int(rl.tokens), wait
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

// GetRetryAfter returns the time when the next token will become available.
// It returns the current time if a token is already available.
// This method is primarily for testing purposes.
func (rl *RateLimiter) GetRetryAfter() time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.refill(now)

	if rl.tokens >= 1 {
		return now
	}

	if rl.rate <= 0 {
		return now.Add(time.Duration(math.MaxInt64))
	}

	missing := 1 - rl.tokens
	wait := time.Duration(missing / rl.rate)

	return now.Add(wait)
}
