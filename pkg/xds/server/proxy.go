package server

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

func ToDataplaneProxy(dataplane *core_mesh.DataplaneResource, meta *xds.DataplaneMetadata, snapshot *MeshSnapshot, zone string) (*xds.Proxy, error) {
	proxyID := xds.FromResourceKey(model.MetaToResourceKey(dataplane.GetMeta()))

	// pick a single the most specific route for each outbound interface
	routeMap := topology.BuildRouteMap(dataplane,
		snapshot.Resources[core_mesh.TrafficRouteType].(*core_mesh.TrafficRouteResourceList).Items)

	// create creates a map of selectors to match other dataplanes reachable via given routes
	destinations := topology.BuildDestinationMap(dataplane, routeMap)

	// resolve all endpoints that match given selectors
	outbound, err := topology.GetOutboundTargets(destinations,
		snapshot.Resources[core_mesh.DataplaneType].(*core_mesh.DataplaneResourceList), zone, snapshot.Mesh)
	if err != nil {
		return nil, err
	}

	permissionMap, err := permissions.BuildTrafficPermissionMap(dataplane,
		snapshot.Mesh, snapshot.Resources[core_mesh.TrafficPermissionType].(*core_mesh.TrafficPermissionResourceList).Items)
	if err != nil {
		return nil, err
	}

	healthChecks := topology.BuildHealthCheckMap(dataplane,
		destinations, snapshot.Resources[core_mesh.HealthCheckType].(*core_mesh.HealthCheckResourceList).Items)

	circuitBreakers := topology.BuildCircuitBreakerMap(dataplane,
		destinations, snapshot.Resources[core_mesh.CircuitBreakerType].(*core_mesh.CircuitBreakerResourceList).Items)

	trafficLogs := logs.BuildTrafficLogMap(dataplane,
		snapshot.Mesh, snapshot.Resources[core_mesh.TrafficLogType].(*core_mesh.TrafficLogResourceList).Items)

	trafficTrace := topology.SelectTrafficTrace(dataplane,
		snapshot.Resources[core_mesh.TrafficTraceType].(*core_mesh.TrafficTraceResourceList).Items)

	var tracingBackend *mesh_proto.TracingBackend
	if trafficTrace != nil {
		tracingBackend = snapshot.Mesh.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
	}

	faultInjections, err := faultinjections.BuildFaultInjectionMap(dataplane,
		snapshot.Mesh, snapshot.Resources[core_mesh.FaultInjectionType].(*core_mesh.FaultInjectionResourceList).Items)
	if err != nil {
		return nil, err
	}

	return &xds.Proxy{
		Id:                 proxyID,
		Dataplane:          dataplane,
		TrafficPermissions: permissionMap,
		Logs:               trafficLogs,
		TrafficRoutes:      routeMap,
		OutboundSelectors:  destinations,
		OutboundTargets:    outbound,
		HealthChecks:       healthChecks,
		CircuitBreakers:    circuitBreakers,
		TrafficTrace:       trafficTrace,
		TracingBackend:     tracingBackend,
		FaultInjections:    faultInjections,
		Metadata:           meta,
	}, nil
}
