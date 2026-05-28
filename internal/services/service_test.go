package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/database/models"
	"flysoft-flight-service/internal/dto"
	"flysoft-flight-service/internal/money"
	"flysoft-flight-service/internal/pricing"
	"flysoft-flight-service/internal/providers"
	"flysoft-flight-service/internal/repository"
	"flysoft-flight-service/internal/services"
)

var fixedNow = time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)

func TestFlightServiceSearch(t *testing.T) {
	request := dto.SearchRequest{
		From:          "DYU",
		To:            "IST",
		DepartureDate: "2026-06-15",
		ReturnDate:    "2026-06-25",
		Adults:        1,
		Children:      1,
		Infants:       0,
		Currency:      "USD",
	}

	tests := []struct {
		name           string
		providerOffers []providers.ProviderOffer
		providerErr    error
		wantCode       string
		assert         func(t *testing.T, repo *fakeFlightOfferRepository, provider *fakeFlightProvider, got []dto.OfferResponse)
	}{
		{
			name: "happy path prices and persists offers",
			providerOffers: []providers.ProviderOffer{
				{
					ProviderRef:  "SUP-001",
					From:         "DYU",
					To:           "IST",
					Airline:      "TK",
					FlightNumber: "TK255",
					Currency:     "USD",
					PriceAdult:   30000,
					PriceChild:   20000,
					PriceInfant:  0,
				},
			},
			assert: func(t *testing.T, repo *fakeFlightOfferRepository, provider *fakeFlightProvider, got []dto.OfferResponse) {
				t.Helper()

				if len(provider.calls) != 1 {
					t.Fatalf("provider calls = %d, want 1", len(provider.calls))
				}
				if provider.calls[0].From != "DYU" || provider.calls[0].To != "IST" || provider.calls[0].Currency != "USD" {
					t.Fatalf("provider request = %+v, want normalized DYU/IST/USD", provider.calls[0])
				}
				if len(repo.created) != 1 {
					t.Fatalf("created offers = %d, want 1", len(repo.created))
				}
				created := repo.created[0]
				if created.OfferID != "OF-123456" {
					t.Fatalf("created OfferID = %q, want OF-123456", created.OfferID)
				}
				if created.Provider != "FakeAir" {
					t.Fatalf("created Provider = %q, want FakeAir", created.Provider)
				}
				if !created.ExpiresAt.Equal(fixedNow.Add(30 * time.Minute)) {
					t.Fatalf("created ExpiresAt = %s, want %s", created.ExpiresAt, fixedNow.Add(30*time.Minute))
				}
				if !created.CreatedAt.Equal(fixedNow) {
					t.Fatalf("created CreatedAt = %s, want %s", created.CreatedAt, fixedNow)
				}
				assertPriceCents(t, "created base", created.BasePrice, 50000)
				assertPriceCents(t, "created commission", created.Commission, 2500)
				assertPriceCents(t, "created service fee", created.ServiceFee, 2500)
				assertPriceCents(t, "created total", created.TotalPrice, 55000)
				assertPriceCents(t, "created profit", created.Profit, 5000)

				if len(got) != 1 {
					t.Fatalf("responses = %d, want 1", len(got))
				}
				response := got[0]
				if response.OfferID != "OF-123456" || response.Provider != "FakeAir" || response.FlightNumber != "TK255" {
					t.Fatalf("response = %+v, want persisted offer identity", response)
				}
				if response.DepartureDate != "2026-06-15" || response.ReturnDate != "2026-06-25" {
					t.Fatalf("response dates = %s/%s, want request dates", response.DepartureDate, response.ReturnDate)
				}
				assertMoneyCents(t, "response base", response.BasePrice, 50000)
				assertMoneyCents(t, "response commission", response.Commission, 2500)
				assertMoneyCents(t, "response service fee", response.ServiceFee, 2500)
				assertMoneyCents(t, "response total", response.TotalPrice, 55000)
				assertMoneyCents(t, "response profit", response.Profit, 5000)

				payload, err := json.Marshal(response)
				if err != nil {
					t.Fatalf("Marshal response: %v", err)
				}
				jsonText := string(payload)
				for _, want := range []string{
					`"base_price":500.00`,
					`"commission":25.00`,
					`"service_fee":25.00`,
					`"total_price":550.00`,
					`"profit":50.00`,
				} {
					if !strings.Contains(jsonText, want) {
						t.Fatalf("response JSON = %s, want to contain %s", jsonText, want)
					}
				}
			},
		},
		{
			name:        "provider error maps to provider unavailable",
			providerErr: errors.New("supplier unavailable"),
			wantCode:    apperror.CodeProviderUnavailable,
		},
		{
			name:           "provider empty maps to provider empty",
			providerOffers: []providers.ProviderOffer{},
			wantCode:       apperror.CodeProviderEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &fakeFlightProvider{
				name:   "FakeAir",
				offers: tt.providerOffers,
				err:    tt.providerErr,
			}
			repo := &fakeFlightOfferRepository{}
			calculator := pricing.NewCalculator(pricing.Config{
				CommissionPercent: 5,
				ServiceFees: pricing.ServiceFees{
					Adult:  money.FromCents(1500),
					Child:  money.FromCents(1000),
					Infant: money.FromCents(0),
				},
			})
			service := services.NewFlightService(
				provider,
				repo,
				&calculator,
				30*time.Minute,
				&sequenceIDGenerator{ids: []string{"OF-123456"}},
				fixedClock{now: fixedNow},
				discardLogger(),
			)

			got, err := service.Search(context.Background(), request)
			if tt.wantCode != "" {
				assertAppErrorCode(t, err, tt.wantCode)
				if len(repo.created) != 0 {
					t.Fatalf("created offers = %d, want 0", len(repo.created))
				}
				return
			}
			if err != nil {
				t.Fatalf("Search() error = %v, want nil", err)
			}
			tt.assert(t, repo, provider, got)
		})
	}
}

func TestBookingServiceCreate(t *testing.T) {
	request := dto.BookingRequest{
		OfferID: "OF-123456",
		Passengers: []dto.PassengerDTO{
			{
				Type:           "adult",
				FirstName:      "Alisher",
				LastName:       "Sabirov",
				DocumentNumber: "A1234567",
			},
		},
	}

	tests := []struct {
		name     string
		offer    *models.FlightOffer
		wantCode string
		assert   func(t *testing.T, repo *fakeBookingRepository, got dto.BookingResponse)
	}{
		{
			name: "happy path persists booking and passengers",
			offer: &models.FlightOffer{
				OfferID:   "OF-123456",
				ExpiresAt: fixedNow.Add(30 * time.Minute),
			},
			assert: func(t *testing.T, repo *fakeBookingRepository, got dto.BookingResponse) {
				t.Helper()

				if got.BookingID != "BK-987654" {
					t.Fatalf("BookingID = %q, want BK-987654", got.BookingID)
				}
				if got.Status != "created" {
					t.Fatalf("Status = %q, want created", got.Status)
				}
				if got.OfferID != "OF-123456" {
					t.Fatalf("OfferID = %q, want OF-123456", got.OfferID)
				}
				if !got.ExpiresAt.Equal(fixedNow.Add(30 * time.Minute)) {
					t.Fatalf("ExpiresAt = %s, want %s", got.ExpiresAt, fixedNow.Add(30*time.Minute))
				}

				if len(repo.created) != 1 {
					t.Fatalf("created bookings = %d, want 1", len(repo.created))
				}
				created := repo.created[0]
				if created.BookingID != "BK-987654" || created.Status != "created" || created.OfferID != "OF-123456" {
					t.Fatalf("created booking = %+v, want BK-987654 created OF-123456", created)
				}
				if !created.CreatedAt.Equal(fixedNow) {
					t.Fatalf("created CreatedAt = %s, want %s", created.CreatedAt, fixedNow)
				}

				if len(repo.passengers) != 1 || len(repo.passengers[0]) != 1 {
					t.Fatalf("created passengers = %+v, want one passenger", repo.passengers)
				}
				passenger := repo.passengers[0][0]
				if passenger.BookingID != created.ID {
					t.Fatalf("passenger BookingID = %d, want internal booking ID %d", passenger.BookingID, created.ID)
				}
				if passenger.Type != "adult" || passenger.FirstName != "Alisher" || passenger.LastName != "Sabirov" || passenger.DocumentNumber != "A1234567" {
					t.Fatalf("passenger = %+v, want request passenger", passenger)
				}
			},
		},
		{
			name:     "offer not found maps to offer not found",
			wantCode: apperror.CodeOfferNotFound,
		},
		{
			name: "expired offer maps to offer expired",
			offer: &models.FlightOffer{
				OfferID:   "OF-123456",
				ExpiresAt: fixedNow,
			},
			wantCode: apperror.CodeOfferExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offerRepo := &fakeFlightOfferRepository{}
			if tt.offer != nil {
				offerRepo.byID = map[string]*models.FlightOffer{
					tt.offer.OfferID: cloneFlightOffer(tt.offer),
				}
			}
			bookingRepo := &fakeBookingRepository{}
			service := services.NewBookingService(
				offerRepo,
				bookingRepo,
				&sequenceIDGenerator{ids: []string{"BK-987654"}},
				fixedClock{now: fixedNow},
				discardLogger(),
			)

			got, err := service.Create(context.Background(), request)
			if tt.wantCode != "" {
				assertAppErrorCode(t, err, tt.wantCode)
				if len(bookingRepo.created) != 0 {
					t.Fatalf("created bookings = %d, want 0", len(bookingRepo.created))
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() error = %v, want nil", err)
			}
			tt.assert(t, bookingRepo, got)
		})
	}
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type sequenceIDGenerator struct {
	ids  []string
	next int
}

func (g *sequenceIDGenerator) Generate(prefix string) string {
	if g.next >= len(g.ids) {
		return prefix + "-000000"
	}

	id := g.ids[g.next]
	g.next++
	return id
}

type fakeFlightProvider struct {
	name   string
	offers []providers.ProviderOffer
	err    error
	calls  []providers.ProviderSearch
}

func (p *fakeFlightProvider) Name() string {
	return p.name
}

func (p *fakeFlightProvider) Search(_ context.Context, req providers.ProviderSearch) ([]providers.ProviderOffer, error) {
	p.calls = append(p.calls, req)
	if p.err != nil {
		return nil, p.err
	}

	offers := make([]providers.ProviderOffer, len(p.offers))
	copy(offers, p.offers)
	return offers, nil
}

type fakeFlightOfferRepository struct {
	created   []*models.FlightOffer
	byID      map[string]*models.FlightOffer
	createErr error
	getErr    error
}

func (r *fakeFlightOfferRepository) Create(_ context.Context, offer *models.FlightOffer) error {
	if r.createErr != nil {
		return r.createErr
	}
	if r.byID == nil {
		r.byID = make(map[string]*models.FlightOffer)
	}

	stored := cloneFlightOffer(offer)
	r.created = append(r.created, stored)
	r.byID[stored.OfferID] = cloneFlightOffer(stored)
	return nil
}

func (r *fakeFlightOfferRepository) GetByOfferID(_ context.Context, offerID string) (*models.FlightOffer, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.byID == nil {
		return nil, repository.ErrNotFound
	}

	offer, ok := r.byID[offerID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cloneFlightOffer(offer), nil
}

type fakeBookingRepository struct {
	created    []*models.Booking
	passengers [][]*models.BookingPassenger
	err        error
	nextID     int64
}

func (r *fakeBookingRepository) CreateWithPassengers(_ context.Context, booking *models.Booking, passengers []*models.BookingPassenger) error {
	if r.err != nil {
		return r.err
	}
	if r.nextID == 0 {
		r.nextID = 1
	}

	booking.ID = r.nextID
	r.nextID++

	storedBooking := *booking
	storedPassengers := make([]*models.BookingPassenger, 0, len(passengers))
	for _, passenger := range passengers {
		passenger.BookingID = booking.ID
		storedPassenger := *passenger
		storedPassengers = append(storedPassengers, &storedPassenger)
	}

	r.created = append(r.created, &storedBooking)
	r.passengers = append(r.passengers, storedPassengers)
	return nil
}

func cloneFlightOffer(offer *models.FlightOffer) *models.FlightOffer {
	if offer == nil {
		return nil
	}

	clone := *offer
	if offer.ReturnDate != nil {
		returnDate := *offer.ReturnDate
		clone.ReturnDate = &returnDate
	}
	return &clone
}

func assertPriceCents(t *testing.T, name string, got int64, want int64) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %d, want %d", name, got, want)
	}
}

func assertMoneyCents(t *testing.T, name string, got money.Money, want int64) {
	t.Helper()

	if got.Cents() != want {
		t.Fatalf("%s = %d, want %d", name, got.Cents(), want)
	}
}

func assertAppErrorCode(t *testing.T, err error, code string) {
	t.Helper()

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %v, want *AppError code %s", err, code)
	}
	if appErr.Code != code {
		t.Fatalf("AppError code = %s, want %s", appErr.Code, code)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
