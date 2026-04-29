package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	util_cache "github.com/kumahq/kuma/v2/pkg/util/cache"
)

type Metrics struct {
	XdsGenerations                         *prometheus.HistogramVec
	XdsGenerationsErrors                   prometheus.Counter
	KubeAuthCache                          *prometheus.CounterVec
	CertExpirationTimestamp                *prometheus.GaugeVec
	SnapshotResources                      *prometheus.HistogramVec
	XdsStreamStaleOwnerTakeovers           *prometheus.CounterVec
	XdsStreamRegistrationInProgressRetries *prometheus.CounterVec
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
	certExpirationTimestamp := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xds_cert_expiration_timestamp_seconds",
		Help: "Unix timestamp (seconds) when the MeshIdentity workload certificate expires. Compute remaining lifetime in PromQL as `xds_cert_expiration_timestamp_seconds - time()`.",
	}, []string{"mesh"})
	snapshotResources := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "xds_snapshot_resources",
		Help:    "Distribution of resource counts per xDS snapshot by resource type.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12),
	}, []string{"resource_type"})
	xdsStreamStaleOwnerTakeovers := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xds_stream_stale_owner_takeovers_total",
		Help: "Total number of stale xDS stream owners replaced by a new stream",
	}, []string{"mesh", "proxy_type"})
	xdsStreamRegistrationInProgressRetries := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xds_stream_registration_in_progress_retries_total",
		Help: "Total number of xDS stream requests rejected because another registration is in progress",
	}, []string{"mesh", "proxy_type"})
	if err := metrics.BulkRegister(
		xdsGenerations,
		xdsGenerationsErrors,
		kubeAuthCache,
		certExpirationTimestamp,
		snapshotResources,
		xdsStreamStaleOwnerTakeovers,
		xdsStreamRegistrationInProgressRetries,
	); err != nil {
		return nil, err
	}

	return &Metrics{
		XdsGenerations:                         xdsGenerations,
		XdsGenerationsErrors:                   xdsGenerationsErrors,
		KubeAuthCache:                          kubeAuthCache,
		CertExpirationTimestamp:                certExpirationTimestamp,
		SnapshotResources:                      snapshotResources,
		XdsStreamStaleOwnerTakeovers:           xdsStreamStaleOwnerTakeovers,
		XdsStreamRegistrationInProgressRetries: xdsStreamRegistrationInProgressRetries,
	}, nil
}
