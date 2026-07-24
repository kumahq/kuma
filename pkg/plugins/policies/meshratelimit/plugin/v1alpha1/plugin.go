package v1alpha1

import (
	"slices"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshratelimit/plugin/xds"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var (
	_   core_plugins.EgressPolicyPlugin = &plugin{}
	log                                 = core.Log.WithName("MeshRateLimit")
)

type plugin struct{}

func (p plugin) Order() int { return api.MeshRateLimitResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRateLimitType, dataplane, resources, opts...)
}

func (p plugin) EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshRateLimitType, tags, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.ZoneEgressProxy != nil {
		return applyToEgress(rs, proxy)
	}
	if proxy.Dataplane == nil {
		return nil
	}

	policies, ok := proxy.Policies.Dynamic[api.MeshRateLimitType]
	if !ok {
		return nil
	}

	listeners := xds.GatherListeners(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy); err != nil {
		return err
	}
	if err := applyToZoneProxyListeners(policies, listeners, proxy); err != nil {
		return err
	}

	return nil
}

func applyToInbounds(
	fromRules core_rules.FromRules,
	inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	proxy *core_xds.Proxy,
) error {
	for _, inbound := range proxy.Dataplane.Spec.GetNetworking().GetInbound() {
		iface := proxy.Dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		inboundRules := fromRules.InboundRules[listenerKey]
		conf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](inboundRules)
		applyCommonConf := len(inboundRules) == 0 || hasCatchAllInboundRule(inboundRules)
		configurer := plugin_xds.ListenerConfigurer{
			Conf:             conf,
			Rules:            inboundRules,
			SkipCommonConfig: !applyCommonConf,
		}
		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}
	}

	return nil
}

func applyToZoneProxyListeners(
	policies core_xds.TypedMatchingPolicies,
	listeners xds.Listeners,
	proxy *core_xds.Proxy,
) error {
	networking := proxy.Dataplane.Spec.GetNetworking()
	if !networking.HasZoneProxyListeners() {
		return nil
	}

	for _, listener := range networking.GetListeners() {
		var (
			envoyListener *envoy_listener.Listener
			ok            bool
		)

		switch listener.GetType() {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			envoyListener, ok = listeners.ZoneIngress[naming.ContextualZoneIngressListenerName(listener.GetSectionName())]
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			envoyListener, ok = listeners.ZoneEgress[naming.ContextualZoneEgressListenerName(listener.GetSectionName())]
		default:
			continue
		}
		if !ok {
			continue
		}

		inboundRules, err := buildListenerScopedInboundRules(policies, listener.GetSectionName())
		if err != nil {
			return err
		}
		if len(inboundRules) == 0 {
			continue
		}

		if err := applyToZoneProxyListener(envoyListener, inboundRules); err != nil {
			return err
		}
	}

	return nil
}

func applyToZoneProxyListener(
	listener *envoy_listener.Listener,
	inboundRules []*rules_inbound.Rule,
) error {
	commonConf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](inboundRules)
	applyCommonConf := hasCatchAllInboundRule(inboundRules)

	for _, filterChain := range listener.FilterChains {
		matchedRules := zoneProxyFilterChainRules(inboundRules, filterChain)
		baseConf, ok, err := effectiveZoneProxyFilterChainConf(commonConf, applyCommonConf, matchedRules)
		if err != nil {
			return err
		}
		if ok {
			if err := plugin_xds.ConfigureFilterChain(baseConf, filterChain); err != nil {
				return err
			}
		}

		if err := plugin_xds.ConfigureMatchedRoutesOnFilterChain(filterChain, baseConf, inboundRules); err != nil {
			return err
		}
	}

	return nil
}

func effectiveZoneProxyFilterChainConf(
	commonConf api.Conf,
	applyCommonConf bool,
	matchedRules []*rules_inbound.Rule,
) (api.Conf, bool, error) {
	confs := make([]api.Conf, 0, 2)
	if applyCommonConf {
		confs = append(confs, commonConf)
	}
	if conf, ok, err := mergeZoneProxyRuleConfs(matchedRules); err != nil {
		return api.Conf{}, false, err
	} else if ok {
		confs = append(confs, conf)
	}

	return mergeRateLimitConfs(confs...)
}

func applyToEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	listeners := xds.GatherListeners(rs)
	if listeners.Egress == nil {
		log.V(1).Info("skip applying MeshRateLimit, Egress has no listener",
			"proxyName", proxy.ZoneEgressProxy.ZoneEgressResource.GetMeta().GetName(),
		)
		return nil
	}
	return nil
}

func hasCatchAllInboundRule(rules []*rules_inbound.Rule) bool {
	for _, rule := range rules {
		if rule.Match == nil {
			return true
		}
	}
	return false
}

func zoneProxyFilterChainRules(inboundRules []*rules_inbound.Rule, filterChain *envoy_listener.FilterChain) []*rules_inbound.Rule {
	var matched []*rules_inbound.Rule
	for _, rule := range inboundRules {
		if matchesZoneProxyFilterChain(rule, filterChain) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func matchesZoneProxyFilterChain(rule *rules_inbound.Rule, filterChain *envoy_listener.FilterChain) bool {
	if rule.Match == nil {
		return false
	}
	serverNames := filterChain.GetFilterChainMatch().GetServerNames()
	if len(serverNames) == 0 {
		return false
	}
	if rule.Match.SpiffeID != nil || rule.Match.SNI == nil {
		return false
	}
	return slices.Contains(serverNames, rule.Match.SNI.Value)
}

func mergeZoneProxyRuleConfs(rules []*rules_inbound.Rule) (api.Conf, bool, error) {
	confs := make([]any, 0, len(rules))
	for _, rule := range rules {
		conf, ok := rule.Conf.(api.Conf)
		if !ok {
			continue
		}
		confs = append(confs, conf)
	}
	if len(confs) == 0 {
		return api.Conf{}, false, nil
	}

	merged, err := merge.Confs(confs)
	if err != nil {
		return api.Conf{}, false, err
	}
	if len(merged) == 0 {
		return api.Conf{}, false, nil
	}

	conf, ok := merged[0].(api.Conf)
	if !ok {
		return api.Conf{}, false, errors.Errorf("unexpected merged zone proxy conf type: %T", merged[0])
	}
	return conf, true, nil
}

func mergeRateLimitConfs(confs ...api.Conf) (api.Conf, bool, error) {
	mergedInputs := make([]any, 0, len(confs))
	for _, conf := range confs {
		if conf.Local == nil {
			continue
		}
		mergedInputs = append(mergedInputs, conf)
	}
	if len(mergedInputs) == 0 {
		return api.Conf{}, false, nil
	}

	merged, err := merge.Confs(mergedInputs)
	if err != nil {
		return api.Conf{}, false, err
	}
	if len(merged) == 0 {
		return api.Conf{}, false, nil
	}

	conf, ok := merged[0].(api.Conf)
	if !ok {
		return api.Conf{}, false, errors.Errorf("unexpected merged rate limit conf type: %T", merged[0])
	}
	return conf, true, nil
}

func buildListenerScopedInboundRules(
	policies core_xds.TypedMatchingPolicies,
	sectionName string,
) ([]*rules_inbound.Rule, error) {
	if len(policies.DataplanePolicies) == 0 {
		return nil, nil
	}

	filtered := api.MeshRateLimitResourceTypeDescriptor.NewList()
	for _, resource := range policies.DataplanePolicies {
		policy, ok := resource.GetSpec().(core_model.Policy)
		if !ok {
			continue
		}
		targetRef := policy.GetTargetRef()
		if targetRef.Kind == common_api.Dataplane {
			if sn := pointer.Deref(targetRef.SectionName); sn != "" && sn != sectionName {
				continue
			}
		}
		if err := filtered.AddItem(resource); err != nil {
			return nil, err
		}
	}

	return rules_inbound.BuildRules(filtered)
}
