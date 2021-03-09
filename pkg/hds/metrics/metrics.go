package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type Metrics struct {
	HdsGenerations          prometheus.Summary
	HdsGenerationsErrors    prometheus.Counter
	ResponsesReceivedMetric prometheus.Counter
	RequestsReceivedMetric  prometheus.Counter

	streamsActiveMux sync.RWMutex
	StreamsActive    int
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	m := &Metrics{}

	m.HdsGenerations = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "hds_generation",
		Help:       "Summary of HDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(m.HdsGenerations); err != nil {
		return nil, err
	}

	m.HdsGenerationsErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hds_generation_errors",
		Help: "Counter of errors during HDS generation",
	})
	if err := metrics.Register(m.HdsGenerationsErrors); err != nil {
		return nil, err
	}

	m.ResponsesReceivedMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hds_responses_sent",
		Help: "Number of EndpointHealthResponses from a client",
	})
	if err := metrics.Register(m.ResponsesReceivedMetric); err != nil {
		return nil, err
	}

	m.RequestsReceivedMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hds_requests_received",
		Help: "Number of HealthCheckRequests from a client",
	})
	if err := metrics.Register(m.RequestsReceivedMetric); err != nil {
		return nil, err
	}

	streamsActive := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "hds_streams_active",
		Help: "Number of active connections between a server and a client",
	}, func() float64 {
		m.streamsActiveMux.RLock()
		defer m.streamsActiveMux.RUnlock()
		return float64(m.StreamsActive)
	})
	if err := metrics.Register(streamsActive); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Metrics) StreamsActiveInc() {
	m.streamsActiveMux.Lock()
	defer m.streamsActiveMux.Unlock()
	m.StreamsActive++
}

func (m *Metrics) StreamsActiveDec() {
	m.streamsActiveMux.Lock()
	defer m.streamsActiveMux.Unlock()
	m.StreamsActive--
}
