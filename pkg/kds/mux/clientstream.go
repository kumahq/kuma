package mux

import (
	"context"
	"io"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type kdsClientStream struct {
	MultiplexStream
	responses chan *envoy_sd.DiscoveryResponse
}

func (k *kdsClientStream) put(response *envoy_sd.DiscoveryResponse) {
	k.responses <- response
}

func (k *kdsClientStream) Send(request *envoy_sd.DiscoveryRequest) error {
	var msg *mesh_proto.Message

	kdsVersion := KDSVersion(k.Context())
	switch kdsVersion {
	case KDSVersionV2:
		msg = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyRequest{LegacyRequest: DiscoveryRequestV2(request)}}
	case KDSVersionV3:
		msg = &mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: request}}
	default:
		return UnsupportedKDSVersion(kdsVersion)
	}

	return k.MultiplexStream.Send(msg)
}

func (k *kdsClientStream) Recv() (*envoy_sd.DiscoveryResponse, error) {
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
