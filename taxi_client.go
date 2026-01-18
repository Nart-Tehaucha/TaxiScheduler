// taxi_client.go - Simulates taxi registration
// Sends 15 taxi creation requests through the Server API, rate limited to 1 per 5 seconds

package main

import (
	"fmt"
	"math/rand"
	"time"
)

// TaxiClient simulates taxis registering with the system.
// Calls the Server API to register taxis at a rate-limited pace.
type TaxiClient struct {
	server    *Server       // Server API gateway
	maxTaxis  int           // Number of taxis to create (15)
	rateLimit time.Duration // Delay between requests (5 seconds)
}

// NewTaxiClient creates a TaxiClient configured to send 15 registrations.
func NewTaxiClient(server *Server) *TaxiClient {
	return &TaxiClient{
		server:    server,
		maxTaxis:  15,              // Per INSTRUCTIONS.md: 15 taxis
		rateLimit: 5 * time.Second, // Per INSTRUCTIONS.md: 1 per 5 seconds
	}
}

// Start begins sending taxi registration requests.
// Sends 15 requests with random locations, 1 every 5 seconds.
// This method blocks until all registrations are sent.
func (tc *TaxiClient) Start() {
	fmt.Println("[TaxiClient] Starting taxi registration...")

	for i := 0; i < tc.maxTaxis; i++ {
		// Generate random location (0-99 grid)
		location := Location{
			X: rand.Intn(100),
			Y: rand.Intn(100),
		}

		// Call Server API to register taxi
		taxiID := tc.server.RegisterTaxi(location)
		fmt.Printf("[TaxiClient] Registered taxi #%d at (%d, %d)\n",
			taxiID, location.X, location.Y)

		// Rate limit: wait 5 seconds before next request (except after last)
		if i < tc.maxTaxis-1 {
			time.Sleep(tc.rateLimit)
		}
	}

	fmt.Println("[TaxiClient] All 15 taxi registrations sent")
}
