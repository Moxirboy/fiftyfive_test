package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const requestIDKey = "request_id"

type (
	loggerContextKey    struct{}
	requestIDContextKey struct{}
)

func New(level string) (*slog.Logger, error) {
	slogLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler), nil
}

func ContextWithLogger(ctx context.Context, log *slog.Logger) context.Context {
	if log == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerContextKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	log, ok := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if !ok || log == nil {
		return slog.Default()
	}

	return log
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func WithRequestID(ctx context.Context, log *slog.Logger) *slog.Logger {
	if log == nil {
		log = slog.Default()
	}

	requestID := RequestIDFromContext(ctx)
	if requestID == "" {
		return log
	}

	return log.With(slog.String(requestIDKey, requestID))
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", level)
	}
}
