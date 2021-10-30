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

func MakeMockClientStream() *MockClientStream {
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
