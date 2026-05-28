package dto

import (
	"strings"
	"time"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/money"
)

const dateLayout = "2006-01-02"

type SearchRequest struct {
	From          string `json:"from" example:"DYU"`
	To            string `json:"to" example:"IST"`
	DepartureDate string `json:"departure_date" example:"2026-06-15"`
	ReturnDate    string `json:"return_date,omitempty" example:"2026-06-25"`
	Adults        int    `json:"adults" example:"1"`
	Children      int    `json:"children" example:"1"`
	Infants       int    `json:"infants" example:"0"`
	Currency      string `json:"currency" example:"USD"`
}

type OfferResponse struct {
	OfferID       string      `json:"offer_id" example:"OF-123456"`
	Provider      string      `json:"provider" example:"MockAvia"`
	From          string      `json:"from" example:"DYU"`
	To            string      `json:"to" example:"IST"`
	DepartureDate string      `json:"departure_date" example:"2026-06-15"`
	ReturnDate    string      `json:"return_date,omitempty" example:"2026-06-25"`
	Airline       string      `json:"airline" example:"TK"`
	FlightNumber  string      `json:"flight_number" example:"TK255"`
	BasePrice     money.Money `json:"base_price" swaggertype:"number" example:"500.00"`
	ServiceFee    money.Money `json:"service_fee" swaggertype:"number" example:"25.00"`
	Commission    money.Money `json:"commission" swaggertype:"number" example:"25.00"`
	TotalPrice    money.Money `json:"total_price" swaggertype:"number" example:"550.00"`
	Profit        money.Money `json:"profit" swaggertype:"number" example:"50.00"`
	Currency      string      `json:"currency" example:"USD"`
}

func (r SearchRequest) Validate() *apperror.AppError {
	return r.ValidateAt(time.Now())
}

func (r SearchRequest) ValidateAt(now time.Time) *apperror.AppError {
	from := strings.TrimSpace(r.From)
	to := strings.TrimSpace(r.To)

	if from == "" {
		return apperror.ValidationError("from is required")
	}
	if to == "" {
		return apperror.ValidationError("to is required")
	}
	if strings.EqualFold(from, to) {
		return apperror.ValidationError("from and to must be different")
	}

	departureDate, err := time.Parse(dateLayout, r.DepartureDate)
	if err != nil {
		return apperror.ValidationError("departure_date must be YYYY-MM-DD")
	}
	if departureDate.Before(dateOnly(now)) {
		return apperror.ValidationError("departure_date must not be in the past")
	}
	if strings.TrimSpace(r.ReturnDate) != "" {
		if _, err := time.Parse(dateLayout, r.ReturnDate); err != nil {
			return apperror.ValidationError("return_date must be YYYY-MM-DD")
		}
	}

	if r.Adults < 0 || r.Children < 0 || r.Infants < 0 {
		return apperror.ValidationError("passenger counts must be non-negative")
	}
	if r.Adults+r.Children+r.Infants <= 0 {
		return apperror.ValidationError("total passengers must be greater than zero")
	}
	if r.Infants > r.Adults {
		return apperror.ValidationError("infants must not exceed adults")
	}
	if strings.TrimSpace(r.Currency) == "" {
		return apperror.ValidationError("currency is required")
	}

	return nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
