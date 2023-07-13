package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	v3 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ core_plugins.PolicyPlugin = &plugin{}
var log = core.Log.WithName("MeshTrafficPermission")

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTrafficPermissionType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		// MeshTrafficPermission policy is applied only on DPP
		// todo(lobkovilya): add support for ExternalService and ZoneEgress, https://github.com/kumahq/kuma/issues/5050
		return nil
	}

	if proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	mtp, ok := proxy.Policies.Dynamic[api.MeshTrafficPermissionType]
	if !ok {
		return nil
	}

	if !ctx.Mesh.Resource.MTLSEnabled() {
		log.V(1).Info("skip applying MeshTrafficPermission, MTLS is disabled",
			"proxyName", proxy.Dataplane.GetMeta().GetName(),
			"mesh", ctx.Mesh.Resource.GetMeta().GetName())
		return nil
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		if res.Origin != generator.OriginInbound {
			continue
		}

		listener := res.Resource.(*envoy_listener.Listener)
		dpAddress := listener.GetAddress().GetSocketAddress()

		key := core_xds.InboundListener{
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
<<<<<<< HEAD
=======

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
			esName, ok := es.Spec.GetTags()[mesh_proto.ServiceTag]
			if !ok {
				continue
			}
			policies, ok := resource.Dynamic[esName]
			if !ok {
				continue
			}
			mtp, ok := policies[api.MeshTrafficPermissionType]
			if !ok {
				continue
			}
			if listeners.Egress == nil {
				log.V(1).Info("skip applying MeshTrafficPermission, Egress has no listener",
					"proxyName", proxy.ZoneEgressProxy.ZoneEgressResource.GetMeta().GetName(),
					"mesh", resource.Mesh.GetMeta().GetName(),
				)
				return nil
			}

			for _, rule := range mtp.FromRules.Rules {
				configurer := &v3.RBACConfigurer{
					StatsName: listeners.Egress.Name,
					Rules:     rule,
					Mesh:      resource.Mesh.GetMeta().GetName(),
				}
				for _, filterChain := range listeners.Egress.FilterChains {
					if filterChain.Name == names.GetEgressFilterChainName(esName, meshName) {
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
>>>>>>> e6ced1005 (fix(MeshTrafficPermission): use serviceName instead of resource name for egress MTP (#7225))
