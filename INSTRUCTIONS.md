# Core entities

`Taxi`
- implemented in `taxi_client.go`, sends requests for new taxi creation to `server.go`.
- Has a location (simple (X,Y) coordinate)
`Client`
- implemented in `user_client.go`, sends requests for taxi rides to `server.go`.
- Has a location (simple (X,Y) coordinate)
`Ride`
- an entity representing a ride
- has `startLocation` and `endLocation`, both are (X,Y) coordinates.
- The lifecycle of a ride is: `CREATED` -> `ASSIGNED`-> `IN_PROGRESS` -> `FINISHED`
`TaxiManager`
- Responsible for managing taxis - creation, update, deletion.
`TaxiAssigner`
- Responsible for assigning available taxis to pending rides.
- Assigns the closest Taxi to the requested ride, using `LocationService`.
`RideScheduler`
- Respobsible for beggining and ending rides.
- The duration of a ride is calculated by: (distance between taxi to `startLocation` + distance between `startLocation` and `endLocation`).
- Rate limited to 1 new rides every 3 seconds.
`TaxiStore`
- Holds all taxi data. Uses mutex for concurrent access.
- Use an efficient data structure for managing taxis (hashtable that points to an array)
`LocationService`
- Responsible for calculating the distance between point A to point B.


# Tasks
- Implement `user_client.go` so that it sends 100 ride requests, rate limited to 1 request every 5 seconds
- Implement `taxi_client.go` so that it sends 15 requests for new taxi creation, rate limited to 1 request every 5 seconds
- Ride requests will be handled through a channel called "rideRequests" that holds pending rides. `RideScheduler` ranges over this channel and handles a request every 3 seconds.

# General instructions
- Stick closely to the KISS principle. I want pure simplicity over robustness or production-readiness.
- Thoroughly add comments to everything. Make the code very easy to understand.
- Use the current codebase as a starting-off point, but feel free to make any necessary changes to it.

