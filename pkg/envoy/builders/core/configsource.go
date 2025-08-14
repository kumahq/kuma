package core

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func NewConfigSource() *Builder[envoy_core.ConfigSource] {
	return &Builder[envoy_core.ConfigSource]{}
}

func Sds() Configurer[envoy_core.ConfigSource] {
	return func(cs *envoy_core.ConfigSource) error {
		cs.ResourceApiVersion = envoy_core.ApiVersion_V3
		cs.ConfigSourceSpecifier = &envoy_core.ConfigSource_Ads{}
		return nil
	}
}

func ApiConfigSource(clusterName string) Configurer[envoy_core.ConfigSource] {
	return func(cs *envoy_core.ConfigSource) error {
		cs.ResourceApiVersion = envoy_core.ApiVersion_V3
		cs.ConfigSourceSpecifier = &envoy_core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &envoy_core.ApiConfigSource{
				ApiType: envoy_core.ApiConfigSource_GRPC,
				GrpcServices: []*envoy_core.GrpcService{
					{
						TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
							EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
								ClusterName: clusterName,
							},
						},
					},
				},
			},
		}
		return nil
	}
}
