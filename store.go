// store.go - Thread-safe taxi storage
// Provides concurrent access protection using sync.RWMutex

package main

import "sync"

// TaxiStore holds all taxi data with concurrent access protection.
// Uses a map for O(1) lookup by TaxiID.
// All public methods are safe for concurrent access from multiple goroutines.
type TaxiStore struct {
	mu     sync.RWMutex   // Read-write mutex for concurrent access
	taxis  map[int]*Taxi  // Map from taxi ID to Taxi pointer
	nextID int            // Auto-incrementing ID counter
}

// NewTaxiStore creates and returns an initialized TaxiStore.
func NewTaxiStore() *TaxiStore {
	return &TaxiStore{
		taxis:  make(map[int]*Taxi),
		nextID: 1,
	}
}

// Add inserts a new taxi at the given location and returns its assigned ID.
// The taxi is marked as available by default.
func (ts *TaxiStore) Add(location Location) int {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	id := ts.nextID
	ts.nextID++

	ts.taxis[id] = &Taxi{
		ID:          id,
		Location:    location,
		IsAvailable: true,
	}

	return id
}

// Get retrieves a taxi by ID. Returns nil if not found.
func (ts *TaxiStore) Get(id int) *Taxi {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.taxis[id]
}

// GetAllAvailable returns a slice of all taxis that can accept rides.
func (ts *TaxiStore) GetAllAvailable() []*Taxi {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	available := make([]*Taxi, 0)
	for _, taxi := range ts.taxis {
		if taxi.IsAvailable {
			available = append(available, taxi)
		}
	}
	return available
}

// SetAvailability updates a taxi's availability status.
// Returns false if the taxi was not found.
func (ts *TaxiStore) SetAvailability(id int, available bool) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	taxi, exists := ts.taxis[id]
	if !exists {
		return false
	}
	taxi.IsAvailable = available
	return true
}

// UpdateLocation updates a taxi's location.
// Returns false if the taxi was not found.
func (ts *TaxiStore) UpdateLocation(id int, location Location) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	taxi, exists := ts.taxis[id]
	if !exists {
		return false
	}
	taxi.Location = location
	return true
}

// Count returns the total number of taxis in the store.
func (ts *TaxiStore) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.taxis)
}
