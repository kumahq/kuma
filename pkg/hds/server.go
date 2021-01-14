package hds

import (
	"context"
	"sync/atomic"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	envoy_cache "github.com/kumahq/kuma/pkg/hds/cache"
)

// inspired by https://github.com/envoyproxy/go-control-plane/blob/master/pkg/server/sotw/v3/server.go

type Stream interface {
	grpc.ServerStream

	Send(specifier *envoy_service_health.HealthCheckSpecifier) error
	Recv() (*envoy_service_health.HealthCheckRequestOrEndpointHealthResponse, error)
}

type Callbacks interface {
	// OnHealthCheckRequest is called when Envoy sends HealthCheckRequest with Node and Capabilities
	OnHealthCheckRequest(streamID int64, request *envoy_service_health.HealthCheckRequest) error

	// OnEndpointHealthResponse is called when there is a response from Envoy with status of endpoints in the cluster
	OnEndpointHealthResponse(streamID int64, response *envoy_service_health.EndpointHealthResponse) error

	// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
	OnStreamClosed(int64)
}

type server struct {
	streamCount int64
	ctx         context.Context
	callbacks   Callbacks
	cache       cache.Cache

	healthChecks       chan cache.Response
	healthChecksCancel func()
}

func NewServer(ctx context.Context, config cache.Cache, callbacks Callbacks) envoy_service_health.HealthDiscoveryServiceServer {
	return &server{
		ctx:       ctx,
		callbacks: callbacks,
		cache:     config,
	}
}

func (s *server) StreamHealthCheck(stream envoy_service_health.HealthDiscoveryService_StreamHealthCheckServer) error {
	return s.StreamHandler(stream)
}

// StreamHandler converts a blocking read call to channels and initiates stream processing
func (s *server) StreamHandler(stream Stream) error {
	// a channel for receiving incoming requests
	reqOrRespCh := make(chan *envoy_service_health.HealthCheckRequestOrEndpointHealthResponse)
	reqStop := int32(0)
	go func() {
		for {
			req, err := stream.Recv()
			if atomic.LoadInt32(&reqStop) != 0 {
				return
			}
			if err != nil {
				close(reqOrRespCh)
				return
			}
			select {
			case reqOrRespCh <- req:
			case <-s.ctx.Done():
				return
			}
		}
	}()

	err := s.process(stream, reqOrRespCh)
	atomic.StoreInt32(&reqStop, 1)

	return err
}

func (s *server) process(stream Stream, reqOrRespCh chan *envoy_service_health.HealthCheckRequestOrEndpointHealthResponse) error {
	streamID := atomic.AddInt64(&s.streamCount, 1)

	send := func(resp cache.Response) error {
		if resp == nil {
			return errors.New("missing response")
		}

		out, err := resp.GetDiscoveryResponse()
		if err != nil {
			return err
		}
		if len(out.Resources) == 0 {
			return nil
		}

		hcs := &envoy_service_health.HealthCheckSpecifier{}
		if err := ptypes.UnmarshalAny(out.Resources[0], hcs); err != nil {
			return err
		}
		return stream.Send(hcs)
	}

	var node = &envoy_core.Node{}
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case resp, more := <-s.healthChecks:
			if !more {
				return status.Error(codes.Unavailable, "healthChecks watch failed")
			}
			if err := send(resp); err != nil {
				return err
			}
		case reqOrResp, more := <-reqOrRespCh:
			if !more {
				return nil
			}
			if reqOrResp == nil {
				return status.Errorf(codes.Unavailable, "empty request")
			}
			if req := reqOrResp.GetHealthCheckRequest(); req != nil {
				if req.Node != nil {
					node = req.Node
				} else {
					req.Node = node
				}

				if s.callbacks != nil {
					if err := s.callbacks.OnHealthCheckRequest(streamID, req); err != nil {
						return err
					}
				}
			}
			if resp := reqOrResp.GetEndpointHealthResponse(); resp != nil {
				if s.callbacks != nil {
					if err := s.callbacks.OnEndpointHealthResponse(streamID, resp); err != nil {
						return err
					}
				}
			}

			if s.healthChecksCancel != nil {
				s.healthChecksCancel()
			}
			s.healthChecks, s.healthChecksCancel = s.cache.CreateWatch(&cache.Request{
				Node:          node,
				TypeUrl:       envoy_cache.HealthCheckSpecifierType,
				ResourceNames: []string{"hcs"},
			})
		}
	}
}

func (s *server) FetchHealthCheck(ctx context.Context, response *envoy_service_health.HealthCheckRequestOrEndpointHealthResponse) (*envoy_service_health.HealthCheckSpecifier, error) {
	panic("not implemented")
}
