package envoy

import (
	"fmt"
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_error "github.com/Kong/kuma/pkg/util/error"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/rbac/v2"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	"github.com/gogo/protobuf/types"

	"github.com/envoyproxy/go-control-plane/envoy/type/matcher"
)

func createRbacFilter(listenerName string, permissions *mesh_core.TrafficPermissionResourceList) listener.Filter {
	rbacRule := createRbacRule(listenerName, permissions)
	rbacMarshalled, err := types.MarshalAny(rbacRule)
	util_error.MustNot(err)
	return listener.Filter{
		Name: "envoy.filters.network.rbac", // TODO(gszr): Change to util.RoleBasedAccessControl after go-control-plane update
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: rbacMarshalled,
		},
	}
}

func createRbacRule(listenerName string, permissions *mesh_core.TrafficPermissionResourceList) *rbac.RBAC {
	policies := make(map[string]*rbac_config.Policy, len(permissions.Items))
	for _, permission := range permissions.Items {
		policyName := fmt.Sprintf("%s.%s", permission.Meta.GetNamespace(), permission.Meta.GetName())
		policies[policyName] = createPolicy(permission)
	}

	return &rbac.RBAC{
		Rules: &rbac_config.RBAC{
			Action:   rbac_config.RBAC_ALLOW,
			Policies: policies,
		},
		StatPrefix: listenerName,
	}
}

func createPolicy(permission *mesh_core.TrafficPermissionResource) *rbac_config.Policy {
	principals := []*rbac_config.Principal{}
	// build principals list: one per sources/destinations rule
	for _, rule := range permission.Spec.Rules {
		for _, source := range rule.Sources {
			service := source.Match["service"]
			principal := &rbac_config.Principal{}
			if service == v1alpha1.MatchAllTag {
				principal.Identifier = &rbac_config.Principal_Any{
					Any: true,
				}
			} else {
				principal.Identifier = &rbac_config.Principal_Authenticated_{
					Authenticated: &rbac_config.Principal_Authenticated{
						PrincipalName: &matcher.StringMatcher{
							MatchPattern: &matcher.StringMatcher_Exact{
								Exact: fmt.Sprintf("spiffe://%s/%s", permission.Meta.GetMesh(), service),
							},
						},
					},
				}
			}
			principals = append(principals, principal)
		}
	}

	return &rbac_config.Policy{
		Permissions: []*rbac_config.Permission{
			{
				Rule: &rbac_config.Permission_Any{
					// todo(jakubdyszkiewicz) for now it matches on any destination port, which means that
					// if dataplane has two services ex. web, web-api. Allowing traffic on web will also work on web-api
					Any: true,
				},
			},
		},
		Principals: principals,
	}
}
