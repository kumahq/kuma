package v1alpha1

import (
	"context"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshroute_gateway "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute/gateway"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type ToRouteRule struct {
	Subset    rules.Subset
	Rules     []api.Rule
	Hostnames []string

	Origins          []core_model.ResourceMeta
	BackendRefOrigin map[common_api.MatchesHash]core_model.ResourceMeta
}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHTTPRouteType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}

	// These policies have already been merged using the custom `GetDefault`
	// method and therefore are of the
	// `ToRouteRule` type, where rules have been appended together.
	policies := proxy.Policies.Dynamic[api.MeshHTTPRouteType]

	// Only fallback if we have TrafficRoutes & No MeshHTTPRoutes
	if len(xdsCtx.Mesh.Resources.TrafficRoutes().Items) != 0 && len(policies.ToRules.Rules) == 0 && len(policies.GatewayRules.ToRules.ByListenerAndHostname) == 0 {
		return nil
	}

	if err := ApplyToOutbounds(proxy, rs, xdsCtx, policies.ToRules); err != nil {
		return err
	}

	ctx := context.TODO()
	if err := ApplyToGateway(ctx, proxy, rs, xdsCtx, policies.GatewayRules); err != nil {
		return err
	}

	return nil
}

func ApplyToOutbounds(proxy *core_xds.Proxy, rs *core_xds.ResourceSet, xdsCtx xds_context.Context, rules rules.ToRules) error {
	tlsReady := xdsCtx.Mesh.GetTLSReadiness()
	servicesAcc := envoy_common.NewServicesAccumulator(tlsReady)

	listeners, err := generateListeners(proxy, rules, servicesAcc, xdsCtx.Mesh)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAcc.Services()

	clusters, err := meshroute.GenerateClusters(proxy, xdsCtx.Mesh, services, xdsCtx.ControlPlane.SystemNamespace)
	if err != nil {
		return errors.Wrap(err, "couldn't generate cluster resources")
	}
	rs.AddSet(clusters)

	endpoints, err := meshroute.GenerateEndpoints(proxy, xdsCtx, services)
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
	rawRules rules.GatewayRules,
) error {
	var limits []plugin_gateway.RuntimeResoureLimitListener

	if len(rawRules.ToRules.ByListenerAndHostname) == 0 {
		return nil
	}

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
		rawRules,
		[]mesh_proto.MeshGateway_Listener_Protocol{mesh_proto.MeshGateway_Listener_HTTP, mesh_proto.MeshGateway_Listener_HTTPS},
		sortRulesToHosts,
	)
	plugin_gateway.SetGatewayListeners(proxy, listeners)

	for _, listener := range listeners {
		cdsResources, err := generateGatewayClusters(ctx, xdsCtx, listener)
		if err != nil {
			return err
		}
		resources.AddSet(cdsResources)

		ldsResources, limit, err := generateGatewayListeners(xdsCtx, listener)
		if err != nil {
			return err
		}
		resources.AddSet(ldsResources)

		if limit != nil {
			limits = append(limits, *limit)
		}

		rdsResources, err := generateGatewayRoutes(xdsCtx, listener)
		if err != nil {
			return err
		}
		resources.AddSet(rdsResources)
	}

	resources.Add(plugin_gateway.GenerateRTDS(limits))

	return nil
}
