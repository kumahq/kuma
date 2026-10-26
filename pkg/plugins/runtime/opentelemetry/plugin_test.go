package opentelemetry

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/kumahq/kuma/v2/pkg/config/tracing"
	otel_extensions "github.com/kumahq/kuma/v2/pkg/plugins/extensions/otel"
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
		DeferCleanup(shutdownAll, processors)
		Expect(processors).To(HaveLen(3))
	})

	It("should skip a failing factory and keep the other processors", func() {
		// given
		failing := func(context.Context) (trace.SpanProcessor, error) {
			return nil, errors.New("boom")
		}

		// when
		processors, err := newSpanProcessors(context.Background(), logr.Discard(), tracing.OpenTelemetry{Enabled: true},
			[]otel_extensions.SpanProcessorFactory{failing, inMemory})

		// then
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(shutdownAll, processors)
		Expect(processors).To(HaveLen(2))
	})
})

func shutdownAll(processors []trace.SpanProcessor) {
	for _, processor := range processors {
		Expect(processor.Shutdown(context.Background())).To(Succeed())
	}
}
