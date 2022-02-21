package v3

import (
	"fmt"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	rbac "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type NetworkRBACConfigurer struct {
	StatsName  string
	Permission *core_mesh.TrafficPermissionResource
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

func createRbacFilter(statsName string, permission *core_mesh.TrafficPermissionResource) (*envoy_listener.Filter, error) {
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

func createRbacRule(statsName string, permission *core_mesh.TrafficPermissionResource) *rbac.RBAC {
	policies := make(map[string]*rbac_config.Policy)
	// We only create policy if Traffic Permission is selected. Otherwise, we still need to build RBAC filter
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

func createPolicy(permission *core_mesh.TrafficPermissionResource) *rbac_config.Policy {
	principals := []*rbac_config.Principal{}

	// build principals list: one per sources/destinations rule
	for _, selector := range permission.Spec.Sources {
		principals = append(principals, principalFromSelector(selector, permission.GetMeta().GetMesh()))
	}

	return &rbac_config.Policy{
		Permissions: []*rbac_config.Permission{
			{
				Rule: &rbac_config.Permission_Any{
					Any: true, // the relation between many selector is OR
				},
			},
		},
		Principals: principals,
	}
}

func principalFromSelector(selector *mesh_proto.Selector, mesh string) *rbac_config.Principal {
	principals := kumaPrincipals(selector)

	service := selector.Match[mesh_proto.ServiceTag]
	if service != "" && service != mesh_proto.MatchAllTag {
		spiffePrincipal := &rbac_config.Principal{
			Identifier: &rbac_config.Principal_Authenticated_{
				Authenticated: &rbac_config.Principal_Authenticated{
					PrincipalName: tls.ServiceSpiffeIDMatcher(mesh, service),
				},
			},
		}
		principals = append(principals, spiffePrincipal)
	}

	switch len(principals) {
	case 0:
		return &rbac_config.Principal{
			Identifier: &rbac_config.Principal_Any{
				Any: true,
			},
		}
	case 1:
		return principals[0]
	default:
		return &rbac_config.Principal{
			Identifier: &rbac_config.Principal_AndIds{ // many tags in selector mean that all of them have to match therefore AND
				AndIds: &rbac_config.Principal_Set{
					Ids: principals,
				},
			},
		}
	}
}

// kumaPrincipals can match any other tag than kuma.io/service tag
func kumaPrincipals(selector *mesh_proto.Selector) []*rbac_config.Principal {
	principals := []*rbac_config.Principal{}
	for tag, value := range selector.Match {
		if tag == mesh_proto.ServiceTag {
			continue // service tag is matched by spiffe principal
		}
		if value == mesh_proto.MatchAllTag {
			continue // '*' can match anything so no need to build principal for it
		}
		principal := &rbac_config.Principal{
			Identifier: &rbac_config.Principal_Authenticated_{
				Authenticated: &rbac_config.Principal_Authenticated{
					PrincipalName: tls.KumaIDMatcher(tag, value),
				},
			},
		}
		principals = append(principals, principal)
	}
	return principals
}
