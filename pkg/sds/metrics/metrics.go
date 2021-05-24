package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type Metrics struct {
	sdsGeneration        *prometheus.SummaryVec
	sdsGenerationsErrors *prometheus.CounterVec
	certGenerations      *prometheus.CounterVec
	Callbacks            util_xds.Callbacks

	sync.RWMutex
	activeWatchdogs        *prometheus.GaugeVec
	activeWatchdogsCounter uint32
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	activeWatchdogs := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sds_watchdogs",
		Help: "Number of active watchdogs",
	}, []string{"apiVersion"})

	sdsGenerations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "sds_generation",
		Help:       "Summary of SDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	}, []string{"apiVersion"})

	sdsGenerationsErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Counter of errors during SDS generation",
		Name: "sds_generation_errors",
	}, []string{"apiVersion"})

	certGenerationsMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Number of generated certificates",
		Name: "sds_cert_generation",
	}, []string{"apiVersion"})

	if err := metrics.BulkRegister(activeWatchdogs, sdsGenerations, sdsGenerationsErrors, certGenerationsMetric); err != nil {
		return nil, err
	}

	statsCallbacks, err := util_xds.NewStatsCallbacks(metrics, "sds")
	if err != nil {
		return nil, err
	}

	return &Metrics{
		activeWatchdogs:      activeWatchdogs,
		sdsGeneration:        sdsGenerations,
		sdsGenerationsErrors: sdsGenerationsErrors,
		certGenerations:      certGenerationsMetric,
		Callbacks:            statsCallbacks,
	}, nil
}

func (s *Metrics) SdsGeneration(apiVersion envoy_common.APIVersion) prometheus.Observer {
	return s.sdsGeneration.WithLabelValues(string(apiVersion))
}

func (s *Metrics) SdsGenerationsErrors(apiVersion envoy_common.APIVersion) prometheus.Counter {
	return s.sdsGenerationsErrors.WithLabelValues(string(apiVersion))
}

func (s *Metrics) CertGenerations(apiVersion envoy_common.APIVersion) prometheus.Counter {
	return s.certGenerations.WithLabelValues(string(apiVersion))
}

func (s *Metrics) IncrementActiveWatchdogs(apiVersion envoy_common.APIVersion) {
	s.Lock()
	defer s.Unlock()
	s.activeWatchdogsCounter += 1
	s.activeWatchdogs.WithLabelValues(string(apiVersion)).Set(float64(s.activeWatchdogsCounter))
}

func (s *Metrics) DecrementActiveWatchdogs(apiVersion envoy_common.APIVersion) {
	s.Lock()
	defer s.Unlock()
	s.activeWatchdogsCounter -= 1
	s.activeWatchdogs.WithLabelValues(string(apiVersion)).Set(float64(s.activeWatchdogsCounter))
}
