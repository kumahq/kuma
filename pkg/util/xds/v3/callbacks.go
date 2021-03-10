package v3

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	_struct "github.com/golang/protobuf/ptypes/struct"

	"github.com/kumahq/kuma/pkg/util/xds"
)

type adapterCallbacks struct {
	callbacks xds.Callbacks
}

// AdaptCallbacks translate Kuma callbacks to real go-control-plane Callbacks
func AdaptCallbacks(callbacks xds.Callbacks) envoy_xds.Callbacks {
	return &adapterCallbacks{
		callbacks: callbacks,
	}
}

var _ envoy_xds.Callbacks = &adapterCallbacks{}

func (a *adapterCallbacks) OnFetchRequest(ctx context.Context, request *envoy_sd.DiscoveryRequest) error {
	return nil
}

func (a *adapterCallbacks) OnFetchResponse(request *envoy_sd.DiscoveryRequest, response *envoy_sd.DiscoveryResponse) {
}

func (a *adapterCallbacks) OnStreamOpen(ctx context.Context, streamID int64, typeURL string) error {
	return a.callbacks.OnStreamOpen(ctx, streamID, typeURL)
}

func (a *adapterCallbacks) OnStreamClosed(streamID int64) {
	a.callbacks.OnStreamClosed(streamID)
}

func (a *adapterCallbacks) OnStreamRequest(streamID int64, request *envoy_sd.DiscoveryRequest) error {
	return a.callbacks.OnStreamRequest(streamID, &discoveryRequest{request})
}

func (a *adapterCallbacks) OnStreamResponse(streamID int64, request *envoy_sd.DiscoveryRequest, response *envoy_sd.DiscoveryResponse) {
	a.callbacks.OnStreamResponse(streamID, &discoveryRequest{request}, &discoveryResponse{response})
}

type discoveryRequest struct {
	*envoy_sd.DiscoveryRequest
}

func (d *discoveryRequest) Metadata() *_struct.Struct {
	return d.GetNode().GetMetadata()
}

func (d *discoveryRequest) NodeId() string {
	return d.GetNode().GetId()
}

func (d *discoveryRequest) HasErrors() bool {
	return d.ErrorDetail != nil
}

func (d *discoveryRequest) ErrorMsg() string {
	return d.GetErrorDetail().GetMessage()
}

var _ xds.DiscoveryRequest = &discoveryRequest{}

type discoveryResponse struct {
	*envoy_sd.DiscoveryResponse
}
