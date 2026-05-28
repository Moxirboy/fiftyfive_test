package apperror

import (
	"errors"
	"net/http"
)

const (
	CodeValidationError     = "VALIDATION_ERROR"
	CodeBadRequest          = "BAD_REQUEST"
	CodeOfferNotFound       = "OFFER_NOT_FOUND"
	CodeOfferExpired        = "OFFER_EXPIRED"
	CodeProviderUnavailable = "PROVIDER_UNAVAILABLE"
	CodeProviderEmpty       = "PROVIDER_EMPTY"
	CodeInternalError       = "INTERNAL_ERROR"
)

type AppError struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"-"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func ValidationError(message string) *AppError {
	return newAppError(CodeValidationError, http.StatusBadRequest, message, "validation error")
}

func BadRequest(message string) *AppError {
	return newAppError(CodeBadRequest, http.StatusBadRequest, message, "bad request")
}

func OfferNotFound(message string) *AppError {
	return newAppError(CodeOfferNotFound, http.StatusNotFound, message, "offer not found")
}

func OfferExpired(message string) *AppError {
	return newAppError(CodeOfferExpired, http.StatusConflict, message, "offer expired")
}

func ProviderUnavailable(message string) *AppError {
	return newAppError(CodeProviderUnavailable, http.StatusBadGateway, message, "provider unavailable")
}

func ProviderEmpty(message string) *AppError {
	return newAppError(CodeProviderEmpty, http.StatusNotFound, message, "provider returned no offers")
}

func InternalError(message string) *AppError {
	return newAppError(CodeInternalError, http.StatusInternalServerError, message, "internal error")
}

func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return InternalError("")
}

func newAppError(code string, status int, message string, fallback string) *AppError {
	if message == "" {
		message = fallback
	}
	return &AppError{
		Code:       code,
		HTTPStatus: status,
		Message:    message,
	}
}
