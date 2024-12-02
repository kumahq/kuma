package service

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
)

type Server interface {
	observability_v1.MonitoringAssignmentDiscoveryServiceServer
}

func NewServer(cache envoy_cache.Cache, callbacks envoy_server.Callbacks) Server {
	return &server{cache: cache, callbacks: callbacks}
}

var _ Server = &server{}

type server struct {
	observability_v1.UnimplementedMonitoringAssignmentDiscoveryServiceServer
	cache     envoy_cache.Cache
	callbacks envoy_server.Callbacks
}

func (s *server) FetchMonitoringAssignments(ctx context.Context, req *envoy_sd.DiscoveryRequest) (*envoy_sd.DiscoveryResponse, error) {
	if s.callbacks != nil {
		if err := s.callbacks.OnFetchRequest(ctx, req); err != nil {
			return nil, err
		}
	}
	resChan := make(chan envoy_cache.Response, 1)
	streamState := stream.NewStreamState(false, nil)
	// When the context has a deadline we want to do long polling
	// We therefore create a watch that will have 2 possible outcome:
	// 1. context reaches deadline, in which nothing changed we're at the end of the long polling time
	// 2. watch resolves, the cache changed, let's return the new info
	// Note that for simplicity we don't use the value returned by the watch and simply fetch from the cache
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		cancelWatch := s.cache.CreateWatch(req, streamState, resChan)
		if cancelWatch != nil {
			defer cancelWatch()
			select { // Wait until either we timeout or the watch triggers
			case <-ctx.Done():
			case <-resChan:
			}
		}
	}
	// Don't include the deadline here as it is used for long polling.
	resp, err := s.cache.Fetch(context.WithoutCancel(ctx), req)
	if err != nil {
		return nil, err
	}
	out, err := resp.GetDiscoveryResponse()
	if err != nil {
		return nil, err
	}
	if s.callbacks != nil {
		s.callbacks.OnFetchResponse(req, out)
	}
	return out, err
}
