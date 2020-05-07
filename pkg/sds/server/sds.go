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

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_resources "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"

	"github.com/Kong/kuma/pkg/core"
)

type SecretDiscoveryHandler interface {
	Handle(ctx context.Context, req envoy.DiscoveryRequest) (*envoy_auth.Secret, error)
}

type Server interface {
	envoy_discovery.SecretDiscoveryServiceServer
}

func NewServer(source SecretDiscoveryHandler, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	return &server{source: source, callbacks: callbacks, log: log}
}

// server is a simplified version of the original XDS server at
// https://github.com/envoyproxy/go-control-plane/blob/master/pkg/server/server.go
type server struct {
	source    SecretDiscoveryHandler
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

type state struct {
	resourceName string

	secretNonce string
}

func createResponse(resp *envoy_cache.Response, typeURL string) (*envoy.DiscoveryResponse, error) {
	if resp == nil {
		return nil, errors.New("missing response")
	}
	resources := make([]*any.Any, len(resp.Resources))
	for i := 0; i < len(resp.Resources); i++ {
		data, err := proto.Marshal(resp.Resources[i])
		if err != nil {
			return nil, err
		}
		resources[i] = &any.Any{
			TypeUrl: typeURL,
			Value:   data,
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
			log.Error(err, "XDS stream terminated with an error")
		}
	}()

	// unique nonce generator for req-resp pairs per xDS stream; the server
	// ignores stale nonces. nonce is only modified within send() function.
	var streamNonce int64

	var state state
	defer func() {
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

		if err := s.validateSdsResponse(&state, out); err != nil {
			return "", err
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

			if len(req.ResourceNames) == 0 {
				// Do not respond to SDS requests with an empty list of resource names.
				// In practice, such requests can be observed when Envoy is removing
				// Listeners and Clusters with TLS configuration that refers to SDS.
				continue
			}

			if err := s.validateSdsRequest(&state, req); err != nil {
				return err
			}

			if state.secretNonce != "" && state.secretNonce == nonce {
				continue // ACK
			}

			secret, err := s.source.Handle(stream.Context(), *req)
			if err != nil {
				return err
			}

			resp := s.toResponse(req, secret)

			nonce, err = send(resp, envoy_resources.SecretType)
			if err != nil {
				return err
			}
			state.secretNonce = nonce
		}
	}
}

func (s *server) toResponse(req *envoy.DiscoveryRequest, secret *envoy_auth.Secret) envoy_cache.Response {
	return envoy_cache.Response{
		Request:   *req,
		Version:   s.version(secret),
		Resources: []envoy_types.Resource{secret},
	}
}

func (s *server) version(msg proto.Message) string {
	return core.NewUUID()
}

func (s *server) validateSdsRequest(state *state, req *envoy.DiscoveryRequest) error {
	if len(req.ResourceNames) != 1 {
		return errors.Errorf("invalid SDS request: expected exactly 1 resourceName, got %d: %+v", len(req.ResourceNames), req)
	}
	resourceName := req.ResourceNames[0]
	if resourceName == "" {
		return errors.Errorf("invalid SDS request: resourceName must be non-empty: %+v", req)
	}
	if state.resourceName == "" {
		state.resourceName = resourceName
	}
	if state.resourceName != resourceName {
		return errors.Errorf("invalid SDS request: resourceName is different from previous requests on that stream: expected %q, got %q: %+v", state.resourceName, resourceName, req)
	}
	return nil
}

func (s *server) validateSdsResponse(state *state, resp *envoy.DiscoveryResponse) error {
	if len(resp.Resources) != 1 {
		return errors.Errorf("invalid SDS response: expected exactly 1 resource, got %d: %+v", len(resp.Resources), resp)
	}
	return nil
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

func (s *server) StreamSecrets(stream envoy_discovery.SecretDiscoveryService_StreamSecretsServer) error {
	return s.handler(stream, envoy_resources.SecretType)
}

func (s *server) FetchSecrets(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *server) DeltaSecrets(_ envoy_discovery.SecretDiscoveryService_DeltaSecretsServer) error {
	return errors.New("not implemented")
}
