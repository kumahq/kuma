package client

import (
	"fmt"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type UpstreamResponse struct {
	ControlPlaneId       string
	Type                 model.ResourceType
	AddedResources       model.ResourceList
	RemovedResourceNames []string
	IsInitialRequest     bool
}

// All methods other than Receive() are non-blocking. It does not wait until the peer CP receives the message.
type KDSStream interface {
	DiscoveryRequest(resourceType model.ResourceType) error
	Receive() (UpstreamResponse, error)
	ACK(typ string) error
	NACK(typ string, err error) error
}

var _ KDSStream = &stream{}

type stream struct {
	streamClient   mesh_proto.KDSSyncService_GlobalToZoneSyncClient
	latestACKed    map[string]*envoy_sd.DeltaDiscoveryResponse
	latestReceived map[string]*envoy_sd.DeltaDiscoveryResponse
	initStateMap   map[string]map[string]string
	clientId       string
	cpConfig       string
}

func NewKDSStream(s mesh_proto.KDSSyncService_GlobalToZoneSyncClient, clientId string, cpConfig string, initStateMap map[string]map[string]string) KDSStream {
	return &stream{
		streamClient:   s,
		latestACKed:    make(map[string]*envoy_sd.DeltaDiscoveryResponse),
		latestReceived: make(map[string]*envoy_sd.DeltaDiscoveryResponse),
		initStateMap:   initStateMap,
		clientId:       clientId,
		cpConfig:       cpConfig,
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
	initialResources := map[string]string{}
	if value, found := s.initStateMap[string(resourceType)]; found {
		initialResources = value
	}

	req := &envoy_sd.DeltaDiscoveryRequest{
		InitialResourceVersions: initialResources,
		ResponseNonce:           "",
		Node: &envoy_core.Node{
			Id: s.clientId,
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					kds.MetadataFieldVersion: {Kind: &structpb.Value_StructValue{StructValue: cpVersion}},
					kds.MetadataFieldConfig:  {Kind: &structpb.Value_StringValue{StringValue: s.cpConfig}},
					kds.MetadataFeatures: {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{
						Values: []*structpb.Value{
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureZoneToken}},
						},
					}}},
				},
			},
		},
		ResourceNamesSubscribe: []string{"*"},
		TypeUrl:                string(resourceType),
	}
	return s.streamClient.Send(req)
}

func (s *stream) Receive() (UpstreamResponse, error) {
	resp, err := s.streamClient.Recv()
	if err != nil {
		return UpstreamResponse{}, err
	}
	rs, resourceAndVersion, err := util.ToDeltaCoreResourceList(resp)
	if err != nil {
		return UpstreamResponse{}, err
	}
	s.latestReceived[string(rs.GetItemType())] = resp

	// when map is empty that means we are doing the first request
	isInitialRequest := len(s.initStateMap) == 0
	typ := string(rs.GetItemType())
	var versions map[string]string
	if value, found := s.initStateMap[typ]; found {
		versions = value
	} else {
		versions = map[string]string{}
	}
	for _, item := range resourceAndVersion {
		versions[item.ResourceName] = item.Version
	}
	s.initStateMap[typ] = versions
	return UpstreamResponse{
		ControlPlaneId:       resp.GetControlPlane().GetIdentifier(),
		Type:                 rs.GetItemType(),
		AddedResources:       rs,
		RemovedResourceNames: resp.RemovedResources,
		IsInitialRequest:     isInitialRequest,
	}, nil
}

func (s *stream) ACK(typ string) error {
	latestReceived := s.latestReceived[typ]
	if latestReceived == nil {
		return nil
	}
	err := s.streamClient.Send(&envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: latestReceived.Nonce,
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
	return s.streamClient.Send(&envoy_sd.DeltaDiscoveryRequest{
		InitialResourceVersions: map[string]string{},
		ResponseNonce:           latestReceived.Nonce,
		TypeUrl:                 typ,
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}
