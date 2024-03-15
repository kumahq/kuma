package filters

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

type RequestMirrorConfigurer struct {
	requestMirror           api.RequestMirror
	backendRefToClusterName map[common_api.BackendRefHash]string
}

func NewRequestMirror(requestMirror api.RequestMirror, backendRefToClusterName map[common_api.BackendRefHash]string) *RequestMirrorConfigurer {
	return &RequestMirrorConfigurer{
		requestMirror:           requestMirror,
		backendRefToClusterName: backendRefToClusterName,
	}
}

func (f *RequestMirrorConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	return UpdateRouteAction(envoyRoute, func(action *envoy_route.RouteAction) error {
		clusterName, found := f.backendRefToClusterName[f.requestMirror.BackendRef.Hash()]
		if !found {
			// this should never happen because we create clusters for all backendRefs
			return errors.Errorf("could not find cluster for backendRef %s", f.requestMirror.BackendRef.Hash())
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
			Cluster:         clusterName,
		})
		return nil
	})
}
