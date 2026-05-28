package booking

import "time"

type PassengerType string

const (
	PassengerTypeAdult  PassengerType = "adult"
	PassengerTypeChild  PassengerType = "child"
	PassengerTypeInfant PassengerType = "infant"
)

type Booking struct {
	ID         int64
	BookingID  string
	OfferID    string
	Status     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	Passengers []Passenger
}

type Passenger struct {
	ID             int64
	Type           PassengerType
	FirstName      string
	LastName       string
	DocumentNumber string
	CreatedAt      time.Time
}
