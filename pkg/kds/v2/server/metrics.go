package server

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
)

const (
	ReasonResync    = "resync"
	ReasonEvent     = "event"
	ResultChanged   = "changed"
	ResultNoChanges = "no_changes"
)

type Metrics struct {
	KdsGenerations             *prometheus.HistogramVec
	KdsGenerationErrors        prometheus.Counter
	KdsZoneActiveConnections   *prometheus.GaugeVec
	KdsNackTotal               *prometheus.CounterVec
	KdsZoneAttributionRewrites *prometheus.CounterVec
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	kdsGenerations := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "kds_delta_generation",
		Help: "Summary of KDS Snapshot generation",
	}, []string{"reason", "result", "zone_name"})

	kdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Help: "Counter of errors during KDS generation",
		Name: "kds_delta_generation_errors",
	})

	kdsZoneActiveConnections := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kds_zone_active_connections",
		Help: "Number of active KDS streams per zone.",
	}, []string{"zone_name"})

	kdsNackTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kds_nack_total",
		Help: "Total KDS NACKs sent by zone and resource type.",
	}, []string{"zone_name", "resource_type"})

	kdsZoneAttributionRewrites := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kds_zone_attribution_rewrites_total",
		Help: "Total zone-to-global synced resources whose zone attribution was rewritten to the connecting zone's client-id because a sender-provided value differed, by resource type.",
	}, []string{"resource_type"})

	if err := metrics.BulkRegister(kdsGenerations, kdsGenerationsErrors, kdsZoneActiveConnections, kdsNackTotal, kdsZoneAttributionRewrites); err != nil {
		return nil, err
	}

	return &Metrics{
		KdsGenerations:             kdsGenerations,
		KdsGenerationErrors:        kdsGenerationsErrors,
		KdsZoneActiveConnections:   kdsZoneActiveConnections,
		KdsNackTotal:               kdsNackTotal,
		KdsZoneAttributionRewrites: kdsZoneAttributionRewrites,
	}, nil
}
