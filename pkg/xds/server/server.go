// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// Package server provides an implementation of a streaming xDS server.
package server

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	gcp_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

// NewServer creates handlers from a config watcher and callbacks.
func NewServer(config envoy_cache.Cache, callbacks gcp_server.Callbacks) gcp_server.Server {
	return &server{cache: config, callbacks: callbacks}
}

type server struct {
	cache     envoy_cache.Cache
	callbacks gcp_server.Callbacks

	// streamCount for counting bi-di streams
	streamCount int64
}

type stream interface {
	grpc.ServerStream

	Send(*v2.DiscoveryResponse) error
	Recv() (*v2.DiscoveryRequest, error)
}

// watches for all xDS resource types
type watches struct {
	endpoints chan envoy_cache.Response
	clusters  chan envoy_cache.Response
	routes    chan envoy_cache.Response
	listeners chan envoy_cache.Response
	secrets   chan envoy_cache.Response
	runtimes  chan envoy_cache.Response

	endpointCancel func()
	clusterCancel  func()
	routeCancel    func()
	listenerCancel func()
	secretCancel   func()
	runtimeCancel  func()

	endpointNonce string
	clusterNonce  string
	routeNonce    string
	listenerNonce string
	secretNonce   string
	runtimeNonce  string

	endpointNonceAcked bool
}

// Cancel all watches
func (values watches) Cancel() {
	if values.endpointCancel != nil {
		values.endpointCancel()
	}
	if values.clusterCancel != nil {
		values.clusterCancel()
	}
	if values.routeCancel != nil {
		values.routeCancel()
	}
	if values.listenerCancel != nil {
		values.listenerCancel()
	}
	if values.secretCancel != nil {
		values.secretCancel()
	}
	if values.runtimeCancel != nil {
		values.runtimeCancel()
	}
}

func createResponse(resp *envoy_cache.Response, typeURL string) (*v2.DiscoveryResponse, error) {
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
	out := &v2.DiscoveryResponse{
		VersionInfo: resp.Version,
		Resources:   resources,
		TypeUrl:     typeURL,
	}
	return out, nil
}

// process handles a bi-di stream request
func (s *server) process(stream stream, reqCh <-chan *v2.DiscoveryRequest, defaultTypeURL string) error {
	// increment stream count
	streamID := atomic.AddInt64(&s.streamCount, 1)

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
		streamNonce = streamNonce + 1
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

	peekResponse := func(req envoy_cache.Request) *envoy_cache.Response {
		versionInfo := req.VersionInfo
		req.VersionInfo = "" // trick cache into returning a response despite unchanged version
		resp, err := s.cache.Fetch(stream.Context(), req)
		if err != nil || resp == nil {
			return nil
		}
		resp.Request.VersionInfo = versionInfo // restore the original value
		return resp
	}
	scheduleResponse := func(resp envoy_cache.Response) (value chan envoy_cache.Response, cancel func()) {
		value = make(chan envoy_cache.Response, 1)
		value <- resp
		cancel = func() {
			close(value)

			// consume channel
			select {
			case <-value:
			default:
			}
		}
		return
	}

	// node may only be set on the first discovery request
	var node = &envoy_core.Node{}

	for {
		select {
		// config watcher can send the requested resources types in any order
		case resp, more := <-values.endpoints:
			if !more {
				return status.Errorf(codes.Unavailable, "endpoints watch failed")
			}
			nonce, err := send(resp, envoy_resource.EndpointType)
			if err != nil {
				return err
			}
			values.endpointNonce = nonce
			values.endpointNonceAcked = false

		case resp, more := <-values.clusters:
			if !more {
				return status.Errorf(codes.Unavailable, "clusters watch failed")
			}
			nonce, err := send(resp, envoy_resource.ClusterType)
			if err != nil {
				return err
			}
			values.clusterNonce = nonce

		case resp, more := <-values.routes:
			if !more {
				return status.Errorf(codes.Unavailable, "routes watch failed")
			}
			nonce, err := send(resp, envoy_resource.RouteType)
			if err != nil {
				return err
			}
			values.routeNonce = nonce

		case resp, more := <-values.listeners:
			if !more {
				return status.Errorf(codes.Unavailable, "listeners watch failed")
			}
			nonce, err := send(resp, envoy_resource.ListenerType)
			if err != nil {
				return err
			}
			values.listenerNonce = nonce

		case resp, more := <-values.secrets:
			if !more {
				return status.Errorf(codes.Unavailable, "secrets watch failed")
			}
			nonce, err := send(resp, envoy_resource.SecretType)
			if err != nil {
				return err
			}
			values.secretNonce = nonce

		case resp, more := <-values.runtimes:
			if !more {
				return status.Errorf(codes.Unavailable, "runtimes watch failed")
			}
			nonce, err := send(resp, envoy_resource.RuntimeType)
			if err != nil {
				return err
			}
			values.runtimeNonce = nonce

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
			if defaultTypeURL == envoy_resource.AnyType {
				if req.TypeUrl == "" {
					return status.Errorf(codes.InvalidArgument, "type URL is required for ADS")
				}
			} else if req.TypeUrl == "" {
				req.TypeUrl = defaultTypeURL
			}

			if s.callbacks != nil {
				if err := s.callbacks.OnStreamRequest(streamID, req); err != nil {
					return err
				}
			}

			// cancel existing watches to (re-)request a newer version
			switch {
			case req.TypeUrl == envoy_resource.EndpointType && (values.endpointNonce == "" || values.endpointNonce == nonce):
				// If Envoy uses the same Nonce for the second time, it probably means that
				// a Cluster has been created or updated and goes through the warming stage.
				// In that case we must respond even if EDS configuration hasn't changed.
				var resp *envoy_cache.Response
				if values.endpointNonceAcked {
					resp = peekResponse(*req)
				}
				values.endpointNonceAcked = values.endpointNonce != "" && values.endpointNonce == nonce
				if values.endpointCancel != nil {
					values.endpointCancel()
				}
				if resp != nil {
					values.endpoints, values.endpointCancel = scheduleResponse(*resp)
				} else {
					values.endpoints, values.endpointCancel = s.cache.CreateWatch(*req)
				}
			case req.TypeUrl == envoy_resource.ClusterType && (values.clusterNonce == "" || values.clusterNonce == nonce):
				if values.clusterCancel != nil {
					values.clusterCancel()
				}
				values.clusters, values.clusterCancel = s.cache.CreateWatch(*req)
			case req.TypeUrl == envoy_resource.RouteType && (values.routeNonce == "" || values.routeNonce == nonce):
				if values.routeCancel != nil {
					values.routeCancel()
				}
				values.routes, values.routeCancel = s.cache.CreateWatch(*req)
			case req.TypeUrl == envoy_resource.ListenerType && (values.listenerNonce == "" || values.listenerNonce == nonce):
				if values.listenerCancel != nil {
					values.listenerCancel()
				}
				values.listeners, values.listenerCancel = s.cache.CreateWatch(*req)
			case req.TypeUrl == envoy_resource.SecretType && (values.secretNonce == "" || values.secretNonce == nonce):
				if values.secretCancel != nil {
					values.secretCancel()
				}
				values.secrets, values.secretCancel = s.cache.CreateWatch(*req)
			case req.TypeUrl == envoy_resource.RuntimeType && (values.runtimeNonce == "" || values.runtimeNonce == nonce):
				if values.runtimeCancel != nil {
					values.runtimeCancel()
				}
				values.runtimes, values.runtimeCancel = s.cache.CreateWatch(*req)
			}
		}
	}
}

// handler converts a blocking read call to channels and initiates stream processing
func (s *server) handler(stream stream, typeURL string) error {
	// a channel for receiving incoming requests
	reqCh := make(chan *v2.DiscoveryRequest)
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

	// prevents writing to a closed channel if send failed on blocked recv
	// TODO(kuat) figure out how to unblock recv through gRPC API
	atomic.StoreInt32(&reqStop, 1)

	return err
}

func (s *server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	return s.handler(stream, envoy_resource.AnyType)
}

func (s *server) StreamEndpoints(stream v2.EndpointDiscoveryService_StreamEndpointsServer) error {
	return s.handler(stream, envoy_resource.EndpointType)
}

func (s *server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	return s.handler(stream, envoy_resource.ClusterType)
}

func (s *server) StreamRoutes(stream v2.RouteDiscoveryService_StreamRoutesServer) error {
	return s.handler(stream, envoy_resource.RouteType)
}

func (s *server) StreamListeners(stream v2.ListenerDiscoveryService_StreamListenersServer) error {
	return s.handler(stream, envoy_resource.ListenerType)
}

func (s *server) StreamSecrets(stream discovery.SecretDiscoveryService_StreamSecretsServer) error {
	return s.handler(stream, envoy_resource.SecretType)
}

func (s *server) StreamRuntime(stream discovery.RuntimeDiscoveryService_StreamRuntimeServer) error {
	return s.handler(stream, envoy_resource.RuntimeType)
}

// Fetch is the universal fetch method.
func (s *server) Fetch(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
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

func (s *server) FetchEndpoints(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = envoy_resource.EndpointType
	return s.Fetch(ctx, req)
}

func (s *server) FetchClusters(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = envoy_resource.ClusterType
	return s.Fetch(ctx, req)
}

func (s *server) FetchRoutes(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = envoy_resource.RouteType
	return s.Fetch(ctx, req)
}

func (s *server) FetchListeners(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = envoy_resource.ListenerType
	return s.Fetch(ctx, req)
}

func (s *server) FetchSecrets(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = envoy_resource.SecretType
	return s.Fetch(ctx, req)
}

func (s *server) FetchRuntime(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = envoy_resource.RuntimeType
	return s.Fetch(ctx, req)
}

func (s *server) DeltaAggregatedResources(_ discovery.AggregatedDiscoveryService_DeltaAggregatedResourcesServer) error {
	return errors.New("not implemented")
}

func (s *server) DeltaEndpoints(_ v2.EndpointDiscoveryService_DeltaEndpointsServer) error {
	return errors.New("not implemented")
}

func (s *server) DeltaClusters(_ v2.ClusterDiscoveryService_DeltaClustersServer) error {
	return errors.New("not implemented")
}

func (s *server) DeltaRoutes(_ v2.RouteDiscoveryService_DeltaRoutesServer) error {
	return errors.New("not implemented")
}

func (s *server) DeltaListeners(_ v2.ListenerDiscoveryService_DeltaListenersServer) error {
	return errors.New("not implemented")
}

func (s *server) DeltaSecrets(_ discovery.SecretDiscoveryService_DeltaSecretsServer) error {
	return errors.New("not implemented")
}

func (s *server) DeltaRuntime(_ discovery.RuntimeDiscoveryService_DeltaRuntimeServer) error {
	return errors.New("not implemented")
}
