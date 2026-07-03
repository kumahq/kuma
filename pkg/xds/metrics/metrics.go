package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	util_cache "github.com/kumahq/kuma/v2/pkg/util/cache"
)

type Metrics struct {
	XdsGenerations             *prometheus.SummaryVec
	XdsGenerationsErrors       prometheus.Counter
	KubeAuthCache              *prometheus.CounterVec
	DataplaneConfigRegenerated *prometheus.CounterVec
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	xdsGenerations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "xds_generation",
		Help:       "Summary of XDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	}, []string{"proxy_type", "result"})
	xdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "xds_generation_errors",
		Help: "Counter of errors during XDS generation",
	})
	kubeAuthCache := util_cache.NewMetric(
		"kube_auth_cache",
		"Number of cache operations for Kubernetes authentication on XDS connection",
	)
	dataplaneConfigRegenerated := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xds_dataplane_config_regenerated_total",
		Help: "Counter of successful Dataplane xDS reconcile attempts triggered by mesh-context hash changes.",
	}, []string{"mesh"})
	if err := metrics.BulkRegister(
		xdsGenerations,
		xdsGenerationsErrors,
		kubeAuthCache,
		dataplaneConfigRegenerated,
	); err != nil {
		return nil, err
	}

	return &Metrics{
		XdsGenerations:             xdsGenerations,
		XdsGenerationsErrors:       xdsGenerationsErrors,
		KubeAuthCache:              kubeAuthCache,
		DataplaneConfigRegenerated: dataplaneConfigRegenerated,
	}, nil
}
