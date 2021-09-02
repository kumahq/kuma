package v3

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	envoy_config_common_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/config/common/matcher/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_common_matching_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/matching/v3"
	envoy_extensions_filters_common_matcher_action_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/matcher/action/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// makeHostHeaderMatcher returns a matcher that skips the filter execution
// unless the given stringMatcher matches against the :authority header.
// This can be used to make HTTP filters conditional on the virtual host.
func makeHostHeaderMatcher(stringMatcher *envoy_type_matcher_v3.StringMatcher) *envoy_config_common_matcher_v3.Matcher {
	skipFilterAction := envoy_config_core_v3.TypedExtensionConfig{
		Name: "skip",
		TypedConfig: util_proto.MustMarshalAny(
			&envoy_extensions_filters_common_matcher_action_v3.SkipFilter{},
		),
	}

	hostHeaderInput := &envoy_config_core_v3.TypedExtensionConfig{
		Name: "match",
		TypedConfig: util_proto.MustMarshalAny(
			&envoy_type_matcher_v3.HttpRequestHeaderMatchInput{
				HeaderName: ":authority",
			},
		),
	}

	matchHostname := envoy_config_common_matcher_v3.Matcher_MatcherList_Predicate_SinglePredicate{
		Input: hostHeaderInput,
		Matcher: &envoy_config_common_matcher_v3.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
			ValueMatch: stringMatcher,
		},
	}

	list := &envoy_config_common_matcher_v3.Matcher_MatcherList{
		Matchers: []*envoy_config_common_matcher_v3.Matcher_MatcherList_FieldMatcher{
			{
				Predicate: &envoy_config_common_matcher_v3.Matcher_MatcherList_Predicate{
					MatchType: &envoy_config_common_matcher_v3.Matcher_MatcherList_Predicate_SinglePredicate_{
						SinglePredicate: &matchHostname,
					},
				},
				OnMatch: nil,
				// XXX(jpeach) This is fucked. Envoy requires an action here, but the only
				// action is "skip". We only want to skip the filter if there is *not* a match
				// (see the on_no_match action). So our only option is Envoy 1.19 where we can
				// have a not_matcher predicate.
			},
		}}

	return &envoy_config_common_matcher_v3.Matcher{
		MatcherType: &envoy_config_common_matcher_v3.Matcher_MatcherList_{
			MatcherList: list,
		},
		OnNoMatch: &envoy_config_common_matcher_v3.Matcher_OnMatch{
			OnMatch: &envoy_config_common_matcher_v3.Matcher_OnMatch_Action{
				Action: &skipFilterAction,
			},
		},
	}
}

func MatchFilterForHostname(hostname string, filter *envoy_hcm.HttpFilter) *envoy_hcm.HttpFilter {
	matcher := envoy_extensions_common_matching_v3.ExtensionWithMatcher{
		Matcher: makeHostHeaderMatcher(
			&envoy_type_matcher_v3.StringMatcher{
				MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
					Exact: hostname,
				},
			},
		),
		ExtensionConfig: &envoy_config_core_v3.TypedExtensionConfig{
			Name:        filter.GetName(),
			TypedConfig: filter.GetTypedConfig(),
		},
	}

	return &envoy_hcm.HttpFilter{
		Name: fmt.Sprintf("%s(%s)", filter.GetName(), hostname),
		ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
			TypedConfig: util_proto.MustMarshalAny(&matcher),
		},
	}
}

func MatchFilterForDomain(domain string, filter *envoy_hcm.HttpFilter) *envoy_hcm.HttpFilter {
	// Force domain into ".foo.com" format so that we can suffix match.
	if !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}

	matcher := envoy_extensions_common_matching_v3.ExtensionWithMatcher{
		Matcher: makeHostHeaderMatcher(
			&envoy_type_matcher_v3.StringMatcher{
				MatchPattern: &envoy_type_matcher_v3.StringMatcher_Suffix{
					Suffix: domain,
				},
			},
		),

		ExtensionConfig: &envoy_config_core_v3.TypedExtensionConfig{
			Name:        filter.GetName(),
			TypedConfig: filter.GetTypedConfig(),
		},
	}

	return &envoy_hcm.HttpFilter{
		Name: fmt.Sprintf("%s(%s)", filter.GetName(), domain),
		ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
			TypedConfig: util_proto.MustMarshalAny(&matcher),
		},
	}
}

func UpdateHTTPConnectionManager(filterChain *envoy_listener.FilterChain, updateFunc func(manager *envoy_hcm.HttpConnectionManager) error) error {
	return UpdateFilterConfig(filterChain, "envoy.filters.network.http_connection_manager", func(filterConfig proto.Message) error {
		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		if !ok {
			return NewUnexpectedFilterConfigTypeError(filterConfig, (*envoy_hcm.HttpConnectionManager)(nil))
		}
		return updateFunc(hcm)
	})
}

func UpdateTCPProxy(filterChain *envoy_listener.FilterChain, updateFunc func(*envoy_tcp.TcpProxy) error) error {
	return UpdateFilterConfig(filterChain, "envoy.filters.network.tcp_proxy", func(filterConfig proto.Message) error {
		tcpProxy, ok := filterConfig.(*envoy_tcp.TcpProxy)
		if !ok {
			return NewUnexpectedFilterConfigTypeError(filterConfig, (*envoy_tcp.TcpProxy)(nil))
		}
		return updateFunc(tcpProxy)
	})
}

func UpdateFilterConfig(filterChain *envoy_listener.FilterChain, filterName string, updateFunc func(proto.Message) error) error {
	for i, filter := range filterChain.Filters {
		if filter.Name == filterName {
			if filter.GetTypedConfig() == nil {
				return errors.Errorf("filters[%d]: config cannot be 'nil'", i)
			}

			msg, err := filter.GetTypedConfig().UnmarshalNew()
			if err != nil {
				return err
			}
			if err := updateFunc(msg); err != nil {
				return err
			}

			any, err := anypb.New(msg)
			if err != nil {
				return err
			}

			filter.ConfigType = &envoy_listener.Filter_TypedConfig{
				TypedConfig: any,
			}
		}
	}
	return nil
}

func NewUnexpectedFilterConfigTypeError(actual, expected proto.Message) error {
	return errors.Errorf("filter config has unexpected type: expected %T, got %T", expected, actual)
}

func ConvertPercentage(percentage *wrapperspb.DoubleValue) *envoy_type.FractionalPercent {
	const tenThousand = 10000
	const million = 1000000

	isInteger := func(f float64) bool {
		return math.Floor(f) == f
	}

	value := percentage.GetValue()
	if isInteger(value) {
		return &envoy_type.FractionalPercent{
			Numerator:   uint32(value),
			Denominator: envoy_type.FractionalPercent_HUNDRED,
		}
	}

	tenThousandTimes := tenThousand * value
	if isInteger(tenThousandTimes) {
		return &envoy_type.FractionalPercent{
			Numerator:   uint32(tenThousandTimes),
			Denominator: envoy_type.FractionalPercent_TEN_THOUSAND,
		}
	}

	return &envoy_type.FractionalPercent{
		Numerator:   uint32(math.Round(million * value)),
		Denominator: envoy_type.FractionalPercent_MILLION,
	}
}

var bandwidthRegex = regexp.MustCompile(`(\d*)\s?([gmk]?bps)`)

func ConvertBandwidthToKbps(bandwidth string) (uint64, error) {
	match := bandwidthRegex.FindStringSubmatch(bandwidth)
	value, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}

	units := match[2]

	var factor int // multiply on factor, to convert into kbps
	switch units {
	case "kbps":
		factor = 1
	case "mbps":
		factor = 1000
	case "gbps":
		factor = 1000000
	default:
		return 0, errors.New("unsupported unit type")
	}

	return uint64(factor * value), nil
}
