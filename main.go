package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	mu      sync.RWMutex
	clients = make(map[string]*TokenBucket)
)

// 1. Define what a TokenBucket is (The Atom)
type TokenBucket struct {
	mu         sync.Mutex
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
	// Mutex lock specific to this *individual* bucket's values
	// Note: You need to add `mu sync.Mutex` to your TokenBucket struct fields if not already there!
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// 1. Calculate lazy refill based on continuous time elapsed
	tb.tokens = tb.tokens + (elapsed * tb.refillRate)
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	// 2. Check if a whole token is available to consume
	if tb.tokens >= 1.0 {
		tb.tokens--
		return true // Request allowed
	}

	return false // Rate limit exceeded!
}

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

func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Identify the client by their network IP address string
		identity := r.RemoteAddr

		// Fetch or build their safe, isolated rate bucket
		bucket := getClientLimiter(identity)

		// Evaluate token availability
		if !bucket.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("429 Too Many Requests: Rate limit exceeded.\n"))
			return // Short-circuit: do not run the actual API handler
		}

		// Success! Pass control down to the actual endpoint code
		next(w, r)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("200 OK: Request processed successfully!\n"))
}

func main() {
	// Register endpoint /api/data wrapped inside our middleware guard
	http.HandleFunc("/api/data", RateLimitMiddleware(helloHandler))

	fmt.Println("Backend server listening on port :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
