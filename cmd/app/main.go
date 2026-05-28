package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "flysoft-flight-service/docs"
	"flysoft-flight-service/internal/config"
	"flysoft-flight-service/internal/database"
	httpserver "flysoft-flight-service/internal/http"
	"flysoft-flight-service/internal/http/handlers"
	applogger "flysoft-flight-service/internal/logger"
	"flysoft-flight-service/internal/money"
	"flysoft-flight-service/internal/pricing"
	"flysoft-flight-service/internal/providers"
	"flysoft-flight-service/internal/providers/mockavia"
	"flysoft-flight-service/internal/repository"
	"flysoft-flight-service/internal/services"
)

const shutdownTimeout = 10 * time.Second

// @title FlySoft Flight Integration Service API
// @version 1.0
// @description Flight search and preliminary booking API with unified success and error envelopes.
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	log, err := applogger.New(cfg.LogLevel)
	if err != nil {
		slog.Error("failed to build logger", slog.Any("error", err))
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		log.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to get database handle", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Error("failed to close database", slog.Any("error", err))
		}
	}()

	flightRepo := repository.NewFlightOfferRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	provider := providers.NewRetry(mockavia.New(), providers.RetryConfig{
		Timeout:    cfg.Provider.Timeout,
		MaxRetries: cfg.Provider.MaxRetries,
		Backoff:    cfg.Provider.RetryBackoff,
	})
	calculator := pricing.NewCalculator(pricing.Config{
		CommissionPercent: cfg.CommissionPercent,
		ServiceFees: pricing.ServiceFees{
			Adult:  money.FromCents(cfg.ServiceFees.Adult),
			Child:  money.FromCents(cfg.ServiceFees.Child),
			Infant: money.FromCents(cfg.ServiceFees.Infant),
		},
	})

	flightService := services.NewFlightService(provider, flightRepo, &calculator, cfg.OfferTTL, nil, nil, log)
	bookingService := services.NewBookingService(flightRepo, bookingRepo, nil, nil, log)

	flightHandler := handlers.NewFlightHandler(flightService, log)
	bookingHandler := handlers.NewBookingHandler(bookingService, log)
	healthHandler := handlers.NewHealthHandler()

	gin.SetMode(gin.ReleaseMode)
	router := httpserver.NewRouter(log, flightHandler, bookingHandler, healthHandler)

	server := &nethttp.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting HTTP server", slog.Int("port", cfg.HTTPPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			log.Error("HTTP server stopped with error", slog.Any("error", err))
			os.Exit(1)
		}
		return
	case <-signalCtx.Done():
		stop()
	}

	log.Info("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shut down HTTP server", slog.Any("error", err))
		os.Exit(1)
	}

	if err := <-serverErr; err != nil {
		log.Error("HTTP server stopped with error", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("HTTP server stopped")
}
