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
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/ingress"
)

type ZoneAvailableServicesTracker struct {
	logger            logr.Logger
	metric            prometheus.Summary
	resManager        manager.ResourceManager
	meshCache         *mesh.Cache
	interval          time.Duration
	ingressTagFilters []string
	zone              string
}

func NewZoneAvailableServicesTracker(
	logger logr.Logger,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	meshCache *mesh.Cache,
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
		meshCache:         meshCache,
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
				t.logger.Error(err, "couldn't update ZoneIngress available services, will retry")
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
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) []*core_mesh.ExternalServiceResource {
	var externalServices []*core_mesh.ExternalServiceResource

	for _, mesh := range aggregatedMeshCtxs.Meshes {
		if !mesh.ZoneEgressEnabled() || mesh.Spec.MeshServicesMode() == mesh_proto.Mesh_MeshServices_Exclusive {
			continue
		}

		meshCtx := aggregatedMeshCtxs.MustGetMeshContext(mesh.GetMeta().GetName())
		// look for external services that are only available in my zone and expose them
		for _, es := range meshCtx.Resources.ExternalServices().Items {
			if es.Spec.Tags[mesh_proto.ZoneTag] == t.zone {
				externalServices = append(externalServices, es)
			}
		}
	}

	return externalServices
}

func (t *ZoneAvailableServicesTracker) updateZoneIngresses(ctx context.Context) error {
	zis := core_mesh.ZoneIngressResourceList{}
	if err := t.resManager.List(ctx, &zis); err != nil {
		return err
	}

	var availableServices []*mesh_proto.ZoneIngress_AvailableService
	aggregatedMeshCtxs, err := xds_context.AggregateMeshContexts(ctx, t.resManager, t.meshCache.GetMeshContext)
	if err != nil {
		return err
	}
	skipAvailableServices := map[xds.MeshName]struct{}{}
	for mesh, meshCtx := range aggregatedMeshCtxs.MeshContextsByName {
		if meshCtx.Resource.Spec.MeshServicesMode() == mesh_proto.Mesh_MeshServices_Exclusive {
			skipAvailableServices[mesh] = struct{}{}
		}
	}
	availableServices = ingress.GetAvailableServices(
		skipAvailableServices,
		aggregatedMeshCtxs.AllDataplanes(),
		aggregatedMeshCtxs.AllMeshGateways(),
		t.getIngressExternalServices(aggregatedMeshCtxs),
		t.ingressTagFilters,
	)

	var names []string
	for _, zi := range zis.Items {
		// Zone is empty for resources in the local zone
		if zi.Spec.Zone != "" && zi.Spec.Zone != t.zone {
			continue
		}
		if availableServicesEqual(availableServices, zi.Spec.GetAvailableServices()) {
			continue
		}
		zi.Spec.AvailableServices = availableServices
		if err := t.resManager.Update(ctx, zi); err != nil {
			t.logger.Error(err, "couldn't update ZoneIngress, will retry", "name", zi.GetMeta().GetName())
		} else {
			names = append(names, zi.GetMeta().GetName())
		}
	}
	if len(names) > 0 {
		t.logger.Info("updated ZoneIngress available services", "ZoneIngresses", names)
	}
	return nil
}
