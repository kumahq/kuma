package grpc

import (
	"context"
	"io"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type MockServerStream struct {
	Ctx    context.Context
	RecvCh chan *envoy_sd.DiscoveryRequest
	SentCh chan *envoy_sd.DiscoveryResponse
	Nonce  int
	grpc.ServerStream
}

func (stream *MockServerStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockServerStream) Send(resp *envoy_sd.DiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.Nonce++
	stream.SentCh <- resp
	return nil
}

func (stream *MockServerStream) Recv() (*envoy_sd.DiscoveryRequest, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, io.EOF
	}
	return req, nil
}

func (stream *MockServerStream) ClientStream(stopCh chan struct{}) *MockClientStream {
	mockClientStream := NewMockClientStream()
	go func() {
		for {
			r, more := <-mockClientStream.SentCh
			if !more {
				close(stream.RecvCh)
				return
			}
			stream.RecvCh <- r
		}
	}()
	go func() {
		for {
			select {
			case <-stopCh:
				close(mockClientStream.RecvCh)
				return
			case r := <-stream.SentCh:
				mockClientStream.RecvCh <- r
			}
		}
	}()
	return mockClientStream
}

func NewMockServerStream() *MockServerStream {
	return &MockServerStream{
		Ctx:    metadata.NewIncomingContext(context.Background(), map[string][]string{}),
		SentCh: make(chan *envoy_sd.DiscoveryResponse, len(registry.Global().ObjectTypes())),
		RecvCh: make(chan *envoy_sd.DiscoveryRequest, len(registry.Global().ObjectTypes())),
	}
}

type MockDeltaServerStream struct {
	Ctx    context.Context
	RecvCh chan *envoy_sd.DeltaDiscoveryRequest
	SentCh chan *envoy_sd.DeltaDiscoveryResponse
	Nonce  int
	grpc.ServerStream
}

func (stream *MockDeltaServerStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockDeltaServerStream) Send(resp *envoy_sd.DeltaDiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.Nonce++
	stream.SentCh <- resp
	return nil
}

func (stream *MockDeltaServerStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, io.EOF
	}
	return req, nil
}

func (stream *MockDeltaServerStream) ClientStream(stopCh chan struct{}) *MockDeltaClientStream {
	mockClientStream := NewMockDeltaClientStream()
	go func() {
		for {
			r, more := <-mockClientStream.SentCh
			if !more {
				close(stream.RecvCh)
				return
			}
			stream.RecvCh <- r
		}
	}()
	go func() {
		for {
			select {
			case <-stopCh:
				close(mockClientStream.RecvCh)
				return
			case r := <-stream.SentCh:
				mockClientStream.RecvCh <- r
			}
		}
	}()
	return mockClientStream
}

func NewMockDeltaServerStream() *MockDeltaServerStream {
	return &MockDeltaServerStream{
		Ctx:    context.Background(),
		SentCh: make(chan *envoy_sd.DeltaDiscoveryResponse, len(registry.Global().ObjectTypes())),
		RecvCh: make(chan *envoy_sd.DeltaDiscoveryRequest, len(registry.Global().ObjectTypes())),
	}
}
