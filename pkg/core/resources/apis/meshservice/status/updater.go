package status

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/util/maps"
)

type StatusUpdater struct {
	roResManager manager.ReadOnlyResourceManager
	resManager   manager.ResourceManager
	logger       logr.Logger
	metric       prometheus.Summary
	interval     time.Duration
	localZone    string
}

var _ component.Component = &StatusUpdater{}

func NewStatusUpdater(
	logger logr.Logger,
	roResManager manager.ReadOnlyResourceManager,
	resManager manager.ResourceManager,
	interval time.Duration,
	metrics core_metrics.Metrics,
	localZone string,
) (component.Component, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_ms_status_updater",
		Help:       "Summary of Inter CP Heartbeat component interval",
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
		localZone:    localZone,
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
				s.logger.Error(err, "could not update status of mesh services")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	msList := &meshservice_api.MeshServiceResourceList{}
	if err := s.roResManager.List(ctx, msList); err != nil {
		return errors.Wrap(err, "could not list of Dataplanes")
	}
	if len(msList.Items) == 0 {
		// skip fetching other resources if MeshService is not used
		return nil
	}
	dpList := mesh.DataplaneResourceList{}
	if err := s.roResManager.List(ctx, &dpList); err != nil {
		return errors.Wrap(err, "could not list of Dataplanes")
	}

	dppsForMs := meshservice.MatchDataplanesWithMeshServices(dpList.Items, msList.Items, false)

	for ms, dpps := range dppsForMs {
		if !ms.IsLocalMeshService(s.localZone) {
			// identities are already computed by the other zone
			continue
		}
		serviceTagIdentities := map[string]struct{}{}
		for _, dpp := range dpps {
			for service := range dpp.Spec.TagSet()[mesh_proto.ServiceTag] {
				serviceTagIdentities[service] = struct{}{}
			}
		}
		var identites []meshservice_api.MeshServiceIdentity
		for _, identity := range maps.SortedKeys(serviceTagIdentities) {
			identites = append(identites, meshservice_api.MeshServiceIdentity{
				Type:  meshservice_api.MeshServiceIdentityServiceTagType,
				Value: identity,
			})
		}
		log := s.logger.WithValues("meshservice", ms.GetMeta().GetName(), "mesh", ms.GetMeta().GetMesh())
		if !reflect.DeepEqual(ms.Spec.Identities, identites) {
			ms.Spec.Identities = identites
			log.Info("updating identities for the service", "identities", identites)
			if err := s.resManager.Update(ctx, ms); err != nil {
				log.Error(err, "could not update identities")
			}
		}
	}
	return nil
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}
