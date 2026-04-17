package dnsproxy

import (
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	RequestDuration             prometheus.Histogram
	UpstreamRequestDuration     prometheus.Histogram
	UpstreamRequestFailureCount prometheus.Counter
	QueriesTotal                *prometheus.CounterVec
	ResponseCodesTotal          *prometheus.CounterVec
	EntriesTotal                prometheus.Gauge
	ConfigReadyWaitSeconds      prometheus.Histogram
	ConfigReadyGateBypassed     prometheus.Counter
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
	queriesTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "kuma_dp_dns_queries_total",
		Help:        "Total DNS queries handled, by query type and source (local map or upstream).",
		ConstLabels: constLabels,
	}, []string{"qtype", "source"})
	registerer.MustRegister(queriesTotal)
	responseCodesTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "kuma_dp_dns_response_codes_total",
		Help:        "Total DNS responses by response code.",
		ConstLabels: constLabels,
	}, []string{"rcode"})
	registerer.MustRegister(responseCodesTotal)
	entriesTotal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "kuma_dp_dns_entries_total",
		Help:        "Current number of hostnames in the DNS proxy map.",
		ConstLabels: constLabels,
	})
	registerer.MustRegister(entriesTotal)
	configReadyWaitSeconds := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "kuma_dp_dns_config_ready_wait_seconds",
		Help:        "Time between DNS proxy start and first configuration received from Envoy.",
		ConstLabels: constLabels,
		Buckets:     []float64{0.1, 0.5, 1, 2, 5, 10, 30},
	})
	registerer.MustRegister(configReadyWaitSeconds)
	configReadyGateBypassed := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "kuma_dp_dns_config_ready_gate_bypassed_total",
		Help:        "Total times the DNS config readiness gate was bypassed due to timeout.",
		ConstLabels: constLabels,
	})
	registerer.MustRegister(configReadyGateBypassed)
	return &metrics{
		RequestDuration:             requestDuration,
		UpstreamRequestDuration:     upstreamRequestDuration,
		UpstreamRequestFailureCount: upstreamRequestFailureCount,
		QueriesTotal:                queriesTotal,
		ResponseCodesTotal:          responseCodesTotal,
		EntriesTotal:                entriesTotal,
		ConfigReadyWaitSeconds:      configReadyWaitSeconds,
		ConfigReadyGateBypassed:     configReadyGateBypassed,
	}
}

func qtypeLabel(qtype uint16) string {
	switch qtype {
	case dns.TypeA:
		return "A"
	case dns.TypeAAAA:
		return "AAAA"
	default:
		return "other"
	}
}

func rcodeLabel(rcode int) string {
	switch rcode {
	case dns.RcodeSuccess:
		return "noerror"
	case dns.RcodeNameError:
		return "nxdomain"
	case dns.RcodeServerFailure:
		return "servfail"
	default:
		return "other"
	}
}
