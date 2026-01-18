// location.go - Distance calculation service
// Provides utilities for calculating distances between locations

package main

// LocationService handles distance calculations between locations.
// Uses Manhattan distance for simplicity (grid-based movement).
type LocationService struct{}

// NewLocationService creates a new LocationService instance.
func NewLocationService() *LocationService {
	return &LocationService{}
}

// CalculateDistance returns the Manhattan distance between two locations.
// Manhattan distance = |x1 - x2| + |y1 - y2|
// This is simpler than Euclidean distance and uses only integer arithmetic.
func (ls *LocationService) CalculateDistance(from, to Location) int {
	xDist := from.X - to.X
	yDist := from.Y - to.Y

	// Calculate absolute values without using math package
	if xDist < 0 {
		xDist = -xDist
	}
	if yDist < 0 {
		yDist = -yDist
	}

	return xDist + yDist
}
