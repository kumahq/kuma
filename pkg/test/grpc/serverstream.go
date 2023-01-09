package grpc

import (
	"context"
	"io"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
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
		Ctx:    context.Background(),
		SentCh: make(chan *envoy_sd.DiscoveryResponse, 10),
		RecvCh: make(chan *envoy_sd.DiscoveryRequest, 10),
	}
}
