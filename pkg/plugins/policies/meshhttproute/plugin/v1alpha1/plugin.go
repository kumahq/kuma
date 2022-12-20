package v1alpha1

import (
	"reflect"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type Route struct {
	Matches     []api.Match
	Filters     []api.Filter
	BackendRefs []api.BackendRef
}

type RuleAcc struct {
	MatchKey  []api.Match
	RuleConfs []api.RuleConf
}

type ToRouteRule struct {
	Subset core_xds.Subset
	Rules  []api.Rule
	Origin []core_model.ResourceMeta
}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHTTPRouteType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	// These policies have alreadu been merged using the custom `GetDefault`
	// method and therefore are of the
	// `ToRouteRule` type, where rules have been appended together.
	policies := proxy.Policies.Dynamic[api.MeshHTTPRouteType]

	var toRules []ToRouteRule
	for _, policy := range policies.ToRules.Rules {
		toRules = append(toRules, ToRouteRule{
			Subset: policy.Subset,
			Rules:  policy.Conf.([]api.Rule),
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
	servicesAcc := envoy_common.NewServicesAccumulator(ctx.Mesh.ServiceTLSReadiness)

	listeners, err := generateListeners(proxy, rules, servicesAcc)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAcc.Services()

	clusters, err := generateClusters(proxy, ctx.Mesh, services)
	if err != nil {
		return errors.Wrap(err, "couldn't generate cluster resources")
	}
	rs.AddSet(clusters)

	endpoints, err := generateEndpoints(proxy, ctx, services)
	if err != nil {
		return errors.Wrap(err, "couldn't generate endpoint resources")
	}
	rs.AddSet(endpoints)

	return nil
}

func FindRoutes(
	rules []ToRouteRule,
	serviceName string,
) []Route {
	var unmergedRules []RuleAcc

	for _, rule := range rules {
		if !rule.Subset.IsSubset(core_xds.MeshService(serviceName)) {
			continue
		}
		// Look through all the rules and accumulate all filters/refs for a
		// given matches value.
		// TODO: this is O(#rules^2*#matches) because Go can't use a list of
		// matches as a key.
		for _, routeRules := range rule.Rules {
			var found bool
			// Treat the list of (matches, filters/refs) as a map
			for i, accRule := range unmergedRules {
				if !reflect.DeepEqual(accRule.MatchKey, routeRules.Matches) {
					continue
				}
				unmergedRules[i] = RuleAcc{
					MatchKey:  accRule.MatchKey,
					RuleConfs: append(accRule.RuleConfs, routeRules.Default),
				}
				found = true
			}
			if !found {
				unmergedRules = append(unmergedRules, RuleAcc{
					MatchKey:  routeRules.Matches,
					RuleConfs: []api.RuleConf{routeRules.Default},
				})
			}
		}
	}

	var routes []Route

	for _, rule := range unmergedRules {
		route := Route{
			Matches: rule.MatchKey,
		}
		for _, conf := range rule.RuleConfs {
			if conf.Filters != nil {
				route.Filters = *conf.Filters
			}
			if conf.BackendRefs != nil {
				route.BackendRefs = *conf.BackendRefs
			}
		}
		routes = append(routes, route)
	}

	// append the default route
	routes = append(routes, Route{
		Matches: []api.Match{{
			Path: api.PathMatch{
				Prefix: "/",
			},
		}},
		Filters: nil,
		BackendRefs: []api.BackendRef{{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: serviceName,
			},
			Weight: 100,
		}},
	})

	return routes
}
