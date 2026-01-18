// assigner.go - Taxi assignment logic
// Assigns the closest available taxi to ride requests

package main

import (
	"fmt"
	"log"
)

// TaxiAssigner handles assigning taxis to rides.
// Uses LocationService to find the nearest available taxi.
type TaxiAssigner struct {
	store           *TaxiStore       // Reference to taxi storage
	locationService *LocationService // For distance calculations
}

// NewTaxiAssigner creates a TaxiAssigner with the given dependencies.
func NewTaxiAssigner(store *TaxiStore, locationService *LocationService) *TaxiAssigner {
	return &TaxiAssigner{
		store:           store,
		locationService: locationService,
	}
}

// AssignClosestTaxi finds and assigns the nearest available taxi to a ride.
// Updates the ride's TaxiID and Status fields.
// Returns the assigned taxi, or nil if no taxis are available.
func (ta *TaxiAssigner) AssignClosestTaxi(ride *Ride) *Taxi {
	availableTaxis := ta.store.GetAllAvailable()

	if len(availableTaxis) == 0 {
		fmt.Printf("[TaxiAssigner] No taxis available for ride #%d\n", ride.ID)
		return nil
	}

	// Find the closest taxi to the ride's start location
	var closestTaxi *Taxi
	closestDistance := -1 // -1 indicates no taxi found yet

	for _, taxi := range availableTaxis {
		distance := ta.locationService.CalculateDistance(taxi.Location, ride.StartLocation)

		if closestDistance == -1 || distance < closestDistance {
			closestDistance = distance
			closestTaxi = taxi
		}
	}

	// Mark the taxi as unavailable and assign it to the ride
	if closestTaxi != nil {
		if !ta.store.SetAvailability(closestTaxi.ID, false) {
			log.Printf("[TaxiAssigner] ERROR: Failed to mark taxi #%d as unavailable\n", closestTaxi.ID)
			return nil
		}
		ride.mu.Lock()
		ride.TaxiID = closestTaxi.ID
		ride.Status = ASSIGNED
		ride.mu.Unlock()
		fmt.Printf("[TaxiAssigner] Assigned taxi #%d to ride #%d (distance: %d)\n",
			closestTaxi.ID, ride.ID, closestDistance)
	}

	return closestTaxi
}

// CalculateRideDuration computes the total duration of a ride.
// Duration = distance(taxi -> pickup) + distance(pickup -> destination)
func (ta *TaxiAssigner) CalculateRideDuration(taxi *Taxi, ride *Ride) int {
	pickupDistance := ta.locationService.CalculateDistance(taxi.Location, ride.StartLocation)
	rideDistance := ta.locationService.CalculateDistance(ride.StartLocation, ride.EndLocation)
	return pickupDistance + rideDistance
}
