package xds

import (
	"errors"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_filter_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/fault/v3"
	envoy_http_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Configurer struct {
	FaultInjections []api.FaultInjectionConf
	From            core_xds.Subset
}

func (c *Configurer) ConfigureHttpListener(filterChain *envoy_listener.FilterChain) error {
	// Do not add new faults if old ones were applied
	for _, filter := range filterChain.Filters {
		if filter.Name == "envoy.filters.http.fault" {
			return nil
		}
	}

	httpRoutes := func(hcm *envoy_hcm.HttpConnectionManager) error {
		var fiFilters []*envoy_hcm.HttpFilter

		for _, fi := range c.FaultInjections {
			config := &envoy_http_fault.HTTPFault{
				Delay: c.convertDelay(fi.Delay),
				Abort: c.convertAbort(fi.Abort),
			}

			if len(c.From) > 0 {
				config.Headers = []*envoy_route.HeaderMatcher{
					c.createHeaders(c.From),
				}
			}

			rrl, err := c.convertResponseRateLimit(fi.ResponseBandwidth)
			if err != nil {
				return err
			}
			config.ResponseRateLimit = rrl

			pbst, err := util_proto.MarshalAnyDeterministic(config)
			if err != nil {
				return err
			}
			fiFilters = append(fiFilters, &envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.fault",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			})
		}
		// envoy.filters.http.router has to be the last filter
		filters := []*envoy_hcm.HttpFilter{}
		for _, filter := range hcm.HttpFilters {
			if filter.Name == "envoy.filters.http.router" {
				filters = append(filters, fiFilters...)
			}
			filters = append(filters, filter)
		}
		hcm.HttpFilters = filters
		return nil
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRoutes); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	return nil
}

func (c *Configurer) createHeaders(from core_xds.Subset) *envoy_route.HeaderMatcher {
	tagsMap := map[string]string{}
	for _, tag := range from {
		tagsMap[tag.Key] = tag.Value
	}

	var selectorRegexs []string
	selectorRegexs = append(selectorRegexs, tags.MatchingRegex(tagsMap))
	regexOR := tags.RegexOR(selectorRegexs...)

	return &envoy_route.HeaderMatcher{
		Name: tags.TagsHeaderName,
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

func (c *Configurer) convertDelay(delay *api.DelayConf) *envoy_filter_fault.FaultDelay {
	if delay == nil {
		return nil
	}
	return &envoy_filter_fault.FaultDelay{
		FaultDelaySecifier: &envoy_filter_fault.FaultDelay_FixedDelay{FixedDelay: util_proto.Duration(delay.Value.Duration)},
		Percentage:         listeners_v3.ConvertPercentage(util_proto.Double(float64(delay.Percentage))),
	}
}

func (c *Configurer) convertAbort(abort *api.AbortConf) *envoy_http_fault.FaultAbort {
	if abort == nil {
		return nil
	}
	return &envoy_http_fault.FaultAbort{
		ErrorType:  &envoy_http_fault.FaultAbort_HttpStatus{HttpStatus: uint32(abort.HttpStatus)},
		Percentage: listeners_v3.ConvertPercentage(util_proto.Double(float64(abort.Percentage))),
	}
}

func (c *Configurer) convertResponseRateLimit(responseBandwidth *api.ResponseBandwidthConf) (*envoy_filter_fault.FaultRateLimit, error) {
	if responseBandwidth == nil {
		return nil, nil
	}

	limitKbps, err := listeners_v3.ConvertBandwidthToKbps(responseBandwidth.Limit)
	if err != nil {
		return nil, err
	}

	return &envoy_filter_fault.FaultRateLimit{
		LimitType: &envoy_filter_fault.FaultRateLimit_FixedLimit_{
			FixedLimit: &envoy_filter_fault.FaultRateLimit_FixedLimit{
				LimitKbps: limitKbps,
			},
		},
		Percentage: listeners_v3.ConvertPercentage(util_proto.Double(float64(responseBandwidth.Percentage))),
	}, nil
}
