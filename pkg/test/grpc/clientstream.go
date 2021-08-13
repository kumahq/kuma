package grpc

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type MockClientStream struct {
	Ctx    context.Context
	SentCh chan *envoy_sd.DiscoveryRequest
	RecvCh chan *envoy_sd.DiscoveryResponse
	grpc.ClientStream
}

func (stream *MockClientStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockClientStream) Send(resp *envoy_sd.DiscoveryRequest) error {
	stream.SentCh <- resp
	return nil
}

func (stream *MockClientStream) Recv() (*envoy_sd.DiscoveryResponse, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, errors.New("empty")
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
	close(stream.SentCh)
	return nil
}
