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
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/util"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
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
type DeltaKDSStream interface {
	DeltaDiscoveryRequest(resourceType model.ResourceType) error
	Receive() (UpstreamResponse, error)
	ACK(typ string) error
	NACK(typ string, err error) error
}

var _ DeltaKDSStream = &stream{}

type stream struct {
	streamClient   mesh_proto.KDSSyncService_GlobalToZoneSyncClient
	latestNonce    map[core_model.ResourceType]string
	deltaInitState cache_v2.ResourceVersionMap
	clientId       string
	cpConfig       string
}

func NewDeltaKDSStream(s mesh_proto.KDSSyncService_GlobalToZoneSyncClient, clientId string, cpConfig string, deltaInitState cache_v2.ResourceVersionMap) DeltaKDSStream {
	return &stream{
		streamClient:   s,
		latestNonce:    make(map[core_model.ResourceType]string),
		deltaInitState: deltaInitState,
		clientId:       clientId,
		cpConfig:       cpConfig,
	}
}

func (s *stream) DeltaDiscoveryRequest(resourceType model.ResourceType) error {
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
	if value, found := s.deltaInitState[resourceType]; found {
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
	rs, nameToVersion, err := util.ToDeltaCoreResourceList(resp)
	if err != nil {
		return UpstreamResponse{}, err
	}
	s.latestNonce[rs.GetItemType()] = resp.Nonce

	// when map is empty that means we are doing the first request
	// it has to be called before `getVersionMap`
	isInitialRequest := len(s.deltaInitState) == 0

	s.deltaInitState[rs.GetItemType()] = s.updateVersionMap(rs.GetItemType(), nameToVersion)
	return UpstreamResponse{
		ControlPlaneId:       resp.GetControlPlane().GetIdentifier(),
		Type:                 rs.GetItemType(),
		AddedResources:       rs,
		RemovedResourceNames: resp.RemovedResources,
		IsInitialRequest:     isInitialRequest,
	}, nil
}

func (s *stream) ACK(typ string) error {
	latestNonce, found := s.latestNonce[core_model.ResourceType(typ)]
	if !found {
		return nil
	}
	err := s.streamClient.Send(&envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: latestNonce,
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		TypeUrl: typ,
	})
	return err
}

func (s *stream) NACK(typ string, err error) error {
	latestNonce, found := s.latestNonce[core_model.ResourceType(typ)]
	if !found {
		return nil
	}
	return s.streamClient.Send(&envoy_sd.DeltaDiscoveryRequest{
		InitialResourceVersions: map[string]string{},
		ResponseNonce:           latestNonce,
		TypeUrl:                 typ,
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

func (s *stream) updateVersionMap(typ core_model.ResourceType, nameToVersion cache_v2.NameToVersion) cache_v2.NameToVersion {
	var versions map[string]string
	if value, found := s.deltaInitState[typ]; found {
		versions = value
	} else {
		return nameToVersion
	}
	for name, version := range nameToVersion {
		versions[name] = version
	}
	return versions
}