package grpc

import (
	"context"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type MockClientStream struct {
	Ctx    context.Context
	SentCh chan *v2.DiscoveryRequest
	RecvCh chan *v2.DiscoveryResponse
	grpc.ClientStream
}

func (stream *MockClientStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockClientStream) Send(resp *v2.DiscoveryRequest) error {
	stream.SentCh <- resp
	return nil
}

func (stream *MockClientStream) Recv() (*v2.DiscoveryResponse, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func MakeMockClientStream() *MockClientStream {
	return &MockClientStream{
		Ctx:    context.Background(),
		RecvCh: make(chan *v2.DiscoveryResponse, 10),
		SentCh: make(chan *v2.DiscoveryRequest, 10),
	}
}

func (stream *MockClientStream) CloseSend() error {
	close(stream.SentCh)
	return nil
}
