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

package server_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/envoyproxy/go-control-plane/pkg/test/resource/v2"

	"github.com/Kong/kuma/pkg/xds/server"
)

type hasher struct {
}

func (h hasher) ID(node *envoy_core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

type mockConfigWatcher struct {
	counts     map[string]int
	responses  map[string][]envoy_cache.Response
	closeWatch bool
}

func (config *mockConfigWatcher) CreateWatch(req v2.DiscoveryRequest) (chan envoy_cache.Response, func()) {
	config.counts[req.TypeUrl] = config.counts[req.TypeUrl] + 1
	out := make(chan envoy_cache.Response, 1)
	if len(config.responses[req.TypeUrl]) > 0 {
		out <- config.responses[req.TypeUrl][0]
		config.responses[req.TypeUrl] = config.responses[req.TypeUrl][1:]
	} else if config.closeWatch {
		close(out)
	}
	return out, func() {}
}

func (config *mockConfigWatcher) Fetch(ctx context.Context, req v2.DiscoveryRequest) (*envoy_cache.Response, error) {
	if len(config.responses[req.TypeUrl]) > 0 {
		out := config.responses[req.TypeUrl][0]
		config.responses[req.TypeUrl] = config.responses[req.TypeUrl][1:]
		return &out, nil
	}
	return nil, errors.New("missing")
}

func (config *mockConfigWatcher) GetStatusInfo(string) envoy_cache.StatusInfo { return nil }
func (config *mockConfigWatcher) GetStatusKeys() []string                     { return nil }

func makeMockConfigWatcher() *mockConfigWatcher {
	return &mockConfigWatcher{
		counts: make(map[string]int),
	}
}

type callbacks struct {
	fetchReq      int
	fetchResp     int
	callbackError bool
}

func (c *callbacks) OnStreamOpen(context.Context, int64, string) error {
	if c.callbackError {
		return errors.New("stream open error")
	}
	return nil
}
func (c *callbacks) OnStreamClosed(int64)                                                {}
func (c *callbacks) OnStreamRequest(int64, *v2.DiscoveryRequest) error                   { return nil }
func (c *callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
func (c *callbacks) OnFetchRequest(context.Context, *v2.DiscoveryRequest) error {
	if c.callbackError {
		return errors.New("fetch request error")
	}
	c.fetchReq++
	return nil
}
func (c *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {
	c.fetchResp++
}

type mockStream struct {
	t         *testing.T
	ctx       context.Context
	recv      chan *v2.DiscoveryRequest
	sent      chan *v2.DiscoveryResponse
	nonce     int
	sendError bool
	grpc.ServerStream
}

func (stream *mockStream) Context() context.Context {
	return stream.ctx
}

func (stream *mockStream) Send(resp *v2.DiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.nonce = stream.nonce + 1
	if resp.Nonce != fmt.Sprintf("%d", stream.nonce) {
		stream.t.Errorf("Nonce => got %q, want %d", resp.Nonce, stream.nonce)
	}
	// check that version is set
	if resp.VersionInfo == "" {
		stream.t.Error("VersionInfo => got none, want non-empty")
	}
	// check resources are non-empty
	if len(resp.Resources) == 0 {
		stream.t.Error("Resources => got none, want non-empty")
	}
	// check that type URL matches in resources
	if resp.TypeUrl == "" {
		stream.t.Error("TypeUrl => got none, want non-empty")
	}
	for _, res := range resp.Resources {
		if res.TypeUrl != resp.TypeUrl {
			stream.t.Errorf("TypeUrl => got %q, want %q", res.TypeUrl, resp.TypeUrl)
		}
	}
	stream.sent <- resp
	if stream.sendError {
		return errors.New("send error")
	}
	return nil
}

func (stream *mockStream) Recv() (*v2.DiscoveryRequest, error) {
	req, more := <-stream.recv
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func makeMockStream(t *testing.T) *mockStream {
	return &mockStream{
		t:    t,
		ctx:  context.Background(),
		sent: make(chan *v2.DiscoveryResponse, 10),
		recv: make(chan *v2.DiscoveryRequest, 10),
	}
}

const (
	clusterName  = "cluster0"
	routeName    = "route0"
	listenerName = "listener0"
)

var (
	node = &envoy_core.Node{
		Id:      "test-id",
		Cluster: "test-cluster",
	}
	endpoint  = resource.MakeEndpoint(clusterName, 8080)
	cluster   = resource.MakeCluster(resource.Ads, clusterName)
	route     = resource.MakeRoute(routeName, clusterName)
	listener  = resource.MakeHTTPListener(resource.Ads, listenerName, 80, routeName)
	testTypes = []string{
		envoy_resource.EndpointType,
		envoy_resource.ClusterType,
		envoy_resource.RouteType,
		envoy_resource.ListenerType,
	}
)

func makeResponses() map[string][]envoy_cache.Response {
	return map[string][]envoy_cache.Response{
		envoy_resource.EndpointType: {{
			Version:   "1",
			Resources: []envoy_types.Resource{endpoint},
		}},
		envoy_resource.ClusterType: {{
			Version:   "2",
			Resources: []envoy_types.Resource{cluster},
		}},
		envoy_resource.RouteType: {{
			Version:   "3",
			Resources: []envoy_types.Resource{route},
		}},
		envoy_resource.ListenerType: {{
			Version:   "4",
			Resources: []envoy_types.Resource{listener},
		}},
	}
}

func TestResponseHandlers(t *testing.T) {
	for _, typ := range testTypes {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcher()
			config.responses = makeResponses()
			s := server.NewServer(config, &callbacks{})

			// make a request
			resp := makeMockStream(t)
			resp.recv <- &v2.DiscoveryRequest{Node: node}
			go func() {
				var err error
				switch typ {
				case envoy_resource.EndpointType:
					err = s.StreamEndpoints(resp)
				case envoy_resource.ClusterType:
					err = s.StreamClusters(resp)
				case envoy_resource.RouteType:
					err = s.StreamRoutes(resp)
				case envoy_resource.ListenerType:
					err = s.StreamListeners(resp)
				}
				if err != nil {
					t.Errorf("Stream() => got %v, want no error", err)
				}
			}()

			// check a response
			select {
			case <-resp.sent:
				close(resp.recv)
				if want := map[string]int{typ: 1}; !reflect.DeepEqual(want, config.counts) {
					t.Errorf("watch counts => got %v, want %v", config.counts, want)
				}
			case <-time.After(1 * time.Second):
				t.Fatalf("got no response")
			}
		})
	}
}

func TestFetch(t *testing.T) {
	config := makeMockConfigWatcher()
	config.responses = makeResponses()
	cb := &callbacks{}
	s := server.NewServer(config, cb)
	if out, err := s.FetchEndpoints(context.Background(), &v2.DiscoveryRequest{Node: node}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for endpoints: %v", err)
	}
	if out, err := s.FetchClusters(context.Background(), &v2.DiscoveryRequest{Node: node}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for clusters: %v", err)
	}
	if out, err := s.FetchRoutes(context.Background(), &v2.DiscoveryRequest{Node: node}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for routes: %v", err)
	}
	if out, err := s.FetchListeners(context.Background(), &v2.DiscoveryRequest{Node: node}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for listeners: %v", err)
	}

	// try again and expect empty results
	if out, err := s.FetchEndpoints(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil {
		t.Errorf("expected empty or error for endpoints: %v", err)
	}
	if out, err := s.FetchClusters(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil {
		t.Errorf("expected empty or error for clusters: %v", err)
	}
	if out, err := s.FetchRoutes(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil {
		t.Errorf("expected empty or error for routes: %v", err)
	}
	if out, err := s.FetchListeners(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil {
		t.Errorf("expected empty or error for listeners: %v", err)
	}

	// try empty requests: not valid in a real gRPC server
	if out, err := s.FetchEndpoints(context.Background(), nil); out != nil {
		t.Errorf("expected empty on empty request: %v", err)
	}
	if out, err := s.FetchClusters(context.Background(), nil); out != nil {
		t.Errorf("expected empty on empty request: %v", err)
	}
	if out, err := s.FetchRoutes(context.Background(), nil); out != nil {
		t.Errorf("expected empty on empty request: %v", err)
	}
	if out, err := s.FetchListeners(context.Background(), nil); out != nil {
		t.Errorf("expected empty on empty request: %v", err)
	}

	// send error from callback
	cb.callbackError = true
	if out, err := s.FetchEndpoints(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchClusters(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchRoutes(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchListeners(context.Background(), &v2.DiscoveryRequest{Node: node}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}

	// verify fetch callbacks
	if want := 8; cb.fetchReq != want {
		t.Errorf("unexpected number of fetch requests: got %d, want %d", cb.fetchReq, want)
	}
	if want := 4; cb.fetchResp != want {
		t.Errorf("unexpected number of fetch responses: got %d, want %d", cb.fetchResp, want)
	}
}

func TestWatchClosed(t *testing.T) {
	for _, typ := range testTypes {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcher()
			config.closeWatch = true
			s := server.NewServer(config, &callbacks{})

			// make a request
			resp := makeMockStream(t)
			resp.recv <- &v2.DiscoveryRequest{
				Node:    node,
				TypeUrl: typ,
			}

			// check that response fails since watch gets closed
			if err := s.StreamAggregatedResources(resp); err == nil {
				t.Error("Stream() => got no error, want watch failed")
			}

			close(resp.recv)
		})
	}
}

func TestSendError(t *testing.T) {
	for _, typ := range testTypes {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcher()
			config.responses = makeResponses()
			s := server.NewServer(config, &callbacks{})

			// make a request
			resp := makeMockStream(t)
			resp.sendError = true
			resp.recv <- &v2.DiscoveryRequest{
				Node:    node,
				TypeUrl: typ,
			}

			// check that response fails since send returns error
			if err := s.StreamAggregatedResources(resp); err == nil {
				t.Error("Stream() => got no error, want send error")
			}

			close(resp.recv)
		})
	}
}

func TestStaleNonce(t *testing.T) {
	for _, typ := range testTypes {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcher()
			config.responses = makeResponses()
			s := server.NewServer(config, &callbacks{})

			resp := makeMockStream(t)
			resp.recv <- &v2.DiscoveryRequest{
				Node:    node,
				TypeUrl: typ,
			}
			stop := make(chan struct{})
			go func() {
				if err := s.StreamAggregatedResources(resp); err != nil {
					t.Errorf("StreamAggregatedResources() => got %v, want no error", err)
				}
				// should be two watches called
				if want := map[string]int{typ: 2}; !reflect.DeepEqual(want, config.counts) {
					t.Errorf("watch counts => got %v, want %v", config.counts, want)
				}
				close(stop)
			}()
			select {
			case <-resp.sent:
				// stale request
				resp.recv <- &v2.DiscoveryRequest{
					Node:          node,
					TypeUrl:       typ,
					ResponseNonce: "xyz",
				}
				// fresh request
				resp.recv <- &v2.DiscoveryRequest{
					VersionInfo:   "1",
					Node:          node,
					TypeUrl:       typ,
					ResponseNonce: "1",
				}
				close(resp.recv)
			case <-time.After(1 * time.Second):
				t.Fatalf("got %d messages on the stream, not 4", resp.nonce)
			}
			<-stop
		})
	}
}

func TestAggregatedHandlers(t *testing.T) {
	config := makeMockConfigWatcher()
	config.responses = makeResponses()
	resp := makeMockStream(t)

	resp.recv <- &v2.DiscoveryRequest{
		Node:    node,
		TypeUrl: envoy_resource.ListenerType,
	}
	resp.recv <- &v2.DiscoveryRequest{
		Node:    node,
		TypeUrl: envoy_resource.ClusterType,
	}
	resp.recv <- &v2.DiscoveryRequest{
		Node:          node,
		TypeUrl:       envoy_resource.EndpointType,
		ResourceNames: []string{clusterName},
	}
	resp.recv <- &v2.DiscoveryRequest{
		Node:          node,
		TypeUrl:       envoy_resource.RouteType,
		ResourceNames: []string{routeName},
	}

	s := server.NewServer(config, &callbacks{})
	go func() {
		if err := s.StreamAggregatedResources(resp); err != nil {
			t.Errorf("StreamAggregatedResources() => got %v, want no error", err)
		}
	}()

	count := 0
	for {
		select {
		case <-resp.sent:
			count++
			if count >= 4 {
				close(resp.recv)
				if want := map[string]int{
					envoy_resource.EndpointType: 1,
					envoy_resource.ClusterType:  1,
					envoy_resource.RouteType:    1,
					envoy_resource.ListenerType: 1,
				}; !reflect.DeepEqual(want, config.counts) {
					t.Errorf("watch counts => got %v, want %v", config.counts, want)
				}

				// got all messages
				return
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("got %d messages on the stream, not 4", count)
		}
	}
}

func TestClusterWarming(t *testing.T) {
	config := envoy_cache.NewSnapshotCache(true, hasher{}, nil)
	err := config.SetSnapshot(node.Id, envoy_cache.NewSnapshot("1", []envoy_types.Resource{endpoint}, nil, nil, nil, nil))
	if err != nil {
		t.Fatalf("got %v, want no error", err)
	}
	resp := makeMockStream(t)

	s := server.NewServer(config, &callbacks{})
	go func() {
		if err := s.StreamAggregatedResources(resp); err != nil {
			t.Errorf("StreamAggregatedResources() => got %v, want no error", err)
		}
	}()

	// simulate initial EDS request

	resp.recv <- &v2.DiscoveryRequest{
		Node:          node,
		TypeUrl:       envoy_resource.EndpointType,
		ResourceNames: []string{clusterName},
	}

	var resp1 *v2.DiscoveryResponse
resp1:
	for {
		select {
		case resp := <-resp.sent:
			if resp.TypeUrl != envoy_resource.EndpointType {
				t.Errorf("TypeUrl => got %v, want %v", resp.TypeUrl, envoy_resource.EndpointType)
			}
			resp1 = resp
			break resp1
		case <-time.After(1 * time.Second):
			t.Fatalf("got %d messages on the stream, not 1", 0)
		}
	}

	// simulate EDS request as part of cluster warming

	req2 := &v2.DiscoveryRequest{
		VersionInfo:   resp1.VersionInfo,
		ResponseNonce: resp1.Nonce,
		Node:          node,
		TypeUrl:       envoy_resource.EndpointType,
		ResourceNames: []string{clusterName},
	}

	resp.recv <- req2 // ACK to resp1
	resp.recv <- req2 // cluster warming

resp2:
	for {
		select {
		case resp := <-resp.sent:
			if resp.TypeUrl != envoy_resource.EndpointType {
				t.Errorf("TypeUrl => got %v, want %v", resp.TypeUrl, envoy_resource.EndpointType)
			}
			break resp2
		case <-time.After(1 * time.Second):
			t.Fatalf("got %d messages on the stream, not 1", 0)
		}
	}
}

func TestAggregateRequestType(t *testing.T) {
	config := makeMockConfigWatcher()
	s := server.NewServer(config, &callbacks{})
	resp := makeMockStream(t)
	resp.recv <- &v2.DiscoveryRequest{Node: node}
	if err := s.StreamAggregatedResources(resp); err == nil {
		t.Error("StreamAggregatedResources() => got nil, want an error")
	}
}

func TestCallbackError(t *testing.T) {
	for _, typ := range testTypes {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcher()
			config.responses = makeResponses()
			s := server.NewServer(config, &callbacks{callbackError: true})

			// make a request
			resp := makeMockStream(t)
			resp.recv <- &v2.DiscoveryRequest{
				Node:    node,
				TypeUrl: typ,
			}

			// check that response fails since stream open returns error
			if err := s.StreamAggregatedResources(resp); err == nil {
				t.Error("Stream() => got no error, want error")
			}

			close(resp.recv)
		})
	}
}
