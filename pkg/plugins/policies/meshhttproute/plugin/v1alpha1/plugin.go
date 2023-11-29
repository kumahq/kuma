package v1alpha1

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type Route struct {
	Matches     []api.Match
	Filters     []api.Filter
	BackendRefs []common_api.BackendRef
}

type RuleAcc struct {
	MatchKey  []api.Match
	RuleConfs []api.RuleConf
}

type ToRouteRule struct {
	Subset rules.Subset
	Rules  []api.Rule
	Origin []core_model.ResourceMeta
}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHTTPRouteType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	// These policies have already been merged using the custom `GetDefault`
	// method and therefore are of the
	// `ToRouteRule` type, where rules have been appended together.
	policies := proxy.Policies.Dynamic[api.MeshHTTPRouteType]

	if len(policies.ToRules.Rules) == 0 {
		return nil
	}

	var toRules []ToRouteRule
	for _, policy := range policies.ToRules.Rules {
		toRules = append(toRules, ToRouteRule{
			Subset: policy.Subset,
			Rules:  policy.Conf.(api.PolicyDefault).Rules,
			Origin: policy.Origin,
		})
	}

	if err := ApplyToOutbounds(proxy, rs, ctx, toRules); err != nil {
		return err
	}
	return nil
}

func ApplyToOutbounds(
	proxy *core_xds.Proxy,
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	rules []ToRouteRule,
) error {
	tlsReady := map[string]bool{}
	for serviceName, info := range ctx.Mesh.ServiceInformations {
		tlsReady[serviceName] = info.TLSReadiness
	}
	servicesAcc := envoy_common.NewServicesAccumulator(tlsReady)

	listeners, err := generateListeners(proxy, rules, servicesAcc, ctx.Mesh.ServiceInformations)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAcc.Services()

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
