package models

import "time"

type BookingPassenger struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	BookingID      int64     `gorm:"column:booking_id"`
	Type           string    `gorm:"column:type"`
	FirstName      string    `gorm:"column:first_name"`
	LastName       string    `gorm:"column:last_name"`
	DocumentNumber string    `gorm:"column:document_number"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (BookingPassenger) TableName() string {
	return "booking_passengers"
}
