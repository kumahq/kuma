package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"google.golang.org/protobuf/types/known/anypb"
)

func ToRateLimitConfiguraiton(rl *mesh_proto.RateLimit) *envoy_common.RateLimitConfiguration{
	headers := []*envoy_common.Headers{}
	for _, header := range rl.GetConf().GetHttp().GetOnRateLimit().GetHeaders(){
		headers = append(headers, &envoy_common.Headers{
			Key: header.GetKey(),
			Value: header.GetValue(),
			Append: header.GetAppend().Value,
		})
	}
	return &envoy_common.RateLimitConfiguration{
		Interval:  rl.GetConf().GetHttp().GetInterval().AsDuration(), 
		Requests: rl.GetConf().GetHttp().GetRequests(),
		OnRateLimit: &envoy_common.OnRateLimit{
			Status: rl.GetConf().GetHttp().GetOnRateLimit().GetStatus().GetValue(),
			Headers: headers,
		},
	}
}

func NewRateLimitConfiguration(rl *envoy_common.RateLimitConfiguration) (*anypb.Any, error) {
	var status *envoy_type_v3.HttpStatus
	var responseHeaders []*envoy_config_core_v3.HeaderValueOption
	if rl.OnRateLimit != nil {
		status = &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(rl.OnRateLimit.Status),
		}
		responseHeaders = []*envoy_config_core_v3.HeaderValueOption{}
		for _, h := range rl.OnRateLimit.Headers {
			responseHeaders = append(responseHeaders, &envoy_config_core_v3.HeaderValueOption{
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   h.Key,
					Value: h.Value,
				},
				Append: util_proto.Bool(h.Append),
			})
		}
	}

	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
		Status:     status,
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens:     rl.Requests,
			TokensPerFill: util_proto.UInt32(rl.Requests),
			FillInterval:  util_proto.Duration(rl.Interval),
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

	return util_proto.MarshalAnyDeterministic(config)
}
