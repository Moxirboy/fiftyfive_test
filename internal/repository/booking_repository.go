package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"flysoft-flight-service/internal/database/models"
)

const uniqueViolationCode = "23505"

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) CreateWithPassengers(ctx context.Context, booking *models.Booking, passengers []*models.BookingPassenger) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(booking).Error; err != nil {
			return err
		}

		if len(passengers) == 0 {
			return nil
		}

		for _, passenger := range passengers {
			passenger.BookingID = booking.ID
		}

		return tx.Create(&passengers).Error
	})

	if isUniqueViolation(err) {
		return ErrDuplicate
	}
	return err
}

func (r *bookingRepository) GetByIdempotencyKey(ctx context.Context, key string) (*models.Booking, error) {
	var booking models.Booking
	if err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &booking, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == uniqueViolationCode
	}
	return false
}
