package server

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/kumahq/kuma/pkg/kds/definitions"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
)

// We are using go-control-plane's server and cache for KDS exchange.
// We are setting TypeURL for DiscoveryRequest/DiscoveryResponse for our resource name like "TrafficRoute" / "Mesh" etc.
// but the actual resource which we are sending is kuma.mesh.v1alpha1.KumaResource
//
// The function which is marshalling DiscoveryResponse
// func (r *RawResponse) GetDiscoveryResponse() (*discovery.DiscoveryResponse, error)
// Ignores the TypeURL from marshalling operation and overrides it with TypeURL of the request.
// If we pass wrong TypeURL in envoy_api.DiscoveryResponse#Resources we won't be able to unmarshall it, therefore we need to adjust the type.
type typeAdjustCallbacks struct {
	util_xds_v2.NoopCallbacks
}

func (c *typeAdjustCallbacks) OnStreamResponse(streamID int64, req *envoy_api.DiscoveryRequest, resp *envoy_api.DiscoveryResponse) {
	for _, res := range resp.Resources {
		res.TypeUrl = definitions.KumaResource
	}
}
