package v3

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

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

type UnexpectedFilterConfigTypeError struct {
	actual   proto.Message
	expected proto.Message
}

func (e *UnexpectedFilterConfigTypeError) Error() string {
	return fmt.Sprintf("filter config has unexpected type: expected %T, got %T", e.expected, e.actual)
}

func NewUnexpectedFilterConfigTypeError(actual, expected proto.Message) error {
	return &UnexpectedFilterConfigTypeError{
		actual:   actual,
		expected: expected,
	}
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
