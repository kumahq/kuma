package server

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type ServerStream interface {
	stream.DeltaStream
}

type serverStream struct {
	stream mesh_proto.KDSSyncService_ZoneToGlobalSyncClient
}

// NewServerStream converts client stream to a server's DeltaStream, so it can be used in DeltaStreamHandler
func NewServerStream(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncClient) ServerStream {
	s := &serverStream{
		stream: stream,
	}
	return s
}

func (k *serverStream) Send(response *envoy_sd.DeltaDiscoveryResponse) error {
	err := k.stream.Send(response)
	return err
}

func (k *serverStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	res, err := k.stream.Recv()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (k *serverStream) SetHeader(metadata.MD) error {
	panic("not implemented")
}

func (k *serverStream) SendHeader(metadata.MD) error {
	panic("not implemented")
}

func (k *serverStream) SetTrailer(metadata.MD) {
	panic("not implemented")
}

func (k *serverStream) Context() context.Context {
	return k.stream.Context()
}

func (k *serverStream) SendMsg(m interface{}) error {
	panic("not implemented")
}

func (k *serverStream) RecvMsg(m interface{}) error {
	panic("not implemented")
}
