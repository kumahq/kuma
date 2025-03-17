package dnsproxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	RequestDuration             prometheus.Summary
	UpstreamRequestDuration     prometheus.Summary
	UpstreamRequestFailureCount prometheus.Counter
}

func newMetrics() *metrics {
	upstreamRequestDuration := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "kuma_dp_embeddeddns_proxy_upstream_request_duration_seconds",
		Help: "The duration of the proxied requests.",
	})
	prometheus.MustRegister(upstreamRequestDuration)
	requestDuration := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "kuma_dp_embeddeddns_proxy_request_duration_seconds",
		Help: "The duration of the request (inclusive of request that use internal DNS map).",
	})
	prometheus.MustRegister(requestDuration)
	upstreamRequestFailureCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kuma_dp_embeddeddns_proxy_upstream_request_failure_count",
		Help: "The total number of failed upstream requests.",
	})

	return &metrics{
		RequestDuration:             requestDuration,
		UpstreamRequestDuration:     upstreamRequestDuration,
		UpstreamRequestFailureCount: upstreamRequestFailureCount,
	}
}
