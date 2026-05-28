package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"flysoft-flight-service/internal/http/handlers"
	"flysoft-flight-service/internal/http/middleware"
)

func NewRouter(
	log *slog.Logger,
	flightHandler *handlers.FlightHandler,
	bookingHandler *handlers.BookingHandler,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(
		middleware.Recovery(log),
		middleware.RequestID(),
		middleware.Logger(log),
	)

	router.GET("/health", healthHandler.Get)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	v1.POST("/flights/search", flightHandler.Search)
	v1.POST("/bookings", bookingHandler.Create)

	return router
}
