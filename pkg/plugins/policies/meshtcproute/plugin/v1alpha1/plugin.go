package v1alpha1

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshroute_gateway "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute/gateway"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(
	dataplane *core_mesh.DataplaneResource,
	resources xds_context.Resources,
	opts ...core_plugins.MatchedPoliciesOption,
) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTCPRouteType, dataplane, resources, opts...)
}

func (p plugin) Apply(
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) error {
	if proxy.Dataplane == nil {
		return nil
	}

	policies := proxy.Policies.Dynamic[api.MeshTCPRouteType]
	// Only fallback if we have TrafficRoutes & No MeshTCPRoutes
	if len(ctx.Mesh.Resources.TrafficRoutes().Items) > 0 && len(policies.ToRules.Rules) == 0 && len(policies.GatewayRules.ToRules.ByListenerAndHostname) == 0 {
		return nil
	}

	if err := ApplyToOutbounds(proxy, rs, ctx, policies); err != nil {
		return err
	}

	if err := ApplyToGateway(context.TODO(), proxy, rs, ctx, policies); err != nil {
		return err
	}

	return nil
}

func ApplyToOutbounds(
	proxy *core_xds.Proxy,
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	policies core_xds.TypedMatchingPolicies,
) error {
	tlsReady := ctx.Mesh.GetTLSReadiness()
	servicesAccumulator := envoy_common.NewServicesAccumulator(tlsReady)

	listeners, err := generateListeners(proxy, policies.ToRules, servicesAccumulator, ctx.Mesh)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAccumulator.Services()

	clusters, err := meshroute.GenerateClusters(proxy, ctx.Mesh, services, ctx.ControlPlane.SystemNamespace)
	if err != nil {
		return errors.Wrap(err, "couldn't generate cluster resources")
	}
	rs.AddSet(clusters)

	endpoints, err := meshroute.GenerateEndpoints(proxy, ctx, services)
	if err != nil {
		return errors.Wrap(err, "couldn't generate endpoint resources")
	}
	rs.AddSet(endpoints)

	return nil
}

func ApplyToGateway(
	ctx context.Context,
	proxy *core_xds.Proxy,
	resources *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	policies core_xds.TypedMatchingPolicies,
) error {
	if len(policies.GatewayRules.ToRules.ByListenerAndHostname) == 0 {
		return nil
	}

	var limits []plugin_gateway.RuntimeResoureLimitListener

	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := xdsCtx.Mesh.Resources.MeshLocalResources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}

	gateway := xds_topology.SelectGateway(gateways.Items, proxy.Dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}

	listeners := meshroute_gateway.CollectListenerInfos(
		ctx,
		xdsCtx.Mesh,
		gateway,
		proxy,
		policies.GatewayRules,
		[]mesh_proto.MeshGateway_Listener_Protocol{mesh_proto.MeshGateway_Listener_TCP, mesh_proto.MeshGateway_Listener_TLS},
		sortRulesToHosts,
	)

	plugin_gateway.SetGatewayListeners(proxy, listeners)

	for _, info := range listeners {
		cdsResources, err := generateGatewayClusters(ctx, xdsCtx, info)
		if err != nil {
			return err
		}
		resources.AddSet(cdsResources)

		ldsResources, limit, err := generateGatewayListeners(xdsCtx, info) // nolint: contextcheck
		if err != nil {
			return err
		}
		resources.AddSet(ldsResources)

		if limit != nil {
			limits = append(limits, *limit)
		}
	}

	resources.Add(plugin_gateway.GenerateRTDS(limits))

	return nil
}

func sortRulesToHosts(
	meshCtx xds_context.MeshContext,
	rawRules rules.GatewayRules,
	address string,
	port uint32,
	protocol mesh_proto.MeshGateway_Listener_Protocol,
	sublisteners []meshroute_gateway.Sublistener,
	resolver model.LabelResourceIdentifierResolver,
) []plugin_gateway.GatewayListenerHostname {
	hostInfosByHostname := map[string]plugin_gateway.GatewayListenerHostname{}
	for _, hostnameTag := range sublisteners {
		host := plugin_gateway.GatewayHost{
			Hostname: hostnameTag.Hostname,
			Routes:   nil,
			Policies: map[model.ResourceType][]match.RankedPolicy{},
			Tags:     hostnameTag.Tags,
		}
		hostInfo := plugin_gateway.GatewayHostInfo{
			Host: host,
		}
		inboundListener := rules.NewInboundListenerHostname(
			address,
			port,
			hostnameTag.Hostname,
		)
		rulesForListener, ok := rawRules.ToRules.ByListenerAndHostname[inboundListener]
		if !ok {
			continue
		}
		hostInfo.AppendEntries(generateEnvoyRouteEntries(meshCtx, host, rulesForListener, resolver))
		meshroute_gateway.AddToListenerByHostname(
			hostInfosByHostname,
			protocol,
			hostnameTag.Hostname,
			hostnameTag.TLS,
			hostInfo,
		)
	}

	return meshroute_gateway.SortByHostname(hostInfosByHostname)
}
