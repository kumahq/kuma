package dnsproxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	RequestDuration             prometheus.Histogram
	UpstreamRequestDuration     prometheus.Histogram
	UpstreamRequestFailureCount prometheus.Counter
}

func newMetrics(registerer prometheus.Registerer, constLabels prometheus.Labels) *metrics {
	upstreamRequestDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "kuma_dp_dns_upstream_request_duration_seconds",
		Help:        "The duration of the proxied requests.",
		ConstLabels: constLabels,
	})
	registerer.MustRegister(upstreamRequestDuration)
	requestDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "kuma_dp_dns_request_duration_seconds",
		Help:        "The duration of the request (inclusive of request that use internal DNS map).",
		ConstLabels: constLabels,
	})
	registerer.MustRegister(requestDuration)
	upstreamRequestFailureCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "kuma_dp_dns_upstream_request_failure_count",
		Help:        "The total number of failed upstream requests.",
		ConstLabels: constLabels,
	})
	registerer.MustRegister(upstreamRequestFailureCount)
	return &metrics{
		RequestDuration:             requestDuration,
		UpstreamRequestDuration:     upstreamRequestDuration,
		UpstreamRequestFailureCount: upstreamRequestFailureCount,
	}
}
