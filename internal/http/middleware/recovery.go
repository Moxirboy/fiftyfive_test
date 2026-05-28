package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/dto"
	applogger "flysoft-flight-service/internal/logger"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}

	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			applogger.WithRequestID(c.Request.Context(), log).Error(
				"panic recovered",
				slog.Any("panic", recovered),
			)
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				dto.NewErrorEnvelope(apperror.InternalError("")),
			)
		}()

		c.Next()
	}
}
