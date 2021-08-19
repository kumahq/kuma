package mux

import (
	"context"
	"io"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
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
	requests chan *envoy_sd.DiscoveryRequest
}

func (k *kdsServerStream) put(request *envoy_sd.DiscoveryRequest) {
	k.requests <- request
}

func (k *kdsServerStream) Send(response *envoy_sd.DiscoveryResponse) error {
	var msg *mesh_proto.Message

	kdsVersion := KDSVersion(k.Context())
	switch kdsVersion {
	case KDSVersionV2:
		msg = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyResponse{LegacyResponse: DiscoveryResponseV2(response)}}
	case KDSVersionV3:
		msg = &mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: response}}
	default:
		return UnsupportedKDSVersion(kdsVersion)
	}

	return k.MultiplexStream.Send(msg)
}

func (k *kdsServerStream) Recv() (*envoy_sd.DiscoveryRequest, error) {
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
