package server

import (
	"context"

	"github.com/pkg/errors"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

func DefaultSecretDiscoveryHandler() SecretDiscoveryHandler {
	return SecretDiscoveryHandlerFunc(func(ctx context.Context, req envoy.DiscoveryRequest) (envoy_cache.Response, error) {
		block := make(chan struct{})
		// block indefinitely
		<-block
		return envoy_cache.Response{}, errors.New("not implemented")
	})
}

type SecretDiscoveryHandlerFunc func(ctx context.Context, req envoy.DiscoveryRequest) (envoy_cache.Response, error)

func (f SecretDiscoveryHandlerFunc) Handle(ctx context.Context, req envoy.DiscoveryRequest) (envoy_cache.Response, error) {
	return f(ctx, req)
}
