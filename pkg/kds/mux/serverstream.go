package mux

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type kdsServerStream struct {
	ctx          context.Context
	bufferStream *bufferStream
}

func (k *kdsServerStream) Send(response *envoy_sd.DiscoveryResponse) error {
	err := k.bufferStream.Send(&mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: response}})
	return err
}

func (k *kdsServerStream) Recv() (*envoy_sd.DiscoveryRequest, error) {
	res, err := k.bufferStream.Recv()
	if err != nil {
		return nil, err
	}
	return res.GetRequest(), nil
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
	return k.ctx
}

func (k *kdsServerStream) SendMsg(m interface{}) error {
	panic("not implemented")
}

func (k *kdsServerStream) RecvMsg(m interface{}) error {
	panic("not implemented")
}
