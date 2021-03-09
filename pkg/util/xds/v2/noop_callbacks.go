package v2

import (
	"context"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

type NoopCallbacks struct {
}

func (c *NoopCallbacks) OnFetchRequest(context.Context, *envoy_api.DiscoveryRequest) error {
	return nil
}

func (c *NoopCallbacks) OnFetchResponse(*envoy_api.DiscoveryRequest, *envoy_api.DiscoveryResponse) {
}

func (c *NoopCallbacks) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (c *NoopCallbacks) OnStreamClosed(int64) {
}

func (c *NoopCallbacks) OnStreamRequest(int64, *envoy_api.DiscoveryRequest) error {
	return nil
}

func (c *NoopCallbacks) OnStreamResponse(int64, *envoy_api.DiscoveryRequest, *envoy_api.DiscoveryResponse) {
}

var _ envoy_xds.Callbacks = &NoopCallbacks{}
