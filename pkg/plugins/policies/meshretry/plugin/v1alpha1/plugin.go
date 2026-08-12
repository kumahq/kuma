package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshretry/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshretry/plugin/xds"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshRetryResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshRetryType]
	if !ok {
		return nil
	}

	rctx := outbound.RootContext[api.Conf](ctx.Mesh.Resource, policies.ToRules.ResourceRules)
	for _, r := range util_slices.Filter(rs.List(), core_xds.HasAssociatedServiceResource) {
		svcCtx := rctx.
			WithID(kri.NoSectionName(r.ResourceOrigin)).
			WithID(r.ResourceOrigin)

		if err := applyToRealResource(svcCtx, r); err != nil {
			return err
		}
	}
	return nil
}

func applyToRealResource(rctx *outbound.ResourceContext[api.Conf], r *core_xds.Resource) error {
	if envoyResource, ok := r.Resource.(*envoy_listener.Listener); ok {
		configurer := plugin_xds.Configurer{Conf: rctx.Conf(), Protocol: r.Protocol}
		if err := configurer.ConfigureListener(envoyResource); err != nil {
			return err
		}

		for _, fc := range envoyResource.FilterChains {
			if err := listeners_v3.UpdateHTTPConnectionManager(fc, func(hcm *envoy_hcm.HttpConnectionManager) error {
				for _, vh := range hcm.GetRouteConfig().VirtualHosts {
					for _, route := range vh.Routes {
						if !kri.IsValid(route.Name) {
							continue
						}

						id, err := kri.FromString(route.Name)
						if err != nil {
							return err
						}

						routeCtx := rctx.
							WithID(kri.NoSectionName(id)).
							WithID(id)

						if err := configureRoute(routeCtx, route, r.Protocol); err != nil {
							return err
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func configureRoute(rctx *outbound.ResourceContext[api.Conf], route *envoy_route.Route, protocol core_meta.Protocol) error {
	policy, err := plugin_xds.GetRouteRetryConfig(pointer.To(rctx.Conf()), protocol)
	if err != nil {
		return err
	}
	if policy == nil {
		return nil
	}

	if a, ok := route.GetAction().(*envoy_route.Route_Route); ok {
		a.Route.RetryPolicy = policy
	}

	return nil
}
