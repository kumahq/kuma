package server

import (
	"context"
	"strconv"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"
	"github.com/Kong/kuma/pkg/mads"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

type Server interface {
	observability_proto.MonitoringAssignmentDiscoveryServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	return &server{cache: config, callbacks: callbacks, log: log}
}

// server is a simplified version of the original XDS server at
// https://github.com/envoyproxy/go-control-plane/blob/master/pkg/server/server.go
type server struct {
	cache     envoy_cache.Cache
	callbacks envoy_server.Callbacks

	// streamCount for counting bi-di streams
	streamCount int64

	log logr.Logger
}

type stream interface {
	grpc.ServerStream

	Send(*envoy.DiscoveryResponse) error
	Recv() (*envoy.DiscoveryRequest, error)
}

// watches for all xDS resource types
type watches struct {
	assignments chan envoy_cache.Response

	assignmentsCancel func()

	assignmentsNonce string
}

// Cancel all watches
func (values watches) Cancel() {
	if values.assignmentsCancel != nil {
		values.assignmentsCancel()
	}
}

func createResponse(resp *envoy_cache.Response, typeURL string) (*envoy.DiscoveryResponse, error) {
	if resp == nil {
		return nil, errors.New("missing response")
	}
	resources := make([]*any.Any, len(resp.Resources))
	for i := 0; i < len(resp.Resources); i++ {
		// Envoy relies on serialized protobuf bytes for detecting changes to the resources.
		// This requires deterministic serialization.
		b := proto.NewBuffer(nil)
		b.SetDeterministic(true)
		err := b.Marshal(resp.Resources[i])
		if err != nil {
			return nil, err
		}
		resources[i] = &any.Any{
			TypeUrl: typeURL,
			Value:   b.Bytes(),
		}
	}
	out := &envoy.DiscoveryResponse{
		VersionInfo: resp.Version,
		Resources:   resources,
		TypeUrl:     typeURL,
	}
	return out, nil
}

// process handles a bi-di stream request
func (s *server) process(stream stream, reqCh <-chan *envoy.DiscoveryRequest, defaultTypeURL string) (err error) {
	// increment stream count
	streamID := atomic.AddInt64(&s.streamCount, 1)

	log := s.log.WithValues("streamID", streamID)
	defer func() {
		if err != nil {
			log.Error(err, "xDS stream terminated with an error")
		}
	}()

	// unique nonce generator for req-resp pairs per xDS stream; the server
	// ignores stale nonces. nonce is only modified within send() function.
	var streamNonce int64

	// a collection of watches per request type
	var values watches
	defer func() {
		values.Cancel()
		if s.callbacks != nil {
			s.callbacks.OnStreamClosed(streamID)
		}
	}()

	// sends a response by serializing to protobuf Any
	send := func(resp envoy_cache.Response, typeURL string) (string, error) {
		out, err := createResponse(&resp, typeURL)
		if err != nil {
			return "", err
		}

		// increment nonce
		streamNonce++
		out.Nonce = strconv.FormatInt(streamNonce, 10)
		if s.callbacks != nil {
			s.callbacks.OnStreamResponse(streamID, &resp.Request, out)
		}

		return out.Nonce, stream.Send(out)
	}

	if s.callbacks != nil {
		if err := s.callbacks.OnStreamOpen(stream.Context(), streamID, defaultTypeURL); err != nil {
			return err
		}
	}

	// node may only be set on the first discovery request
	var node = &envoy_core.Node{}

	for {
		select {
		case resp, more := <-values.assignments:
			if !more {
				return status.Errorf(codes.Unavailable, "MonitoringAssignment watch failed")
			}
			nonce, err := send(resp, mads.MonitoringAssignmentType)
			if err != nil {
				return err
			}
			values.assignmentsNonce = nonce

		case req, more := <-reqCh:
			// input stream ended or errored out
			if !more {
				return nil
			}
			if req == nil {
				return status.Errorf(codes.Unavailable, "empty request")
			}

			// node field in discovery request is delta-compressed
			if req.Node != nil {
				node = req.Node
			} else {
				req.Node = node
			}

			// nonces can be reused across streams; we verify nonce only if nonce is not initialized
			nonce := req.GetResponseNonce()

			// type URL is required for ADS but is implicit for xDS
			if req.TypeUrl == "" {
				req.TypeUrl = defaultTypeURL
			}

			if s.callbacks != nil {
				if err := s.callbacks.OnStreamRequest(streamID, req); err != nil {
					return err
				}
			}

			// cancel existing watches to (re-)request a newer version
			switch {
			case req.TypeUrl == mads.MonitoringAssignmentType && (values.assignmentsNonce == "" || values.assignmentsNonce == nonce):
				if values.assignmentsCancel != nil {
					values.assignmentsCancel()
				}
				values.assignments, values.assignmentsCancel = s.cache.CreateWatch(*req)
			}
		}
	}
}

// handler converts a blocking read call to channels and initiates stream processing
func (s *server) handler(stream stream, typeURL string) error {
	// a channel for receiving incoming requests
	reqCh := make(chan *envoy.DiscoveryRequest)
	reqStop := int32(0)
	go func() {
		for {
			req, err := stream.Recv()
			if atomic.LoadInt32(&reqStop) != 0 {
				return
			}
			if err != nil {
				close(reqCh)
				return
			}
			reqCh <- req
		}
	}()

	err := s.process(stream, reqCh, typeURL)

	atomic.StoreInt32(&reqStop, 1)

	return err
}

func (s *server) StreamMonitoringAssignments(stream observability_proto.MonitoringAssignmentDiscoveryService_StreamMonitoringAssignmentsServer) error {
	return s.handler(stream, mads.MonitoringAssignmentType)
}

func (s *server) DeltaMonitoringAssignments(_ observability_proto.MonitoringAssignmentDiscoveryService_DeltaMonitoringAssignmentsServer) error {
	return errors.New("not implemented")
}

func (s *server) FetchMonitoringAssignments(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = mads.MonitoringAssignmentType
	return s.Fetch(ctx, req)
}

// Fetch is the universal fetch method.
func (s *server) Fetch(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	if s.callbacks != nil {
		if err := s.callbacks.OnFetchRequest(ctx, req); err != nil {
			return nil, err
		}
	}
	resp, err := s.cache.Fetch(ctx, *req)
	if err != nil {
		return nil, err
	}
	out, err := createResponse(resp, req.TypeUrl)
	if s.callbacks != nil {
		s.callbacks.OnFetchResponse(req, out)
	}
	return out, err
}
