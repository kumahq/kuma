package v3

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

// controlPlaneIdCallbacks adds Control Plane ID to the DiscoveryResponse
type controlPlaneIdCallbacks struct {
	NoopCallbacks
	id string
}

var _ envoy_xds.Callbacks = &controlPlaneIdCallbacks{}

func NewControlPlaneIdCallbacks(id string) envoy_xds.Callbacks {
	return &controlPlaneIdCallbacks{
		id: id,
	}
}

func (c *controlPlaneIdCallbacks) OnStreamResponse(ctx context.Context, streamID int64, request *envoy_discovery.DiscoveryRequest, response *envoy_discovery.DiscoveryResponse) {
	if c.id != "" {
		response.ControlPlane = &envoy_core.ControlPlane{
			Identifier: c.id,
		}
	}
}
