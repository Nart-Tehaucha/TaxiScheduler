# TaxiScheduler System Summary

## Entity Summary

| Entity | Purpose | Depends On | Depended On By |
|--------|---------|------------|----------------|
| **Location** | Simple (X,Y) coordinate struct | None | Taxi, Ride, RideRequest, TaxiRegistration, LocationService |
| **RideStatus** | Enum for ride lifecycle (CREATED→ASSIGNED→IN_PROGRESS→FINISHED) | None | Ride, RideScheduler |
| **Taxi** | Vehicle with ID, location, availability | Location | TaxiStore, TaxiAssigner, RideScheduler |
| **Ride** | Ride request with start/end locations and status | Location, RideStatus | RideScheduler, TaxiAssigner |
| **RideRequest** | Channel message for new ride requests | Location | UserClient, RideScheduler |
| **TaxiRegistration** | Channel message for new taxi registrations | Location | TaxiClient, Server |
| **LocationService** | Calculates Manhattan distance between points | Location | TaxiAssigner, RideScheduler |
| **TaxiStore** | Thread-safe taxi storage (map + RWMutex) | Taxi, Location | TaxiManager, TaxiAssigner, RideScheduler |
| **TaxiManager** | CRUD operations for taxis | TaxiStore, Location | Server |
| **TaxiAssigner** | Finds & assigns closest available taxi | TaxiStore, LocationService, Ride, Taxi | RideScheduler |
| **RideScheduler** | Processes ride requests (1/3s rate limit) | TaxiAssigner, TaxiStore, LocationService, RideRequest, Ride | Server |
| **Server** | API gateway for all client requests | TaxiManager, RideScheduler, channels | TaxiClient, UserClient |
| **TaxiClient** | Sends 15 taxi registrations (1/5s) | Server | main() |
| **UserClient** | Sends 100 ride requests (1/5s) | Server | main() |

---

## System Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              main() / server.go                             │
│  - Initializes all components                                               │
│  - Creates channels                                                         │
│  - Starts goroutines                                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
          ┌───────────────────────────┼───────────────────────────┐
          │                           │                           │
          ▼                           ▼                           ▼
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│   TaxiClient     │      │   UserClient     │      │  RideScheduler   │
│  (goroutine)     │      │  (main thread)   │      │   (goroutine)    │
│                  │      │                  │      │                  │
│ Sends 15 taxis   │      │ Sends 100 rides  │      │ Processes rides  │
│ @ 1 per 5 sec    │      │ @ 1 per 5 sec    │      │ @ 1 per 3 sec    │
└────────┬─────────┘      └────────┬─────────┘      └────────┬─────────┘
         │                         │                         │
         │                         │                         │
         ▼                         ▼                         │
┌─────────────────────────────────────────────────┐          │
│                    Server                       │          │
│               (API Gateway)                     │          │
│                                                 │          │
│  RegisterTaxi(location) ──► TaxiManager         │          │
│  RequestRide(req) ──────────► rideRequests ─────┼──────────┘
│                                 channel         │
└─────────────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│  TaxiManager     │─────▶│   TaxiStore      │◀─────│  TaxiAssigner    │
│                  │      │                  │      │                  │
│ CreateTaxi()     │      │ map[int]*Taxi    │      │ AssignClosest-   │
│ GetTaxi()        │      │ sync.RWMutex     │      │ Taxi()           │
│ UpdateLocation() │      │                  │      │ CalculateRide-   │
└──────────────────┘      │ Add()            │      │ Duration()       │
                          │ Get()            │      └────────┬─────────┘
                          │ GetAllAvailable()│               │
                          │ SetAvailability()│               │
                          │ UpdateLocation() │               │
                          └──────────────────┘               │
                                   ▲                         │
                                   │                         ▼
                                   │              ┌──────────────────┐
                                   │              │ LocationService  │
                                   │              │                  │
                                   │              │ CalculateDistance│
                                   │              │ (Manhattan)      │
                                   │              └──────────────────┘
                                   │
                                   │
┌──────────────────────────────────┴──────────────────────────────────┐
│                         Ride Lifecycle                              │
│                                                                     │
│  ┌─────────┐    ┌──────────┐    ┌─────────────┐    ┌──────────┐    │
│  │ CREATED │───▶│ ASSIGNED │───▶│ IN_PROGRESS │───▶│ FINISHED │    │
│  └─────────┘    └──────────┘    └─────────────┘    └──────────┘    │
│       │              │                 │                 │          │
│  RideScheduler  TaxiAssigner     RideScheduler    RideScheduler    │
│  creates ride   assigns taxi     starts ride      ends ride,       │
│                 marks taxi       (goroutine)      frees taxi       │
│                 unavailable                                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Summary

1. **Taxi Registration Flow**:
   ```
   TaxiClient → Server.RegisterTaxi() → TaxiManager → TaxiStore
   ```

2. **Ride Request Flow**:
   ```
   UserClient → Server.RequestRide() → rideRequests channel → RideScheduler → TaxiAssigner → TaxiStore
                                                                   │
                                                                   ▼
                                                             Ride goroutine → endRide() → TaxiStore (free taxi)
   ```

3. **Distance Calculation Flow**:
   ```
   TaxiAssigner/RideScheduler → LocationService.CalculateDistance(from, to)
   ```

---

## Rate Limits

| Component | Rate | Count | Duration |
|-----------|------|-------|----------|
| TaxiClient | 1 per 5s | 15 | ~75s |
| UserClient | 1 per 5s | 100 | ~500s |
| RideScheduler | 1 per 3s | - | processes as received |
