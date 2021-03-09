package v3

import (
	"fmt"

	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type NetworkRBACConfigurer struct {
	StatsName  string
	Permission *mesh_core.TrafficPermissionResource
}

func (c *NetworkRBACConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filter, err := createRbacFilter(c.StatsName, c.Permission)
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
		Name: "envoy.filters.network.rbac",
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
					PrincipalName: tls.ServiceSpiffeIDMatcher(permission.Meta.GetMesh(), service),
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
