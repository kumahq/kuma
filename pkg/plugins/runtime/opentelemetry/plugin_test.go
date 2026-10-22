package opentelemetry

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/kumahq/kuma/v3/pkg/config/tracing"
	otel_extensions "github.com/kumahq/kuma/v3/pkg/plugins/extensions/otel"
)

var _ = Describe("OpenTelemetry tracing", func() {
	inMemory := func(context.Context) (trace.SpanProcessor, error) {
		return trace.NewSimpleSpanProcessor(tracetest.NewInMemoryExporter()), nil
	}

	It("should create the primary processor plus one per factory", func() {
		// when
		processors, err := newSpanProcessors(context.Background(), logr.Discard(), tracing.OpenTelemetry{Enabled: true},
			[]otel_extensions.SpanProcessorFactory{inMemory, inMemory})

		// then
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() {
			for _, processor := range processors {
				Expect(processor.Shutdown(context.Background())).To(Succeed())
			}
		})
		Expect(processors).To(HaveLen(3))
	})

	It("should fail and shut down created processors when a factory fails", func() {
		// given
		created := &shutdownRecorder{SpanProcessor: trace.NewSimpleSpanProcessor(tracetest.NewInMemoryExporter())}
		recorded := func(context.Context) (trace.SpanProcessor, error) {
			return created, nil
		}
		failing := func(context.Context) (trace.SpanProcessor, error) {
			return nil, errors.New("boom")
		}

		// when
		_, err := newSpanProcessors(context.Background(), logr.Discard(), tracing.OpenTelemetry{Enabled: true},
			[]otel_extensions.SpanProcessorFactory{recorded, failing})

		// then
		Expect(err).To(MatchError(ContainSubstring("additional span processor 1: boom")))
		Expect(created.shutdown).To(BeTrue())
	})

	It("should deliver every span to every processor", func() {
		// given
		first := tracetest.NewInMemoryExporter()
		second := tracetest.NewInMemoryExporter()
		provider := newTracerProvider(resource.Empty(), []trace.SpanProcessor{
			trace.NewSimpleSpanProcessor(first),
			trace.NewBatchSpanProcessor(second),
		})
		defer func() {
			Expect(provider.Shutdown(context.Background())).To(Succeed())
		}()

		// when
		_, span := provider.Tracer("test").Start(context.Background(), "op")
		span.End()
		Expect(provider.ForceFlush(context.Background())).To(Succeed())

		// then
		Expect(first.GetSpans()).To(HaveLen(1))
		Expect(second.GetSpans()).To(HaveLen(1))
		Expect(first.GetSpans()[0].SpanContext.TraceID()).To(Equal(second.GetSpans()[0].SpanContext.TraceID()))
	})
})

type shutdownRecorder struct {
	trace.SpanProcessor
	shutdown bool
}

func (r *shutdownRecorder) Shutdown(ctx context.Context) error {
	r.shutdown = true
	return r.SpanProcessor.Shutdown(ctx)
}
