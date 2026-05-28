package repository

import (
	"context"
	"errors"

	"flysoft-flight-service/internal/database/models"
)

var (
	ErrNotFound  = errors.New("repository: not found")
	ErrDuplicate = errors.New("repository: duplicate")
)

type FlightOfferRepository interface {
	Create(ctx context.Context, offer *models.FlightOffer) error
	GetByOfferID(ctx context.Context, offerID string) (*models.FlightOffer, error)
}

type BookingRepository interface {
	CreateWithPassengers(ctx context.Context, booking *models.Booking, passengers []*models.BookingPassenger) error
	GetByIdempotencyKey(ctx context.Context, key string) (*models.Booking, error)
}
