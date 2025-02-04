package xds

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
)

type LoggingCallbacks struct {
	Log logr.Logger
}

var _ Callbacks = LoggingCallbacks{}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb LoggingCallbacks) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	cb.Log.V(1).Info("OnStreamOpen", "context", ctx, "streamid", streamID, "type", typ)
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb LoggingCallbacks) OnStreamClosed(streamID int64) {
	cb.Log.V(1).Info("OnStreamClosed", "streamid", streamID)
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb LoggingCallbacks) OnStreamRequest(streamID int64, req DiscoveryRequest) error {
	cb.Log.V(1).Info("OnStreamRequest", "streamid", streamID, "req", req)
	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (cb LoggingCallbacks) OnStreamResponse(streamID int64, req DiscoveryRequest, resp DiscoveryResponse) {
	cb.Log.V(1).Info("OnStreamResponse", "streamid", streamID, "req", req, "resp", resp)
}

// OnDeltaStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnDeltaStreamOpen will still be called.
func (cb LoggingCallbacks) OnDeltaStreamOpen(ctx context.Context, streamID int64, typ string) error {
	cb.Log.V(1).Info("OnDeltaStreamOpen", "context", ctx, "streamid", streamID, "type", typ)
	return nil
}

// OnDeltaStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb LoggingCallbacks) OnDeltaStreamClosed(streamID int64) {
	cb.Log.V(1).Info("OnDeltaStreamClosed", "streamid", streamID)
}

// OnStreamDeltaRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamDeltaRequest will still be called.
func (cb LoggingCallbacks) OnStreamDeltaRequest(streamID int64, req DeltaDiscoveryRequest) error {
	if req.ErrorMsg() != "" {
		cb.Log.Error(errors.New(req.ErrorMsg()), "resource was rejected", "streamid", streamID, "req", req)
	}
	cb.Log.V(1).Info("OnStreamDeltaRequest", "streamid", streamID, "req", req)
	return nil
}

// OnStreamDeltaResponse is called immediately prior to sending a response on a stream.
func (cb LoggingCallbacks) OnStreamDeltaResponse(streamID int64, req DeltaDiscoveryRequest, resp DeltaDiscoveryResponse) {
	cb.Log.V(1).Info("OnStreamDeltaResponse", "streamid", streamID, "req", req, "resp", resp)
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb LoggingCallbacks) OnFetchRequest(ctx context.Context, req DiscoveryRequest) error {
	cb.Log.V(1).Info("OnFetchRequest", "context", ctx, "req", req)
	return nil
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
// OnFetchResponse is called immediately prior to sending a response.
func (cb LoggingCallbacks) OnFetchResponse(req DiscoveryRequest, resp DiscoveryResponse) {
	cb.Log.V(1).Info("OnFetchResponse", "req", req, "resp", resp)
}
