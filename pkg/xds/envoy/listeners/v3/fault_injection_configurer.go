package v3

import (
	"time"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_filter_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/fault/v3"
	envoy_http_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type FaultInjectionConfiguration struct {
	SourceTags        []map[string]string
	Abort             *Abort
	Delay             *Delay
	ResponseBandwidth *ResponseBandwidth
}

type Abort struct {
	HttpStatus uint32
	Percentage float64
}

type Delay struct {
	Value      time.Duration
	Percentage float64
}

type ResponseBandwidth struct {
	Limit      string
	Percentage float64
}

type FaultInjectionConfigurer struct {
	FaultInjections []*FaultInjectionConfiguration
}

func FaultInjectionConfigurationFromProto(fi *mesh_proto.FaultInjection) *FaultInjectionConfiguration {
	faultInjection := FaultInjectionConfiguration{}
	if fi.GetConf() == nil {
		return &faultInjection
	}

	if fi.GetConf().GetAbort() != nil {
		faultInjection.Abort = &Abort{
			HttpStatus: fi.GetConf().GetAbort().GetHttpStatus().GetValue(),
			Percentage: fi.GetConf().GetAbort().GetPercentage().GetValue(),	
		}
	}
	if fi.GetConf().GetDelay() != nil {
		faultInjection.Delay = &Delay{
			Value: fi.GetConf().GetDelay().GetValue().AsDuration(),
			Percentage: fi.GetConf().GetDelay().GetPercentage().GetValue(),
		}
	}
	if fi.GetConf().GetResponseBandwidth() != nil {
		faultInjection.ResponseBandwidth = &ResponseBandwidth{
			Limit: fi.GetConf().GetResponseBandwidth().GetLimit().GetValue(),
			Percentage: fi.GetConf().GetResponseBandwidth().GetPercentage().GetValue(),
		}
	}
	sources := []map[string]string{}
	if fi.GetSources() != nil{
		for _, selector := range fi.GetSources() {
			sources = append(sources, selector.GetMatch())
		}
		faultInjection.SourceTags = sources
	}
	return &faultInjection
}

func (f *FaultInjectionConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	var httpFilters []*envoy_hcm.HttpFilter

	// Iterate over FaultInjections and generate the relevant HTTP filters.
	// We do assume that the FaultInjections resource is sorted, so the most
	// specific source matches come first.
	for _, fi := range f.FaultInjections {
		config := &envoy_http_fault.HTTPFault{
			Delay: convertDelay(fi.Delay),
			Abort: convertAbort(fi.Abort),
			Headers: []*envoy_route.HeaderMatcher{
				createHeaders(fi.SourceTags),
			},
		}

		rrl, err := convertResponseRateLimit(fi.ResponseBandwidth)
		if err != nil {
			return err
		}
		config.ResponseRateLimit = rrl

		pbst, err := util_proto.MarshalAnyDeterministic(config)
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

func createHeaders(selectors []map[string]string) *envoy_route.HeaderMatcher {
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

func convertDelay(delay *Delay) *envoy_filter_fault.FaultDelay {
	if delay == nil {
		return nil
	}
	return &envoy_filter_fault.FaultDelay{
		FaultDelaySecifier: &envoy_filter_fault.FaultDelay_FixedDelay{FixedDelay: util_proto.Duration(delay.Value)},
		Percentage:         ConvertPercentage(util_proto.Double(delay.Percentage)),
	}
}

func convertAbort(abort *Abort) *envoy_http_fault.FaultAbort {
	if abort == nil {
		return nil
	}
	return &envoy_http_fault.FaultAbort{
		ErrorType:  &envoy_http_fault.FaultAbort_HttpStatus{HttpStatus: abort.HttpStatus},
		Percentage: ConvertPercentage(util_proto.Double(abort.Percentage)),
	}
}

func convertResponseRateLimit(responseBandwidth *ResponseBandwidth) (*envoy_filter_fault.FaultRateLimit, error) {
	if responseBandwidth == nil {
		return nil, nil
	}

	limitKbps, err := ConvertBandwidthToKbps(responseBandwidth.Limit)
	if err != nil {
		return nil, err
	}

	return &envoy_filter_fault.FaultRateLimit{
		LimitType: &envoy_filter_fault.FaultRateLimit_FixedLimit_{
			FixedLimit: &envoy_filter_fault.FaultRateLimit_FixedLimit{
				LimitKbps: limitKbps,
			},
		},
		Percentage: ConvertPercentage(util_proto.Double(responseBandwidth.Percentage)),
	}, nil
}
