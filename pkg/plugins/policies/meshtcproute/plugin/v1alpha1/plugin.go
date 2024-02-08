package v1alpha1

import (
	"github.com/pkg/errors"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(
	dataplane *core_mesh.DataplaneResource,
	resources xds_context.Resources,
) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTCPRouteType, dataplane, resources)
}

func (p plugin) Apply(
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) error {
	tcpRules := proxy.Policies.Dynamic[api.MeshTCPRouteType].ToRules.Rules
	if len(tcpRules) == 0 {
		return nil
	}

	servicesAccumulator := envoy_common.NewServicesAccumulator(
		ctx.Mesh.ServiceTLSReadiness)

	listeners, err := generateListeners(proxy, tcpRules, servicesAccumulator)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAccumulator.Services()

	clusters, err := meshroute.GenerateClusters(proxy, ctx.Mesh, services)
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
<<<<<<< HEAD
=======

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

	listeners := meshroute.CollectListenerInfos(
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
	meshLocalResources xds_context.ResourceMap,
	rawRules rules.GatewayRules,
	address string,
	listener *mesh_proto.MeshGateway_Listener,
	sublisteners []meshroute.Sublistener,
) []plugin_gateway.GatewayListenerHostname {
	hostInfosByHostname := map[string]plugin_gateway.GatewayListenerHostname{}
	for _, hostnameTag := range sublisteners {
		host := plugin_gateway.GatewayHost{
			Hostname: hostnameTag.Hostname,
			Routes:   nil,
			Policies: map[model.ResourceType][]match.RankedPolicy{},
			TLS:      listener.Tls,
			Tags:     hostnameTag.Tags,
		}
		hostInfo := plugin_gateway.GatewayHostInfo{
			Host: host,
		}
		inboundListener := rules.NewInboundListenerHostname(
			address,
			listener.GetPort(),
			hostnameTag.Hostname,
		)
		rulesForListener, ok := rawRules.ToRules.ByListenerAndHostname[inboundListener]
		if !ok {
			continue
		}
		hostInfo.AppendEntries(generateEnvoyRouteEntries(host, rulesForListener))
		meshroute.AddToListenerByHostname(
			hostInfosByHostname,
			listener.Protocol,
			hostnameTag.Hostname,
			listener.Tls,
			hostInfo,
		)
	}

	return meshroute.SortByHostname(hostInfosByHostname)
}
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
