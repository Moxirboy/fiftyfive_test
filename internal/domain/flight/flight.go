package flight

import (
	"time"

	"flysoft-flight-service/internal/money"
)

type SearchCriteria struct {
	From          string
	To            string
	DepartureDate time.Time
	ReturnDate    *time.Time
	Adults        int
	Children      int
	Infants       int
	Currency      string
}

type Offer struct {
	ID            int64
	OfferID       string
	Provider      string
	From          string
	To            string
	DepartureDate time.Time
	ReturnDate    *time.Time
	Airline       string
	FlightNumber  string
	BasePrice     money.Money
	Commission    money.Money
	ServiceFee    money.Money
	TotalPrice    money.Money
	Profit        money.Money
	Currency      string
	ExpiresAt     time.Time
	CreatedAt     time.Time
}
