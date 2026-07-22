package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// contextKey is a private type to avoid context key collisions.
type contextKey string

// TraceIDKey is the context key under which the request/trace ID is stored.
const TraceIDKey contextKey = "trace_id"

// Logger is the application-wide structured logger.
var Logger *slog.Logger

func init() {
	// Default to info level; overridden by Configure.
	Logger = newJSONLogger(slog.LevelInfo)
}

// Configure (re)initializes the global logger with the given level string
// (e.g. "debug", "info", "warn", "error"). Invalid values default to info.
func Configure(level string) {
	Logger = newJSONLogger(parseLevel(level))
}

func newJSONLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithTrace returns a context carrying the given trace ID.
func WithTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// TraceFrom extracts the trace ID from the context, or "" if absent.
func TraceFrom(ctx context.Context) string {
	if v, ok := ctx.Value(TraceIDKey).(string); ok {
		return v
	}
	return ""
}

// FromContext returns a logger that includes the trace ID from ctx as a field.
// Falls back to the global logger when no trace ID is present.
func FromContext(ctx context.Context) *slog.Logger {
	traceID := TraceFrom(ctx)
	if traceID == "" {
		return Logger
	}
	return Logger.With(slog.String("trace_id", traceID))
}

// FromGin returns a logger bound to a Gin request context. It safely handles
// test contexts where Request may be nil by falling back to Background.
func FromGin(ctx *gin.Context) *slog.Logger {
	if ctx == nil {
		return Logger
	}
	if ctx.Request == nil {
		return FromContext(context.Background())
	}
	return FromContext(ctx.Request.Context())
}
