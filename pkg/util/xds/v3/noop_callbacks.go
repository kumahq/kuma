package v3

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

type NoopCallbacks struct {
}

func (c *NoopCallbacks) OnFetchRequest(context.Context, *envoy_sd.DiscoveryRequest) error {
	return nil
}

func (c *NoopCallbacks) OnFetchResponse(*envoy_sd.DiscoveryRequest, *envoy_sd.DiscoveryResponse) {
}

func (c *NoopCallbacks) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (c *NoopCallbacks) OnStreamClosed(int64) {
}

func (c *NoopCallbacks) OnStreamRequest(int64, *envoy_sd.DiscoveryRequest) error {
	return nil
}

func (c *NoopCallbacks) OnStreamResponse(context.Context, int64, *envoy_sd.DiscoveryRequest, *envoy_sd.DiscoveryResponse) {
}

func (c *NoopCallbacks) OnDeltaStreamOpen(ctx context.Context, i int64, s string) error {
	return nil
}

func (c *NoopCallbacks) OnDeltaStreamClosed(i int64) {
}

func (c *NoopCallbacks) OnStreamDeltaRequest(i int64, request *envoy_sd.DeltaDiscoveryRequest) error {
	return nil
}

func (c *NoopCallbacks) OnStreamDeltaResponse(i int64, request *envoy_sd.DeltaDiscoveryRequest, response *envoy_sd.DeltaDiscoveryResponse) {
}

var _ envoy_xds.Callbacks = &NoopCallbacks{}
