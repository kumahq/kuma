package server

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

const (
	ReasonResync    = "resync"
	ReasonEvent     = "event"
	ResultChanged   = "changed"
	ResultNoChanges = "no_changes"
)

type Metrics struct {
	KdsGenerations      *prometheus.SummaryVec
	KdsGenerationErrors prometheus.Counter
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

	if err := metrics.BulkRegister(kdsGenerations, kdsGenerationsErrors); err != nil {
		return nil, err
	}

	return &Metrics{
		KdsGenerations:      kdsGenerations,
		KdsGenerationErrors: kdsGenerationsErrors,
	}, nil
}
