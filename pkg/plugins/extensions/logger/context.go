package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type spanLogValuesProcessorKey struct{}

// SpanLogValuesProcessor should be a function which process received
// trace.Span. Returned []]interface{} will be later added as logger values.
type SpanLogValuesProcessor func(trace.Span) []any

func FromSpanLogValuesProcessorContext(ctx context.Context) (SpanLogValuesProcessor, bool) {
	fn, ok := ctx.Value(spanLogValuesProcessorKey{}).(SpanLogValuesProcessor)
	return fn, ok
}
