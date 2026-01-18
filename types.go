// types.go - Core data structures for the TaxiScheduler system
// All shared types are defined here to prevent circular dependencies

package main

import "sync"

// Location represents a simple 2D coordinate in the system.
// Used for taxi positions and ride start/end points.
type Location struct {
	X int // X coordinate
	Y int // Y coordinate
}

// RideStatus represents the lifecycle state of a ride.
// A ride progresses through these states in order: CREATED -> ASSIGNED -> IN_PROGRESS -> FINISHED
type RideStatus int

const (
	CREATED     RideStatus = iota // Ride has been requested, awaiting taxi assignment
	ASSIGNED                      // Taxi has been assigned, ride not yet started
	IN_PROGRESS                   // Ride is currently happening
	FINISHED                      // Ride has been completed
)

// Taxi represents a taxi vehicle in the system.
type Taxi struct {
	ID          int      // Unique identifier for the taxi
	Location    Location // Current (X,Y) position of the taxi
	IsAvailable bool     // Whether the taxi can accept new rides
}

// Ride represents a ride request and its current state.
// The mu mutex protects concurrent access to Status and TaxiID fields.
type Ride struct {
	mu            sync.Mutex // Protects Status and TaxiID
	ID            int        // Unique identifier for the ride
	ClientID      int        // ID of the client who requested the ride
	TaxiID        int        // ID of the assigned taxi (0 if unassigned)
	StartLocation Location   // Pickup point
	EndLocation   Location   // Destination
	Status        RideStatus // Current lifecycle state
}

// RideRequest is sent through the rideRequests channel for processing.
// Contains the information needed to create a new Ride.
type RideRequest struct {
	ClientID      int      // ID of the requesting client
	StartLocation Location // Pickup point
	EndLocation   Location // Destination
}
