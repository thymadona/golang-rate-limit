package main

import (
	"sync"
	"time"
)

// 1. Define what a TokenBucket is (The Atom)
type TokenBucket struct {
	capacity   float64
	refillRate float64   // How many tokens added per second
	tokens     float64   // Current available tokens
	lastRefill time.Time // The last time we updated this bucket
}

// 2. Define the Constructor function (The Factory)
func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity, // Start completely full
		lastRefill: time.Now(),
	}
}

// 3. Define the evaluation logic (The Gatekeeper)
func (tb *TokenBucket) Allow() bool {
	// (We will connect this to your map lookup next)
	return true
}

var (
	mu      sync.RWMutex
	clients = make(map[string]*TokenBucket)
)

func getClientLimiter(identity string) *TokenBucket {
	// 1. Acquire a Read Lock (multiple goroutines can read simultaneously)
	mu.RLock()
	bucket, exists := clients[identity]
	mu.RUnlock()

	// 2. If the user already has a bucket, return it immediately
	if exists {
		return bucket
	}

	// 3. If not, acquire a full Write Lock to modify the map safely
	mu.Lock()
	defer mu.Unlock()

	// Double-check existence in case another thread created it
	// while we were upgrading from Read Lock to Write Lock
	if bucket, exists = clients[identity]; exists {
		return bucket
	}

	// 4. Create a fresh bucket for the new user (e.g., 60 req/min capacity)
	newBucket := NewTokenBucket(60, 1.0)
	clients[identity] = newBucket

	return newBucket
}

func main() {

}
