package server

import (
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"github.com/kumahq/kuma/pkg/kds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

// We are using go-control-plane's server and cache for KDS exchange.
// We are setting TypeURL for DeltaDiscoveryRequest/DeltaDiscoveryResponse for our resource name like "TrafficRoute" / "Mesh" etc.
// but the actual resource which we are sending is kuma.mesh.v1alpha1.KumaResource
//
// The function which is marshaling DeltaDiscoveryResponse
// func (r *RawDeltaResponse) GetDeltaDiscoveryResponse() (*discovery.DeltaDiscoveryResponse, error)
// Ignores the TypeURL from marshaling operation and overrides it with TypeURL of the request.
// If we pass wrong TypeURL in envoy_api.DeltaDiscoveryResponse#Resources we won't be able to unmarshall it, therefore we need to adjust the type.
type typeAdjustCallbacks struct {
	util_xds_v3.NoopCallbacks
}

func (c *typeAdjustCallbacks) OnStreamDeltaResponse(streamID int64, req *envoy_sd.DeltaDiscoveryRequest, resp *envoy_sd.DeltaDiscoveryResponse) {
	for _, res := range resp.GetResources() {
		res.Resource.TypeUrl = kds.KumaResource
	}
}
