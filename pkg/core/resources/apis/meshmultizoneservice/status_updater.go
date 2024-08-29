package meshmultizoneservice

import (
	"context"
	"reflect"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

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
		Name:       "component_mzms_status_updater",
		Help:       "Summary of MeshMultizoneService Updater component",
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
	s.logger.Info("starting")
	ticker := time.NewTicker(s.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := s.updateStatus(ctx); err != nil {
				s.logger.Error(err, "could not update status of mesh multizone services")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	mzSvcList := &meshmzservice_api.MeshMultiZoneServiceResourceList{}
	if err := s.roResManager.List(ctx, mzSvcList); err != nil {
		return errors.Wrap(err, "could not list of MeshMultiZoneServices")
	}
	if len(mzSvcList.Items) == 0 {
		// skip fetching other resources if MeshMultiZoneService is not used
		return nil
	}

	msList := &meshservice_api.MeshServiceResourceList{}
	if err := s.roResManager.List(ctx, msList); err != nil {
		return errors.Wrap(err, "could not list of MeshServices")
	}

	for _, mzSvc := range mzSvcList.Items {
		var matched []meshmzservice_api.MatchedMeshService
		ports := map[uint32]meshservice_api.Port{}
		for _, svc := range msList.Items {
			if matchesService(mzSvc, svc) {
				ri := model.NewResourceIdentifier(svc)
				matched = append(matched, meshmzservice_api.MatchedMeshService{
					Name:      ri.Name,
					Namespace: ri.Namespace,
					Zone:      ri.Zone,
					Mesh:      ri.Mesh,
				})
				for _, port := range svc.Spec.Ports {
					ports[port.Port] = port
				}
			}
		}

		sort.Slice(matched, func(i, j int) bool {
			return matched[i].Name < matched[j].Name
		})

		if !reflect.DeepEqual(mzSvc.Status.MeshServices, matched) {
			log := s.logger.WithValues("meshmultizoneservice", mzSvc.Meta.GetName())
			mzSvc.Status.MeshServices = matched
			log.Info("updating matched mesh services", "matchedMeshServices", matched)
			if err := s.resManager.Update(ctx, mzSvc); err != nil {
				log.Error(err, "could not update ports and mesh services")
			}
		}
	}
	return nil
}

func matchesService(mzSvc *meshmzservice_api.MeshMultiZoneServiceResource, svc *meshservice_api.MeshServiceResource) bool {
	for label, value := range mzSvc.Spec.Selector.MeshService.MatchLabels {
		if svc.Meta.GetLabels()[label] != value {
			return false
		}
	}
	return true
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}
