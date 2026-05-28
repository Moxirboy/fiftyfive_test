package dto

import "flysoft-flight-service/internal/apperror"

type SearchEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Data    []OfferResponse `json:"data"`
}

type BookingEnvelope struct {
	Success bool            `json:"success" example:"true"`
	Booking BookingResponse `json:"booking"`
}

type ErrorEnvelope struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"from is required"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

func NewSearchEnvelope(data []OfferResponse) SearchEnvelope {
	return SearchEnvelope{Success: true, Data: data}
}

func NewBookingEnvelope(booking BookingResponse) BookingEnvelope {
	return BookingEnvelope{Success: true, Booking: booking}
}

func NewErrorEnvelope(err *apperror.AppError) ErrorEnvelope {
	if err == nil {
		err = apperror.InternalError("")
	}

	return ErrorEnvelope{
		Success: false,
		Error: ErrorBody{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}
