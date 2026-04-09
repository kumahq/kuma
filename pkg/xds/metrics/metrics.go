package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	util_cache "github.com/kumahq/kuma/v2/pkg/util/cache"
)

type Metrics struct {
	XdsGenerations          *prometheus.HistogramVec
	XdsGenerationsErrors    prometheus.Counter
	KubeAuthCache           *prometheus.CounterVec
	CertExpirationRemaining *prometheus.GaugeVec
	SnapshotResourcesTotal  *prometheus.HistogramVec
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	xdsGenerations := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "xds_generation",
		Help: "Summary of XDS Snapshot generation",
	}, []string{"proxy_type", "result"})
	xdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "xds_generation_errors",
		Help: "Counter of errors during XDS generation",
	})
	kubeAuthCache := util_cache.NewMetric(
		"kube_auth_cache",
		"Number of cache operations for Kubernetes authentication on XDS connection",
	)
	certExpirationRemaining := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cert_expiration_remaining_seconds",
		Help: "Seconds until the MeshIdentity workload certificate expires.",
	}, []string{"mesh"})
	snapshotResourcesTotal := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "xds_snapshot_resources_total",
		Help:    "Distribution of resource counts per xDS snapshot by resource type.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12),
	}, []string{"resource_type"})
	if err := metrics.BulkRegister(xdsGenerations, xdsGenerationsErrors, kubeAuthCache, certExpirationRemaining, snapshotResourcesTotal); err != nil {
		return nil, err
	}

	return &Metrics{
		XdsGenerations:          xdsGenerations,
		XdsGenerationsErrors:    xdsGenerationsErrors,
		KubeAuthCache:           kubeAuthCache,
		CertExpirationRemaining: certExpirationRemaining,
		SnapshotResourcesTotal:  snapshotResourcesTotal,
	}, nil
}
