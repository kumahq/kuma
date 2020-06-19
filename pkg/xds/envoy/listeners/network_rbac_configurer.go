package listeners

import (
	"fmt"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/rbac/v2"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/util/proto"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

func NetworkRBAC(statsName string, rbacEnabled bool, permission *mesh_core.TrafficPermissionResource) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if rbacEnabled {
			config.Add(&NetworkRBACConfigurer{
				statsName:  statsName,
				permission: permission,
			})
		}
	})
}

type NetworkRBACConfigurer struct {
	statsName  string
	permission *mesh_core.TrafficPermissionResource
}

func (c *NetworkRBACConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filter, err := createRbacFilter(c.statsName, c.permission)
	if err != nil {
		return err
	}

	// RBAC filter should be the first in the chain
	filterChain.Filters = append([]*envoy_listener.Filter{filter}, filterChain.Filters...)
	return nil
}

func createRbacFilter(statsName string, permission *mesh_core.TrafficPermissionResource) (*envoy_listener.Filter, error) {
	rbacRule := createRbacRule(statsName, permission)
	rbacMarshalled, err := proto.MarshalAnyDeterministic(rbacRule)
	if err != nil {
		return nil, err
	}
	return &envoy_listener.Filter{
		Name: envoy_wellknown.RoleBasedAccessControl,
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: rbacMarshalled,
		},
	}, nil
}

func createRbacRule(statsName string, permission *mesh_core.TrafficPermissionResource) *rbac.RBAC {
	policies := make(map[string]*rbac_config.Policy)
	// We only create policy if Traffic Permission is selected. Otherwise we still need to build RBAC filter
	// to restrict all the traffic coming to the dataplane.
	if permission != nil {
		policies[permission.GetMeta().GetName()] = createPolicy(permission)
	}

	return &rbac.RBAC{
		Rules: &rbac_config.RBAC{
			Action:   rbac_config.RBAC_ALLOW,
			Policies: policies,
		},
		StatPrefix: fmt.Sprintf("%s.", util_xds.SanitizeMetric(statsName)), // we include dot to change "inbound:127.0.0.1:21011rbac.allowed" metric to "inbound:127.0.0.1:21011.rbac.allowed"
	}
}

func createPolicy(permission *mesh_core.TrafficPermissionResource) *rbac_config.Policy {
	principals := []*rbac_config.Principal{}

	// build principals list: one per sources/destinations rule
	for _, source := range permission.Spec.Sources {
		service := source.Match[mesh_proto.ServiceTag]
		principal := &rbac_config.Principal{}
		if service == mesh_proto.MatchAllTag {
			principal.Identifier = &rbac_config.Principal_Any{
				Any: true,
			}
		} else {
			principal.Identifier = &rbac_config.Principal_Authenticated_{
				Authenticated: &rbac_config.Principal_Authenticated{
					PrincipalName: envoy.ServiceSpiffeIDMatcher(permission.Meta.GetMesh(), service),
				},
			}
		}
		principals = append(principals, principal)
	}

	return &rbac_config.Policy{
		Permissions: []*rbac_config.Permission{
			{
				Rule: &rbac_config.Permission_Any{
					Any: true,
				},
			},
		},
		Principals: principals,
	}
}
