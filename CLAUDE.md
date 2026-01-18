# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TaxiScheduler is a Go-based taxi dispatch and scheduling system. It simulates concurrent taxi clients and user clients making requests to a central server that manages ride assignment and lifecycle.

## Build and Run

```bash
# Run the application
go run .

# Run with race detection (recommended during development)
go run -race *.go

# Run tests (when test files exist)
go test ./...
```

## Architecture

### Core Entities
- **Taxi**: Vehicle with (X,Y) location coordinates and availability status
- **Client**: User with (X,Y) location requesting rides
- **Ride**: Request with startLocation, endLocation, and lifecycle status

### Ride Lifecycle
`CREATED` → `ASSIGNED` → `IN_PROGRESS` → `FINISHED`

### Key Components
- **TaxiManager**: CRUD operations for taxis
- **TaxiAssigner**: Assigns closest available taxi to pending rides using LocationService
- **RideScheduler**: Manages ride lifecycle, rate-limited to 1 ride every 3 seconds
- **TaxiStore**: Thread-safe taxi storage using mutex with hashtable+array structure
- **LocationService**: Distance calculations between coordinates

### Concurrency Model
- Ride requests flow through a `rideRequests` channel
- RideScheduler ranges over this channel, processing one request every 3 seconds
- TaxiStore uses mutex for concurrent access protection

## Implementation Requirements

From INSTRUCTIONS.md:
- `user_client.go`: Send 100 ride requests, rate-limited to 1 per 5 seconds
- `taxi_client.go`: Send 15 taxi creation requests, rate-limited to 1 per 5 seconds

## Code Style

- Follow KISS principle: simplicity over robustness or production-readiness
- Add thorough comments to make code easy to understand
- Use existing codebase patterns as reference
