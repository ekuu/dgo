package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const traceName = "dgo"

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	opts = append(
		opts,
		trace.WithSpanKind(trace.SpanKindServer),
	)
	return otel.Tracer(traceName).Start(ctx, spanName, opts...)
}

func RecordError(span trace.Span, err error, recordErrorOpts ...trace.EventOption) error {
	if err != nil {
		span.RecordError(err, append(recordErrorOpts, trace.WithStackTrace(true))...)
	}
	return err
}

func Template(ctx context.Context, spanName string, fn func(ctx context.Context) error, recordErrorOpts ...trace.EventOption) error {
	ctx, span := Start(ctx, spanName)
	defer span.End()
	return RecordError(span, fn(ctx), recordErrorOpts...)
}

func Template2[T any](ctx context.Context, spanName string, fn func(ctx context.Context) (T, error), recordErrorOpts ...trace.EventOption) (T, error) {
	ctx, span := Start(ctx, spanName)
	defer span.End()
	t, err := fn(ctx)
	if err != nil {
		return t, RecordError(span, err, recordErrorOpts...)
	}
	return t, nil
}
