package server

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
	v2 "github.com/kumahq/kuma/pkg/xds/server/v2"
	v3 "github.com/kumahq/kuma/pkg/xds/server/v3"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	// HashMeshExcludedResources defines Mesh-scoped resources that are not used in XDS therefore when counting hash mesh we can skip them
	HashMeshExcludedResources = map[core_model.ResourceType]bool{
		core_mesh.DataplaneInsightType:  true,
		core_mesh.DataplaneOverviewType: true,
		core_mesh.ServiceInsightType:    true,
		core_system.ConfigType:          true,
	}
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

func RegisterXDS(rt core_runtime.Runtime, server *grpc.Server) error {
	// Build common dependencies for V2 and V3 servers.
	// We want to have same metrics (we cannot register one metric twice) and same caches for both V2 and V3.
	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds")
	if err != nil {
		return err
	}
	xdsMetrics, err := xds_metrics.NewMetrics(rt.Metrics())
	if err != nil {
		return err
	}
	meshSnapshotCache, err := mesh.NewCache(rt.ReadOnlyResourceManager(), rt.Config().Store.Cache.ExpirationTime, meshResourceTypes(HashMeshExcludedResources), rt.LookupIP(), rt.Metrics())
	if err != nil {
		return err
	}
	claCache, err := cla.NewCache(rt.ReadOnlyResourceManager(), rt.DataSourceLoader(), rt.Config().Multizone.Remote.Zone, rt.Config().Store.Cache.ExpirationTime, rt.LookupIP(), rt.Metrics())
	if err != nil {
		return err
	}
	envoyCpCtx, err := xds_context.BuildControlPlaneContext(rt.Config(), claCache)
	if err != nil {
		return err
	}

	if err := v2.RegisterXDS(statsCallbacks, xdsMetrics, meshSnapshotCache, envoyCpCtx, rt, server); err != nil {
		return errors.Wrap(err, "could not register V2 XDS")
	}
	if err := v3.RegisterXDS(statsCallbacks, xdsMetrics, meshSnapshotCache, envoyCpCtx, rt, server); err != nil {
		return errors.Wrap(err, "could not register V3 XDS")
	}
	return nil
}
