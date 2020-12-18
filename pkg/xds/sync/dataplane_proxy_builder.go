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

func (p *dataplaneProxyBuilder) Build(key core_model.ResourceKey, streamId int64) (*xds.Proxy, *xds_context.MeshContext, error) {
	ctx := context.Background()
	dataplane := core_mesh.NewDataplaneResource()
	proxyID := xds.FromResourceKey(key)

	if err := p.ResManager.Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
		return nil, nil, err
	}

	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return nil, nil, err
	}
	dataplane = resolvedDp

	mesh := core_mesh.NewMeshResource()
	if err := p.ResManager.Get(ctx, mesh, core_store.GetByKey(proxyID.Mesh, core_model.NoMesh)); err != nil {
		return nil, nil, err
	}

	dataplanes, err := xds_topology.GetDataplanes(syncLog, ctx, p.ResManager, p.LookupIP, dataplane.Meta.GetMesh())
	if err != nil {
		return nil, nil, err
	}
	externalServices := &core_mesh.ExternalServiceResourceList{}
	if err := p.ResManager.List(ctx, externalServices, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, nil, err
	}

	// pick a single the most specific route for each outbound interface
	routes, err := xds_topology.GetRoutes(ctx, dataplane, p.ResManager)
	if err != nil {
		return nil, nil, err
	}

	// create creates a map of selectors to match other dataplanes reachable via given routes
	destinations := xds_topology.BuildDestinationMap(dataplane, routes)

	// resolve all endpoints that match given selectors
	outbound := xds_topology.BuildEndpointMap(
		mesh, p.Zone,
		dataplanes.Items, externalServices.Items, p.DataSourceLoader)

	healthChecks, err := xds_topology.GetHealthChecks(ctx, dataplane, destinations, p.ResManager)
	if err != nil {
		return nil, nil, err
	}

	circuitBreakers, err := xds_topology.GetCircuitBreakers(ctx, dataplane, destinations, p.ResManager)
	if err != nil {
		return nil, nil, err
	}

	trafficTrace, err := xds_topology.GetTrafficTrace(ctx, dataplane, p.ResManager)
	if err != nil {
		return nil, nil, err
	}
	var tracingBackend *mesh_proto.TracingBackend
	if trafficTrace != nil {
		tracingBackend = mesh.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
	}

	matchedPermissions, err := p.PermissionMatcher.Match(ctx, dataplane, mesh)
	if err != nil {
		return nil, nil, err
	}

	matchedLogs, err := p.LogsMatcher.Match(ctx, dataplane)
	if err != nil {
		return nil, nil, err
	}

	faultInjection, err := p.FaultInjectionMatcher.Match(ctx, dataplane, mesh)
	if err != nil {
		return nil, nil, err
	}

	meshCtx := xds_context.MeshContext{
		Resource:   mesh,
		Dataplanes: dataplanes,
	}
	proxy := xds.Proxy{
		Id:                 proxyID,
		Dataplane:          dataplane,
		TrafficPermissions: matchedPermissions,
		TrafficRoutes:      routes,
		OutboundSelectors:  destinations,
		OutboundTargets:    outbound,
		HealthChecks:       healthChecks,
		CircuitBreakers:    circuitBreakers,
		Logs:               matchedLogs,
		TrafficTrace:       trafficTrace,
		TracingBackend:     tracingBackend,
		FaultInjections:    faultInjection,
		Metadata:           p.MetadataTracker.Metadata(streamId),
	}
	return &proxy, &meshCtx, nil
}
