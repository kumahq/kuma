package mux

import (
	"context"
	"io"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type kdsClientStream struct {
	MultiplexStream
	responses chan *envoy_api_v2.DiscoveryResponse
}

func (k *kdsClientStream) put(response *envoy_api_v2.DiscoveryResponse) {
	k.responses <- response
}

func (k *kdsClientStream) Send(request *envoy_api_v2.DiscoveryRequest) error {
	return k.MultiplexStream.Send(&mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: request}})
}

func (k *kdsClientStream) Recv() (*envoy_api_v2.DiscoveryResponse, error) {
	if r, ok := <-k.responses; ok {
		return r, nil
	}
	return nil, io.EOF
}

func (k *kdsClientStream) Header() (metadata.MD, error) {
	panic("not implemented")
}

func (k *kdsClientStream) Trailer() metadata.MD {
	panic("not implemented")
}

func (k *kdsClientStream) CloseSend() error {
	panic("not implemented")
}

func (k *kdsClientStream) Context() context.Context {
	return k.MultiplexStream.Context()
}

func (k *kdsClientStream) SendMsg(m interface{}) error {
	panic("not implemented")
}

func (k *kdsClientStream) RecvMsg(m interface{}) error {
	panic("not implemented")
}
