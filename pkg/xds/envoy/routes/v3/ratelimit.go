package v3

import (
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	ratelimit_api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
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
	headers := []*Headers{}
	for _, header := range rl.GetConf().GetHttp().GetOnRateLimit().GetHeaders() {
		headers = append(headers, &Headers{
			Key:    header.GetKey(),
			Value:  header.GetValue(),
			Append: header.GetAppend().Value,
		})
	}
	return &RateLimitConfiguration{
		Interval: rl.GetConf().GetHttp().GetInterval().AsDuration(),
		Requests: rl.GetConf().GetHttp().GetRequests(),
		OnRateLimit: &OnRateLimit{
			Status:  rl.GetConf().GetHttp().GetOnRateLimit().GetStatus().GetValue(),
			Headers: headers,
		},
	}
}

func RateLimitConfigurationFromPolicy(rl *ratelimit_api.LocalHTTP) *RateLimitConfiguration {
	headers := []*Headers{}
	if rl.OnRateLimit != nil {
		for _, h := range rl.OnRateLimit.Headers {
			header := &Headers{
				Key:   h.Key,
				Value: h.Value,
			}
			if h.Append != nil {
				header.Append = *h.Append
			}
			headers = append(headers, header)
		}
	}
	var status uint32
	if rl.OnRateLimit != nil && rl.OnRateLimit.Status != nil {
		status = *rl.OnRateLimit.Status
	}
	return &RateLimitConfiguration{
		Interval: rl.Interval.Duration,
		Requests: rl.Requests,
		OnRateLimit: &OnRateLimit{
			Status:  status,
			Headers: headers,
		},
	}
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
			responseHeaders = append(responseHeaders, &envoy_config_core_v3.HeaderValueOption{
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   h.Key,
					Value: h.Value,
				},
				Append: proto.Bool(h.Append),
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
