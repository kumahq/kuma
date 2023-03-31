package grpc

import (
	"context"
	"io"
	"sync"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
)

type MockClientStream struct {
	Ctx    context.Context
	SentCh chan *envoy_sd.DiscoveryRequest
	RecvCh chan *envoy_sd.DiscoveryResponse
	grpc.ClientStream
	closed bool
	sync.RWMutex
}

func (stream *MockClientStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockClientStream) Send(resp *envoy_sd.DiscoveryRequest) error {
	stream.RLock()
	defer stream.RUnlock()
	if stream.closed {
		return io.EOF
	}
	stream.SentCh <- resp
	return nil
}

func (stream *MockClientStream) Recv() (*envoy_sd.DiscoveryResponse, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, io.EOF
	}
	return req, nil
}

func NewMockClientStream() *MockClientStream {
	return &MockClientStream{
		Ctx:    context.Background(),
		RecvCh: make(chan *envoy_sd.DiscoveryResponse, 10),
		SentCh: make(chan *envoy_sd.DiscoveryRequest, 10),
	}
}

func (stream *MockClientStream) CloseSend() error {
	stream.Lock()
	defer stream.Unlock()
	close(stream.SentCh)
	stream.closed = true
	return nil
}

type MockDeltaClientStream struct {
	Ctx    context.Context
	SentCh chan *envoy_sd.DeltaDiscoveryRequest
	RecvCh chan *envoy_sd.DeltaDiscoveryResponse
	grpc.ClientStream
	closed bool
	sync.RWMutex
}

func (stream *MockDeltaClientStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockDeltaClientStream) Send(resp *envoy_sd.DeltaDiscoveryRequest) error {
	stream.RLock()
	defer stream.RUnlock()
	if stream.closed {
		return io.EOF
	}
	stream.SentCh <- resp
	return nil
}

func (stream *MockDeltaClientStream) Recv() (*envoy_sd.DeltaDiscoveryResponse, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, io.EOF
	}
	return req, nil
}

func NewMockDeltaClientStream() *MockDeltaClientStream {
	return &MockDeltaClientStream{
		Ctx:    context.Background(),
		RecvCh: make(chan *envoy_sd.DeltaDiscoveryResponse, 10),
		SentCh: make(chan *envoy_sd.DeltaDiscoveryRequest, 10),
	}
}

func (stream *MockDeltaClientStream) CloseSend() error {
	stream.Lock()
	defer stream.Unlock()
	close(stream.SentCh)
	stream.closed = true
	return nil
}
