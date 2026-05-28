package repository

import (
	"context"

	"gorm.io/gorm"

	"flysoft-flight-service/internal/database/models"
)

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) CreateWithPassengers(ctx context.Context, booking *models.Booking, passengers []*models.BookingPassenger) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
}
