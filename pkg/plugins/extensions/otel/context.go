package otel

import (
	"context"

	"go.opentelemetry.io/otel/sdk/trace"
)

type spanProcessorFactoriesKey struct{}

// A factory, not a processor: the provider is rebuilt on component restart
// and its shutdown takes the processors with it.
type SpanProcessorFactory func(context.Context) (trace.SpanProcessor, error)

func WithSpanProcessorFactories(ctx context.Context, factories ...SpanProcessorFactory) context.Context {
	existing := SpanProcessorFactoriesFromContext(ctx)
	merged := make([]SpanProcessorFactory, 0, len(existing)+len(factories))
	merged = append(merged, existing...)
	merged = append(merged, factories...)
	return context.WithValue(ctx, spanProcessorFactoriesKey{}, merged)
}

func SpanProcessorFactoriesFromContext(ctx context.Context) []SpanProcessorFactory {
	factories, _ := ctx.Value(spanProcessorFactoriesKey{}).([]SpanProcessorFactory)
	return factories
}
