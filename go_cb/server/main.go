package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	interval  time.Duration
	ticker    *time.Ticker
}

func NewRateLimiter(maxTokens int, refillInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		interval:  refillInterval,
		ticker:    time.NewTicker(refillInterval),
	}

	go rl.refill()

	return rl
}

func (rl *RateLimiter) refill() {
	for range rl.ticker.C {
		rl.mu.Lock()
		if rl.tokens < rl.maxTokens {
			rl.tokens++
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func rateLimitedHandler(rl *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rl.Allow() {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}
	}
}

func main() {
	rateLimiter := NewRateLimiter(5, time.Second) // Allow 5 requests per second

	http.HandleFunc("/", rateLimitedHandler(rateLimiter, hello))

	fmt.Println("Server listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
