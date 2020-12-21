package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
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

func (c *controlPlaneIdCallbacks) OnStreamResponse(streamID int64, request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
	if c.id != "" {
		response.ControlPlane = &envoy_core.ControlPlane{
			Identifier: c.id,
		}
	}
}
