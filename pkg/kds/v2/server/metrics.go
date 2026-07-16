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
<<<<<<< HEAD
	KdsGenerations      *prometheus.SummaryVec
	KdsGenerationErrors prometheus.Counter
=======
	KdsGenerations             *prometheus.HistogramVec
	KdsGenerationErrors        prometheus.Counter
	KdsZoneActiveConnections   *prometheus.GaugeVec
	KdsNackTotal               *prometheus.CounterVec
	KdsZoneAttributionRewrites *prometheus.CounterVec
>>>>>>> db8309dd90 (fix(kds): attribute zone-to-global synced resources by authenticated zone (#17456))
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	kdsGenerations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "kds_delta_generation",
		Help:       "Summary of KDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	}, []string{"reason", "result"})

	kdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Help: "Counter of errors during KDS generation",
		Name: "kds_delta_generation_errors",
	})

<<<<<<< HEAD
	if err := metrics.BulkRegister(kdsGenerations, kdsGenerationsErrors); err != nil {
=======
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
>>>>>>> db8309dd90 (fix(kds): attribute zone-to-global synced resources by authenticated zone (#17456))
		return nil, err
	}

	return &Metrics{
<<<<<<< HEAD
		KdsGenerations:      kdsGenerations,
		KdsGenerationErrors: kdsGenerationsErrors,
=======
		KdsGenerations:             kdsGenerations,
		KdsGenerationErrors:        kdsGenerationsErrors,
		KdsZoneActiveConnections:   kdsZoneActiveConnections,
		KdsNackTotal:               kdsNackTotal,
		KdsZoneAttributionRewrites: kdsZoneAttributionRewrites,
>>>>>>> db8309dd90 (fix(kds): attribute zone-to-global synced resources by authenticated zone (#17456))
	}, nil
}
