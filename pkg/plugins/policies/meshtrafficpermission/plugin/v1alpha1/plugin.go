package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	v3 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
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

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTrafficPermissionType, dataplane, resources, opts...)
}

func (p plugin) EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshTrafficPermissionType, tags, resources, opts...)
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
			"proxyName", proxy.Dataplane.GetMeta().GetName(),
			"mesh", ctx.Mesh.Resource.GetMeta().GetName())
		return nil
	}

	mtp := proxy.Policies.Dynamic[api.MeshTrafficPermissionType]
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
			if len(proxy.Policies.TrafficPermissions) == 0 {
				rules = p.denyRules()
			} else {
				continue
			}
		}

		configurer := &v3.RBACConfigurer{
			StatsName: res.Name,
			Rules:     rules,
			Mesh:      proxy.Dataplane.GetMeta().GetMesh(),
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
	}
	return nil
}

func (p plugin) denyRules() core_rules.Rules {
	return core_rules.Rules{
		&core_rules.Rule{
			Subset: subsetutils.MeshSubset(),
			Conf: api.Conf{
				Action: api.Deny,
			},
		},
	}
}

func (p plugin) allowRules() core_rules.Rules {
	return core_rules.Rules{
		&core_rules.Rule{
			Subset: subsetutils.MeshSubset(),
			Conf: api.Conf{
				Action: api.Allow,
			},
		},
	}
}

func (p plugin) configureEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	listeners := policies_xds.GatherListeners(rs)
	for _, resource := range proxy.ZoneEgressProxy.MeshResourcesList {
		meshName := resource.Mesh.GetMeta().GetName()
		if listeners.Egress == nil {
			log.V(1).Info("skip applying MeshTrafficPermission, Egress has no listener",
				"proxyName", proxy.ZoneEgressProxy.ZoneEgressResource.GetMeta().GetName(),
				"mesh", meshName,
			)
			return nil
		}
		if !resource.Mesh.MTLSEnabled() {
			log.V(1).Info("skip applying MeshTrafficPermission, MTLS is disabled", "mesh", meshName)
			continue
		}

		esNames := []string{}
		for _, es := range resource.ExternalServices {
			name := es.Spec.GetService()
			if name != "" {
				esNames = append(esNames, es.Spec.GetService())
			}
		}
		// egress is configured for all meshes so we cannot use mesh context in this case
		mesNames := []string{}
		for _, mes := range resource.ListOrEmpty(meshexternalservice_api.MeshExternalServiceType).GetItems() {
			meshExtSvc := mes.(*meshexternalservice_api.MeshExternalServiceResource)
			mesNames = append(mesNames, meshExtSvc.DestinationName(uint32(meshExtSvc.Spec.Match.Port)))
		}

		for _, esName := range esNames {
			var rules core_rules.FromRules
			if policies, ok := resource.Dynamic[esName]; ok {
				if mtp, ok := policies[api.MeshTrafficPermissionType]; ok {
					rules = mtp.FromRules
				}
			}
			if len(rules.Rules) == 0 {
				if resource.ExternalServicePermissionMap[esName] == nil {
					rules = core_rules.FromRules{
						Rules: map[core_rules.InboundListener]core_rules.Rules{
							{}: p.denyRules(),
						},
					}
				} else {
					continue
				}
			}

			for _, rule := range rules.Rules {
				configurer := &v3.RBACConfigurer{
					StatsName: listeners.Egress.Name,
					Rules:     rule,
					Mesh:      meshName,
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

		for _, mesName := range mesNames {
			rule := p.allowRules()
			if resource.Mesh.Spec.GetRouting() != nil && resource.Mesh.Spec.GetRouting().DefaultForbidMeshExternalServiceAccess {
				rule = p.denyRules()
			}
			rules := core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{}: rule,
				},
			}

			for _, rule := range rules.Rules {
				configurer := &v3.RBACConfigurer{
					StatsName: listeners.Egress.Name,
					Rules:     rule,
					Mesh:      meshName,
				}
				for _, filterChain := range listeners.Egress.FilterChains {
					if filterChain.Name == mesName {
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
