package grpc

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
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
		return nil, errors.New("empty")
	}
	return req, nil
}

func (stream *MockServerStream) ClientStream(stopCh chan struct{}) *MockClientStream {
	sentCh := make(chan *envoy_sd.DiscoveryRequest)
	recvCh := make(chan *envoy_sd.DiscoveryResponse)
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
		SentCh: make(chan *envoy_sd.DiscoveryResponse, 10),
		RecvCh: make(chan *envoy_sd.DiscoveryRequest, 10),
	}
}
