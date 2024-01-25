package v1alpha1

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
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
	if proxy.Dataplane == nil {
		return nil
	}

	policies := proxy.Policies.Dynamic[api.MeshTCPRouteType]
	// Only fallback if we have TrafficRoutes & No MeshTCPRoutes
	if len(ctx.Mesh.Resources.TrafficRoutes().Items) > 0 && len(policies.ToRules.Rules) == 0 && len(policies.GatewayRules.ToRules.ByListener) == 0 {
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

	listeners, err := generateListeners(proxy, policies.ToRules.Rules, servicesAccumulator, ctx.Mesh)
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

func ApplyToGateway(
	ctx context.Context,
	proxy *core_xds.Proxy,
	resources *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	policies core_xds.TypedMatchingPolicies,
) error {
	if len(policies.GatewayRules.ToRules.ByListener) == 0 {
		return nil
	}

	var limits []plugin_gateway.RuntimeResoureLimitListener

	listeners := plugin_gateway.ExtractGatewayListeners(proxy)
	for listenerIndex, info := range listeners {
		if info.Listener.Protocol != mesh_proto.MeshGateway_Listener_TCP {
			continue
		}
		address := proxy.Dataplane.Spec.GetNetworking().Address
		port := info.Listener.Port
		var hostInfos []plugin_gateway.GatewayHostInfo
		for _, info := range info.HostInfos {
			inboundListener := rules.NewInboundListenerHostname(
				address,
				port,
				info.Host.Hostname,
			)
			routes, ok := policies.GatewayRules.ToRules.ByListenerAndHostname[inboundListener]
			if !ok {
				continue
			}

			info.AppendEntries(GenerateEnvoyRouteEntries(info.Host, routes))
			hostInfos = append(hostInfos, info)
		}
		info.HostInfos = hostInfos
		listeners[listenerIndex] = info
		plugin_gateway.SetGatewayListeners(proxy, listeners)

		cdsResources, err := generateGatewayClusters(ctx, xdsCtx, info, hostInfos)
		if err != nil {
			return err
		}
		resources.AddSet(cdsResources)

		ldsResources, limit, err := generateGatewayListeners(xdsCtx, info, hostInfos) // nolint: contextcheck
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
