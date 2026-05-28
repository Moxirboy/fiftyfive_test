package services

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/database/models"
	"flysoft-flight-service/internal/dto"
	applogger "flysoft-flight-service/internal/logger"
	"flysoft-flight-service/internal/money"
	"flysoft-flight-service/internal/pricing"
	"flysoft-flight-service/internal/providers"
	"flysoft-flight-service/internal/repository"
)

const searchDateLayout = "2006-01-02"

type FlightService struct {
	provider   providers.FlightProvider
	offers     repository.FlightOfferRepository
	calculator *pricing.Calculator
	offerTTL   time.Duration
	ids        IDGenerator
	clock      Clock
	logger     *slog.Logger
}

func NewFlightService(
	provider providers.FlightProvider,
	offers repository.FlightOfferRepository,
	calculator *pricing.Calculator,
	offerTTL time.Duration,
	ids IDGenerator,
	clock Clock,
	logger *slog.Logger,
) *FlightService {
	if ids == nil {
		ids = RandomIDGenerator{}
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &FlightService{
		provider:   provider,
		offers:     offers,
		calculator: calculator,
		offerTTL:   offerTTL,
		ids:        ids,
		clock:      clock,
		logger:     logger,
	}
}

func (s *FlightService) Search(ctx context.Context, req dto.SearchRequest) ([]dto.OfferResponse, error) {
	now := s.clock.Now()
	if err := req.ValidateAt(now); err != nil {
		return nil, err
	}

	departureDate, returnDate, err := parseSearchDates(req)
	if err != nil {
		return nil, err
	}

	providerOffers, err := s.provider.Search(ctx, providers.ProviderSearch{
		From:          strings.ToUpper(strings.TrimSpace(req.From)),
		To:            strings.ToUpper(strings.TrimSpace(req.To)),
		DepartureDate: departureDate.Format(searchDateLayout),
		ReturnDate:    strings.TrimSpace(req.ReturnDate),
		Currency:      strings.ToUpper(strings.TrimSpace(req.Currency)),
		Adults:        req.Adults,
		Children:      req.Children,
		Infants:       req.Infants,
	})
	if err != nil {
		s.log(ctx).Error("flight provider search failed", slog.Any("error", err))
		return nil, apperror.ProviderUnavailable("provider unavailable")
	}
	if len(providerOffers) == 0 {
		return nil, apperror.ProviderEmpty("provider returned no offers")
	}

	responses := make([]dto.OfferResponse, 0, len(providerOffers))
	for _, providerOffer := range providerOffers {
		breakdown := s.pricingCalculator().Calculate(
			pricing.PassengerCounts{
				Adults:   int64(req.Adults),
				Children: int64(req.Children),
				Infants:  int64(req.Infants),
			},
			pricing.PassengerPrices{
				Adult:  money.FromCents(providerOffer.PriceAdult),
				Child:  money.FromCents(providerOffer.PriceChild),
				Infant: money.FromCents(providerOffer.PriceInfant),
			},
		)

		offer := &models.FlightOffer{
			OfferID:       s.ids.Generate("OF"),
			Provider:      s.provider.Name(),
			Origin:        providerOffer.From,
			Destination:   providerOffer.To,
			DepartureDate: departureDate,
			ReturnDate:    returnDate,
			Airline:       providerOffer.Airline,
			FlightNumber:  providerOffer.FlightNumber,
			BasePrice:     breakdown.Base.Cents(),
			Commission:    breakdown.Commission.Cents(),
			ServiceFee:    breakdown.ServiceFee.Cents(),
			TotalPrice:    breakdown.Total.Cents(),
			Profit:        breakdown.Profit.Cents(),
			Currency:      providerOffer.Currency,
			ExpiresAt:     now.Add(s.offerTTL),
			CreatedAt:     now,
		}

		if err := s.offers.Create(ctx, offer); err != nil {
			s.log(ctx).Error("failed to persist flight offer", slog.Any("error", err))
			return nil, apperror.InternalError("failed to save offer")
		}

		responses = append(responses, flightOfferResponse(offer))
	}

	return responses, nil
}

func (s *FlightService) pricingCalculator() *pricing.Calculator {
	if s.calculator == nil {
		return &pricing.Calculator{}
	}
	return s.calculator
}

func (s *FlightService) log(ctx context.Context) *slog.Logger {
	return applogger.WithRequestID(ctx, s.logger)
}

func parseSearchDates(req dto.SearchRequest) (time.Time, *time.Time, error) {
	departureDate, err := time.Parse(searchDateLayout, req.DepartureDate)
	if err != nil {
		return time.Time{}, nil, apperror.ValidationError("departure_date must be YYYY-MM-DD")
	}

	var returnDate *time.Time
	if strings.TrimSpace(req.ReturnDate) != "" {
		parsedReturnDate, err := time.Parse(searchDateLayout, req.ReturnDate)
		if err != nil {
			return time.Time{}, nil, apperror.ValidationError("return_date must be YYYY-MM-DD")
		}
		returnDate = &parsedReturnDate
	}

	return departureDate, returnDate, nil
}

func flightOfferResponse(offer *models.FlightOffer) dto.OfferResponse {
	response := dto.OfferResponse{
		OfferID:       offer.OfferID,
		Provider:      offer.Provider,
		From:          offer.Origin,
		To:            offer.Destination,
		DepartureDate: offer.DepartureDate.Format(searchDateLayout),
		Airline:       offer.Airline,
		FlightNumber:  offer.FlightNumber,
		BasePrice:     money.FromCents(offer.BasePrice),
		ServiceFee:    money.FromCents(offer.ServiceFee),
		Commission:    money.FromCents(offer.Commission),
		TotalPrice:    money.FromCents(offer.TotalPrice),
		Profit:        money.FromCents(offer.Profit),
		Currency:      offer.Currency,
	}
	if offer.ReturnDate != nil {
		response.ReturnDate = offer.ReturnDate.Format(searchDateLayout)
	}

	return response
}
