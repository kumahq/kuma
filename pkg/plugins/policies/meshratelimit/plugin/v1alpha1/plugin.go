package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ core_plugins.PolicyPlugin = &plugin{}
var log = core.Log.WithName("MeshRateLimit")

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRateLimitType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshRateLimitType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	return nil
}

func applyToInbounds(fromRules core_xds.FromRules, inboundListeners map[core_xds.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource) error {
	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := xds.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		rules, ok := fromRules.Rules[listenerKey]
		if !ok {
			continue
		}

		if err := configure(rules, xds.MeshSubset(), iface.DataplanePort, listener, dataplane); err != nil {
			return err
		}
	}

	return nil
}

func configure(rules core_xds.Rules,
	subset core_xds.Subset,
	dataplanePort uint32,
	listener *envoy_listener.Listener,
	dataplane *core_mesh.DataplaneResource) error {
	configurer := plugin_xds.Configurer{
		From:        rules,
		ClusterName: envoy_names.GetLocalClusterName(dataplanePort),
		Dataplane:   dataplane,
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}
	return nil
}
