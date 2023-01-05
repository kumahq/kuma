package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"net/http"
	"strconv"
	"strings"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

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

type Configurer struct {
	Retry    *api.Conf
	Protocol core_mesh.Protocol
}

func genGrpcRetryPolicy(conf *api.GRPC) *envoy_route.RetryPolicy {
	if conf == nil {
		return nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn: GrpcRetryOnDefault,
	}

	if conf.PerTryTimeout != nil {
		policy.PerTryTimeout = util_proto.Duration(conf.PerTryTimeout.Duration)

	}

	if conf.NumRetries != nil {
		policy.NumRetries = util_proto.UInt32(*conf.NumRetries)
	}

	retryOn := GrpcRetryOn(conf.RetryOn)
	if retryOn != "" {
		policy.RetryOn = retryOn
	}

	if conf.BackOff != nil {
		policy.RetryBackOff = &envoy_route.RetryPolicy_RetryBackOff{}

		if conf.BackOff.BaseInterval != nil {
			policy.RetryBackOff.BaseInterval = util_proto.Duration(conf.BackOff.BaseInterval.Duration)
		}
		if conf.BackOff.MaxInterval != nil {
			policy.RetryBackOff.MaxInterval = util_proto.Duration(conf.BackOff.MaxInterval.Duration)
		}
	}

	return &policy
}

func genHttpRetryPolicy(conf *api.HTTP) *envoy_route.RetryPolicy {
	if conf == nil {
		return nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn: HttpRetryOnDefault,
	}

	if conf.PerTryTimeout != nil {
		policy.PerTryTimeout = util_proto.Duration(conf.PerTryTimeout.Duration)
	}

	if conf.NumRetries != nil {
		policy.NumRetries = util_proto.UInt32(*conf.NumRetries)
	}

	if conf.BackOff != nil {
		retryBackOff := &envoy_route.RetryPolicy_RetryBackOff{}

		if conf.BackOff.BaseInterval != nil {
			retryBackOff.BaseInterval = util_proto.Duration(conf.BackOff.BaseInterval.Duration)
		}

		if conf.BackOff.MaxInterval != nil {
			retryBackOff.MaxInterval = util_proto.Duration(conf.BackOff.MaxInterval.Duration)
		}
	}

	retryOn, retriableStatusCodes, retriableMethods := splitRetryOn(conf.RetryOn)
	if retryOn != "" {
		policy.RetryOn = retryOn
	}
	if len(retriableStatusCodes) != 0 {
		policy.RetriableStatusCodes = retriableStatusCodes
		policy.RetryOn = ensureRetriableStatusCodes(policy.RetryOn)
	}

	for _, method := range retriableMethods {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: method,
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

func splitRetryOn(conf *[]api.HTTPRetryOn) (string, []uint32, []string) {
	if conf == nil {
		return "", []uint32{}, nil
	}
	var retryOn []string
	var retriableStatusCodes []uint32
	var retriableMethods []string

	for _, item := range *conf {
		key := string(item)
		// As `retryOn` is an enum value, and as in protobuf we can't use
		// hyphens we are using underscores instead, but as envoy expect
		// values with hyphens it's being changed here
		statusCode, err := strconv.Atoi(key)
		if err == nil && http.StatusText(statusCode) != "" {
			retriableStatusCodes = append(retriableStatusCodes, uint32(statusCode))
		} else if strings.HasPrefix(key, "HTTP_METHOD_") {
			method := strings.TrimPrefix(key, "HTTP_METHOD_")
			retriableMethods = append(retriableMethods, method)
		} else {
			retryOn = append(retryOn, strings.ReplaceAll(string(item), "_", "-"))
		}

	}
	return strings.Join(retryOn, ","), retriableStatusCodes, retriableMethods
}

func GrpcRetryOn(conf *[]api.GRPCRetryOn) string {
	if conf == nil || len(*conf) == 0 {
		return ""
	}
	var retryOn []string

	for _, item := range *conf {
		// As `retryOn` is an enum value, and as in protobuf we can't use
		// hyphens we are using underscores instead, but as envoy expect
		// values with hyphens it's being changed here
		retryOn = append(retryOn, strings.ReplaceAll(string(item), "_", "-"))
	}

	return strings.Join(retryOn, ",")
}

func (c *Configurer) Configure(
	filterChain *envoy_listener.FilterChain,
) error {
	if c.Retry == nil {
		return nil
	}

	updateFunc := func(manager *envoy_hcm.HttpConnectionManager) error {
		var policy *envoy_route.RetryPolicy

		switch c.Protocol {
		case "http":
			policy = genHttpRetryPolicy(c.Retry.HTTP)
		case "grpc":
			policy = genGrpcRetryPolicy(c.Retry.GRPC)
		default:
			return nil
		}

		for _, virtualHost := range manager.GetRouteConfig().VirtualHosts {
			virtualHost.RetryPolicy = policy
		}

		return nil
	}

	return v3.UpdateHTTPConnectionManager(filterChain, updateFunc)
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
