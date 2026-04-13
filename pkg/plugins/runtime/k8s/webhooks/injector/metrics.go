package injector

import (
	"github.com/prometheus/client_golang/prometheus"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
)

type injectionMetrics struct {
	InjectionTotal *prometheus.CounterVec
}

func classifyInjectionError(err error) string {
	if kube_errors.IsNotFound(err) {
		return "mesh_not_found"
	}
	return "config_error"
}

func newInjectionMetrics(metrics core_metrics.Metrics) (*injectionMetrics, error) {
	injectionTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sidecar_injection_total",
		Help: "Total sidecar injection decisions by result (success/skipped/failed) and reason.",
	}, []string{"result", "reason"})
	if err := metrics.Register(injectionTotal); err != nil {
		return nil, err
	}
	return &injectionMetrics{InjectionTotal: injectionTotal}, nil
}
