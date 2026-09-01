package main

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu         sync.Mutex
	capacity   float64
	refillRate float64
	tokens     float64

	lastRefill time.Time
}

func NewRateLimiter(capacity int, refillRate float64) *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		capacity:   float64(capacity),
		refillRate: refillRate,
		tokens:     float64(capacity),
		lastRefill: now,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.capacity {
		rl.tokens = rl.capacity
	}

	rl.lastRefill = now

	if rl.tokens < 1 {
		return false
	}

	rl.tokens--
	return true
}

type RateLimitMap struct {
	mu         sync.Mutex
	limiters   map[string]*RateLimiter
	capacity   int
	refillRate float64
}

func NewRateLimitMap(capacity int, refillRate float64) *RateLimitMap {
	return &RateLimitMap{
		limiters:   make(map[string]*RateLimiter),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (m *RateLimitMap) Allow(key string) bool {
	m.mu.Lock()
	rl, ok := m.limiters[key]
	if !ok {
		rl = NewRateLimiter(m.capacity, m.refillRate)
		m.limiters[key] = rl
	}
	m.mu.Unlock()

	return rl.Allow()
}
