package meshopentelemetrybackend

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	meshaccesslog_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshmetric_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshtrace_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_time "github.com/kumahq/kuma/v2/pkg/util/time"
	otel_status "github.com/kumahq/kuma/v2/pkg/xds/otel/status"
)

type StatusUpdater struct {
	roResManager manager.ReadOnlyResourceManager
	resManager   manager.ResourceManager
	logger       logr.Logger
	metric       prometheus.Histogram
	interval     time.Duration
}

var _ component.Component = &StatusUpdater{}

const (
	otelSignalStateReady     = otel_status.SignalStateReady
	otelSignalStateBlocked   = otel_status.SignalStateBlocked
	otelSignalStateMissing   = otel_status.SignalStateMissing
	otelSignalStateAmbiguous = otel_status.SignalStateAmbiguous
)

type backendRuntimeSummary struct {
	reportingDataplanes          int
	readyDataplanes              int
	blockedDataplanes            int
	missingRequiredEnvDataplanes int
	ambiguousDataplanes          int
}

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

	dpInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := s.roResManager.List(ctx, dpInsights); err != nil {
		return errors.Wrap(err, "could not list DataplaneInsights")
	}

	backendsByMesh := buildBackendNameIndex(motbList.Items)

	if err := s.updateMeshMetricStatuses(ctx, meshMetrics, backendsByMesh); err != nil {
		return err
	}
	if err := s.updateMeshTraceStatuses(ctx, meshTraces, backendsByMesh); err != nil {
		return err
	}
	if err := s.updateMeshAccessLogStatuses(ctx, meshAccessLogs, backendsByMesh); err != nil {
		return err
	}

	if len(motbList.Items) == 0 {
		return nil
	}

	refCounts := countBackendRefs(meshMetrics, meshTraces, meshAccessLogs, backendsByMesh)
	runtimeSummaries := buildBackendRuntimeSummaries(dpInsights.Items)
	for _, motb := range motbList.Items {
		name := motb.GetMeta().GetName()
		mesh := motb.GetMeta().GetMesh()
		count := refCounts[mesh+"/"+name]
		summary := runtimeSummaries[mesh+"/"+name]
		conditions := []common_api.Condition{
			buildReferencedByCondition(count),
			buildReadyCondition(summary),
			buildBlockedCondition(summary),
			buildMissingRequiredEnvCondition(summary),
			buildAmbiguousCondition(summary),
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
			log.V(1).Info("updating OTEL backend status", "policyRefCount", count, "reportingDataplanes", summary.reportingDataplanes)
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
func countBackendRefs(
	meshMetrics *meshmetric_api.MeshMetricResourceList,
	meshTraces *meshtrace_api.MeshTraceResourceList,
	meshAccessLogs *meshaccesslog_api.MeshAccessLogResourceList,
	backendsByMesh map[string]*backendNameIndex,
) map[string]int {
	counts := map[string]int{}

	for _, mm := range meshMetrics.Items {
		mesh := mm.GetMeta().GetMesh()
		for _, backend := range pointer.Deref(mm.Spec.Default.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				if resolved, ok := resolveBackendRef(backendsByMesh, mesh, backend.OpenTelemetry.BackendRef.Name); ok {
					counts[mesh+"/"+resolved]++
				}
			}
		}
	}

	for _, mt := range meshTraces.Items {
		mesh := mt.GetMeta().GetMesh()
		for _, backend := range pointer.Deref(mt.Spec.Default.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				if resolved, ok := resolveBackendRef(backendsByMesh, mesh, backend.OpenTelemetry.BackendRef.Name); ok {
					counts[mesh+"/"+resolved]++
				}
			}
		}
	}

	for _, mal := range meshAccessLogs.Items {
		mesh := mal.GetMeta().GetMesh()
		collectAccessLogBackendRefs(mesh, mal.Spec, counts, backendsByMesh)
	}

	return counts
}

// collectAccessLogBackendRefs extracts backendRef names from all MeshAccessLog conf locations.
// Keys written into counts use "<mesh>/<name>" format.
func collectAccessLogBackendRefs(
	mesh string,
	spec *meshaccesslog_api.MeshAccessLog,
	counts map[string]int,
	backendsByMesh map[string]*backendNameIndex,
) {
	collectFromConf := func(conf meshaccesslog_api.Conf) {
		for _, backend := range pointer.Deref(conf.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				if resolved, ok := resolveBackendRef(backendsByMesh, mesh, backend.OpenTelemetry.BackendRef.Name); ok {
					counts[mesh+"/"+resolved]++
				}
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

type backendNameIndex struct {
	byName        map[string]struct{}
	byDisplayName map[string][]string
}

func buildBackendNameIndex(motbs []*motb_api.MeshOpenTelemetryBackendResource) map[string]*backendNameIndex {
	indexByMesh := map[string]*backendNameIndex{}
	for _, motb := range motbs {
		mesh := motb.GetMeta().GetMesh()
		if indexByMesh[mesh] == nil {
			indexByMesh[mesh] = &backendNameIndex{
				byName:        map[string]struct{}{},
				byDisplayName: map[string][]string{},
			}
		}

		name := motb.GetMeta().GetName()
		indexByMesh[mesh].byName[name] = struct{}{}

		displayName := motb.GetMeta().GetLabels()[mesh_proto.DisplayName]
		if displayName != "" && displayName != name {
			indexByMesh[mesh].byDisplayName[displayName] = append(indexByMesh[mesh].byDisplayName[displayName], name)
		}
	}

	return indexByMesh
}

func resolveBackendRef(indexByMesh map[string]*backendNameIndex, mesh string, refName string) (string, bool) {
	idx := indexByMesh[mesh]
	if idx == nil {
		return "", false
	}

	if _, exists := idx.byName[refName]; exists {
		return refName, true
	}

	matchingDisplayNames := idx.byDisplayName[refName]
	if len(matchingDisplayNames) == 1 {
		return matchingDisplayNames[0], true
	}

	return "", false
}

func unresolvedBackendRefs(mesh string, refs []string, indexByMesh map[string]*backendNameIndex) []string {
	if len(refs) == 0 {
		return nil
	}

	unresolved := map[string]struct{}{}
	for _, ref := range refs {
		if _, ok := resolveBackendRef(indexByMesh, mesh, ref); !ok {
			unresolved[ref] = struct{}{}
		}
	}

	res := make([]string, 0, len(unresolved))
	for ref := range unresolved {
		res = append(res, ref)
	}
	sort.Strings(res)

	return res
}

func buildBackendRefsResolvedCondition(
	refs []string,
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

func meshMetricBackendRefs(mm *meshmetric_api.MeshMetricResource) []string {
	refs := map[string]struct{}{}
	for _, backend := range pointer.Deref(mm.Spec.Default.Backends) {
		if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
			refs[backend.OpenTelemetry.BackendRef.Name] = struct{}{}
		}
	}

	result := make([]string, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	sort.Strings(result)

	return result
}

func meshTraceBackendRefs(mt *meshtrace_api.MeshTraceResource) []string {
	refs := map[string]struct{}{}
	for _, backend := range pointer.Deref(mt.Spec.Default.Backends) {
		if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
			refs[backend.OpenTelemetry.BackendRef.Name] = struct{}{}
		}
	}

	result := make([]string, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	sort.Strings(result)

	return result
}

func meshAccessLogBackendRefs(mal *meshaccesslog_api.MeshAccessLogResource) []string {
	refs := map[string]struct{}{}
	collectFromConf := func(conf meshaccesslog_api.Conf) {
		for _, backend := range pointer.Deref(conf.Backends) {
			if backend.OpenTelemetry != nil && backend.OpenTelemetry.BackendRef != nil {
				refs[backend.OpenTelemetry.BackendRef.Name] = struct{}{}
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

	result := make([]string, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	sort.Strings(result)

	return result
}

func (s *StatusUpdater) updateMeshMetricStatuses(
	ctx context.Context,
	meshMetrics *meshmetric_api.MeshMetricResourceList,
	indexByMesh map[string]*backendNameIndex,
) error {
	for _, mm := range meshMetrics.Items {
		refs := meshMetricBackendRefs(mm)
		unresolved := unresolvedBackendRefs(mm.GetMeta().GetMesh(), refs, indexByMesh)
		condition := buildBackendRefsResolvedCondition(
			refs,
			unresolved,
			meshmetric_api.BackendRefsResolvedCondition,
			meshmetric_api.AllBackendRefsResolvedReason,
			meshmetric_api.UnresolvedBackendRefsReason,
		)

		if mm.Status == nil {
			mm.Status = &meshmetric_api.MeshMetricStatus{}
		}
		if conditionEquals(mm.Status.Conditions, condition) {
			continue
		}

		mm.Status.Conditions = updateConditions(mm.Status.Conditions, condition)
		log := s.logger.WithValues("meshmetric", mm.GetMeta().GetName(), "mesh", mm.GetMeta().GetMesh())
		if err := s.resManager.Update(ctx, mm); err != nil {
			if store.IsConflict(err) {
				log.Info("couldn't update MeshMetric, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
			} else {
				log.Error(err, "could not update MeshMetric status")
			}
		}
	}

	return nil
}

func (s *StatusUpdater) updateMeshTraceStatuses(
	ctx context.Context,
	meshTraces *meshtrace_api.MeshTraceResourceList,
	indexByMesh map[string]*backendNameIndex,
) error {
	for _, mt := range meshTraces.Items {
		refs := meshTraceBackendRefs(mt)
		unresolved := unresolvedBackendRefs(mt.GetMeta().GetMesh(), refs, indexByMesh)
		condition := buildBackendRefsResolvedCondition(
			refs,
			unresolved,
			meshtrace_api.BackendRefsResolvedCondition,
			meshtrace_api.AllBackendRefsResolvedReason,
			meshtrace_api.UnresolvedBackendRefsReason,
		)

		if mt.Status == nil {
			mt.Status = &meshtrace_api.MeshTraceStatus{}
		}
		if conditionEquals(mt.Status.Conditions, condition) {
			continue
		}

		mt.Status.Conditions = updateConditions(mt.Status.Conditions, condition)
		log := s.logger.WithValues("meshtrace", mt.GetMeta().GetName(), "mesh", mt.GetMeta().GetMesh())
		if err := s.resManager.Update(ctx, mt); err != nil {
			if store.IsConflict(err) {
				log.Info("couldn't update MeshTrace, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
			} else {
				log.Error(err, "could not update MeshTrace status")
			}
		}
	}

	return nil
}

func (s *StatusUpdater) updateMeshAccessLogStatuses(
	ctx context.Context,
	meshAccessLogs *meshaccesslog_api.MeshAccessLogResourceList,
	indexByMesh map[string]*backendNameIndex,
) error {
	for _, mal := range meshAccessLogs.Items {
		refs := meshAccessLogBackendRefs(mal)
		unresolved := unresolvedBackendRefs(mal.GetMeta().GetMesh(), refs, indexByMesh)
		condition := buildBackendRefsResolvedCondition(
			refs,
			unresolved,
			meshaccesslog_api.BackendRefsResolvedCondition,
			meshaccesslog_api.AllBackendRefsResolvedReason,
			meshaccesslog_api.UnresolvedBackendRefsReason,
		)

		if mal.Status == nil {
			mal.Status = &meshaccesslog_api.MeshAccessLogStatus{}
		}
		if conditionEquals(mal.Status.Conditions, condition) {
			continue
		}

		mal.Status.Conditions = updateConditions(mal.Status.Conditions, condition)
		log := s.logger.WithValues("meshaccesslog", mal.GetMeta().GetName(), "mesh", mal.GetMeta().GetMesh())
		if err := s.resManager.Update(ctx, mal); err != nil {
			if store.IsConflict(err) {
				log.Info("couldn't update MeshAccessLog, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
			} else {
				log.Error(err, "could not update MeshAccessLog status")
			}
		}
	}

	return nil
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

func buildBackendRuntimeSummaries(insights []*core_mesh.DataplaneInsightResource) map[string]backendRuntimeSummary {
	summaries := map[string]backendRuntimeSummary{}

	for _, insight := range insights {
		if insight == nil || insight.Spec == nil || !insight.Spec.IsOnline() || insight.Spec.GetOpenTelemetry() == nil {
			continue
		}

		mesh := insight.GetMeta().GetMesh()
		for _, backend := range insight.Spec.GetOpenTelemetry().GetBackends() {
			if backend == nil {
				continue
			}

			reported, ready, blocked, missingRequiredEnv, ambiguous := summarizeBackendRuntime(backend)
			if !reported {
				continue
			}

			key := mesh + "/" + backend.GetName()
			summary := summaries[key]
			summary.reportingDataplanes++
			if ready {
				summary.readyDataplanes++
			}
			if blocked {
				summary.blockedDataplanes++
			}
			if missingRequiredEnv {
				summary.missingRequiredEnvDataplanes++
			}
			if ambiguous {
				summary.ambiguousDataplanes++
			}
			summaries[key] = summary
		}
	}

	return summaries
}

func summarizeBackendRuntime(backend *mesh_proto.DataplaneInsight_OpenTelemetry_Backend) (bool, bool, bool, bool, bool) {
	signals := []*mesh_proto.DataplaneInsight_OpenTelemetry_Signal{
		backend.GetTraces(),
		backend.GetLogs(),
		backend.GetMetrics(),
	}

	reported := false
	ready := true
	blocked := false
	missingRequiredEnv := false
	ambiguous := false

	for _, signal := range signals {
		if signal == nil || !signal.GetEnabled() {
			continue
		}

		reported = true

		switch signal.GetState() {
		case otelSignalStateReady:
			// Soft blocks produce state "ready" but should still count as blocked
			reasons := signal.GetBlockedReasons()
			if slices.Contains(reasons, core_xds.OtelBlockedReasonEnvDisabledByPlatform) ||
				slices.Contains(reasons, core_xds.OtelBlockedReasonEnvDisabledByPolicy) ||
				slices.Contains(reasons, core_xds.OtelBlockedReasonSignalOverridesBlocked) {
				blocked = true
			}
		case otelSignalStateAmbiguous:
			ready = false
			ambiguous = true
		case otelSignalStateBlocked:
			ready = false
			blocked = true
		case otelSignalStateMissing:
			ready = false
			if slices.Contains(signal.GetBlockedReasons(), core_xds.OtelBlockedReasonRequiredEnvMissing) {
				missingRequiredEnv = true
			}
		default:
			ready = false
		}
	}

	return reported, ready, blocked, missingRequiredEnv, ambiguous
}

func buildReadyCondition(summary backendRuntimeSummary) common_api.Condition {
	if summary.reportingDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.ReadyCondition,
			Status:  kube_meta.ConditionUnknown,
			Reason:  motb_api.NoDataplaneReportsReason,
			Message: "No online dataplane has reported OTEL runtime status for this backend",
		}
	}

	if summary.readyDataplanes == summary.reportingDataplanes {
		return common_api.Condition{
			Type:    motb_api.ReadyCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  motb_api.AllReportingDataplanesReadyReason,
			Message: fmt.Sprintf("All %d reporting dataplane(s) are ready", summary.reportingDataplanes),
		}
	}

	return common_api.Condition{
		Type:    motb_api.ReadyCondition,
		Status:  kube_meta.ConditionFalse,
		Reason:  motb_api.SomeReportingDataplanesNotReadyReason,
		Message: fmt.Sprintf("%d of %d reporting dataplane(s) are ready", summary.readyDataplanes, summary.reportingDataplanes),
	}
}

func buildBlockedCondition(summary backendRuntimeSummary) common_api.Condition {
	if summary.reportingDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.DataplanesBlockedCondition,
			Status:  kube_meta.ConditionUnknown,
			Reason:  motb_api.NoDataplaneReportsReason,
			Message: "No online dataplane has reported OTEL runtime status for this backend",
		}
	}

	if summary.blockedDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.DataplanesBlockedCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NoReportingDataplanesBlockedReason,
			Message: "No reporting dataplanes are blocked by OTEL env policy",
		}
	}

	return common_api.Condition{
		Type:    motb_api.DataplanesBlockedCondition,
		Status:  kube_meta.ConditionTrue,
		Reason:  motb_api.SomeReportingDataplanesBlockedReason,
		Message: fmt.Sprintf("%d reporting dataplane(s) are blocked by OTEL env policy", summary.blockedDataplanes),
	}
}

func buildMissingRequiredEnvCondition(summary backendRuntimeSummary) common_api.Condition {
	if summary.reportingDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.DataplanesMissingRequiredEnvCondition,
			Status:  kube_meta.ConditionUnknown,
			Reason:  motb_api.NoDataplaneReportsReason,
			Message: "No online dataplane has reported OTEL runtime status for this backend",
		}
	}

	if summary.missingRequiredEnvDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.DataplanesMissingRequiredEnvCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NoReportingDataplanesMissingRequiredEnvReason,
			Message: "No reporting dataplanes are missing required OTEL env input",
		}
	}

	return common_api.Condition{
		Type:    motb_api.DataplanesMissingRequiredEnvCondition,
		Status:  kube_meta.ConditionTrue,
		Reason:  motb_api.SomeReportingDataplanesMissingRequiredEnvReason,
		Message: fmt.Sprintf("%d reporting dataplane(s) are missing required OTEL env input", summary.missingRequiredEnvDataplanes),
	}
}

func buildAmbiguousCondition(summary backendRuntimeSummary) common_api.Condition {
	if summary.reportingDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.DataplanesAmbiguousCondition,
			Status:  kube_meta.ConditionUnknown,
			Reason:  motb_api.NoDataplaneReportsReason,
			Message: "No online dataplane has reported OTEL runtime status for this backend",
		}
	}

	if summary.ambiguousDataplanes == 0 {
		return common_api.Condition{
			Type:    motb_api.DataplanesAmbiguousCondition,
			Status:  kube_meta.ConditionFalse,
			Reason:  motb_api.NoReportingDataplanesAmbiguousReason,
			Message: "No reporting dataplanes are ambiguous",
		}
	}

	return common_api.Condition{
		Type:    motb_api.DataplanesAmbiguousCondition,
		Status:  kube_meta.ConditionTrue,
		Reason:  motb_api.SomeReportingDataplanesAmbiguousReason,
		Message: fmt.Sprintf("%d reporting dataplane(s) are ambiguous", summary.ambiguousDataplanes),
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
