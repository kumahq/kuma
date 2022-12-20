package util

import (
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

func RateLimitConfigurationFromPolicy(rl *api.LocalHTTP) *envoy_routes.RateLimitConfiguration {
	headers := []*envoy_routes.Headers{}
	if rl.OnRateLimit != nil {
		for _, h := range rl.OnRateLimit.Headers {
			header := &envoy_routes.Headers{
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
	return &envoy_routes.RateLimitConfiguration{
		Interval: rl.Interval.Duration,
		Requests: rl.Requests,
		OnRateLimit: &envoy_routes.OnRateLimit{
			Status:  status,
			Headers: headers,
		},
	}
}
