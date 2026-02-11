package tracing_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/trace"

	"github.com/kumahq/kuma/v2/pkg/core/tracing"
)

// panicSpan is a fake span that panics when End() is called
type panicSpan struct {
	trace.Span
	endCalled    bool
	endPanics    bool
	receivedOpts []trace.SpanEndOption
}

func (s *panicSpan) End(opts ...trace.SpanEndOption) {
	s.endCalled = true
	s.receivedOpts = opts
	if s.endPanics {
		panic("simulated span.End() panic during OTel shutdown")
	}
}

var _ = Describe("SafeSpanEnd", func() {
	It("should handle nil span gracefully", func() {
		// when/then - should not panic
		tracing.SafeSpanEnd(nil)
	})

	It("should call End() on valid span", func() {
		// given
		span := &panicSpan{}

		// when
		tracing.SafeSpanEnd(span)

		// then
		Expect(span.endCalled).To(BeTrue())
	})

	It("should recover from panic during End()", func() {
		// given
		span := &panicSpan{endPanics: true}

		// when/then - should not propagate panic
		tracing.SafeSpanEnd(span)

		// and span.End() was attempted
		Expect(span.endCalled).To(BeTrue())
	})

	It("should forward SpanEndOptions to End()", func() {
		// given
		span := &panicSpan{}
		opt1 := trace.WithTimestamp(time.Now())
		opt2 := trace.WithStackTrace(true)

		// when
		tracing.SafeSpanEnd(span, opt1, opt2)

		// then
		Expect(span.endCalled).To(BeTrue())
		Expect(span.receivedOpts).To(HaveLen(2))
	})

	It("should forward options even if End() panics", func() {
		// given
		span := &panicSpan{endPanics: true}
		opt := trace.WithStackTrace(true)

		// when
		tracing.SafeSpanEnd(span, opt)

		// then
		Expect(span.endCalled).To(BeTrue())
		Expect(span.receivedOpts).To(HaveLen(1))
	})
})
