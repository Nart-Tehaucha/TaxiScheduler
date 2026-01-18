# TaxiScheduler Implementation Plan

## Overview

Go-based taxi dispatch system with concurrent clients sending requests through channels to a central scheduler.

## File Structure

```
TaxiScheduler/
├── go.mod              # Module: taxischeduler
├── types.go            # All type definitions
├── location.go         # LocationService (distance calculation)
├── store.go            # TaxiStore (mutex-protected storage)
├── manager.go          # TaxiManager (CRUD operations)
├── assigner.go         # TaxiAssigner (closest taxi logic)
├── scheduler.go        # RideScheduler (channel processing, rate-limited)
├── taxi_client.go      # Sends 15 taxi registrations (1 per 5s)
├── user_client.go      # Sends 100 ride requests (1 per 5s)
└── server.go           # Main orchestration
```

## Implementation Order

### Step 1: go.mod
```
module taxischeduler
go 1.21
```

### Step 2: types.go
```go
type Location struct { X, Y int }

type RideStatus int
const (
    CREATED RideStatus = iota
    ASSIGNED
    IN_PROGRESS
    FINISHED
)

type Taxi struct {
    ID          int
    Location    Location
    IsAvailable bool
}

type Ride struct {
    ID            int
    ClientID      int
    TaxiID        int
    StartLocation Location
    EndLocation   Location
    Status        RideStatus
}

type RideRequest struct {
    ClientID      int
    StartLocation Location
    EndLocation   Location
}

type TaxiRegistration struct {
    Location Location
}
```

### Step 3: location.go
- `LocationService` struct
- `CalculateDistance(from, to Location) int` - Manhattan distance: `|x1-x2| + |y1-y2|`

### Step 4: store.go
- `TaxiStore` with `sync.RWMutex` and `map[int]*Taxi`
- Methods: `Add(location) int`, `Get(id) *Taxi`, `GetAllAvailable() []*Taxi`, `SetAvailability(id, bool)`, `UpdateLocation(id, Location)`

### Step 5: manager.go
- `TaxiManager` wraps `TaxiStore`
- Methods: `CreateTaxi(location) int`, `GetTaxi(id) *Taxi`

### Step 6: assigner.go
- `TaxiAssigner` with `TaxiStore` and `LocationService`
- `AssignClosestTaxi(ride *Ride) *Taxi` - finds nearest available taxi, marks unavailable
- `CalculateRideDuration(taxi, ride) int` - distance(taxi→start) + distance(start→end)

### Step 7: scheduler.go
- `RideScheduler` with `chan RideRequest`, `TaxiAssigner`, `TaxiStore`, `LocationService`
- `Start()` - ranges over channel with `time.Ticker(3 * time.Second)` rate limit
- `processRequest(RideRequest)` - creates Ride, assigns taxi, starts ride goroutine
- `endRide(ride, taxi)` - marks FINISHED, updates taxi location, sets available

### Step 8: taxi_client.go
- `TaxiClient` with `chan TaxiRegistration`
- `Start()` - sends 15 registrations with random locations, `time.Sleep(5 * time.Second)` between each

### Step 9: user_client.go
- `UserClient` with `chan RideRequest`
- `Start()` - sends 100 requests with random start/end locations, `time.Sleep(5 * time.Second)` between each

### Step 10: server.go (main)
```go
func main() {
    rand.Seed(time.Now().UnixNano())

    // Initialize
    locationService := NewLocationService()
    taxiStore := NewTaxiStore()
    taxiManager := NewTaxiManager(taxiStore)
    taxiAssigner := NewTaxiAssigner(taxiStore, locationService)

    // Channels
    taxiRegistrations := make(chan TaxiRegistration, 20)
    rideRequests := make(chan RideRequest, 150)

    // Components
    rideScheduler := NewRideScheduler(rideRequests, taxiAssigner, taxiStore, locationService)
    taxiClient := NewTaxiClient(taxiRegistrations)
    userClient := NewUserClient(rideRequests)

    // Start goroutines
    go handleTaxiRegistrations(taxiManager, taxiRegistrations)
    go rideScheduler.Start()
    go taxiClient.Start()

    // Wait for some taxis before accepting rides
    time.Sleep(10 * time.Second)

    userClient.Start()  // Blocks until done

    // Wait for rides to complete
    time.Sleep(30 * time.Second)
}
```

## Concurrency Model

```
TaxiClient ──[taxiRegistrations]──▶ handleTaxiRegistrations ──▶ TaxiManager ──▶ TaxiStore
                                                                                    ▲
UserClient ──[rideRequests]──▶ RideScheduler (1/3s) ──▶ TaxiAssigner ───────────────┘
                                     │
                                     ▼
                              ride goroutines (simulate duration, then endRide)
```

## Bugs to Fix in Existing server.go

1. `for t := range rm.taxis` - iterates indices, not values; change to `for _, t := range`
2. `rm.taxis[taxi_id] = ...` - will panic; slice not pre-allocated
3. Missing Location field in Taxi struct
4. Missing StartLocation/EndLocation in Ride struct
5. Incomplete AssignTaxi and BeginRide methods

## Rate Limits

| Component | Rate | Count | Duration |
|-----------|------|-------|----------|
| TaxiClient | 1/5s | 15 | ~75s |
| UserClient | 1/5s | 100 | ~500s |
| RideScheduler | 1/3s | - | processes as received |

## Code Style

- KISS principle: simplicity over robustness
- Thorough comments on all types, methods, and non-obvious logic
- Use `fmt.Printf` for logging (no external dependencies)
