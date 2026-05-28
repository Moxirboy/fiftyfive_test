package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	applogger "flysoft-flight-service/internal/logger"
)

func Logger(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}

	return func(c *gin.Context) {
		ctx := applogger.ContextWithLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		latencyMS := float64(time.Since(start).Microseconds()) / 1000
		applogger.WithRequestID(c.Request.Context(), log).Info(
			"http request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Float64("latency_ms", latencyMS),
		)
	}
}
