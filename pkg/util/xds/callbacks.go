package xds

import (
	"context"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// DiscoveryRequest defines interface over real Envoy's DiscoveryRequest.
type DiscoveryRequest interface {
	NodeId() string
	// Node returns either a v2 or v3 Node
	Node() interface{}
	Metadata() *structpb.Struct
	VersionInfo() string
	GetTypeUrl() string
	GetResponseNonce() string
	GetResourceNames() []string
	HasErrors() bool
	ErrorMsg() string
}

// DiscoveryResponse defines interface over real Envoy's DiscoveryResponse.
type DiscoveryResponse interface {
	GetTypeUrl() string
	VersionInfo() string
	GetResources() []*anypb.Any
	GetNonce() string
}

// Callbacks defines Callbacks for xDS streaming requests. The difference over real go-control-plane Callbacks is that it takes an DiscoveryRequest / DiscoveryResponse interface.
// It helps us to implement Callbacks once for many different versions of Envoy API.
type Callbacks interface {
	// OnStreamOpen is called once an xDS stream is opened with a stream ID and the type URL (or "" for ADS).
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

// RestCallbacks defines rest.Callbacks for xDS fetch requests. The difference over real go-control-plane
// Callbacks is that it takes an DiscoveryRequest / DiscoveryResponse interface.
// It helps us to implement Callbacks once for many different versions of Envoy API.
type RestCallbacks interface {
	// OnFetchRequest is called when a new rest request comes in.
	// Returning an error will end processing. OnFetchResponse will not be called.
	OnFetchRequest(ctx context.Context, request DiscoveryRequest) error
	// OnFetchResponse is called immediately prior to sending a rest response.
	OnFetchResponse(request DiscoveryRequest, response DiscoveryResponse)
}

// MultiCallbacks implements callbacks for both rest and streaming xDS requests.
type MultiCallbacks interface {
	Callbacks
	RestCallbacks
}
