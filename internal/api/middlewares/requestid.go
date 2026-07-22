package middlewares

import (
	"log/slog"
	"time"

	"gitxyz/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceHeader is the response/request header carrying the trace ID.
const TraceHeader = "X-Trace-Id"

// RequestID generates a trace ID per request, stores it in the context and
// response header, and emits a structured JSON access log on completion.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Reuse an incoming trace id when provided (e.g. via gateway/proxy).
		traceID := ctx.GetHeader(TraceHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		ctx.Set("trace_id", traceID)
		ctx.Header(TraceHeader, traceID)
		ctx.Request = ctx.Request.WithContext(logger.WithTrace(ctx.Request.Context(), traceID))

		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		latency := time.Since(start)
		status := ctx.Writer.Status()
		clientIP := ctx.ClientIP()
		method := ctx.Request.Method
		if raw != "" {
			path = path + "?" + raw
		}

		log := logger.FromContext(ctx.Request.Context())
		attrs := []any{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", ctx.Request.UserAgent()),
		}
		if len(ctx.Errors) > 0 {
			attrs = append(attrs, slog.String("error", ctx.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		switch {
		case status >= 500:
			log.Error("request completed", attrs...)
		case status >= 400:
			log.Warn("request completed", attrs...)
		default:
			log.Info("request completed", attrs...)
		}
	}
}
