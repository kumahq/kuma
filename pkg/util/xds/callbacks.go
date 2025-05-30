package xds

import (
	"context"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type Request interface {
	NodeId() string
	Node() interface{}
	Metadata() *structpb.Struct
	GetResponseNonce() string
	GetTypeUrl() string
	HasErrors() bool
	ErrorMsg() string
	VersionInfo() string
	GetResourceNames() []string
}

type Response interface {
	GetTypeUrl() string
	GetResources() []*anypb.Any
	GetNonce() string
	VersionInfo() string
	GetNumberOfResources() int
}

// DiscoveryRequest defines interface over real Envoy's DiscoveryRequest.
type DiscoveryRequest interface {
	Request
}

// DiscoveryResponse defines interface over real Envoy's DiscoveryResponse.
type DiscoveryResponse interface {
	Response
}

type DeltaDiscoveryRequest interface {
	Request
	GetResourceNamesSubscribe() []string
	GetInitialResourceVersions() map[string]string
}

// DeltaDiscoveryResponse defines interface over real Envoy's DeltaDiscoveryResponse.
type DeltaDiscoveryResponse interface {
	Response
	GetRemovedResources() []string
}

type XdsMode string

const (
	DELTA_GRPC XdsMode = "DELTA_GRPC"
	GRPC       XdsMode = "GRPC"
)

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

type DeltaCallbacks interface {
	// OnDeltaStreamOpen is called once an xDS stream is opened with a stream ID and the type URL (or "" for ADS).
	// Returning an error will end processing and close the stream. OnDeltaStreamClosed will still be called.
	OnDeltaStreamOpen(context.Context, int64, string) error
	// OnDeltaStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
	OnDeltaStreamClosed(int64)
	// OnStreamDeltaRequest is called once a request is received on a stream.
	// Returning an error will end processing and close the stream. OnDeltaStreamClosed will still be called.
	OnStreamDeltaRequest(int64, DeltaDiscoveryRequest) error
	// OnStreamDeltaResponse is called immediately prior to sending a response on a stream.
	OnStreamDeltaResponse(int64, DeltaDiscoveryRequest, DeltaDiscoveryResponse)
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
	DeltaCallbacks
}

type MultiXDSCallbacks interface {
	Callbacks
	DeltaCallbacks
}
