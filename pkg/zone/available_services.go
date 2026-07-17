package zone

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/ingress"
)

type ZoneAvailableServicesTracker struct {
	logger            logr.Logger
	metric            prometheus.Histogram
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
	metric := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "component_zone_available_services",
		Help: "Summary of available services tracker component interval",
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

func (t *ZoneAvailableServicesTracker) updateZoneIngresses(ctx context.Context) error {
	zis := core_mesh.ZoneIngressResourceList{}
	if err := t.resManager.List(ctx, &zis); err != nil {
		return err
	}

	aggregatedMeshCtxs, err := xds_context.AggregateMeshContexts(ctx, t.resManager, t.meshCache.GetMeshContext)
	if err != nil {
		return err
	}
	// kuma.io/service based dataplane services are represented by MeshService now that
	// meshServices.mode is always Exclusive, so skip every mesh's legacy AvailableServices.
	skipAvailableServices := map[xds.MeshName]struct{}{}
	for mesh := range aggregatedMeshCtxs.MeshContextsByName {
		skipAvailableServices[mesh] = struct{}{}
	}
	availableServices := ingress.GetAvailableServices(
		skipAvailableServices,
		aggregatedMeshCtxs.AllDataplanes(),
		nil,
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
