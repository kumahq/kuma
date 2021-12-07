package v3

import (
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

const (
	HttpRetryOnDefault = "gateway-error,connect-failure," +
		"refused-stream"
	HttpRetryOnRetriableStatusCodes = "connect-failure,refused-stream," +
		"retriable-status-codes"
	GrpcRetryOnDefault = "cancelled,connect-failure," +
		"gateway-error,refused-stream,reset,resource-exhausted,unavailable"
)

type RetryConfigurer struct {
	Retry    *core_mesh.RetryResource
	Protocol core_mesh.Protocol
}

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

	if conf.BackOff != nil {
		policy.RetryBackOff = &envoy_route.RetryPolicy_RetryBackOff{
			BaseInterval: conf.BackOff.BaseInterval,
			MaxInterval:  conf.BackOff.MaxInterval,
		}
	}

	if conf.RetryOn != nil {
		var retryOn []string

		for _, item := range conf.RetryOn {
			// As `retryOn` is an enum value, and as in protobuf we can't use
			// hyphens we are using underscores instead, but as envoy expect
			// values with hyphens it's being changed here
			retryOn = append(retryOn, strings.ReplaceAll(item.String(), "_", "-"))
		}

		policy.RetryOn = strings.Join(retryOn, ",")
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

	if conf.RetriableStatusCodes != nil {
		policy.RetryOn = HttpRetryOnRetriableStatusCodes
		policy.RetriableStatusCodes = conf.RetriableStatusCodes
	}

	for _, method := range conf.GetRetriableMethods() {
		if method == mesh_proto.HttpMethod_NONE {
			continue
		}

		policy.RetriableRequestHeaders = append(policy.RetriableRequestHeaders,
			&envoy_route.HeaderMatcher{
				Name:                 ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_ExactMatch{ExactMatch: method.String()},
				InvertMatch:          false,
			})
	}

	return &policy
}

func (c *RetryConfigurer) Configure(
	filterChain *envoy_listener.FilterChain,
) error {
	if c.Retry == nil {
		return nil
	}

	updateFunc := func(manager *envoy_hcm.HttpConnectionManager) error {
		var policy *envoy_route.RetryPolicy

		switch c.Protocol {
		case "http":
			policy = genHttpRetryPolicy(c.Retry.Spec.Conf.GetHttp())
		case "grpc":
			policy = genGrpcRetryPolicy(c.Retry.Spec.Conf.GetGrpc())
		default:
			return nil
		}

		for _, virtualHost := range manager.GetRouteConfig().VirtualHosts {
			virtualHost.RetryPolicy = policy
		}

		return nil
	}

	return UpdateHTTPConnectionManager(filterChain, updateFunc)
}
