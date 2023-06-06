package common

import (
	"strings"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

const (
	HttpRetryOnDefault = "gateway-error,connect-failure," +
		"refused-stream"
	HttpRetryOnRetriableStatusCodes = "retriable-status-codes"
	GrpcRetryOnDefault              = "cancelled,connect-failure," +
		"gateway-error,refused-stream,reset,resource-exhausted,unavailable"
)

func genGrpcRetryPolicy(
	conf *mesh_proto.Retry_Conf_Grpc,
) *envoy_route.RetryPolicy {
	if conf == nil {
		return nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn:       GrpcRetryOnDefault,
		PerTryTimeout: conf.PerTryTimeout,
	}

	if conf.NumRetries != nil {
		policy.NumRetries = util_proto.UInt32(conf.NumRetries.Value)
	}

	retryOn := GrpcRetryOn(conf.GetRetryOn())
	if retryOn != "" {
		policy.RetryOn = retryOn
	}

	if conf.BackOff != nil {
		policy.RetryBackOff = &envoy_route.RetryPolicy_RetryBackOff{
			BaseInterval: conf.BackOff.BaseInterval,
			MaxInterval:  conf.BackOff.MaxInterval,
		}
	}

	return &policy
}

func genHttpRetryPolicy(
	conf *mesh_proto.Retry_Conf_Http,
) *envoy_route.RetryPolicy {
	if conf == nil {
		return nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn:       HttpRetryOnDefault,
		PerTryTimeout: conf.PerTryTimeout,
	}

	if conf.NumRetries != nil {
		policy.NumRetries = util_proto.UInt32(conf.NumRetries.Value)
	}

	if conf.BackOff != nil {
		policy.RetryBackOff = &envoy_route.RetryPolicy_RetryBackOff{
			BaseInterval: conf.BackOff.BaseInterval,
			MaxInterval:  conf.BackOff.MaxInterval,
		}
	}

	retryOn := HttpRetryOn(conf.GetRetryOn())
	if retryOn != "" {
		policy.RetryOn = retryOn
	}

	if conf.RetriableStatusCodes != nil {
		policy.RetriableStatusCodes = conf.RetriableStatusCodes
		policy.RetryOn = ensureRetriableStatusCodes(policy.RetryOn)
	}

	for _, method := range conf.GetRetriableMethods() {
		if method == mesh_proto.HttpMethod_NONE {
			continue
		}

		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: method.String(),
			},
		}
		policy.RetriableRequestHeaders = append(policy.RetriableRequestHeaders,
			&envoy_route.HeaderMatcher{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
				},
				InvertMatch: false,
			})
	}

	return &policy
}

func HttpRetryOn(conf []mesh_proto.HttpRetryOn) string {
	if conf == nil {
		return ""
	}
	var retryOn []string

	for _, item := range conf {
		key := item.String()
		// Protobuf fields cannot start with a number so convert to the correct
		// value before appending
		if key == "all_5xx" {
			key = "5xx"
		}
		// As `retryOn` is an enum value, and as in protobuf we can't use
		// hyphens we are using underscores instead, but as envoy expect
		// values with hyphens it's being changed here
		retryOn = append(retryOn, strings.ReplaceAll(key, "_", "-"))
	}
	return strings.Join(retryOn, ",")
}

func GrpcRetryOn(conf []mesh_proto.Retry_Conf_Grpc_RetryOn) string {
	if len(conf) == 0 {
		return ""
	}
	var retryOn []string

	for _, item := range conf {
		// As `retryOn` is an enum value, and as in protobuf we can't use
		// hyphens we are using underscores instead, but as envoy expect
		// values with hyphens it's being changed here
		retryOn = append(retryOn, strings.ReplaceAll(item.String(), "_", "-"))
	}

	return strings.Join(retryOn, ",")
}

func RetryConfig(retry *core_mesh.RetryResource, protocol core_mesh.Protocol) *envoy_route.RetryPolicy {
	if retry == nil {
		return nil
	}
	var policy *envoy_route.RetryPolicy
	switch protocol {
	case "http":
		policy = genHttpRetryPolicy(retry.Spec.Conf.GetHttp())
	case "grpc":
		policy = genGrpcRetryPolicy(retry.Spec.Conf.GetGrpc())
	default:
		return nil
	}
	return policy	
}

func ensureRetriableStatusCodes(policyRetryOn string) string {
	policyRetrySplit := strings.Split(policyRetryOn, ",")
	seenRetriable := false
	for _, r := range policyRetrySplit {
		if r == HttpRetryOnRetriableStatusCodes {
			seenRetriable = true
			break
		}
	}
	if !seenRetriable {
		policyRetrySplit = append(policyRetrySplit, HttpRetryOnRetriableStatusCodes)
	}
	policyRetryOn = strings.Join(policyRetrySplit, ",")
	return policyRetryOn
}
