package grpc

import (
	"context"
	"fmt"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type MockStream struct {
	Ctx       context.Context
	RecvCh    chan *v2.DiscoveryRequest
	SentCh    chan *v2.DiscoveryResponse
	Nonce     int
	SendError bool
	grpc.ServerStream
}

func (stream *MockStream) Context() context.Context {
	return stream.Ctx
}

func (stream *MockStream) Send(resp *v2.DiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.Nonce++
	Expect(resp.Nonce).To(Equal(fmt.Sprintf("%d", stream.Nonce)))
	Expect(resp.VersionInfo).ToNot(BeEmpty())
	Expect(resp.Resources).ToNot(BeEmpty())
	Expect(resp.TypeUrl).ToNot(BeEmpty())
	for _, res := range resp.Resources {
		Expect(res.TypeUrl).To(Equal(resp.TypeUrl))
	}

	stream.SentCh <- resp
	if stream.SendError {
		return errors.New("send error")
	}
	return nil
}

func (stream *MockStream) Recv() (*v2.DiscoveryRequest, error) {
	req, more := <-stream.RecvCh
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func MakeMockStream() *MockStream {
	return &MockStream{
		Ctx:    context.Background(),
		SentCh: make(chan *v2.DiscoveryResponse, 10),
		RecvCh: make(chan *v2.DiscoveryRequest, 10),
	}
}
