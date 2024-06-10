package zone

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/ingress"
)

type ZoneAvailableServicesTracker struct {
	logger            logr.Logger
	metric            prometheus.Summary
	resManager        manager.ResourceManager
	interval          time.Duration
	ingressTagFilters []string
	zone              string
}

func NewZoneAvailableServicesTracker(
	logger logr.Logger,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	interval time.Duration,
	ingressTagFilters []string,
	zone string,
) (*ZoneAvailableServicesTracker, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_zone_available_services",
		Help:       "Summary of available services tracker component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &ZoneAvailableServicesTracker{
		logger:            logger,
		metric:            metric,
		resManager:        resManager,
		interval:          interval,
		ingressTagFilters: ingressTagFilters,
		zone:              zone,
	}, nil
}

func (t *ZoneAvailableServicesTracker) Start(stop <-chan struct{}) error {
	t.logger.Info("starting")
	ticker := time.NewTicker(t.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := t.updateZoneIngresses(ctx); err != nil {
				t.logger.Error(err, "couldn't update ZoneIngress available services")
			}
			t.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			t.logger.Info("stopping")
			return nil
		}
	}
}

func (t *ZoneAvailableServicesTracker) NeedLeaderElection() bool {
	return true
}

func availableServicesEqual(services []*mesh_proto.ZoneIngress_AvailableService, other []*mesh_proto.ZoneIngress_AvailableService) bool {
	if len(services) != len(other) {
		return false
	}
	for i := range services {
		if !proto.Equal(services[i], other[i]) {
			return false
		}
	}
	return true
}

func (t *ZoneAvailableServicesTracker) getIngressExternalServices(
	ctx context.Context,
	meshes []*core_mesh.MeshResource,
) (core_mesh.ExternalServiceResourceList, error) {
	var externalServices []*core_mesh.ExternalServiceResource
	allMeshExternalServices := core_mesh.ExternalServiceResourceList{}

	for _, mesh := range meshes {
		if !mesh.ZoneEgressEnabled() {
			continue
		}

		ess := core_mesh.ExternalServiceResourceList{}
		if err := t.resManager.List(ctx, &ess, store.ListByMesh(mesh.GetMeta().GetName())); err != nil {
			return core_mesh.ExternalServiceResourceList{}, err
		}

		// look for external services that are only available in my zone and expose them
		for _, es := range ess.Items {
			if es.Spec.Tags[mesh_proto.ZoneTag] == t.zone {
				externalServices = append(externalServices, es)
			}
		}
	}

	allMeshExternalServices.Items = externalServices
	return allMeshExternalServices, nil
}

func (t *ZoneAvailableServicesTracker) updateZoneIngresses(ctx context.Context) error {
	meshes := core_mesh.MeshResourceList{}
	if err := t.resManager.List(ctx, &meshes); err != nil {
		return err
	}

	dps := core_mesh.DataplaneResourceList{}
	if err := t.resManager.List(ctx, &dps); err != nil {
		return err
	}

	mgws := core_mesh.MeshGatewayResourceList{}
	if err := t.resManager.List(ctx, &mgws); err != nil {
		return err
	}

	ess, err := t.getIngressExternalServices(ctx, meshes.Items)
	if err != nil {
		return err
	}

	zis := core_mesh.ZoneIngressResourceList{}
	if err := t.resManager.List(ctx, &zis); err != nil {
		return err
	}

	availableServices := ingress.GetAvailableServices(
		dps.Items,
		mgws.Items,
		ess.Items,
		t.ingressTagFilters,
	)
	for _, zi := range zis.Items {
		// Initially zone is empty
		if zi.Spec.Zone != "" && zi.Spec.Zone != t.zone {
			continue
		}
		if availableServicesEqual(availableServices, zi.Spec.GetAvailableServices()) {
			continue
		}
		zi.Spec.AvailableServices = availableServices
		if err := t.resManager.Update(ctx, zi); err != nil {
			t.logger.Error(err, "couldn't update ZoneIngress", "name", zi.GetMeta().GetName())
		}
	}
	return nil
}
