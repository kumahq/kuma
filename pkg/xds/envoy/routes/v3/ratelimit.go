package v3

import (
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type RateLimitConfiguration struct {
	Interval    time.Duration
	Requests    uint32
	OnRateLimit *OnRateLimit
}

type OnRateLimit struct {
	Status  uint32
	Headers []*Headers
}

type Headers struct {
	Key    string
	Value  string
	Append bool
}

func RateLimitConfigurationFromProto(rl *mesh_proto.RateLimit) *RateLimitConfiguration {
	if rl.GetConf() == nil || rl.GetConf().GetHttp() == nil {
		return &RateLimitConfiguration{}
	}
	rateLimit := &RateLimitConfiguration{
		Interval:    rl.GetConf().GetHttp().GetInterval().AsDuration(),
		Requests:    rl.GetConf().GetHttp().GetRequests(),
		OnRateLimit: &OnRateLimit{},
	}
	if rl.GetConf().GetHttp().GetOnRateLimit() != nil {
		headers := []*Headers{}
		for _, h := range rl.GetConf().GetHttp().GetOnRateLimit().GetHeaders() {
			header := &Headers{
				Key:   h.GetKey(),
				Value: h.GetValue(),
			}
			if h.GetAppend() != nil {
				header.Append = h.GetAppend().Value
			}
			headers = append(headers, header)
		}
		rateLimit.OnRateLimit.Headers = headers

		if rl.GetConf().GetHttp().GetOnRateLimit().GetStatus() != nil {
			rateLimit.OnRateLimit.Status = rl.GetConf().GetHttp().GetOnRateLimit().GetStatus().GetValue()
		}
	}
	return rateLimit
}

func NewRateLimitConfiguration(rlHttp *RateLimitConfiguration) (*anypb.Any, error) {
	var status *envoy_type_v3.HttpStatus
	var responseHeaders []*envoy_config_core_v3.HeaderValueOption
	if rlHttp.OnRateLimit != nil {
		if rlHttp.OnRateLimit.Status != 0 {
			status = &envoy_type_v3.HttpStatus{
				Code: envoy_type_v3.StatusCode(rlHttp.OnRateLimit.Status),
			}
		}
		responseHeaders = []*envoy_config_core_v3.HeaderValueOption{}
		for _, h := range rlHttp.OnRateLimit.Headers {
			appendAction := envoy_config_core_v3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
			if h.Append {
				appendAction = envoy_config_core_v3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
			}
			responseHeaders = append(responseHeaders, &envoy_config_core_v3.HeaderValueOption{
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   h.Key,
					Value: h.Value,
				},
				AppendAction: appendAction,
			})
		}
	}

	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
		Status:     status,
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens:     rlHttp.Requests,
			TokensPerFill: proto.UInt32(rlHttp.Requests),
			FillInterval:  proto.Duration(rlHttp.Interval),
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
