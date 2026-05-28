package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/database/models"
	"flysoft-flight-service/internal/dto"
	applogger "flysoft-flight-service/internal/logger"
	"flysoft-flight-service/internal/repository"
)

const bookingStatusCreated = "created"

type BookingService struct {
	offers   repository.FlightOfferRepository
	bookings repository.BookingRepository
	ids      IDGenerator
	clock    Clock
	logger   *slog.Logger
}

func NewBookingService(
	offers repository.FlightOfferRepository,
	bookings repository.BookingRepository,
	ids IDGenerator,
	clock Clock,
	logger *slog.Logger,
) *BookingService {
	if ids == nil {
		ids = RandomIDGenerator{}
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &BookingService{
		offers:   offers,
		bookings: bookings,
		ids:      ids,
		clock:    clock,
		logger:   logger,
	}
}

func (s *BookingService) Create(ctx context.Context, req dto.BookingRequest) (dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return dto.BookingResponse{}, err
	}

	offerID := strings.TrimSpace(req.OfferID)
	offer, err := s.offers.GetByOfferID(ctx, offerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.BookingResponse{}, apperror.OfferNotFound("offer not found")
		}

		s.log(ctx).Error("failed to load flight offer", slog.Any("error", err))
		return dto.BookingResponse{}, apperror.InternalError("failed to load offer")
	}

	now := s.clock.Now()
	if !offer.ExpiresAt.After(now) {
		return dto.BookingResponse{}, apperror.OfferExpired("offer expired")
	}

	booking := &models.Booking{
		BookingID: s.ids.Generate("BK"),
		OfferID:   offer.OfferID,
		Status:    bookingStatusCreated,
		ExpiresAt: offer.ExpiresAt,
		CreatedAt: now,
	}
	passengers := make([]*models.BookingPassenger, 0, len(req.Passengers))
	for _, passenger := range req.Passengers {
		passengers = append(passengers, &models.BookingPassenger{
			Type:           strings.TrimSpace(passenger.Type),
			FirstName:      strings.TrimSpace(passenger.FirstName),
			LastName:       strings.TrimSpace(passenger.LastName),
			DocumentNumber: strings.TrimSpace(passenger.DocumentNumber),
			CreatedAt:      now,
		})
	}

	if err := s.bookings.CreateWithPassengers(ctx, booking, passengers); err != nil {
		s.log(ctx).Error("failed to persist booking", slog.Any("error", err))
		return dto.BookingResponse{}, apperror.InternalError("failed to save booking")
	}

	return dto.BookingResponse{
		BookingID: booking.BookingID,
		Status:    booking.Status,
		OfferID:   booking.OfferID,
		ExpiresAt: booking.ExpiresAt,
	}, nil
}

func (s *BookingService) log(ctx context.Context) *slog.Logger {
	return applogger.WithRequestID(ctx, s.logger)
}
