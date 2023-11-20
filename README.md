# Theater Booking API

## Overview

The Theater Booking API is a Go-based RESTful API for booking seats in a theater. It is designed to handle concurrent seat booking requests using a mutex to prevent race conditions.

## Features

- **Seamless Booking:** Book your favorite seats for the latest movies hassle-free.

- **Concurrency Control:** Utilizes a mutex to handle race conditions and ensure safe and accurate seat bookings, even with concurrent requests.


## API Endpoints

- **Book a Seat:**

    ```http
    POST /book
    ```

    Example request body:

    ```json
    {
        "room_no": 10,
        "seat_no": "A10",
        "movie": "Hulk",
        "time": "2023-11-19 14:30:00"
    }
    ```

    Example response:

    ```json
    {
        "message": "Booking successful for seat A10 in room 10 for movie Hulk at 2023-11-19 14:30:00"
    }
    ```

## Concurrency Control

To ensure seamless booking in a multi-user environment, the API uses a mutex to handle race conditions when updating the booking status of seats. This prevents conflicts and guarantees that a seat is booked only once, even with concurrent requests.

```go
var (
    mu sync.Mutex
)

func performBooking(roomNo, seatNo, movieName, movieTime string) error {
    // Lock the mutex to handle race conditions
    mu.Lock()
    defer mu.Unlock()

    // Your booking logic here...

    return nil
}
