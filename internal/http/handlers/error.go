package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/dto"
	applogger "flysoft-flight-service/internal/logger"
)

func writeError(c *gin.Context, log *slog.Logger, err error) {
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		applogger.WithRequestID(c.Request.Context(), log).Error(
			"unexpected error",
			slog.Any("error", err),
		)
		appErr = apperror.InternalError("")
	}

	status := appErr.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}

	c.AbortWithStatusJSON(status, dto.NewErrorEnvelope(appErr))
}
