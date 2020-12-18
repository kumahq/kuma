package sync

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var (
	xdsServerLog  = core.Log.WithName("xds-server")
	meshResources = meshResourceTypes(map[core_model.ResourceType]bool{
		core_mesh.DataplaneInsightType:  true,
		core_mesh.DataplaneOverviewType: true,
		core_mesh.ServiceInsightType:    true,
		core_system.ConfigType:          true,
	})
)

func meshResourceTypes(exclude map[core_model.ResourceType]bool) []core_model.ResourceType {
	types := []core_model.ResourceType{}
	for _, typ := range registry.Global().ListTypes() {
		r, err := registry.Global().NewObject(typ)
		if err != nil {
			panic(err)
		}
		if r.Scope() == core_model.ScopeMesh && !exclude[typ] {
			types = append(types, typ)
		}
	}
	return types
}

func defaultDataplaneProxyBuilder(rt core_runtime.Runtime, metadataTracker DataplaneMetadataTracker) *dataplaneProxyBuilder {
	return &dataplaneProxyBuilder{
		ResManager:            rt.ReadOnlyResourceManager(),
		LookupIP:              rt.LookupIP(),
		DataSourceLoader:      rt.DataSourceLoader(),
		MetadataTracker:       metadataTracker,
		PermissionMatcher:     permissions.TrafficPermissionsMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		LogsMatcher:           logs.TrafficLogsMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		FaultInjectionMatcher: faultinjections.FaultInjectionMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		Zone:                  rt.Config().Multizone.Remote.Zone,
	}
}

func defaultIngressProxyBuilder(rt core_runtime.Runtime, metadataTracker DataplaneMetadataTracker) *ingressProxyBuilder {
	return &ingressProxyBuilder{
		ResManager:         rt.ResourceManager(),
		ReadOnlyResManager: rt.ReadOnlyResourceManager(),
		LookupIP:           rt.LookupIP(),
		MetadataTracker:    metadataTracker,
	}
}

func DefaultDataplaneWatchdogFactory(
	rt core_runtime.Runtime,
	metadataTracker DataplaneMetadataTracker,
	connectionInfoTracker ConnectionInfoTracker,
	dataplaneReconciler SnapshotReconciler,
	ingressReconciler SnapshotReconciler,
) (DataplaneWatchdogFactory, error) {
	claCache, err := cla.NewCache(rt.ReadOnlyResourceManager(), rt.DataSourceLoader(), rt.Config().Multizone.Remote.Zone, rt.Config().Store.Cache.ExpirationTime, rt.LookupIP(), rt.Metrics())
	if err != nil {
		return nil, err
	}
	dataplaneProxyBuilder := defaultDataplaneProxyBuilder(rt, metadataTracker)
	ingressProxyBuilder := defaultIngressProxyBuilder(rt, metadataTracker)
	meshSnapshotCache, err := mesh.NewCache(rt.ReadOnlyResourceManager(), rt.Config().Store.Cache.ExpirationTime, meshResources, rt.LookupIP(), rt.Metrics())
	if err != nil {
		return nil, err
	}

	envoyCpCtx, err := xds_context.BuildControlPlaneContext(rt.Config())
	if err != nil {
		return nil, err
	}

	deps := DataplaneWatchdogDependencies{
		resManager:            rt.ResourceManager(),
		dataplaneProxyBuilder: dataplaneProxyBuilder,
		dataplaneReconciler:   dataplaneReconciler,
		ingressProxyBuilder:   ingressProxyBuilder,
		ingressReconciler:     ingressReconciler,
		connectionInfoTracker: connectionInfoTracker,
		envoyCpCtx:            envoyCpCtx,
		claCache:              claCache,
		meshCache:             meshSnapshotCache,
	}
	return NewDataplaneWatchdogFactory(
		rt.Metrics(),
		rt.Config().XdsServer.DataplaneConfigurationRefreshInterval,
		deps,
	)
}
