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
	if proxy.Dataplane == nil {
		return nil
	}
	policies := proxy.Policies.Dynamic[api.MeshTCPRouteType]
	// Only fallback if we have TrafficRoutes & No MeshTCPRoutes
	if len(ctx.Mesh.Resources.TrafficRoutes().Items) > 0 && len(policies.ToRules.Rules) == 0 && len(policies.GatewayRules.Rules) == 0 {
		return nil
	}

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
