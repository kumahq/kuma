package v1alpha1

import (
	"fmt"
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("MeshPassthrough")
)

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshPassthroughType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}
	policies, ok := proxy.Policies.Dynamic[api.MeshPassthroughType]
	if !ok {
		return nil
	}
	if proxy.Dataplane != nil && proxy.Dataplane.Spec.Networking.TransparentProxying == nil {
		policies.Warnings = append(policies.Warnings, fmt.Sprintf("policy doesn't support proxy running without transparent-proxy"))
		return nil
	}
	
	listeners := policies_xds.GatherListeners(rs)
	if err := applyToInbounds(ctx, policies.SingleItemRules, listeners.Ipv4Passthrough, proxy.Dataplane); err != nil {
		return err
	}
	return nil
}

func applyToInbounds(
	ctx xds_context.Context,
	rules core_rules.SingleItemRules,
	passthrough *envoy_listener.Listener,
	dataplane *core_mesh.DataplaneResource,
) error {
	if len(rules.Rules) == 0 {
		return nil
	}
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(api.Conf)

	if ctx.Mesh.Resource.Spec.IsPassthrough() && !pointer.Deref[bool](conf.Enabled) && len(conf.AppendMatch) == 0 {
		// remove cluster
	} else if !ctx.Mesh.Resource.Spec.IsPassthrough() && pointer.Deref[bool](conf.Enabled) {
		// add cluster
	} else if ctx.Mesh.Resource.Spec.IsPassthrough() && !pointer.Deref[bool](conf.Enabled) && len(conf.AppendMatch) > 0 {
		// remove default and add matchers
		_ = orderRules(conf)
		// generateMatchers := generateMatchers(ordered, passthrough)
	} else if !ctx.Mesh.Resource.Spec.IsPassthrough() && !pointer.Deref[bool](conf.Enabled) && len(conf.AppendMatch) > 0{
		// ordered := orderRules(conf)
	}

	return nil
}

func generateMatchers(orderedRules map[string]map[int]map[MatchInfo][]api.Match, passthrough *envoy_listener.Listener) error {
	configurer := &v3.TLSInspectorConfigurer{}
	err := configurer.Configure(passthrough)
	if err != nil {
		return err
	}

	// passthrough.FilterChainMatcher := &v32.Matcher{
	// 	MatcherType: &v32.Matcher_MatcherTree_{
	// 		MatcherTree: &v32.Matcher_MatcherTree{
	// 			Input: &v3.TypedExtensionConfig{

	// 			}},
	// 	},
	// }


	return nil
}

func orderRules(conf api.Conf) map[string]map[int]map[MatchInfo][]api.Match {
	ordered := map[string]map[int]map[MatchInfo][]api.Match{}
	for _, match := range conf.AppendMatch{
		matchInfo := MatchInfo{
			Type: match.Type,
			IsWildcard: strings.HasPrefix(match.Value, "*"),
		}
		switch match.Protocol {
		case "tls":
			if _, ok := ordered["tls"]; !ok {
				ordered["tls"] = map[int]map[MatchInfo][]api.Match{}
			}
			port := 0
			if match.Port != nil {
				port = *match.Port
			}
			if _, ok := ordered["tls"][port]; !ok {
				ordered["tls"][port] = map[MatchInfo][]api.Match{}
			}
			if _, ok := ordered["tls"][port][matchInfo]; !ok {
				ordered["tls"][port][matchInfo] = []api.Match{}
			}
			ordered["tls"][port][matchInfo] = append(ordered["tls"][port][matchInfo], match)
		default:
			if _, ok := ordered["raw_buffer"]; !ok {
				ordered["raw_buffer"] = map[int]map[MatchInfo][]api.Match{}
			}
			port := 0
			if match.Port != nil {
				port = *match.Port
			}
			if _, ok := ordered["raw_buffer"][port]; !ok {
				ordered["raw_buffer"][port] = map[MatchInfo][]api.Match{}
			}
			if _, ok := ordered["raw_buffer"][port][matchInfo]; !ok {
				ordered["raw_buffer"][port][matchInfo] = []api.Match{}
			}
			ordered["raw_buffer"][port][matchInfo] = append(ordered["raw_buffer"][port][matchInfo], match)
		}
	}

	// for _, protocol := range ordered {
	// 	for _, port := range ordered[protocol] {
	// 		orderedList := []MatchInfo{}

	// 		for orderedList
			
	// 	}
	// }
	return ordered
}

type MatchInfo struct {
	Type api.MatchType
	IsWildcard bool
}

// tls 
// 		port
//			server name
//			ip address
//		ip address
// raw_buffer
// 		port
//			http listener/hostheader
//			source ip
//		source ip
// 

// map[protocol]map[port][] string