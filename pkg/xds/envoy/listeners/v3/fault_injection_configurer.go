package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_filter_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/fault/v3"
	envoy_http_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type FaultInjectionConfigurer struct {
	FaultInjections []*core_mesh.FaultInjectionResource
}

func (f *FaultInjectionConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	var httpFilters []*envoy_hcm.HttpFilter

	// Iterate over FaultInjections and generate the relevant HTTP filters.
	// We do assume that the FaultInjections resource is sorted, so the most
	// specific source matches come first.
	for _, fi := range f.FaultInjections {
		config := &envoy_http_fault.HTTPFault{
			Delay: convertDelay(fi.Spec.Conf.GetDelay()),
			Abort: convertAbort(fi.Spec.Conf.GetAbort()),
			Headers: []*envoy_route.HeaderMatcher{
				createHeaders(fi.Spec.SourceTags()),
			},
		}

		rrl, err := convertResponseRateLimit(fi.Spec.Conf.GetResponseBandwidth())
		if err != nil {
			return err
		}
		config.ResponseRateLimit = rrl

		pbst, err := proto.MarshalAnyDeterministic(config)
		if err != nil {
			return err
		}
		httpFilters = append(httpFilters, &envoy_hcm.HttpFilter{
			Name: "envoy.filters.http.fault",
			ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: pbst,
			},
		})
	}

	if len(httpFilters) == 0 {
		return nil
	}

	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append(manager.HttpFilters, httpFilters...)
		return nil
	})
}

func createHeaders(selectors []mesh_proto.SingleValueTagSet) *envoy_route.HeaderMatcher {
	var selectorRegexs []string
	for _, selector := range selectors {
		selectorRegexs = append(selectorRegexs, tags.MatchingRegex(selector))
	}
	regexOR := tags.RegexOR(selectorRegexs...)

	return &envoy_route.HeaderMatcher{
		Name: envoy_routes.TagsHeaderName,
		HeaderMatchSpecifier: &envoy_route.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: &envoy_type_matcher.RegexMatcher{
				EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
				},
				Regex: regexOR,
			},
		},
	}
}

func convertDelay(delay *mesh_proto.FaultInjection_Conf_Delay) *envoy_filter_fault.FaultDelay {
	if delay == nil {
		return nil
	}
	return &envoy_filter_fault.FaultDelay{
		FaultDelaySecifier: &envoy_filter_fault.FaultDelay_FixedDelay{FixedDelay: delay.GetValue()},
		Percentage:         ConvertPercentage(delay.GetPercentage()),
	}
}

func convertAbort(abort *mesh_proto.FaultInjection_Conf_Abort) *envoy_http_fault.FaultAbort {
	if abort == nil {
		return nil
	}
	return &envoy_http_fault.FaultAbort{
		ErrorType:  &envoy_http_fault.FaultAbort_HttpStatus{HttpStatus: abort.HttpStatus.GetValue()},
		Percentage: ConvertPercentage(abort.GetPercentage()),
	}
}

func convertResponseRateLimit(responseBandwidth *mesh_proto.FaultInjection_Conf_ResponseBandwidth) (*envoy_filter_fault.FaultRateLimit, error) {
	if responseBandwidth == nil {
		return nil, nil
	}

	limitKbps, err := ConvertBandwidthToKbps(responseBandwidth.GetLimit().GetValue())
	if err != nil {
		return nil, err
	}

	return &envoy_filter_fault.FaultRateLimit{
		LimitType: &envoy_filter_fault.FaultRateLimit_FixedLimit_{
			FixedLimit: &envoy_filter_fault.FaultRateLimit_FixedLimit{
				LimitKbps: limitKbps,
			},
		},
		Percentage: ConvertPercentage(responseBandwidth.GetPercentage()),
	}, nil
}
