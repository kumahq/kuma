package meshopentelemetrybackend

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	meshaccesslog_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshmetric_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshtrace_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/v2/pkg/util/maps"
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
	defer ticker.Stop()
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

	meshMetrics := &meshmetric_api.MeshMetricResourceList{}
	if err := s.roResManager.List(ctx, meshMetrics); err != nil {
		return errors.Wrap(err, "could not list MeshMetrics")
	}

	meshTraces := &meshtrace_api.MeshTraceResourceList{}
	if err := s.roResManager.List(ctx, meshTraces); err != nil {
		return errors.Wrap(err, "could not list MeshTraces")
	}

	meshAccessLogs := &meshaccesslog_api.MeshAccessLogResourceList{}
	if err := s.roResManager.List(ctx, meshAccessLogs); err != nil {
		return errors.Wrap(err, "could not list MeshAccessLogs")
	}

	backendsByMesh := buildBackendIndex(motbList.Items)

	s.updateMeshMetricStatuses(ctx, meshMetrics, backendsByMesh)
	s.updateMeshTraceStatuses(ctx, meshTraces, backendsByMesh)
	s.updateMeshAccessLogStatuses(ctx, meshAccessLogs, backendsByMesh)

	if len(motbList.Items) == 0 {
		return nil
	}

	refCounts := countBackendRefs(meshMetrics, meshTraces, meshAccessLogs, backendsByMesh)
	for _, motb := range motbList.Items {
		name := motb.GetMeta().GetName()
		mesh := motb.GetMeta().GetMesh()
		count := refCounts[mesh+"/"+name]
		conditions := []common_api.Condition{
			buildReferencedByCondition(count),
		}

		updatedConditions := motb.Status.Conditions
		changed := false
		for _, condition := range conditions {
			if conditionEquals(updatedConditions, condition) {
				continue
			}
			updatedConditions = updateConditions(updatedConditions, condition)
			changed = true
		}

		if changed {
			log := s.logger.WithValues("meshopentelemetrybackend", name)
			motb.Status.Conditions = updatedConditions
			log.V(1).Info("updating OTEL backend status", "policyRefCount", count)
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
// policies reference each MeshOpenTelemetryBackend, keyed by "<mesh>/<name>".
func countBackendRefs(
	meshMetrics *meshmetric_api.MeshMetricResourceList,
	meshTraces *meshtrace_api.MeshTraceResourceList,
	meshAccessLogs *meshaccesslog_api.MeshAccessLogResourceList,
	indexByMesh map[string]*backendIndex,
) map[string]int {
	counts := map[string]int{}
	countRef := func(mesh string, ref *common_api.BackendResourceRef) {
		if resolved, ok := resolveRef(indexByMesh, mesh, ref); ok {
			counts[mesh+"/"+resolved]++
		}
	}

	for _, mm := range meshMetrics.Items {
		for _, ref := range meshMetricRefs(mm) {
			countRef(mm.GetMeta().GetMesh(), ref)
		}
	}
	for _, mt := range meshTraces.Items {
		for _, ref := range meshTraceRefs(mt) {
			countRef(mt.GetMeta().GetMesh(), ref)
		}
	}
	for _, mal := range meshAccessLogs.Items {
		for _, ref := range meshAccessLogRefs(mal) {
			countRef(mal.GetMeta().GetMesh(), ref)
		}
	}
	return counts
}

type motbEntry struct {
	name         string
	labels       map[string]string
	creationTime time.Time
}

type backendIndex struct {
	byName    map[string]struct{}
	resources []motbEntry
}

func buildBackendIndex(motbs []*motb_api.MeshOpenTelemetryBackendResource) map[string]*backendIndex {
	indexByMesh := map[string]*backendIndex{}
	for _, motb := range motbs {
		mesh := motb.GetMeta().GetMesh()
		if indexByMesh[mesh] == nil {
			indexByMesh[mesh] = &backendIndex{
				byName: map[string]struct{}{},
			}
		}
		name := motb.GetMeta().GetName()
		indexByMesh[mesh].byName[name] = struct{}{}
		indexByMesh[mesh].resources = append(indexByMesh[mesh].resources, motbEntry{
			name:         name,
			labels:       motb.GetMeta().GetLabels(),
			creationTime: motb.GetMeta().GetCreationTime(),
		})
	}
	return indexByMesh
}

// resolveRef resolves a BackendResourceRef to a MOTB name within a mesh.
// Handles both Name-based (direct lookup) and Labels-based (label matching,
// oldest wins) references.
func resolveRef(indexByMesh map[string]*backendIndex, mesh string, ref *common_api.BackendResourceRef) (string, bool) {
	idx := indexByMesh[mesh]
	if idx == nil {
		return "", false
	}
	if ref.Name != "" {
		if _, exists := idx.byName[ref.Name]; exists {
			return ref.Name, true
		}
		return "", false
	}
	if len(ref.Labels) > 0 {
		return matchByLabels(idx.resources, ref.Labels)
	}
	return "", false
}

func matchByLabels(resources []motbEntry, selector map[string]string) (string, bool) {
	ls := common_api.LabelSelector{MatchLabels: &selector}
	var best *motbEntry
	for i := range resources {
		if ls.Matches(resources[i].labels) {
			if best == nil || resources[i].creationTime.Before(best.creationTime) {
				best = &resources[i]
			}
		}
	}
	if best != nil {
		return best.name, true
	}
	return "", false
}

// backendRefKey returns a string key for a BackendResourceRef for use in
// deduplication maps. Name-based refs use the name directly, labels-based
// refs use a sorted label representation.
func backendRefKey(ref *common_api.BackendResourceRef) string {
	if ref.Name != "" {
		return ref.Name
	}
	parts := make([]string, 0, len(ref.Labels))
	for _, k := range util_maps.SortedKeys(ref.Labels) {
		parts = append(parts, k+"="+ref.Labels[k])
	}
	return "labels:" + strings.Join(parts, ",")
}

func unresolvedBackendRefs(mesh string, refs []*common_api.BackendResourceRef, indexByMesh map[string]*backendIndex) []string {
	if len(refs) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var unresolved []string
	for _, ref := range refs {
		key := backendRefKey(ref)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, ok := resolveRef(indexByMesh, mesh, ref); !ok {
			unresolved = append(unresolved, key)
		}
	}
	slices.Sort(unresolved)
	return unresolved
}

func buildBackendRefsResolvedCondition(
	refs []*common_api.BackendResourceRef,
	unresolved []string,
	conditionType string,
	resolvedReason string,
	unresolvedReason string,
) common_api.Condition {
	if len(refs) == 0 {
		return common_api.Condition{
			Type:    conditionType,
			Status:  kube_meta.ConditionTrue,
			Reason:  resolvedReason,
			Message: "No MeshOpenTelemetryBackend references configured",
		}
	}

	if len(unresolved) == 0 {
		return common_api.Condition{
			Type:    conditionType,
			Status:  kube_meta.ConditionTrue,
			Reason:  resolvedReason,
			Message: "All MeshOpenTelemetryBackend references are resolved",
		}
	}

	return common_api.Condition{
		Type:    conditionType,
		Status:  kube_meta.ConditionFalse,
		Reason:  unresolvedReason,
		Message: fmt.Sprintf("Unresolved MeshOpenTelemetryBackend references: %s", strings.Join(unresolved, ", ")),
	}
}

func meshMetricRefs(mm *meshmetric_api.MeshMetricResource) []*common_api.BackendResourceRef {
	var refs []*common_api.BackendResourceRef
	for _, backend := range pointer.Deref(mm.Spec.Default.Backends) {
		if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
			refs = append(refs, backend.OpenTelemetry.BackendRef)
		}
	}
	return refs
}

func meshTraceRefs(mt *meshtrace_api.MeshTraceResource) []*common_api.BackendResourceRef {
	var refs []*common_api.BackendResourceRef
	for _, backend := range pointer.Deref(mt.Spec.Default.Backends) {
		if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
			refs = append(refs, backend.OpenTelemetry.BackendRef)
		}
	}
	return refs
}

func meshAccessLogRefs(mal *meshaccesslog_api.MeshAccessLogResource) []*common_api.BackendResourceRef {
	var refs []*common_api.BackendResourceRef
	collectFromConf := func(conf meshaccesslog_api.Conf) {
		for _, backend := range pointer.Deref(conf.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				refs = append(refs, backend.OpenTelemetry.BackendRef)
			}
		}
	}
	for _, to := range pointer.Deref(mal.Spec.To) {
		collectFromConf(to.Default)
	}
	for _, from := range pointer.Deref(mal.Spec.From) {
		collectFromConf(from.Default)
	}
	for _, rule := range pointer.Deref(mal.Spec.Rules) {
		collectFromConf(rule.Default)
	}
	return refs
}

func (s *StatusUpdater) updateMeshMetricStatuses(
	ctx context.Context,
	meshMetrics *meshmetric_api.MeshMetricResourceList,
	indexByMesh map[string]*backendIndex,
) {
	for _, mm := range meshMetrics.Items {
		refs := meshMetricRefs(mm)
		condition := buildBackendRefsResolvedCondition(refs,
			unresolvedBackendRefs(mm.GetMeta().GetMesh(), refs, indexByMesh),
			common_api.BackendRefsResolvedCondition,
			common_api.AllBackendRefsResolvedReason,
			common_api.UnresolvedBackendRefsReason,
		)
		if mm.Status == nil {
			mm.Status = &meshmetric_api.MeshMetricStatus{}
		}
		s.updateResourceCondition(ctx, mm, &mm.Status.Conditions, condition, "meshmetric", "MeshMetric")
	}
}

func (s *StatusUpdater) updateMeshTraceStatuses(
	ctx context.Context,
	meshTraces *meshtrace_api.MeshTraceResourceList,
	indexByMesh map[string]*backendIndex,
) {
	for _, mt := range meshTraces.Items {
		refs := meshTraceRefs(mt)
		condition := buildBackendRefsResolvedCondition(refs,
			unresolvedBackendRefs(mt.GetMeta().GetMesh(), refs, indexByMesh),
			common_api.BackendRefsResolvedCondition,
			common_api.AllBackendRefsResolvedReason,
			common_api.UnresolvedBackendRefsReason,
		)
		if mt.Status == nil {
			mt.Status = &meshtrace_api.MeshTraceStatus{}
		}
		s.updateResourceCondition(ctx, mt, &mt.Status.Conditions, condition, "meshtrace", "MeshTrace")
	}
}

func (s *StatusUpdater) updateMeshAccessLogStatuses(
	ctx context.Context,
	meshAccessLogs *meshaccesslog_api.MeshAccessLogResourceList,
	indexByMesh map[string]*backendIndex,
) {
	for _, mal := range meshAccessLogs.Items {
		refs := meshAccessLogRefs(mal)
		condition := buildBackendRefsResolvedCondition(refs,
			unresolvedBackendRefs(mal.GetMeta().GetMesh(), refs, indexByMesh),
			common_api.BackendRefsResolvedCondition,
			common_api.AllBackendRefsResolvedReason,
			common_api.UnresolvedBackendRefsReason,
		)
		if mal.Status == nil {
			mal.Status = &meshaccesslog_api.MeshAccessLogStatus{}
		}
		s.updateResourceCondition(ctx, mal, &mal.Status.Conditions, condition, "meshaccesslog", "MeshAccessLog")
	}
}

func (s *StatusUpdater) updateResourceCondition(
	ctx context.Context,
	resource core_model.Resource,
	conditions *[]common_api.Condition,
	condition common_api.Condition,
	logKey, typeName string,
) {
	if conditionEquals(*conditions, condition) {
		return
	}
	*conditions = updateConditions(*conditions, condition)
	log := s.logger.WithValues(logKey, resource.GetMeta().GetName(), "mesh", resource.GetMeta().GetMesh())
	if err := s.resManager.Update(ctx, resource); err != nil {
		if store.IsConflict(err) {
			log.Info(fmt.Sprintf("couldn't update %s, will retry next interval", typeName), "interval", s.interval)
		} else {
			log.Error(err, fmt.Sprintf("could not update %s status", typeName))
		}
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
