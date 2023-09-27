package trace

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const traceIDKey = "traceID"

type slogHandler struct {
	slog.Handler
}

func SlogHandler(h slog.Handler) slog.Handler {
	return &slogHandler{Handler: h}
}

// Handle implements [slog.Handler].
func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		record.AddAttrs(slog.String(traceIDKey, spanCtx.TraceID().String()))
	}
	return h.Handler.Handle(ctx, record)
}
