package models

import "time"

type FlightOffer struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	OfferID       string     `gorm:"column:offer_id"`
	Provider      string     `gorm:"column:provider"`
	Origin        string     `gorm:"column:origin"`
	Destination   string     `gorm:"column:destination"`
	DepartureDate time.Time  `gorm:"column:departure_date;type:date"`
	ReturnDate    *time.Time `gorm:"column:return_date;type:date"`
	Airline       string     `gorm:"column:airline"`
	FlightNumber  string     `gorm:"column:flight_number"`
	BasePrice     int64      `gorm:"column:base_price"`
	Commission    int64      `gorm:"column:commission"`
	ServiceFee    int64      `gorm:"column:service_fee"`
	TotalPrice    int64      `gorm:"column:total_price"`
	Profit        int64      `gorm:"column:profit"`
	Currency      string     `gorm:"column:currency"`
	ExpiresAt     time.Time  `gorm:"column:expires_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (FlightOffer) TableName() string {
	return "flight_offers"
}
