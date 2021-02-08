package client

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/kds/util"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/genproto/googleapis/rpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type KDSStream interface {
	DiscoveryRequest(resourceType model.ResourceType) error
	Receive() (string, model.ResourceList, error)
	ACK(typ string) error
	NACK(typ string, err error) error
	Close() error
}

var _ KDSStream = &stream{}

type stream struct {
	streamClient   mesh_proto.KumaDiscoveryService_StreamKumaResourcesClient
	latestACKed    map[string]*envoy.DiscoveryResponse
	latestReceived map[string]*envoy.DiscoveryResponse
	clientId       string
}

func NewKDSStream(s mesh_proto.KumaDiscoveryService_StreamKumaResourcesClient, clientId string) KDSStream {
	return &stream{
		streamClient:   s,
		latestACKed:    make(map[string]*envoy.DiscoveryResponse),
		latestReceived: make(map[string]*envoy.DiscoveryResponse),
		clientId:       clientId,
	}
}

func (s *stream) DiscoveryRequest(resourceType model.ResourceType) error {
	cpVersion, err := util_proto.ToStruct(&system_proto.Version{
		KumaCp: &system_proto.KumaCpVersion{
			Version:   kuma_version.Build.Version,
			GitTag:    kuma_version.Build.GitTag,
			GitCommit: kuma_version.Build.GitCommit,
			BuildDate: kuma_version.Build.BuildDate,
		},
	})
	if err != nil {
		return err
	}
	return s.streamClient.Send(&envoy.DiscoveryRequest{
		VersionInfo:   "",
		ResponseNonce: "",
		Node: &envoy_core.Node{
			Id: s.clientId,
			Metadata: &pstruct.Struct{
				Fields: map[string]*pstruct.Value{
					"version": {Kind: &pstruct.Value_StructValue{StructValue: cpVersion}},
				},
			},
		},
		ResourceNames: []string{},
		TypeUrl:       string(resourceType),
	})
}

func (s *stream) Receive() (string, model.ResourceList, error) {
	resp, err := s.streamClient.Recv()
	if err != nil {
		return "", nil, err
	}
	rs, err := util.ToCoreResourceList(resp)
	if err != nil {
		return "", nil, err
	}
	s.latestReceived[string(rs.GetItemType())] = resp
	return resp.GetControlPlane().GetIdentifier(), rs, nil
}

func (s *stream) ACK(typ string) error {
	latestReceived := s.latestReceived[typ]
	if latestReceived == nil {
		return nil
	}
	err := s.streamClient.Send(&envoy.DiscoveryRequest{
		VersionInfo:   latestReceived.VersionInfo,
		ResponseNonce: latestReceived.Nonce,
		ResourceNames: []string{},
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		TypeUrl: typ,
	})
	if err == nil {
		s.latestACKed[typ] = latestReceived
	}
	return err
}

func (s *stream) NACK(typ string, err error) error {
	latestReceived := s.latestReceived[typ]
	if latestReceived == nil {
		return nil
	}
	latestACKed := s.latestACKed[typ]
	return s.streamClient.Send(&envoy.DiscoveryRequest{
		VersionInfo:   latestACKed.GetVersionInfo(),
		ResponseNonce: latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       typ,
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

func (s *stream) Close() error {
	return s.streamClient.CloseSend()
}
