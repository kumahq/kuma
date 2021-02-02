package callbacks

import (
	"context"

	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
)

type Chain []Callbacks

var _ Callbacks = Chain{}

func (chain Chain) OnStreamOpen(ctx context.Context, streamID int64) error {
	for _, cb := range chain {
		if err := cb.OnStreamOpen(ctx, streamID); err != nil {
			return err
		}
	}
	return nil
}

func (chain Chain) OnHealthCheckRequest(streamID int64, request *envoy_service_health.HealthCheckRequest) error {
	for _, cb := range chain {
		if err := cb.OnHealthCheckRequest(streamID, request); err != nil {
			return err
		}
	}
	return nil
}

func (chain Chain) OnEndpointHealthResponse(streamID int64, response *envoy_service_health.EndpointHealthResponse) error {
	for _, cb := range chain {
		if err := cb.OnEndpointHealthResponse(streamID, response); err != nil {
			return err
		}
	}
	return nil
}

func (chain Chain) OnStreamClosed(streamID int64) {
	for i := len(chain) - 1; i >= 0; i-- {
		cb := chain[i]
		cb.OnStreamClosed(streamID)
	}
}
