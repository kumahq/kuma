package xds

import (
	"context"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"

	"go.uber.org/multierr"
)

type CallbacksChain []envoy_xds.Callbacks

var _ envoy_xds.Callbacks = CallbacksChain{}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (chain CallbacksChain) OnStreamOpen(ctx context.Context, streamID int64, typ string) (errs error) {
	for _, cb := range chain {
		errs = multierr.Append(errs, cb.OnStreamOpen(ctx, streamID, typ))
	}
	return
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (chain CallbacksChain) OnStreamClosed(streamID int64) {
	for i := len(chain) - 1; i >= 0; i-- {
		cb := chain[i]
		cb.OnStreamClosed(streamID)
	}
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (chain CallbacksChain) OnStreamRequest(streamID int64, req *envoy.DiscoveryRequest) (errs error) {
	for _, cb := range chain {
		errs = multierr.Append(errs, cb.OnStreamRequest(streamID, req))
	}
	return
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (chain CallbacksChain) OnStreamResponse(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
	for i := len(chain) - 1; i >= 0; i-- {
		cb := chain[i]
		cb.OnStreamResponse(streamID, req, resp)
	}
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (chain CallbacksChain) OnFetchRequest(ctx context.Context, req *envoy.DiscoveryRequest) (errs error) {
	for _, cb := range chain {
		errs = multierr.Append(errs, cb.OnFetchRequest(ctx, req))
	}
	return
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
// OnFetchResponse is called immediately prior to sending a response.
func (chain CallbacksChain) OnFetchResponse(req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
	for i := len(chain) - 1; i >= 0; i-- {
		cb := chain[i]
		cb.OnFetchResponse(req, resp)
	}
}
