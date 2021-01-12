package xds

import (
	"context"

	_struct "github.com/golang/protobuf/ptypes/struct"
)

// DiscoveryRequest defines interface over real Envoy's DiscoveryRequest.
type DiscoveryRequest interface {
	NodeId() string
	Metadata() *_struct.Struct
	GetTypeUrl() string
	GetResponseNonce() string
	HasErrors() bool
}

// DiscoveryResponse defines interface over real Envoy's DiscoveryResponse.
type DiscoveryResponse interface {
	GetTypeUrl() string
}

// Callbacks defines Callbacks for XDS requests. The difference over real go-control-plane Callbacks is that it takes an DiscoveryRequest / DiscoveryResponse interface.
// It helps us to implement Callbacks once for many different versions of Envoy API.
type Callbacks interface {
	// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
	// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
	OnStreamOpen(context.Context, int64, string) error
	// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
	OnStreamClosed(int64)
	// OnStreamRequest is called once a request is received on a stream.
	// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
	OnStreamRequest(int64, DiscoveryRequest) error
	// OnStreamResponse is called immediately prior to sending a response on a stream.
	OnStreamResponse(int64, DiscoveryRequest, DiscoveryResponse)
}
