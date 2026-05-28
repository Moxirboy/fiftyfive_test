package providers

import "context"

type ProviderSearch struct {
	From          string
	To            string
	DepartureDate string
	ReturnDate    string
	Currency      string
	Adults        int
	Children      int
	Infants       int
}

type ProviderOffer struct {
	ProviderRef  string
	From         string
	To           string
	Airline      string
	FlightNumber string
	Currency     string
	PriceAdult   int64
	PriceChild   int64
	PriceInfant  int64
}

type FlightProvider interface {
	Search(ctx context.Context, req ProviderSearch) ([]ProviderOffer, error)
	Name() string
}
