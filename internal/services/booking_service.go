package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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

// Create creates a preliminary booking for an existing, unexpired offer.
//
// When idempotencyKey is non-empty, the operation is safe to retry: the first
// request stores the key alongside a hash of the request body, and any later
// request with the same key returns the original booking. Reusing a key with a
// different body is rejected as a conflict.
func (s *BookingService) Create(ctx context.Context, req dto.BookingRequest, idempotencyKey string) (dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return dto.BookingResponse{}, err
	}

	idempotencyKey = strings.TrimSpace(idempotencyKey)
	requestHash := hashBookingRequest(req)

	if idempotencyKey != "" {
		existing, err := s.bookings.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			return s.replay(existing, requestHash)
		}
		if !errors.Is(err, repository.ErrNotFound) {
			s.log(ctx).Error("failed to load booking by idempotency key", slog.Any("error", err))
			return dto.BookingResponse{}, apperror.InternalError("failed to load booking")
		}
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
	if idempotencyKey != "" {
		key := idempotencyKey
		booking.IdempotencyKey = &key
		booking.RequestHash = requestHash
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
		// A concurrent request with the same idempotency key won the race; replay it.
		if idempotencyKey != "" && errors.Is(err, repository.ErrDuplicate) {
			if existing, getErr := s.bookings.GetByIdempotencyKey(ctx, idempotencyKey); getErr == nil {
				return s.replay(existing, requestHash)
			}
		}

		s.log(ctx).Error("failed to persist booking", slog.Any("error", err))
		return dto.BookingResponse{}, apperror.InternalError("failed to save booking")
	}

	return bookingResponse(booking), nil
}

// replay returns a previously stored booking for an idempotency key, rejecting
// the request if the same key was first used for a different body.
func (s *BookingService) replay(existing *models.Booking, requestHash string) (dto.BookingResponse, error) {
	if existing.RequestHash != "" && existing.RequestHash != requestHash {
		return dto.BookingResponse{}, apperror.IdempotencyConflict("idempotency key already used for a different request")
	}

	return bookingResponse(existing), nil
}

func (s *BookingService) log(ctx context.Context) *slog.Logger {
	return applogger.WithRequestID(ctx, s.logger)
}

func bookingResponse(booking *models.Booking) dto.BookingResponse {
	return dto.BookingResponse{
		BookingID: booking.BookingID,
		Status:    booking.Status,
		OfferID:   booking.OfferID,
		ExpiresAt: booking.ExpiresAt,
	}
}

// hashBookingRequest produces a deterministic fingerprint of the booking body so
// that reusing an idempotency key with a different payload can be detected.
func hashBookingRequest(req dto.BookingRequest) string {
	h := sha256.New()
	fmt.Fprintf(h, "offer=%s", strings.TrimSpace(req.OfferID))
	for _, passenger := range req.Passengers {
		fmt.Fprintf(h, "|p=%s,%s,%s,%s",
			strings.TrimSpace(passenger.Type),
			strings.TrimSpace(passenger.FirstName),
			strings.TrimSpace(passenger.LastName),
			strings.TrimSpace(passenger.DocumentNumber),
		)
	}

	return hex.EncodeToString(h.Sum(nil))
}
