package mockavia_test

import (
	"context"
	"errors"
	"testing"

	"flysoft-flight-service/internal/providers"
	"flysoft-flight-service/internal/providers/mockavia"
)

func TestProviderSearchNormalizesSupplierRoute(t *testing.T) {
	provider := mockavia.New()

	offers, err := provider.Search(context.Background(), providers.ProviderSearch{
		From: "DYU",
		To:   "IST",
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(offers) != 1 {
		t.Fatalf("Search() returned %d offers, want 1", len(offers))
	}

	got := offers[0]
	if got.ProviderRef != "SUP-001" {
		t.Fatalf("ProviderRef = %q, want %q", got.ProviderRef, "SUP-001")
	}
	if got.From != "DYU" {
		t.Fatalf("From = %q, want %q", got.From, "DYU")
	}
	if got.To != "IST" {
		t.Fatalf("To = %q, want %q", got.To, "IST")
	}
	if got.Airline != "TK" {
		t.Fatalf("Airline = %q, want %q", got.Airline, "TK")
	}
	if got.FlightNumber != "TK255" {
		t.Fatalf("FlightNumber = %q, want %q", got.FlightNumber, "TK255")
	}
	if got.PriceAdult != 30000 {
		t.Fatalf("PriceAdult = %d, want %d", got.PriceAdult, int64(30000))
	}
	if got.PriceChild != 20000 {
		t.Fatalf("PriceChild = %d, want %d", got.PriceChild, int64(20000))
	}
	if got.PriceInfant != 0 {
		t.Fatalf("PriceInfant = %d, want %d", got.PriceInfant, int64(0))
	}
	if got.Currency != "USD" {
		t.Fatalf("Currency = %q, want %q", got.Currency, "USD")
	}
}

func TestProviderSearchEmptyMatchReturnsEmptySlice(t *testing.T) {
	provider := mockavia.New()

	offers, err := provider.Search(context.Background(), providers.ProviderSearch{
		From: "LHR",
		To:   "JFK",
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if offers == nil {
		t.Fatal("Search() returned nil, want empty slice")
	}
	if len(offers) != 0 {
		t.Fatalf("Search() returned %d offers, want 0", len(offers))
	}
}

func TestProviderSearchRespectsCanceledContext(t *testing.T) {
	provider := mockavia.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Search(ctx, providers.ProviderSearch{
		From: "DYU",
		To:   "IST",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Search() error = %v, want %v", err, context.Canceled)
	}
}
