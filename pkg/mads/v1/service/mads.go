package service

import (
	"context"
	"errors"

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
	// Because we want to do long polling we use the watch system if there's a deadline on the context
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		cancelWatch := s.cache.CreateWatch(req, streamState, resChan)
		defer cancelWatch()
		select { // Wait until either we timeout or the watch triggers
		case <-ctx.Done():
		case <-resChan:
		}
	}
	resp, err := s.cache.Fetch(context.WithoutCancel(ctx), req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("missing response")
	}
	out, err := resp.GetDiscoveryResponse()
	if s.callbacks != nil {
		s.callbacks.OnFetchResponse(req, out)
	}
	return out, err
}
