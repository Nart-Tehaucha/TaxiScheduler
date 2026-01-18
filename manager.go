// manager.go - Taxi management operations
// Provides a business logic layer over TaxiStore for taxi CRUD operations

package main

import "fmt"

// TaxiManager handles taxi creation, update, and deletion.
// Acts as a wrapper around TaxiStore with logging.
type TaxiManager struct {
	store *TaxiStore // Reference to the underlying taxi storage
}

// NewTaxiManager creates a TaxiManager with the given store.
func NewTaxiManager(store *TaxiStore) *TaxiManager {
	return &TaxiManager{store: store}
}

// CreateTaxi registers a new taxi at the given location.
// Returns the new taxi's ID.
func (tm *TaxiManager) CreateTaxi(location Location) int {
	id := tm.store.Add(location)
	fmt.Printf("[TaxiManager] Created taxi #%d at (%d, %d)\n", id, location.X, location.Y)
	return id
}

// GetTaxi retrieves a taxi by ID. Returns nil if not found.
func (tm *TaxiManager) GetTaxi(id int) *Taxi {
	return tm.store.Get(id)
}

// UpdateTaxiLocation moves a taxi to a new location.
// Returns an error if the taxi was not found.
func (tm *TaxiManager) UpdateTaxiLocation(id int, location Location) error {
	if !tm.store.UpdateLocation(id, location) {
		return fmt.Errorf("taxi #%d not found", id)
	}
	fmt.Printf("[TaxiManager] Taxi #%d moved to (%d, %d)\n", id, location.X, location.Y)
	return nil
}

// GetAvailableTaxis returns all taxis that can accept rides.
func (tm *TaxiManager) GetAvailableTaxis() []*Taxi {
	return tm.store.GetAllAvailable()
}
