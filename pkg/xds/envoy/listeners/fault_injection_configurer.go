package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_api_v2_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_filter_fault "github.com/envoyproxy/go-control-plane/envoy/config/filter/fault/v2"
	envoy_http_fault "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/fault/v2"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/envoy/routes"
	"github.com/Kong/kuma/pkg/xds/envoy/tags"
)

func FaultInjection(faultInjection *mesh_proto.FaultInjection) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&FaultInjectionConfigurer{
			faultInjection: faultInjection,
		})
	})
}

type FaultInjectionConfigurer struct {
	faultInjection *mesh_proto.FaultInjection
}

func (f *FaultInjectionConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if f.faultInjection == nil {
		return nil
	}

	config := &envoy_http_fault.HTTPFault{
		Delay: convertDelay(f.faultInjection.Conf.GetDelay()),
		Abort: convertAbort(f.faultInjection.Conf.GetAbort()),
		Headers: []*envoy_api_v2_route.HeaderMatcher{
			createHeaders(f.faultInjection.SourceTags()),
		},
	}

	rrl, err := convertResponseRateLimit(f.faultInjection.Conf.GetResponseBandwidth())
	if err != nil {
		return err
	}
	config.ResponseRateLimit = rrl

	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append([]*envoy_hcm.HttpFilter{
			{
				Name: envoy_wellknown.Fault,
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			},
		}, manager.HttpFilters...)
		return nil
	})
}

func createHeaders(selectors []mesh_proto.SingleValueTagSet) *envoy_api_v2_route.HeaderMatcher {
	var selectorRegexs []string
	for _, selector := range selectors {
		selectorRegexs = append(selectorRegexs, tags.MatchingRegex(selector))
	}
	regexOR := tags.RegexOR(selectorRegexs...)

	return &envoy_api_v2_route.HeaderMatcher{
		Name: routes.TagsHeaderName,
		HeaderMatchSpecifier: &envoy_api_v2_route.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: &envoy_type_matcher.RegexMatcher{
				EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{
						MaxProgramSize: &wrappers.UInt32Value{Value: 500},
					},
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
