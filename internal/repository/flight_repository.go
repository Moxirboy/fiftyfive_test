package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"flysoft-flight-service/internal/database/models"
)

type flightOfferRepository struct {
	db *gorm.DB
}

func NewFlightOfferRepository(db *gorm.DB) FlightOfferRepository {
	return &flightOfferRepository{db: db}
}

func (r *flightOfferRepository) Create(ctx context.Context, offer *models.FlightOffer) error {
	return r.db.WithContext(ctx).Create(offer).Error
}

func (r *flightOfferRepository) GetByOfferID(ctx context.Context, offerID string) (*models.FlightOffer, error) {
	var offer models.FlightOffer
	if err := r.db.WithContext(ctx).Where("offer_id = ?", offerID).First(&offer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &offer, nil
}
