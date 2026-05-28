package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/dto"
)

const idempotencyKeyHeader = "Idempotency-Key"

type BookingCreator interface {
	Create(ctx context.Context, req dto.BookingRequest, idempotencyKey string) (dto.BookingResponse, error)
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
// @Description Creates a preliminary booking for an existing, unexpired offer. Pass an Idempotency-Key header to make retries safe.
// @Tags bookings
// @Accept json
// @Produce json
// @Param Idempotency-Key header string false "Idempotency key; repeating it returns the original booking"
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

	idempotencyKey := strings.TrimSpace(c.GetHeader(idempotencyKeyHeader))
	booking, err := h.service.Create(c.Request.Context(), req, idempotencyKey)
	if err != nil {
		writeError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewBookingEnvelope(booking))
}
