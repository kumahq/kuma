package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	util_cache "github.com/kumahq/kuma/v2/pkg/util/cache"
)

type Metrics struct {
<<<<<<< HEAD
	XdsGenerations       *prometheus.SummaryVec
	XdsGenerationsErrors prometheus.Counter
	KubeAuthCache        *prometheus.CounterVec
=======
	XdsGenerations             *prometheus.HistogramVec
	XdsGenerationsErrors       prometheus.Counter
	KubeAuthCache              *prometheus.CounterVec
	CertExpirationTimestamp    *prometheus.GaugeVec
	SnapshotResources          *prometheus.HistogramVec
	DataplaneConfigRegenerated *prometheus.CounterVec
>>>>>>> c222418355 (fix(xds): ignore workload status in mesh hash (#17064))
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
<<<<<<< HEAD
	if err := metrics.BulkRegister(xdsGenerations, xdsGenerationsErrors, kubeAuthCache); err != nil {
=======
	certExpirationTimestamp := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xds_cert_expiration_timestamp_seconds",
		Help: "Unix timestamp (seconds) when the MeshIdentity workload certificate expires. Compute remaining lifetime in PromQL as `xds_cert_expiration_timestamp_seconds - time()`.",
	}, []string{"mesh"})
	snapshotResources := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "xds_snapshot_resources",
		Help:    "Distribution of resource counts per xDS snapshot by resource type.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12),
	}, []string{"resource_type"})
	dataplaneConfigRegenerated := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xds_dataplane_config_regenerated_total",
		Help: "Counter of successful Dataplane xDS reconcile attempts triggered by mesh-context hash changes.",
	}, []string{"mesh"})
	if err := metrics.BulkRegister(
		xdsGenerations,
		xdsGenerationsErrors,
		kubeAuthCache,
		certExpirationTimestamp,
		snapshotResources,
		dataplaneConfigRegenerated,
	); err != nil {
>>>>>>> c222418355 (fix(xds): ignore workload status in mesh hash (#17064))
		return nil, err
	}

	return &Metrics{
<<<<<<< HEAD
		XdsGenerations:       xdsGenerations,
		XdsGenerationsErrors: xdsGenerationsErrors,
		KubeAuthCache:        kubeAuthCache,
=======
		XdsGenerations:             xdsGenerations,
		XdsGenerationsErrors:       xdsGenerationsErrors,
		KubeAuthCache:              kubeAuthCache,
		CertExpirationTimestamp:    certExpirationTimestamp,
		SnapshotResources:          snapshotResources,
		DataplaneConfigRegenerated: dataplaneConfigRegenerated,
>>>>>>> c222418355 (fix(xds): ignore workload status in mesh hash (#17064))
	}, nil
}
