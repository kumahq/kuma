package xds

import (
	"context"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

type controlPlanIdCallbacks struct {
	id string
}

func (c *controlPlanIdCallbacks) OnFetchRequest(ctx context.Context, request *envoy_api.DiscoveryRequest) error {
	return nil
}

func (c *controlPlanIdCallbacks) OnFetchResponse(request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
}

func (c *controlPlanIdCallbacks) OnStreamOpen(ctx context.Context, streamID int64, s string) error {
	return nil
}

func (c *controlPlanIdCallbacks) OnStreamClosed(i int64) {
}

func (c *controlPlanIdCallbacks) OnStreamRequest(i int64, request *envoy_api.DiscoveryRequest) error {
	return nil
}

func (c *controlPlanIdCallbacks) OnStreamResponse(streamID int64, request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
	if c.id != "" {
		response.ControlPlane = &envoy_core.ControlPlane{
			Identifier: c.id,
		}
	}
}

func NewControlPlanIdCallbacks(id string) envoy_xds.Callbacks {
	return &controlPlanIdCallbacks{
		id: id,
	}
}

var _ envoy_xds.Callbacks = &controlPlanIdCallbacks{}
