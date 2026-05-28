package dto

import (
	"strings"
	"time"

	"flysoft-flight-service/internal/apperror"
)

type BookingRequest struct {
	OfferID    string         `json:"offer_id" example:"OF-123456"`
	Passengers []PassengerDTO `json:"passengers"`
}

type PassengerDTO struct {
	Type           string `json:"type" example:"adult"`
	FirstName      string `json:"first_name" example:"Alisher"`
	LastName       string `json:"last_name" example:"Sabirov"`
	DocumentNumber string `json:"document_number" example:"A1234567"`
}

type BookingResponse struct {
	BookingID string    `json:"booking_id" example:"BK-987654"`
	Status    string    `json:"status" example:"created"`
	OfferID   string    `json:"offer_id" example:"OF-123456"`
	ExpiresAt time.Time `json:"expires_at" example:"2026-05-28T15:30:00Z"`
}

func (r BookingRequest) Validate() *apperror.AppError {
	if strings.TrimSpace(r.OfferID) == "" {
		return apperror.ValidationError("offer_id is required")
	}
	if len(r.Passengers) == 0 {
		return apperror.ValidationError("passengers must not be empty")
	}

	for _, passenger := range r.Passengers {
		if !validPassengerType(passenger.Type) {
			return apperror.ValidationError("passenger type must be adult, child, or infant")
		}
		if strings.TrimSpace(passenger.FirstName) == "" {
			return apperror.ValidationError("passenger first_name is required")
		}
		if strings.TrimSpace(passenger.LastName) == "" {
			return apperror.ValidationError("passenger last_name is required")
		}
		if strings.TrimSpace(passenger.DocumentNumber) == "" {
			return apperror.ValidationError("passenger document_number is required")
		}
	}

	return nil
}

func validPassengerType(passengerType string) bool {
	switch strings.TrimSpace(passengerType) {
	case "adult", "child", "infant":
		return true
	default:
		return false
	}
}
