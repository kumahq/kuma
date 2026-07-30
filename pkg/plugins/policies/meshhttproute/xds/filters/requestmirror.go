package filters

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
)

type RequestMirrorConfigurer struct {
	requestMirror api.RequestMirror
	split         envoy_common.Split
}

func NewRequestMirror(requestMirror api.RequestMirror, split envoy_common.Split) *RequestMirrorConfigurer {
	return &RequestMirrorConfigurer{
		requestMirror: requestMirror,
		split:         split,
	}
}

func (f *RequestMirrorConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	return UpdateRouteAction(envoyRoute, func(action *envoy_route.RouteAction) error {
		// no split means no cluster was created for the mirror backendRef, because
		// it points to a destination that doesn't exist or doesn't speak HTTP, so
		// there is nothing to mirror to
		if f.split == nil {
			return nil
		}

		var runtimeFraction *envoy_config_core.RuntimeFractionalPercent
		if f.requestMirror.Percentage != nil {
			decimal, err := common_api.NewDecimalFromIntOrString(*f.requestMirror.Percentage)
			if err != nil {
				return err
			}
			value, _ := decimal.Float64()
			runtimeFraction = &envoy_config_core.RuntimeFractionalPercent{
				DefaultValue: envoy_listeners.ConvertPercentage(util_proto.Double(value)),
			}
		}

		action.RequestMirrorPolicies = append(action.RequestMirrorPolicies, &envoy_route.RouteAction_RequestMirrorPolicy{
			RuntimeFraction: runtimeFraction,
			Cluster:         f.split.ClusterName(),
		})
		return nil
	})
}
