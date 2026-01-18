// scheduler.go - Ride request processing
// Handles ride lifecycle from CREATED to FINISHED with rate limiting

package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// RideScheduler processes ride requests from a channel.
// Rate-limited to handle 1 new ride every 3 seconds.
type RideScheduler struct {
	rideRequests    <-chan RideRequest // Input channel for ride requests
	assigner        *TaxiAssigner      // For assigning taxis to rides
	store           *TaxiStore         // For updating taxi state after rides
	locationService *LocationService   // For calculating ride durations
	mu              sync.Mutex         // Protects nextRideID
	nextRideID      int                // Auto-incrementing ride ID
}

// NewRideScheduler creates a RideScheduler with the given dependencies.
func NewRideScheduler(
	rideRequests <-chan RideRequest,
	assigner *TaxiAssigner,
	store *TaxiStore,
	locationService *LocationService,
) *RideScheduler {
	return &RideScheduler{
		rideRequests:    rideRequests,
		assigner:        assigner,
		store:           store,
		locationService: locationService,
		nextRideID:      1,
	}
}

// Start begins processing ride requests from the channel.
// This method blocks and should be run as a goroutine.
// Processes one ride every 3 seconds (rate limited).
func (rs *RideScheduler) Start() {
	fmt.Println("[RideScheduler] Started - waiting for ride requests...")

	// Rate limiter: 1 request every 3 seconds
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for request := range rs.rideRequests {
		// Wait for rate limit tick before processing
		<-ticker.C
		rs.processRequest(request)
	}

	fmt.Println("[RideScheduler] Channel closed, stopping...")
}

// processRequest handles a single ride request.
// Creates a ride, assigns a taxi, and starts the ride simulation.
func (rs *RideScheduler) processRequest(request RideRequest) {
	// Get next ride ID with mutex protection
	rs.mu.Lock()
	rideID := rs.nextRideID
	rs.nextRideID++
	rs.mu.Unlock()

	// Create a new ride with CREATED status
	ride := &Ride{
		ID:            rideID,
		ClientID:      request.ClientID,
		StartLocation: request.StartLocation,
		EndLocation:   request.EndLocation,
		Status:        CREATED,
	}

	fmt.Printf("[RideScheduler] Created ride #%d for client #%d: (%d,%d) -> (%d,%d)\n",
		ride.ID, ride.ClientID,
		ride.StartLocation.X, ride.StartLocation.Y,
		ride.EndLocation.X, ride.EndLocation.Y)

	// Try to assign a taxi
	taxi := rs.assigner.AssignClosestTaxi(ride)
	if taxi == nil {
		fmt.Printf("[RideScheduler] Ride #%d could not be assigned (no available taxis)\n", ride.ID)
		return
	}

	// Calculate ride duration and start the ride
	duration := rs.assigner.CalculateRideDuration(taxi, ride)
	rs.startRide(ride, taxi, duration)
}

// startRide begins a ride and schedules its completion.
// The ride completion is simulated in a separate goroutine.
func (rs *RideScheduler) startRide(ride *Ride, taxi *Taxi, duration int) {
	ride.mu.Lock()
	ride.Status = IN_PROGRESS
	ride.mu.Unlock()

	fmt.Printf("[RideScheduler] Ride #%d IN_PROGRESS - taxi #%d, duration: %d units\n",
		ride.ID, taxi.ID, duration)

	// Simulate ride completion in a goroutine
	// Duration is converted to seconds for simulation (1 unit = 100ms for faster demo)
	go func(r *Ride, t *Taxi, d int) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[RideScheduler] ERROR: Panic in endRide goroutine for ride #%d: %v\n", r.ID, err)
			}
		}()
		time.Sleep(time.Duration(d) * 100 * time.Millisecond)
		rs.endRide(r, t)
	}(ride, taxi, duration)
}

// endRide completes a ride and frees the taxi.
// Updates the taxi's location to the ride destination and marks it available.
func (rs *RideScheduler) endRide(ride *Ride, taxi *Taxi) {
	ride.mu.Lock()
	ride.Status = FINISHED
	ride.mu.Unlock()

	// Update taxi location to ride destination and mark available
	if !rs.store.UpdateLocation(taxi.ID, ride.EndLocation) {
		log.Printf("[RideScheduler] ERROR: Failed to update location for taxi #%d\n", taxi.ID)
	}
	if !rs.store.SetAvailability(taxi.ID, true) {
		log.Printf("[RideScheduler] ERROR: Failed to set availability for taxi #%d\n", taxi.ID)
	}

	fmt.Printf("[RideScheduler] Ride #%d FINISHED - taxi #%d now at (%d, %d) and available\n",
		ride.ID, taxi.ID, ride.EndLocation.X, ride.EndLocation.Y)
}
