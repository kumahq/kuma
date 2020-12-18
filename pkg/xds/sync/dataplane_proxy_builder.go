package sync

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var syncLog = core.Log.WithName("sync")

type dataplaneProxyBuilder struct {
	ResManager            manager.ReadOnlyResourceManager
	LookupIP              lookup.LookupIPFunc
	DataSourceLoader      datasource.Loader
	MetadataTracker       DataplaneMetadataTracker
	PermissionMatcher     permissions.TrafficPermissionsMatcher
	LogsMatcher           logs.TrafficLogsMatcher
	FaultInjectionMatcher faultinjections.FaultInjectionMatcher

	Zone string
}

func (p *dataplaneProxyBuilder) build(key core_model.ResourceKey, streamId int64) (*xds.Proxy, *xds_context.MeshContext, error) {
	ctx := context.Background()
	proxy := &xds.Proxy{
		Id: xds.FromResourceKey(key),
		Metadata: p.MetadataTracker.Metadata(streamId),
	}

	if err := p.resolveDataplane(ctx, key, proxy); err != nil {
		return nil, nil, err
	}

	meshCtx, err := p.buildMeshContext(ctx, key.Mesh)
	if err != nil {
		return nil, nil, err
	}

	if err := p.resolveRouting(ctx, meshCtx, proxy); err != nil {
		return nil, nil, err
	}

	if err := p.matchPolicies(ctx, meshCtx, proxy); err != nil {
		return nil, nil, err
	}

	return proxy, meshCtx, nil
}

func (p *dataplaneProxyBuilder) resolveDataplane(ctx context.Context, key core_model.ResourceKey, proxy *xds.Proxy) error {
	dataplane := core_mesh.NewDataplaneResource()

	// we use non-cached ResourceManager to always fetch fresh version of the Dataplane.
	// Otherwise, technically MeshCache can use newer version because it uses List operation instead of Get
	if err := p.ResManager.Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
		return err
	}

	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return err
	}
	dataplane = resolvedDp

	proxy.Dataplane = resolvedDp
	return nil
}

func (p *dataplaneProxyBuilder) buildMeshContext(ctx context.Context, meshName string) (*xds_context.MeshContext, error) {
	mesh := core_mesh.NewMeshResource()
	if err := p.ResManager.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
		return nil, err
	}

	dataplanes, err := xds_topology.GetDataplanes(syncLog, ctx, p.ResManager, p.LookupIP, meshName)
	if err != nil {
		return nil, err
	}
	meshCtx := xds_context.MeshContext{
		Resource:   mesh,
		Dataplanes: dataplanes,
	}
	return &meshCtx, nil
}

func (p *dataplaneProxyBuilder) resolveRouting(ctx context.Context, meshContext *xds_context.MeshContext, proxy *xds.Proxy) error {
	externalServices := &core_mesh.ExternalServiceResourceList{}
	if err := p.ResManager.List(ctx, externalServices, core_store.ListByMesh(proxy.Dataplane.Meta.GetMesh())); err != nil {
		return err
	}

	// pick a single the most specific route for each outbound interface
	routes, err := xds_topology.GetRoutes(ctx, proxy.Dataplane, p.ResManager)
	if err != nil {
		return err
	}

	// create creates a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(proxy.Dataplane, routes)
	proxy.OutboundSelectors = destinations

	// resolve all endpoints that match given selectors
	outbound := xds_topology.BuildEndpointMap(meshContext.Resource, p.Zone, meshContext.Dataplanes.Items, externalServices.Items, p.DataSourceLoader)
	proxy.OutboundTargets = outbound

	return nil
}

func (p *dataplaneProxyBuilder) matchPolicies(ctx context.Context, meshContext *xds_context.MeshContext, proxy *xds.Proxy) error {
	healthChecks, err := xds_topology.GetHealthChecks(ctx, proxy.Dataplane, proxy.OutboundSelectors, p.ResManager)
	if err != nil {
		return err
	}
	proxy.HealthChecks = healthChecks

	circuitBreakers, err := xds_topology.GetCircuitBreakers(ctx, proxy.Dataplane, proxy.OutboundSelectors, p.ResManager)
	if err != nil {
		return err
	}
	proxy.CircuitBreakers = circuitBreakers

	trafficTrace, err := xds_topology.GetTrafficTrace(ctx, proxy.Dataplane, p.ResManager)
	if err != nil {
		return err
	}
	var tracingBackend *mesh_proto.TracingBackend
	if trafficTrace != nil {
		tracingBackend = meshContext.Resource.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
	}
	proxy.TracingBackend = tracingBackend
	proxy.TrafficTrace = trafficTrace

	matchedPermissions, err := p.PermissionMatcher.Match(ctx, proxy.Dataplane, meshContext.Resource)
	if err != nil {
		return err
	}
	proxy.TrafficPermissions = matchedPermissions

	matchedLogs, err := p.LogsMatcher.Match(ctx, proxy.Dataplane)
	if err != nil {
		return err
	}
	proxy.Logs = matchedLogs

	faultInjection, err := p.FaultInjectionMatcher.Match(ctx, proxy.Dataplane, meshContext.Resource)
	if err != nil {
		return err
	}
	proxy.FaultInjections = faultInjection

	return nil
}
