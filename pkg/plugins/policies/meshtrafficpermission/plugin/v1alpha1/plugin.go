package v1alpha1

import (
	"fmt"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	v3 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var (
	_   core_plugins.EgressPolicyPlugin = &plugin{}
	log                                 = core.Log.WithName("MeshTrafficPermission")
)

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTrafficPermissionType, dataplane, resources)
}

func (p plugin) EgressMatchedPolicies(es *core_mesh.ExternalServiceResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshTrafficPermissionType, es, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.ZoneEgressProxy != nil {
		return p.configureEgress(rs, proxy)
	}

	if proxy.Dataplane == nil || proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	if !ctx.Mesh.Resource.MTLSEnabled() {
		log.V(1).Info("skip applying MeshTrafficPermission, MTLS is disabled",
			"mesh", ctx.Mesh.Resource.GetMeta().GetName())
		return nil
	}

	mtp, ok := proxy.Policies.Dynamic[api.MeshTrafficPermissionType]
	if !ok {
		return nil
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		if res.Origin != generator.OriginInbound {
			continue
		}

		listener := res.Resource.(*envoy_listener.Listener)
		dpAddress := listener.GetAddress().GetSocketAddress()

		key := core_rules.InboundListener{
			Address: dpAddress.GetAddress(),
			Port:    dpAddress.GetPortValue(),
		}
		rules, ok := mtp.FromRules.Rules[key]
		if !ok {
			continue
		}

		configurer := &v3.RBACConfigurer{
			StatsName: res.Name,
			Rules:     rules,
			Mesh:      proxy.Dataplane.GetMeta().GetMesh(),
		}
		for _, filterChain := range listener.FilterChains {
			if err := configurer.Configure(filterChain); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p plugin) configureEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	listeners := policies_xds.GatherListeners(rs)
	for _, resource := range proxy.ZoneEgressProxy.MeshResourcesList {
		if !resource.Mesh.MTLSEnabled() {
			log.V(1).Info("skip applying MeshTrafficPermission, MTLS is disabled",
				"mesh", resource.Mesh.GetMeta().GetName())
			continue
		}
		for _, es := range resource.ExternalServices {
			meshName := resource.Mesh.GetMeta().GetName()
			esName := es.Meta.GetName()
			policies, ok := resource.Dynamic[esName]
			if !ok {
				continue
			}
			mtp, ok := policies[api.MeshTrafficPermissionType]
			if !ok {
				continue
			}
			l := listeners.Egress[core_rules.InboundListener{
				Address: proxy.ZoneEgressProxy.ZoneEgressResource.Spec.Networking.Address,
				Port:    proxy.ZoneEgressProxy.ZoneEgressResource.Spec.Networking.Port,
			}]

			for _, rule := range mtp.FromRules.Rules {
				configurer := &v3.RBACConfigurer{
					StatsName: l.Name,
					Rules:     rule,
					Mesh:      resource.Mesh.GetMeta().GetName(),
				}
				for _, filterChain := range l.FilterChains {
					if filterChain.Name == fmt.Sprintf("%s_%s", esName, meshName) {
						if err := configurer.Configure(filterChain); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
