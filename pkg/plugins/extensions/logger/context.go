package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type spanLogValuesProcessorKey struct{}

// SpanLogValuesProcessor should be a function which process received
// trace.Span. Returned []]interface{} will be later added as logger values.
type SpanLogValuesProcessor func(trace.Span) []interface{}

// NewSpanLogValuesProcessorContext will enrich the provided context with
// the provided spanLogValuesProcessor. It may be useful for any application
// which depends on Kuma, but wants to for example transform trace/span ids
// from otel to datadog format.
func NewSpanLogValuesProcessorContext(
	ctx context.Context,
	fn SpanLogValuesProcessor,
) context.Context {
	return context.WithValue(ctx, spanLogValuesProcessorKey{}, fn)
}

func FromSpanLogValuesProcessorContext(ctx context.Context) (SpanLogValuesProcessor, bool) {
	fn, ok := ctx.Value(spanLogValuesProcessorKey{}).(SpanLogValuesProcessor)
	return fn, ok
}
