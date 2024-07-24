package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

var cb *gobreaker.CircuitBreaker

func init() {
	settings := gobreaker.Settings{
		Name: "HTTP GET",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			fmt.Printf("Counts: %v, FailRatio: %f\n", counts, failRatio)
			return counts.Requests >= 10 && failRatio >= 0.4
		},
		Timeout: 10 * time.Second,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker state change: %s -> %s\n", from, to)
		},
	}
	cb = gobreaker.NewCircuitBreaker(settings)
}

func makeRequest(simulateFailure bool) error {
	if simulateFailure {
		return fmt.Errorf("simulated failure")
	}

	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}

func main() {
	for i := 0; i < 20; i++ {
		simulateFailure := i >= 10 // Simulate failure for the last 10 requests
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, makeRequest(simulateFailure)
		})

		if err != nil {
			fmt.Println("Error:", err)
		}

		time.Sleep(500 * time.Millisecond) // Increased sleep time to observe state transitions
	}

	// Wait for timeout period to allow circuit breaker to transition to half-open state
	fmt.Println("Waiting for circuit breaker timeout...")
	time.Sleep(15 * time.Second) // Wait for longer than the timeout period

	// Make another set of requests to observe the half-open and closed states
	for i := 0; i < 10; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, makeRequest(false)
		})

		if err != nil {
			fmt.Println("Error:", err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
