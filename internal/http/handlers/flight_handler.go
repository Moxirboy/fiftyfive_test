package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/dto"
)

type FlightSearcher interface {
	Search(ctx context.Context, req dto.SearchRequest) ([]dto.OfferResponse, error)
}

type FlightHandler struct {
	service FlightSearcher
	logger  *slog.Logger
}

func NewFlightHandler(service FlightSearcher, logger *slog.Logger) *FlightHandler {
	return &FlightHandler{
		service: service,
		logger:  logger,
	}
}

// Search godoc
//
// @Summary Search flight offers
// @Description Searches the mock flight provider, prices matching offers, persists them, and returns normalized offers.
// @Tags flights
// @Accept json
// @Produce json
// @Param request body dto.SearchRequest true "Flight search request"
// @Success 200 {object} dto.SearchEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 404 {object} dto.ErrorEnvelope
// @Failure 502 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /api/v1/flights/search [post]
func (h *FlightHandler) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, h.logger, apperror.BadRequest("invalid JSON body"))
		return
	}

	offers, err := h.service.Search(c.Request.Context(), req)
	if err != nil {
		writeError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewSearchEnvelope(offers))
}
