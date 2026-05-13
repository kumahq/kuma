package status

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	resource_status "github.com/kumahq/kuma/v2/pkg/core/resources/status"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	util_time "github.com/kumahq/kuma/v2/pkg/util/time"
)

// StatusUpdater keeps the MeshExternalService status in sync. Today its only
// job is the SNICompliant condition; more checks can be added here as
// MeshExternalService grows.
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
		Name: "component_mes_status_updater",
		Help: "Summary of MeshExternalService status updater run duration in ms",
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
	util_time.SleepUpTo(s.interval)
	s.logger.Info("starting")
	ticker := time.NewTicker(s.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := s.updateStatus(ctx); err != nil {
				s.logger.Error(err, "could not update status of mesh external services")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	mesList := &meshexternalservice_api.MeshExternalServiceResourceList{}
	if err := s.roResManager.List(ctx, mesList); err != nil {
		return errors.Wrap(err, "could not list MeshExternalServices")
	}
	for _, mes := range mesList.Items {
		log := s.logger.WithValues("meshexternalservice", mes.GetMeta().GetName(), "mesh", mes.GetMeta().GetMesh())
		sniCondition := resource_status.BuildSNICompliantCondition(kri.From(mes), mes.GetPorts())
		if resource_status.ConditionEquals(mes.Status.Conditions, sniCondition) {
			continue
		}
		mes.Status.Conditions = resource_status.UpdateConditions(mes.Status.Conditions, sniCondition)
		log.V(1).Info("updating mesh external service status", "reason", "sni compliance", "condition", sniCondition)
		if err := s.resManager.Update(ctx, mes); err != nil {
			if store.IsConflict(err) {
				log.Info("couldn't update mesh external service, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
			} else {
				log.Error(err, "could not update mesh external service status")
			}
		}
	}
	return nil
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}
