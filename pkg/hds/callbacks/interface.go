package callbacks

import (
	"context"

	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
)

type Callbacks interface {
	// OnStreamOpen is called once an HDS stream is open with a stream ID and context
	// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
	OnStreamOpen(ctx context.Context, streamID int64) error

	// OnHealthCheckRequest is called when Envoy sends HealthCheckRequest with Node and Capabilities
	OnHealthCheckRequest(streamID int64, request *envoy_service_health.HealthCheckRequest) error

	// OnEndpointHealthResponse is called when there is a response from Envoy with status of endpoints in the cluster
	OnEndpointHealthResponse(streamID int64, response *envoy_service_health.EndpointHealthResponse) error

	// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
	OnStreamClosed(int64)
}
