// user_client.go - Simulates users requesting rides
// Sends 100 ride requests through the Server API, rate limited to 1 per 5 seconds

package main

import (
	"fmt"
	"math/rand"
	"time"
)

// UserClient simulates users requesting rides.
// Calls the Server API to request rides at a rate-limited pace.
type UserClient struct {
	server    *Server       // Server API gateway
	maxRides  int           // Number of rides to request (100)
	rateLimit time.Duration // Delay between requests (5 seconds)
}

// NewUserClient creates a UserClient configured to send 100 ride requests.
func NewUserClient(server *Server) *UserClient {
	return &UserClient{
		server:    server,
		maxRides:  100,             // Per INSTRUCTIONS.md: 100 rides
		rateLimit: 5 * time.Second, // Per INSTRUCTIONS.md: 1 per 5 seconds
	}
}

// Start begins sending ride requests.
// Sends 100 requests with random start/end locations, 1 every 5 seconds.
// This method blocks until all requests are sent.
func (uc *UserClient) Start() {
	fmt.Println("[UserClient] Starting ride requests...")

	for i := 0; i < uc.maxRides; i++ {
		clientID := i + 1

		// Generate random start and end locations (0-99 grid)
		startLocation := Location{
			X: rand.Intn(100),
			Y: rand.Intn(100),
		}
		endLocation := Location{
			X: rand.Intn(100),
			Y: rand.Intn(100),
		}

		// Call Server API to request ride
		if !uc.server.RequestRide(clientID, startLocation, endLocation) {
			fmt.Printf("[UserClient] Client #%d request rejected (server shutting down)\n", clientID)
			continue
		}
		fmt.Printf("[UserClient] Client #%d requested ride: (%d,%d) -> (%d,%d)\n",
			clientID,
			startLocation.X, startLocation.Y,
			endLocation.X, endLocation.Y)

		// Rate limit: wait 5 seconds before next request (except after last)
		if i < uc.maxRides-1 {
			time.Sleep(uc.rateLimit)
		}
	}

	fmt.Println("[UserClient] All 100 ride requests sent")
}
