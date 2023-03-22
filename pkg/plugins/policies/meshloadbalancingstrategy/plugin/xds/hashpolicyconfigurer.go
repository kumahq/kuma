package xds

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds/filters"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type HashPolicyConfigurer struct {
	HashPolicies []api.HashPolicy
}

func (c *HashPolicyConfigurer) Configure(route *envoy_route.Route) error {
	return filters.UpdateRouteAction(route, func(action *envoy_route.RouteAction) error {
		for _, hp := range c.HashPolicies {
			var envoyHP *envoy_route.RouteAction_HashPolicy
			switch hp.Type {
			case api.HeaderType:
				envoyHP = &envoy_route.RouteAction_HashPolicy{
					PolicySpecifier: &envoy_route.RouteAction_HashPolicy_Header_{
						Header: &envoy_route.RouteAction_HashPolicy_Header{
							HeaderName: hp.Header.Name,
						},
					},
				}
			case api.CookieType:
				var ttl *durationpb.Duration
				if hp.Cookie.TTL != nil {
					ttl = util_proto.Duration(hp.Cookie.TTL.Duration)
				}
				var path string
				if hp.Cookie.Path != nil {
					path = *hp.Cookie.Path
				}
				envoyHP = &envoy_route.RouteAction_HashPolicy{
					PolicySpecifier: &envoy_route.RouteAction_HashPolicy_Cookie_{
						Cookie: &envoy_route.RouteAction_HashPolicy_Cookie{
							Name: hp.Cookie.Name,
							Ttl:  ttl,
							Path: path,
						},
					},
				}
			case api.ConnectionType:
				envoyHP = &envoy_route.RouteAction_HashPolicy{
					PolicySpecifier: &envoy_route.RouteAction_HashPolicy_ConnectionProperties_{
						ConnectionProperties: &envoy_route.RouteAction_HashPolicy_ConnectionProperties{
							SourceIp: pointer.Deref(hp.Connection.SourceIP),
						},
					},
				}
			case api.QueryParameterType:
				envoyHP = &envoy_route.RouteAction_HashPolicy{
					PolicySpecifier: &envoy_route.RouteAction_HashPolicy_QueryParameter_{
						QueryParameter: &envoy_route.RouteAction_HashPolicy_QueryParameter{
							Name: hp.QueryParameter.Name,
						},
					},
				}
			case api.FilterStateType:
				envoyHP = &envoy_route.RouteAction_HashPolicy{
					PolicySpecifier: &envoy_route.RouteAction_HashPolicy_FilterState_{
						FilterState: &envoy_route.RouteAction_HashPolicy_FilterState{
							Key: hp.FilterState.Key,
						},
					},
				}
			}

			envoyHP.Terminal = pointer.Deref(hp.Terminal)
			action.HashPolicy = append(action.HashPolicy, envoyHP)
		}
		return nil
	})
}
