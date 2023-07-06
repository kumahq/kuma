package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_cache "github.com/kumahq/kuma/pkg/util/cache"
)

type Metrics struct {
	XdsGenerations       prometheus.Summary
	XdsGenerationsErrors prometheus.Counter
	KubeAuthCache        *prometheus.CounterVec
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	xdsGenerations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "xds_generation",
		Help:       "Summary of XDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	})
	xdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "xds_generation_errors",
		Help: "Counter of errors during XDS generation",
	})
	kubeAuthCache := util_cache.NewMetric(
		"kube_auth_cache",
		"Number of cache operations for Kubernetes authentication on XDS connection",
	)
	if err := metrics.BulkRegister(xdsGenerations, xdsGenerationsErrors, kubeAuthCache); err != nil {
		return nil, err
	}

	return &Metrics{
		XdsGenerations:       xdsGenerations,
		XdsGenerationsErrors: xdsGenerationsErrors,
		KubeAuthCache:        kubeAuthCache,
	}, nil
}
