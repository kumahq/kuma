package status

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	util_time "github.com/kumahq/kuma/v2/pkg/util/time"
)

// StatusUpdater periodically updates Workload resource status
// based on associated DataplaneResource and DataplaneInsightResource.
type StatusUpdater struct {
	roResManager manager.ReadOnlyResourceManager
	resManager   manager.ResourceManager
	logger       logr.Logger
	metric       prometheus.Summary
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
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_workload_status_updater",
		Help:       "Summary of Workload status updater component interval",
		Objectives: core_metrics.DefaultObjectives,
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
				s.logger.Error(err, "could not update status of workloads")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	workloadList := &workload_api.WorkloadResourceList{}
	if err := s.roResManager.List(ctx, workloadList); err != nil {
		return errors.Wrap(err, "could not list Workloads")
	}
	if len(workloadList.Items) == 0 {
		// skip if no workloads
		return nil
	}

	dpInsightsList := core_mesh.DataplaneInsightResourceList{}
	if err := s.roResManager.List(ctx, &dpInsightsList); err != nil {
		return errors.Wrap(err, "could not list DataplaneInsights")
	}
	insightsByKey := core_model.IndexByKey(dpInsightsList.Items)

	allDpList := core_mesh.DataplaneResourceList{}
	if err := s.roResManager.List(ctx, &allDpList); err != nil {
		return errors.Wrap(err, "could not list Dataplanes")
	}

	dpsByMeshAndWorkload := make(map[string]map[string][]*core_mesh.DataplaneResource)
	for _, dp := range allDpList.Items {
		workloadLabel, ok := dp.GetMeta().GetLabels()[metadata.KumaWorkload]
		if !ok || workloadLabel == "" {
			continue
		}
		mesh := dp.GetMeta().GetMesh()
		if dpsByMeshAndWorkload[mesh] == nil {
			dpsByMeshAndWorkload[mesh] = make(map[string][]*core_mesh.DataplaneResource)
		}
		dpsByMeshAndWorkload[mesh][workloadLabel] = append(dpsByMeshAndWorkload[mesh][workloadLabel], dp)
	}

	for _, workload := range workloadList.Items {
		if !workload.IsLocalWorkload() {
			// status is already computed by the other zone
			continue
		}

		log := s.logger.WithValues("workload", workload.GetMeta().GetName(), "mesh", workload.GetMeta().GetMesh())

		workloadIdentifier := extractWorkloadIdentifier(workload.GetMeta().GetName())

		var matchingDps []*core_mesh.DataplaneResource
		if meshDps, ok := dpsByMeshAndWorkload[workload.GetMeta().GetMesh()]; ok {
			matchingDps = meshDps[workloadIdentifier]
		}

		var changeReasons []string

		dataplaneProxies := buildDataplaneProxies(matchingDps, insightsByKey)
		if !reflect.DeepEqual(workload.Status.DataplaneProxies, dataplaneProxies) {
			changeReasons = append(changeReasons, "data plane proxies")
			workload.Status.DataplaneProxies = dataplaneProxies
		}

		if len(changeReasons) > 0 {
			log.Info("updating workload", "reason", changeReasons)
			if err := s.resManager.Update(ctx, workload); err != nil {
				if store.IsConflict(err) {
					log.Info("couldn't update workload, will try again in next interval")
				} else {
					log.Error(err, "could not update workload")
				}
			}
		}
	}
	return nil
}

func extractWorkloadIdentifier(fullName string) string {
	// For Kubernetes workloads, name format is "{name}.{namespace}"
	// but kuma.io/workload label on dataplanes is just "{name}"
	if idx := strings.Index(fullName, "."); idx > 0 {
		return fullName[:idx]
	}
	return fullName
}

func buildDataplaneProxies(
	dataplanes []*core_mesh.DataplaneResource,
	insightsByKey map[core_model.ResourceKey]*core_mesh.DataplaneInsightResource,
) workload_api.DataplaneProxies {
	result := workload_api.DataplaneProxies{}
	for _, dp := range dataplanes {
		result.Total++
		if insight := insightsByKey[core_model.MetaToResourceKey(dp.Meta)]; insight != nil {
			if insight.Spec.IsOnline() {
				result.Connected++
			}
		}
		// A dataplane is healthy if it has at least one healthy inbound
		if len(dp.Spec.GetNetworking().GetHealthyInbounds()) > 0 {
			result.Healthy++
		}
	}
	return result
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}
