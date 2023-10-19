package telemetry

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"go.opentelemetry.io/otel/trace"

	"github.com/chatbot-go/app/library/ctxkey"
)

func SetLogger(development bool, attrs ...slog.Attr) {
	var handler *slogHandler

	if development {
		handler = &slogHandler{
			Handler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			}),
		}
	} else {
		handler = &slogHandler{
			Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelInfo,
			}),
		}
	}

	handlerWithAttrs := handler.WithAttrs(attrs)

	logger := slog.New(handlerWithAttrs)

	slog.SetDefault(logger)
}

type slogHandler struct {
	slog.Handler
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0)

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		traceID := sc.TraceID().String()
		spanID := sc.SpanID().String()

		attrs = append(attrs,
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),

			// https://docs.datadoghq.com/tracing/other_telemetry/connect_logs_and_traces/opentelemetry/?tab=go
			slog.String("dd.trace_id", convertToDatadog(traceID)),
			slog.String("dd.span_id", convertToDatadog(spanID)),
		)
	}

	if requestID, ok := ctxkey.GetRequestID(ctx); ok {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	if idempotencyKey, ok := ctxkey.GetIdempotencyKey(ctx); ok {
		attrs = append(attrs, slog.String("idempotency_key", idempotencyKey))
	}

	return h.Handler.WithAttrs(attrs).Handle(ctx, record) //nolint:wrapcheck
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{h.Handler.WithAttrs(attrs)}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{h.Handler.WithGroup(name)}
}

func convertToDatadog(id string) string {
	const maxLen = 16

	if len(id) < maxLen {
		return ""
	}

	if len(id) > maxLen {
		id = id[maxLen:]
	}

	intValue, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return ""
	}

	return strconv.FormatUint(intValue, 10)
}
