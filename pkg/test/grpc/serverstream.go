package grpc

import (
	"context"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type MockServerStream struct {
	Ctx    context.Context
	RecvCh chan *v2.DiscoveryRequest
	SentCh chan *v2.DiscoveryResponse
	Nonce  int
	grpc.ServerStream
}

func (stream *MockServerStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockServerStream) Send(resp *v2.DiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.Nonce++
	stream.SentCh <- resp
	return nil
}

func (stream *MockServerStream) Recv() (*v2.DiscoveryRequest, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func (stream *MockServerStream) ClientStream(stopCh chan struct{}) *MockClientStream {
	sentCh := make(chan *v2.DiscoveryRequest)
	recvCh := make(chan *v2.DiscoveryResponse)
	go func() {
		for {
			r, more := <-sentCh
			if more {
				stream.RecvCh <- r
			} else {
				close(stream.RecvCh)
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case <-stopCh:
				close(recvCh)
				return
			case r := <-stream.SentCh:
				recvCh <- r
			}
		}
	}()
	return &MockClientStream{
		Ctx:    stream.Ctx,
		SentCh: sentCh,
		RecvCh: recvCh,
	}
}

func MakeMockStream() *MockServerStream {
	return &MockServerStream{
		Ctx:    context.Background(),
		SentCh: make(chan *v2.DiscoveryResponse, 10),
		RecvCh: make(chan *v2.DiscoveryRequest, 10),
	}
}
