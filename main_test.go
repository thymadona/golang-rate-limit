package main

import (
	"sync"
	"testing"
)

func TestTokenBucketConcurrency(t *testing.T) {
	// 1. Setup a controlled client bucket: Capacity 10, No auto-refill
	identity := "127.0.0.1:8080"
	capacity := 10.0
	limiter := NewTokenBucket(capacity, 0.0)

	// Inject this controlled bucket directly into our global state
	mu.Lock()
	clients[identity] = limiter
	mu.Unlock()

	// 2. Setup orchestration primitives for our parallel attack
	var wg sync.WaitGroup
	concurrentRequests := 50 // Far exceeds our capacity of 10

	// Thread-safe channel to collect evaluation results
	allowedChan := make(chan bool, concurrentRequests)

	// 3. Launch 50 independent routines at the exact same millisecond
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Concurrently look up the client and try to consume a token
			bucket := getClientLimiter(identity)
			allowedChan <- bucket.Allow()
		}()
	}

	// Block until all 50 routines finish executing
	wg.Wait()
	close(allowedChan)

	// 4. Analyze the collected results
	allowedCount := 0
	deniedCount := 0
	for allowed := range allowedChan {
		if allowed {
			allowedCount++
		} else {
			deniedCount++
		}
	}

	// 5. Assert atomic correctness from First Principles
	// Out of 50 simultaneous attacks, exactly 10 must pass. No more, no less.
	if allowedCount != int(capacity) {
		t.Errorf("Race Condition Violation: Expected exactly %d allowed requests, but got %d", int(capacity), allowedCount)
	}

	expectedDenied := concurrentRequests - int(capacity)
	if deniedCount != expectedDenied {
		t.Errorf("Drop Counter Mismatch: Expected exactly %d denied requests, but got %d", expectedDenied, deniedCount)
	}
}
