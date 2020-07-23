package mux

import (
	"context"
	"io"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type MultiplexStream interface {
	Send(*mesh_proto.Message) error
	Recv() (*mesh_proto.Message, error)
	Context() context.Context
}

type kdsServerStream struct {
	MultiplexStream
	requests chan *envoy_api_v2.DiscoveryRequest
}

func (k *kdsServerStream) put(request *envoy_api_v2.DiscoveryRequest) {
	k.requests <- request
}

func (k *kdsServerStream) Send(response *envoy_api_v2.DiscoveryResponse) error {
	return k.MultiplexStream.Send(&mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: response}})
}

func (k *kdsServerStream) Recv() (*envoy_api_v2.DiscoveryRequest, error) {
	if r, ok := <-k.requests; ok {
		return r, nil
	}
	return nil, io.EOF
}

func (k *kdsServerStream) SetHeader(metadata.MD) error {
	panic("not implemented")
}

func (k *kdsServerStream) SendHeader(metadata.MD) error {
	panic("not implemented")
}

func (k *kdsServerStream) SetTrailer(metadata.MD) {
	panic("not implemented")
}

func (k *kdsServerStream) Context() context.Context {
	return k.MultiplexStream.Context()
}

func (k *kdsServerStream) SendMsg(m interface{}) error {
	panic("not implemented")
}

func (k *kdsServerStream) RecvMsg(m interface{}) error {
	panic("not implemented")
}
