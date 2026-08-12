package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	v3 "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/xds"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("MeshTrafficPermission")
)

type plugin struct{}

func (p plugin) Order() int { return api.MeshTrafficPermissionResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil || proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	mtp := proxy.Policies.Dynamic[api.MeshTrafficPermissionType]

	if proxy.WorkloadIdentity == nil {
		log.V(1).Info("skip applying MeshTrafficPermission, the proxy has no workload identity",
			"proxyName", proxy.Dataplane.GetMeta().GetName(),
			"mesh", ctx.Mesh.Resource.GetMeta().GetName())
		return nil
	}

	if err := p.configureZoneEgressListeners(rs, mtp); err != nil {
		return err
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		if res.Origin != metadata.OriginInbound {
			continue
		}
		listener := res.Resource.(*envoy_listener.Listener)
		dpAddress := listener.GetAddress().GetSocketAddress()

		key := core_rules.InboundListener{
			Address: dpAddress.GetAddress(),
			Port:    dpAddress.GetPortValue(),
		}

		// an inbound with no matching rule gets an RBAC filter with no matchers,
		// which denies everything
		inboundRules := mtp.FromRules.InboundRules[key]
		configurer := &v3.RBACConfigurer{
			StatsName:    res.Name,
			InboundRules: inboundRules,
		}
		for _, filterChain := range listener.FilterChains {
			if filterChain.TransportSocket.GetName() != wellknown.TransportSocketTLS {
				// we only want to configure RBAC on listeners protected by Kuma's TLS
				continue
			}
			if err := configurer.Configure(filterChain); err != nil {
				return err
			}
		}
		if hasSNIMatch(inboundRules) {
			if err := envoy_listeners_v3.EnsureTLSInspector(listener); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p plugin) configureZoneEgressListeners(rs *core_xds.ResourceSet, mtp core_xds.TypedMatchingPolicies) error {
	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		if res.Origin != metadata.OriginEgress {
			continue
		}
		listener := res.Resource.(*envoy_listener.Listener)
		dpAddress := listener.GetAddress().GetSocketAddress()
		key := core_rules.InboundListener{
			Address: dpAddress.GetAddress(),
			Port:    dpAddress.GetPortValue(),
		}
		configurer := &v3.RBACConfigurer{
			StatsName:    res.Name,
			InboundRules: mtp.FromRules.InboundRules[key],
		}
		for _, filterChain := range listener.FilterChains {
			if filterChain.TransportSocket.GetName() != wellknown.TransportSocketTLS {
				continue
			}
			if err := configurer.Configure(filterChain); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasSNIMatch(rules []*inbound.Rule) bool {
	hasSNI := func(matches *[]common_api.Match) bool {
		for _, m := range pointer.Deref(matches) {
			if m.SNI != nil {
				return true
			}
		}
		return false
	}
	for _, rule := range rules {
		conf, ok := rule.Conf.(api.RuleConf)
		if !ok {
			continue
		}
		if hasSNI(conf.Deny) || hasSNI(conf.Allow) || hasSNI(conf.AllowWithShadowDeny) {
			return true
		}
	}
	return false
}
