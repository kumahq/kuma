package v1alpha1

import (
	"context"
	"sort"
	"strings"

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
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type Route struct {
	Matches     []api.Match
	Filters     []api.Filter
	BackendRefs []common_api.BackendRef
}

type ToRouteRule struct {
	Subset    rules.Subset
	Rules     []api.Rule
	Hostnames []string
	Origin    []core_model.ResourceMeta
}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHTTPRouteType, dataplane, resources)
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
	if len(xdsCtx.Mesh.Resources.TrafficRoutes().Items) != 0 && len(policies.ToRules.Rules) == 0 && len(policies.GatewayRules.ToRules) == 0 {
		return nil
	}

	if err := ApplyToOutbounds(proxy, rs, xdsCtx, policies.ToRules.Rules); err != nil {
		return err
	}

	ctx := context.TODO()
	if err := ApplyToGateway(ctx, proxy, rs, xdsCtx, policies.GatewayRules); err != nil {
		return err
	}

	return nil
}

func ApplyToOutbounds(
	proxy *core_xds.Proxy,
	rs *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	rules rules.Rules,
) error {
	tlsReady := xdsCtx.Mesh.GetTLSReadiness()
	servicesAcc := envoy_common.NewServicesAccumulator(tlsReady)

	listeners, err := generateListeners(proxy, rules, servicesAcc, xdsCtx.Mesh)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAcc.Services()

	clusters, err := meshroute.GenerateClusters(proxy, xdsCtx.Mesh, services)
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

	if len(rawRules.ToRules) == 0 {
		return nil
	}

	for _, info := range plugin_gateway.ExtractGatewayListeners(proxy) {
		address := proxy.Dataplane.Spec.GetNetworking().Address
		port := info.Listener.Port
		inboundListener := rules.InboundListener{
			Address: address,
			Port:    port,
		}
		rawRules, ok := rawRules.ToRules[inboundListener]
		if !ok {
			continue
		}

		var keys []string
		rulesByHostname := map[string][]ToRouteRule{}
		for _, rawRule := range rawRules {
			conf := rawRule.Conf.(api.PolicyDefault)
			rule := ToRouteRule{
				Subset:    rawRule.Subset,
				Rules:     conf.Rules,
				Hostnames: conf.Hostnames,
				Origin:    rawRule.Origin,
			}
			hostnames := rule.Hostnames
			if len(rule.Hostnames) == 0 {
				hostnames = []string{"*"}
			}
			for _, hostname := range hostnames {
				accRule, ok := rulesByHostname[hostname]
				if !ok {
					keys = append(keys, hostname)
				}
				rulesByHostname[hostname] = append(accRule, rule)
			}
		}

		var hostInfos []plugin_gateway.GatewayHostInfo
		if info.Listener.Protocol == mesh_proto.MeshGateway_Listener_TCP {
			continue
		}

		for _, info := range info.HostInfos {
			listenerHostname := info.Host.Hostname
			separateHostnames := map[string][]ToRouteRule{}
			for _, routeHostname := range keys {
				if !(listenerHostname == "*" || routeHostname == "*" || match.Hostnames(listenerHostname, routeHostname)) {
					continue
				}
				// We need to take the most specific hostname
				hostnameKey := listenerHostname
				if strings.HasPrefix(listenerHostname, "*") && !strings.HasPrefix(routeHostname, "*") {
					hostnameKey = routeHostname
				}
				separateHostnames[hostnameKey] = append(separateHostnames[hostnameKey], rulesByHostname[routeHostname]...)
			}
			for hostname, rules := range separateHostnames {
				host := info.Host
				host.Hostname = hostname
				hostInfos = append(hostInfos, plugin_gateway.GatewayHostInfo{
					Host:    host,
					Entries: GenerateEnvoyRouteEntries(host, rules),
				})
			}
		}
		sort.Slice(hostInfos, func(i, j int) bool {
			return hostInfos[i].Host.Hostname > hostInfos[j].Host.Hostname
		})

		cdsResources, err := generateGatewayClusters(ctx, xdsCtx, info, hostInfos)
		if err != nil {
			return err
		}
		resources.AddSet(cdsResources)

		ldsResources, limit, err := generateGatewayListeners(xdsCtx, info, hostInfos)
		if err != nil {
			return err
		}
		resources.AddSet(ldsResources)

		if limit != nil {
			limits = append(limits, *limit)
		}

		rdsResources, err := generateGatewayRoutes(xdsCtx, info, hostInfos)
		if err != nil {
			return err
		}
		resources.AddSet(rdsResources)
	}

	resources.Add(plugin_gateway.GenerateRTDS(limits))

	return nil
}
