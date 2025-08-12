package dynconf

import (
	"crypto/sha256"
	"encoding/hex"

	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	http_connection_managerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	Origin       = "dynamic-config"
	ListenerName = "_kuma:dynamicconfig"
)

func AddConfigRoute(proxy *core_xds.Proxy, rs *core_xds.ResourceSet, name string, bytes []byte) error {
	var listener *envoy_listener.Listener
	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		if res.Origin == Origin {
			// Listener already exists only add the new route.
			listener = res.Resource.(*envoy_listener.Listener)
			break
		}
	}
	if listener == nil {
		nr, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion, ListenerName).
			Configure(envoy_listeners.PipeListener(core_xds.MeshMetricsDynamicConfigurationSocketName(proxy.Metadata.WorkDir))).
			Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
					Configure(
						envoy_listeners.DirectResponse(ListenerName, []v3.DirectResponseEndpoints{}, core_xds.LocalHostAddresses),
					),
			)).Build()
		listener = nr.(*envoy_listener.Listener)
		if err != nil {
			return err
		}
		r := &core_xds.Resource{
			Name:     nr.GetName(),
			Origin:   Origin,
			Resource: listener,
		}
		rs.Add(r)
	}
	hash := sha256.Sum256(bytes)
	err := v3.UpdateHTTPConnectionManager(listener.FilterChains[0], func(manager *http_connection_managerv3.HttpConnectionManager) error {
		routeConfig := manager.GetRouteConfig()
		routeConfig.VirtualHosts[0].Routes = append(routeConfig.VirtualHosts[0].Routes,
			&envoy_route.Route{
				Match: &envoy_route.RouteMatch{
					// Add a route for etag matching
					PathSpecifier: &envoy_route.RouteMatch_Path{
						Path: name,
					},
					Headers: []*envoy_route.HeaderMatcher{
						{
							Name: "If-None-Match",
							HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_Exact{
										Exact: hex.EncodeToString(hash[:]),
									},
								},
							},
						},
					},
				},
				Action: &envoy_route.Route_DirectResponse{
					DirectResponse: &envoy_route.DirectResponseAction{
						Status: 304,
					},
				},
			},
			&envoy_route.Route{
				Name: ListenerName + ":" + name,
				Match: &envoy_route.RouteMatch{
					PathSpecifier: &envoy_route.RouteMatch_Path{
						Path: name,
					},
				},
				ResponseHeadersToAdd: []*envoy_core_v3.HeaderValueOption{
					{
						Header: &envoy_core_v3.HeaderValue{
							Key:   "Etag",
							Value: hex.EncodeToString(hash[:]),
						},
					},
				},
				Action: &envoy_route.Route_DirectResponse{
					DirectResponse: &envoy_route.DirectResponseAction{
						Status: 200,
						Body: &envoy_core_v3.DataSource{
							Specifier: &envoy_core_v3.DataSource_InlineString{InlineString: string(bytes)},
						},
					},
				},
			},
		)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
