package meshopentelemetrybackend

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	meshaccesslog_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshmetric_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshtrace_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_time "github.com/kumahq/kuma/v2/pkg/util/time"
)

type StatusUpdater struct {
	roResManager manager.ReadOnlyResourceManager
	resManager   manager.ResourceManager
	logger       logr.Logger
	metric       prometheus.Histogram
	interval     time.Duration
}

var _ component.Component = &StatusUpdater{}

func NewStatusUpdater(
	logger logr.Logger,
	roResManager manager.ReadOnlyResourceManager,
	resManager manager.ResourceManager,
	interval time.Duration,
	metrics core_metrics.Metrics,
) (component.Component, error) {
	metric := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "component_motb_status_updater",
		Help: "Summary of MeshOpenTelemetryBackend Updater component",
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &StatusUpdater{
		roResManager: roResManager,
		resManager:   resManager,
		logger:       logger,
		metric:       metric,
		interval:     interval,
	}, nil
}

func (s *StatusUpdater) Start(stop <-chan struct{}) error {
	// sleep to mitigate update conflicts with other components
	util_time.SleepUpTo(s.interval)
	s.logger.Info("starting")
	ticker := time.NewTicker(s.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := s.updateStatus(ctx); err != nil {
				s.logger.Error(err, "could not update status of MeshOpenTelemetryBackends")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	motbList := &motb_api.MeshOpenTelemetryBackendResourceList{}
	if err := s.roResManager.List(ctx, motbList); err != nil {
		return errors.Wrap(err, "could not list MeshOpenTelemetryBackends")
	}
	if len(motbList.Items) == 0 {
		return nil
	}

	refCounts, err := s.countBackendRefs(ctx)
	if err != nil {
		return err
	}

	for _, motb := range motbList.Items {
		displayName := motb.GetMeta().GetLabels()[mesh_proto.DisplayName]
		name := motb.GetMeta().GetName()
		mesh := motb.GetMeta().GetLabels()[mesh_proto.MeshTag]
		// backendRef.name uses the display name (e.g. "main-collector"),
		// but GetName() includes the K8s namespace suffix (e.g. "main-collector.kuma-system").
		// Check both to handle all environments correctly.
		// Keys are "<mesh>/<name>" to avoid counting refs from other meshes.
		count := refCounts[mesh+"/"+displayName] + refCounts[mesh+"/"+name]
		if displayName == name {
			count = refCounts[mesh+"/"+name] // avoid double counting when they're the same
		}
		condition := buildReferencedByCondition(count)

		if !conditionEquals(motb.Status.Conditions, condition) {
			log := s.logger.WithValues("meshopentelemetrybackend", name)
			motb.Status.Conditions = updateConditions(motb.Status.Conditions, condition)
			log.V(1).Info("updating referenced-by condition", "count", count)
			if err := s.resManager.Update(ctx, motb); err != nil {
				if store.IsConflict(err) {
					log.Info("couldn't update MeshOpenTelemetryBackend, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
				} else {
					log.Error(err, "could not update MeshOpenTelemetryBackend status")
				}
			}
		}
	}
	return nil
}

// countBackendRefs scans all three observability policy types and counts how many
// OTel backends reference each MeshOpenTelemetryBackend by name, keyed by "<mesh>/<name>"
// to avoid counting refs from a different mesh.
func (s *StatusUpdater) countBackendRefs(ctx context.Context) (map[string]int, error) {
	counts := map[string]int{}

	// MeshMetric
	meshMetrics := &meshmetric_api.MeshMetricResourceList{}
	if err := s.roResManager.List(ctx, meshMetrics); err != nil {
		return nil, errors.Wrap(err, "could not list MeshMetrics")
	}
	for _, mm := range meshMetrics.Items {
		mesh := mm.GetMeta().GetLabels()[mesh_proto.MeshTag]
		for _, backend := range pointer.Deref(mm.Spec.Default.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				counts[mesh+"/"+pointer.Deref(backend.OpenTelemetry.BackendRef.Name)]++
			}
		}
	}

	// MeshTrace
	meshTraces := &meshtrace_api.MeshTraceResourceList{}
	if err := s.roResManager.List(ctx, meshTraces); err != nil {
		return nil, errors.Wrap(err, "could not list MeshTraces")
	}
	for _, mt := range meshTraces.Items {
		mesh := mt.GetMeta().GetLabels()[mesh_proto.MeshTag]
		for _, backend := range pointer.Deref(mt.Spec.Default.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				counts[mesh+"/"+pointer.Deref(backend.OpenTelemetry.BackendRef.Name)]++
			}
		}
	}

	// MeshAccessLog
	meshAccessLogs := &meshaccesslog_api.MeshAccessLogResourceList{}
	if err := s.roResManager.List(ctx, meshAccessLogs); err != nil {
		return nil, errors.Wrap(err, "could not list MeshAccessLogs")
	}
	for _, mal := range meshAccessLogs.Items {
		mesh := mal.GetMeta().GetLabels()[mesh_proto.MeshTag]
		collectAccessLogBackendRefs(mesh, mal.Spec, counts)
	}

	return counts, nil
}

// collectAccessLogBackendRefs extracts backendRef names from all MeshAccessLog conf locations.
// Keys written into counts use "<mesh>/<name>" format.
func collectAccessLogBackendRefs(mesh string, spec *meshaccesslog_api.MeshAccessLog, counts map[string]int) {
	collectFromConf := func(conf meshaccesslog_api.Conf) {
		for _, backend := range pointer.Deref(conf.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				counts[mesh+"/"+pointer.Deref(backend.OpenTelemetry.BackendRef.Name)]++
			}
		}
	}
	for _, to := range pointer.Deref(spec.To) {
		collectFromConf(to.Default)
	}
	for _, from := range pointer.Deref(spec.From) {
		collectFromConf(from.Default)
	}
	for _, rule := range pointer.Deref(spec.Rules) {
		collectFromConf(rule.Default)
	}
}

func buildReferencedByCondition(count int) common_api.Condition {
	if count == 0 {
		return common_api.Condition{
			Type:    motb_api.ReferencedByPoliciesCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NotReferencedReason,
			Message: "Not referenced by any observability policy",
		}
	}
	return common_api.Condition{
		Type:    motb_api.ReferencedByPoliciesCondition,
		Status:  kube_meta.ConditionTrue,
		Reason:  motb_api.ReferencedReason,
		Message: fmt.Sprintf("Referenced by %d policy backend(s)", count),
	}
}

func conditionEquals(conditions []common_api.Condition, newCondition common_api.Condition) bool {
	for _, c := range conditions {
		if c.Type == newCondition.Type {
			return c.Status == newCondition.Status &&
				c.Reason == newCondition.Reason &&
				c.Message == newCondition.Message
		}
	}
	return false
}

func updateConditions(conditions []common_api.Condition, newCondition common_api.Condition) []common_api.Condition {
	for i, c := range conditions {
		if c.Type == newCondition.Type {
			conditions[i] = newCondition
			return conditions
		}
	}
	return append(conditions, newCondition)
}
