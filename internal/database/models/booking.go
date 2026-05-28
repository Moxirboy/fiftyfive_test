package models

import "time"

type Booking struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	BookingID      string    `gorm:"column:booking_id"`
	OfferID        string    `gorm:"column:offer_id"`
	Status         string    `gorm:"column:status"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	IdempotencyKey *string   `gorm:"column:idempotency_key"`
	RequestHash    string    `gorm:"column:request_hash"`
}

func (Booking) TableName() string {
	return "bookings"
}
