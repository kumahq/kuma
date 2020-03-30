package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_api_v2_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_filter_fault "github.com/envoyproxy/go-control-plane/envoy/config/filter/fault/v2"
	envoy_http_fault "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/fault/v2"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/xds/envoy/routes"
	"github.com/Kong/kuma/pkg/xds/envoy/tags"
)

func FaultInjection(faultInjection *mesh.FaultInjectionResource) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&FaultInjectionConfigurer{
			faultInjection: faultInjection,
		})
	})
}

type FaultInjectionConfigurer struct {
	faultInjection *mesh.FaultInjectionResource
}

func (f *FaultInjectionConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if f.faultInjection == nil {
		return nil
	}

	config := &envoy_http_fault.HTTPFault{
		Delay:   convertDelay(f.faultInjection.Spec.Conf.GetDelay()),
		Abort:   convertAbort(f.faultInjection.Spec.Conf.GetAbort()),
		Headers: createHeaders(f.faultInjection.Spec.SourceTags()),
	}

	rrl, err := convertResponseRateLimit(f.faultInjection.Spec.Conf.GetResponseBandwidth())
	if err != nil {
		return err
	}
	config.ResponseRateLimit = rrl

	pbst, err := ptypes.MarshalAny(config)
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

func createHeaders(tagSets []v1alpha1.SingleValueTagSet) (matchers []*envoy_api_v2_route.HeaderMatcher) {
	for _, t := range tagSets {
		matchers = append(matchers, &envoy_api_v2_route.HeaderMatcher{
			Name: routes.TagsHeaderName,
			HeaderMatchSpecifier: &envoy_api_v2_route.HeaderMatcher_SafeRegexMatch{
				SafeRegexMatch: &envoy_type_matcher.RegexMatcher{
					EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{
							MaxProgramSize: &wrappers.UInt32Value{Value: 500},
						},
					},
					Regex: tags.MatchingRegex(t),
				},
			},
		})
	}
	return
}

func convertDelay(delay *v1alpha1.FaultInjection_Conf_Delay) *envoy_filter_fault.FaultDelay {
	if delay == nil {
		return nil
	}
	return &envoy_filter_fault.FaultDelay{
		FaultDelaySecifier: &envoy_filter_fault.FaultDelay_FixedDelay{FixedDelay: delay.GetValue()},
		Percentage:         ConvertPercentage(delay.GetPercentage()),
	}
}

func convertAbort(abort *v1alpha1.FaultInjection_Conf_Abort) *envoy_http_fault.FaultAbort {
	if abort == nil {
		return nil
	}
	return &envoy_http_fault.FaultAbort{
		ErrorType:  &envoy_http_fault.FaultAbort_HttpStatus{HttpStatus: abort.HttpStatus.GetValue()},
		Percentage: ConvertPercentage(abort.GetPercentage()),
	}
}

func convertResponseRateLimit(responseBandwidth *v1alpha1.FaultInjection_Conf_ResponseBandwidth) (*envoy_filter_fault.FaultRateLimit, error) {
	if responseBandwidth == nil {
		return nil, nil
	}

	limitKbps, err := ConvertBandwidth(responseBandwidth.GetLimit())
	if err != nil {
		return nil, err
	}

	return &envoy_filter_fault.FaultRateLimit{
		LimitType:  limitKbps,
		Percentage: ConvertPercentage(responseBandwidth.GetPercentage()),
	}, nil
}
