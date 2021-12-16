package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/golang/protobuf/ptypes/any"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

func NewRateLimitConfiguration(rlHttp *v1alpha1.RateLimit_Conf_Http) (*any.Any, error) {
	var status *envoy_type_v3.HttpStatus
	var responseHeaders []*envoy_config_core_v3.HeaderValueOption
	if rlHttp.GetOnRateLimit() != nil {
		status = &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(rlHttp.GetOnRateLimit().GetStatus().GetValue()),
		}
		responseHeaders = []*envoy_config_core_v3.HeaderValueOption{}
		for _, h := range rlHttp.GetOnRateLimit().GetHeaders() {
			responseHeaders = append(responseHeaders, &envoy_config_core_v3.HeaderValueOption{
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   h.GetKey(),
					Value: h.GetValue(),
				},
				Append: h.GetAppend(),
			})
		}
	}

	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
		Status:     status,
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens:     rlHttp.GetRequests(),
			TokensPerFill: proto.UInt32(rlHttp.GetRequests()),
			FillInterval:  rlHttp.GetInterval(),
		},
		FilterEnabled: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enabled",
		},
		FilterEnforced: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enforced",
		},
		ResponseHeadersToAdd: responseHeaders,
	}

	return proto.MarshalAnyDeterministic(config)
}
