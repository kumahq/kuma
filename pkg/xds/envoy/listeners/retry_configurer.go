package listeners

import (
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

const (
	HttpRetryOn5XX                  = "5XX"
	HttpRetryOnRetriableStatusCodes = "retriable-status-codes"
	GrpcRetryOnAll                  = "cancelled,deadline-exceeded,internal," +
		"resource-exhausted,unavailable"
)

func Retry(retry *core_mesh.RetryResource, protocol core_mesh.Protocol) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if retry != nil {
			config.Add(&RetryConfigurer{
				retry:    retry,
				protocol: protocol,
			})
		}
	})
}

type RetryConfigurer struct {
	retry    *core_mesh.RetryResource
	protocol core_mesh.Protocol
}

func genGrpcRetryPolicy(
	conf *mesh_proto.Retry_Conf_Grpc,
) *envoy_route.RetryPolicy {
	if conf == nil {
		return nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn:       GrpcRetryOnAll,
		PerTryTimeout: conf.PerTryTimeout,
	}

	if conf.NumRetries != nil {
		policy.NumRetries = &wrappers.UInt32Value{
			Value: conf.NumRetries.Value,
		}
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
		RetryOn:       HttpRetryOn5XX,
		PerTryTimeout: conf.PerTryTimeout,
	}

	if conf.NumRetries != nil {
		policy.NumRetries = &wrappers.UInt32Value{
			Value: conf.NumRetries.Value,
		}
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

	return &policy
}

func (c *RetryConfigurer) Configure(
	filterChain *envoy_listener.FilterChain,
) error {
	if c.retry == nil {
		return nil
	}

	updateFunc := func(manager *envoy_hcm.HttpConnectionManager) error {
		var policy *envoy_route.RetryPolicy

		switch c.protocol {
		case "http":
			policy = genHttpRetryPolicy(c.retry.Spec.Conf.GetHttp())
		case "grpc":
			policy = genGrpcRetryPolicy(c.retry.Spec.Conf.GetGrpc())
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
