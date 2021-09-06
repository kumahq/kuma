package mux

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type kdsClientStream struct {
	ctx          context.Context
	bufferStream *bufferStream
}

func (k *kdsClientStream) Send(request *envoy_sd.DiscoveryRequest) error {
	err := k.bufferStream.Send(&mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: request}})
	return err
}

func (k *kdsClientStream) Recv() (*envoy_sd.DiscoveryResponse, error) {
	res, err := k.bufferStream.Recv()
	if err != nil {
		return nil, err
	}
	return res.GetResponse(), nil
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
	return k.ctx
}

func (k *kdsClientStream) SendMsg(m interface{}) error {
	panic("not implemented")
}

func (k *kdsClientStream) RecvMsg(m interface{}) error {
	panic("not implemented")
}
