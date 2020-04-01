package listeners

import (
	"math"
	"regexp"
	"strconv"

	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

func UpdateHTTPConnectionManager(filterChain *envoy_listener.FilterChain, updateFunc func(manager *envoy_hcm.HttpConnectionManager) error) error {
	return UpdateFilterConfig(filterChain, envoy_wellknown.HTTPConnectionManager, func(filterConfig proto.Message) error {
		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		if !ok {
			return NewUnexpectedFilterConfigTypeError(filterConfig, (*envoy_hcm.HttpConnectionManager)(nil))
		}
		return updateFunc(hcm)
	})
}

func UpdateTCPProxy(filterChain *envoy_listener.FilterChain, updateFunc func(*envoy_tcp.TcpProxy) error) error {
	return UpdateFilterConfig(filterChain, envoy_wellknown.TCPProxy, func(filterConfig proto.Message) error {
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

			var dany ptypes.DynamicAny
			if err := ptypes.UnmarshalAny(filter.GetTypedConfig(), &dany); err != nil {
				return err
			}
			if err := updateFunc(dany.Message); err != nil {
				return err
			}

			pbst, err := ptypes.MarshalAny(dany.Message)
			if err != nil {
				return err
			}

			filter.ConfigType = &envoy_listener.Filter_TypedConfig{
				TypedConfig: pbst,
			}
		}
	}
	return nil
}

func NewUnexpectedFilterConfigTypeError(actual, expected proto.Message) error {
	return errors.Errorf("filter config has unexpected type: expected %T, got %T", expected, actual)
}

func ConvertPercentage(percentage *wrappers.DoubleValue) *envoy_type.FractionalPercent {
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
