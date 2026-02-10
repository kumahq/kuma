package tracing

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/kumahq/kuma/v2/pkg/core"
)

var log = core.Log.WithName("tracing")

// SafeSpanEnd ends a span, recovering from any panics to prevent crashes
// during OTel provider initialization/shutdown race conditions.
//
// This function should be used instead of calling span.End() directly to
// handle race conditions where the OTel provider might be shutting down.
func SafeSpanEnd(span trace.Span) {
	if span == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.V(1).Info("recovered from panic in span.End()", "panic", r)
		}
	}()
	span.End()
}
