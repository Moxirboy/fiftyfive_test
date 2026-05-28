package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flysoft-flight-service/internal/dto"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Get godoc
//
// @Summary Health check
// @Description Returns service liveness status.
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /health [get]
func (h *HealthHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}
