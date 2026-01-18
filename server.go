// server.go - API gateway and main entry point
// Provides a centralized API for all client requests

package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Server is the central API gateway for the taxi scheduling system.
// All client requests (taxi registration, ride requests) go through the Server.
type Server struct {
	taxiManager     *TaxiManager     // For taxi CRUD operations
	rideRequests    chan RideRequest // Channel for ride requests to scheduler
	locationService *LocationService // For distance calculations
	taxiStore       *TaxiStore       // For direct store access if needed
	mu              sync.Mutex       // Protects shutdown flag
	shutdown        bool             // Prevents sends to closed channel
}

// NewServer creates and initializes a new Server with all dependencies.
func NewServer() *Server {
	// Initialize core services
	locationService := NewLocationService()
	taxiStore := NewTaxiStore()
	taxiManager := NewTaxiManager(taxiStore)
	taxiAssigner := NewTaxiAssigner(taxiStore, locationService)

	// Create ride requests channel (buffered to prevent blocking)
	rideRequests := make(chan RideRequest, 150)

	// Create and start the ride scheduler
	rideScheduler := NewRideScheduler(rideRequests, taxiAssigner, taxiStore, locationService)
	go rideScheduler.Start()

	return &Server{
		taxiManager:     taxiManager,
		rideRequests:    rideRequests,
		locationService: locationService,
		taxiStore:       taxiStore,
	}
}

// RegisterTaxi registers a new taxi at the given location.
// Returns the new taxi's ID.
func (s *Server) RegisterTaxi(location Location) int {
	return s.taxiManager.CreateTaxi(location)
}

// RequestRide submits a ride request to the system.
// The request is queued and processed by the RideScheduler.
// Returns false if the server is shutting down.
func (s *Server) RequestRide(clientID int, startLocation, endLocation Location) bool {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		log.Printf("[Server] Rejecting ride request from client #%d, server is shutting down\n", clientID)
		return false
	}
	s.mu.Unlock()

	request := RideRequest{
		ClientID:      clientID,
		StartLocation: startLocation,
		EndLocation:   endLocation,
	}
	s.rideRequests <- request
	fmt.Printf("[Server] Received ride request from client #%d: (%d,%d) -> (%d,%d)\n",
		clientID, startLocation.X, startLocation.Y, endLocation.X, endLocation.Y)
	return true
}

// GetTaxiCount returns the number of registered taxis.
func (s *Server) GetTaxiCount() int {
	return s.taxiStore.Count()
}

// GetAvailableTaxiCount returns the number of available taxis.
func (s *Server) GetAvailableTaxiCount() int {
	return len(s.taxiStore.GetAllAvailable())
}

// Shutdown closes the ride requests channel to signal shutdown.
// Sets shutdown flag first to prevent new requests from being sent.
func (s *Server) Shutdown() {
	s.mu.Lock()
	s.shutdown = true
	s.mu.Unlock()

	close(s.rideRequests)
	fmt.Println("[Server] Shutdown initiated")
}

func main() {
	// Note: As of Go 1.20, the global random generator is automatically seeded

	fmt.Println("=== TaxiScheduler System Starting ===")
	fmt.Println()

	// Create the server (API gateway)
	server := NewServer()

	// Create clients that use the server API
	taxiClient := NewTaxiClient(server)
	userClient := NewUserClient(server)

	// Start taxi client in background (15 taxis, 1 per 5 seconds = ~75 seconds)
	go taxiClient.Start()

	// Wait for some taxis to register before accepting rides
	fmt.Println("[Main] Waiting 10 seconds for initial taxis to register...")
	time.Sleep(10 * time.Second)

	// Start user client (blocks until all 100 requests sent)
	userClient.Start()

	// Shutdown the server
	server.Shutdown()

	// Wait for remaining rides to complete
	fmt.Println("[Main] All requests sent, waiting for rides to complete...")
	time.Sleep(30 * time.Second)

	fmt.Println()
	fmt.Println("=== TaxiScheduler System Finished ===")
}
