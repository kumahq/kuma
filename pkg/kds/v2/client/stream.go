package client

import (
	"fmt"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/util"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ DeltaKDSStream = &stream{}

type latestReceived struct {
	nonce         string
	nameToVersion cache_v2.NameToVersion
}

type stream struct {
	streamClient   KDSSyncServiceStream
	latestACKed    map[core_model.ResourceType]string
	latestReceived map[core_model.ResourceType]*latestReceived
	clientId       string
	cpConfig       string
	runtimeInfo    core_runtime.RuntimeInfo
}

type KDSSyncServiceStream interface {
	Send(*envoy_sd.DeltaDiscoveryRequest) error
	Recv() (*envoy_sd.DeltaDiscoveryResponse, error)
}

func NewDeltaKDSStream(s KDSSyncServiceStream, clientId string, runtimeInfo core_runtime.RuntimeInfo, cpConfig string) DeltaKDSStream {
	return &stream{
		streamClient:   s,
		runtimeInfo:    runtimeInfo,
		latestACKed:    make(map[core_model.ResourceType]string),
		latestReceived: make(map[core_model.ResourceType]*latestReceived),
		clientId:       clientId,
		cpConfig:       cpConfig,
	}
}

func (s *stream) DeltaDiscoveryRequest(resourceType core_model.ResourceType) error {
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

	req := &envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: "",
		Node: &envoy_core.Node{
			Id: s.clientId,
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					kds.MetadataFieldVersion:   {Kind: &structpb.Value_StructValue{StructValue: cpVersion}},
					kds.MetadataFieldConfig:    {Kind: &structpb.Value_StringValue{StringValue: s.cpConfig}},
					kds.MetadataControlPlaneId: {Kind: &structpb.Value_StringValue{StringValue: s.runtimeInfo.GetInstanceId()}},
					kds.MetadataFeatures: {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{
						Values: []*structpb.Value{
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureZoneToken}},
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureHashSuffix}},
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureHostnameGeneratorMzSelector}},
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
	// when there isn't nonce it means it's the first request
	isInitialRequest := true
	if _, found := s.latestACKed[rs.GetItemType()]; found {
		isInitialRequest = false
	}
	s.latestReceived[rs.GetItemType()] = &latestReceived{
		nonce:         resp.Nonce,
		nameToVersion: nameToVersion,
	}
	return UpstreamResponse{
		ControlPlaneId:      resp.GetControlPlane().GetIdentifier(),
		Type:                rs.GetItemType(),
		AddedResources:      rs,
		RemovedResourcesKey: s.mapRemovedResources(resp.RemovedResources),
		IsInitialRequest:    isInitialRequest,
	}, err
}

func (s *stream) ACK(resourceType core_model.ResourceType) error {
	latestReceived := s.latestReceived[resourceType]
	if latestReceived == nil {
		return nil
	}
	err := s.streamClient.Send(&envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: latestReceived.nonce,
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		TypeUrl: string(resourceType),
	})
	if err == nil {
		s.latestACKed[resourceType] = latestReceived.nonce
	}
	return err
}

func (s *stream) NACK(resourceType core_model.ResourceType, err error) error {
	latestReceived, found := s.latestReceived[resourceType]
	if !found {
		return nil
	}
	return s.streamClient.Send(&envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce:          latestReceived.nonce,
		ResourceNamesSubscribe: []string{"*"},
		TypeUrl:                string(resourceType),
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

// go-contro-plane cache keeps them as a <resource_name>.<mesh_name>
func (s *stream) mapRemovedResources(removedResourceNames []string) []core_model.ResourceKey {
	removed := []core_model.ResourceKey{}
	for _, resourceName := range removedResourceNames {
		index := strings.LastIndex(resourceName, ".")
		var rk core_model.ResourceKey
		if index != -1 {
			rk = core_model.WithMesh(resourceName[index+1:], resourceName[:index])
		} else {
			rk = core_model.WithoutMesh(resourceName)
		}
		removed = append(removed, rk)
	}
	return removed
}
