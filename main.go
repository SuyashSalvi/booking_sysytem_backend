//go get -u github.com/lib/pq
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"github.com/lib/pq"
	"strings"
)

// MovieBooking represents a movie booking
type MovieBooking struct {
	RoomNo      int
	SeatNo      string
	BookingFlag bool
	MovieName   string
	MovieTime   time.Time
}

// Database connection details
const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "root"
	dbName     = "postgres"
)

var (
	db  *sql.DB
	mu  sync.Mutex
)

// Movie represents a movie
type Movie struct {
	MovieName string
	About     string
	Rating    float64
	Hours     int64
	Lang      pq.StringArray `json:"lang"`
	Genre     pq.StringArray `json:"genre"`
}

// TheatreBooking represents a booking in the theatre
type TheatreBooking struct {
	RoomNo      int
	SeatNo      string
	BookingFlag bool
	MovieName   string
	MovieTime   time.Time
}


func init() {
	// Initialize the database connection
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the database connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func parseArray(s string) []string {
    // This is a simple example; you might need to handle more complex cases
    // based on your specific PostgreSQL array format
    values := strings.Split(s, ",")
    for i := range values {
        values[i] = strings.TrimSpace(values[i])
    }
    return values
}

// Handler function for the /movies endpoint
func getMoviesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM movies")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		var lang, genre string
		err := rows.Scan(&movie.MovieName, &movie.About, &movie.Rating, &movie.Hours,&lang, &genre)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		movie.Lang = parseArray(lang)
        movie.Genre = parseArray(genre)
		movies = append(movies, movie)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// Handler function for the /theatre endpoint
func getTheatreHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM theatre")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var theatreBookings []TheatreBooking
	for rows.Next() {
		var booking TheatreBooking
		err := rows.Scan(&booking.RoomNo, &booking.SeatNo, &booking.BookingFlag, &booking.MovieName, &booking.MovieTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		theatreBookings = append(theatreBookings, booking)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(theatreBookings)
}

func bookMovie(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	roomNo := r.FormValue("room_no")
	seatNo := r.FormValue("seat_no")
	movieName := r.FormValue("movie")
	movieTime := r.FormValue("movie_time")

	// Perform booking
	err := performBooking(roomNo, seatNo, movieName, movieTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success
	fmt.Fprintf(w, "Booking successful for seat %s in room %s for movie %s at %s\n", seatNo, roomNo, movieName, movieTime)
}

func performBooking(roomNo, seatNo, movieName, movieTime string) error {
	// Lock the mutex to handle race conditions
	mu.Lock()
	defer mu.Unlock()

	// Convert roomNo to int
	roomNoInt := 0
	fmt.Sscanf(roomNo, "%d", &roomNoInt)

	// Check if the seat is available
	bookingFlag, err := isSeatAvailable(roomNoInt, seatNo)
	if err != nil {
		return err
	}

	if !bookingFlag {
		return fmt.Errorf("Seat %s in room %s is not available", seatNo, roomNo)
	}

	// Update the booking_flag to true
	_, err = db.Exec("UPDATE theatre SET booking_flag = true WHERE room_no = $1 AND seat_no = $2", roomNoInt, seatNo)
	if err != nil {
		return err
	}

	return nil
}

func isSeatAvailable(roomNo int, seatNo string) (bool, error) {
	var bookingFlag bool
	err := db.QueryRow("SELECT booking_flag FROM theatre WHERE room_no = $1 AND seat_no = $2", roomNo, seatNo).Scan(&bookingFlag)
	if err != nil {
		fmt.Println("No booking found for room", roomNo, "seat", seatNo)
		return false, err
	}
	return !bookingFlag, nil
}

func main() {
	// Create the theatre table if it doesn't exist
	// _, err := db.Exec(`CREATE TABLE IF NOT EXISTS theatre (
	// 	room_no INT,
	// 	seat_no VARCHAR(50),
	// 	booking_flag BOOLEAN,
	// 	movie_name VARCHAR(100),
	// 	movie_time TIME
	// )`)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Handle booking requests
	http.HandleFunc("/book", bookMovie)
	http.HandleFunc("/movies", getMoviesHandler)
	http.HandleFunc("/theatre", getTheatreHandler)

	// Start the server
	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
