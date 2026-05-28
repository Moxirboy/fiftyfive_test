package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/dto"
)

type BookingCreator interface {
	Create(ctx context.Context, req dto.BookingRequest) (dto.BookingResponse, error)
}

type BookingHandler struct {
	service BookingCreator
	logger  *slog.Logger
}

func NewBookingHandler(service BookingCreator, logger *slog.Logger) *BookingHandler {
	return &BookingHandler{
		service: service,
		logger:  logger,
	}
}

// Create godoc
//
// @Summary Create a preliminary booking
// @Description Creates a preliminary booking for an existing, unexpired offer.
// @Tags bookings
// @Accept json
// @Produce json
// @Param request body dto.BookingRequest true "Booking request"
// @Success 200 {object} dto.BookingEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 404 {object} dto.ErrorEnvelope
// @Failure 409 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /api/v1/bookings [post]
func (h *BookingHandler) Create(c *gin.Context) {
	var req dto.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, h.logger, apperror.BadRequest("invalid JSON body"))
		return
	}

	booking, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		writeError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewBookingEnvelope(booking))
}
