package client

import (
	"fmt"
	"strings"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/kds"
	"github.com/kumahq/kuma/v2/pkg/kds/util"
	cache_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/cache"
	kds_util "github.com/kumahq/kuma/v2/pkg/kds/v2/util"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/v2/pkg/version"
)

var _ DeltaKDSStream = &stream{}

// All methods other than Receive() are non-blocking. It does not wait until the peer CP receives the message.
type DeltaKDSStream interface {
	SendMsg(*envoy_sd.DeltaDiscoveryRequest) error
	BuildDeltaSubScribeRequest(resourceType core_model.ResourceType) *envoy_sd.DeltaDiscoveryRequest
	Receive() (kds_util.UpstreamResponse, error)
	BuildACKRequest(resourceType core_model.ResourceType) *envoy_sd.DeltaDiscoveryRequest
	BuildNACKRequest(resourceType core_model.ResourceType, err error) *envoy_sd.DeltaDiscoveryRequest
	MarkInitialRequestDone(resourceType core_model.ResourceType)
}

type latestReceived struct {
	nonce         string
	nameToVersion cache_v2.NameToVersion
}

type stream struct {
	sync.Mutex
	streamClient       KDSSyncServiceStream
	initialRequestDone map[core_model.ResourceType]bool
	latestReceived     map[core_model.ResourceType]*latestReceived
	clientID           string
	cpConfig           string
	instanceID         string
}

type KDSSyncServiceStream interface {
	Send(*envoy_sd.DeltaDiscoveryRequest) error
	Recv() (*envoy_sd.DeltaDiscoveryResponse, error)
}

func NewDeltaKDSStream(s KDSSyncServiceStream, clientID string, instanceID string, cpConfig string) DeltaKDSStream {
	return &stream{
		streamClient:       s,
		initialRequestDone: make(map[core_model.ResourceType]bool),
		latestReceived:     make(map[core_model.ResourceType]*latestReceived),
		clientID:           clientID,
		cpConfig:           cpConfig,
		instanceID:         instanceID,
	}
}

func (s *stream) SendMsg(request *envoy_sd.DeltaDiscoveryRequest) error {
	return s.streamClient.Send(request)
}

func (s *stream) BuildDeltaSubScribeRequest(resourceType core_model.ResourceType) *envoy_sd.DeltaDiscoveryRequest {
	cpVersion := util_proto.MustToStruct(&system_proto.Version{
		KumaCp: &system_proto.KumaCpVersion{
			Version:   kuma_version.Build.Version,
			GitTag:    kuma_version.Build.GitTag,
			GitCommit: kuma_version.Build.GitCommit,
			BuildDate: kuma_version.Build.BuildDate,
		},
	})

	req := &envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: "",
		Node: &envoy_core.Node{
			Id: s.clientID,
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					kds.MetadataFieldVersion:   {Kind: &structpb.Value_StructValue{StructValue: cpVersion}},
					kds.MetadataFieldConfig:    {Kind: &structpb.Value_StringValue{StringValue: s.cpConfig}},
					kds.MetadataControlPlaneId: {Kind: &structpb.Value_StringValue{StringValue: s.instanceID}},
					kds.MetadataFeatures: {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{
						Values: []*structpb.Value{
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureZoneToken}},
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureHashSuffix}},
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureHostnameGeneratorMzSelector}},
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureProducerPolicyFlow}},
							{Kind: &structpb.Value_StringValue{StringValue: kds.FeatureOptionalTopLevelTargetRef}},
						},
					}}},
				},
			},
		},
		ResourceNamesSubscribe: []string{"*"},
		TypeUrl:                string(resourceType),
	}
	return req
}

func (s *stream) Receive() (kds_util.UpstreamResponse, error) {
	resp, err := s.streamClient.Recv()
	if err != nil {
		return kds_util.UpstreamResponse{}, err
	}
	rs, nameToVersion, err := util.ToDeltaCoreResourceList(resp)
	if err != nil {
		return kds_util.UpstreamResponse{}, err
	}
	// when there isn't nonce it means it's the first request
	s.Lock()
	isInitialRequest := !s.initialRequestDone[rs.GetItemType()]
	s.Unlock()
	s.latestReceived[rs.GetItemType()] = &latestReceived{
		nonce:         resp.Nonce,
		nameToVersion: nameToVersion,
	}
	return kds_util.UpstreamResponse{
		ControlPlaneId:      resp.GetControlPlane().GetIdentifier(),
		Type:                rs.GetItemType(),
		AddedResources:      rs,
		RemovedResourcesKey: s.mapRemovedResources(resp.RemovedResources),
		IsInitialRequest:    isInitialRequest,
		InvalidResourcesKey: []core_model.ResourceKey{},
	}, err
}

func (s *stream) BuildACKRequest(resourceType core_model.ResourceType) *envoy_sd.DeltaDiscoveryRequest {
	latestReceived := s.latestReceived[resourceType]
	if latestReceived == nil {
		return nil
	}

	req := &envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: latestReceived.nonce,
		Node: &envoy_core.Node{
			Id: s.clientID,
		},
		TypeUrl: string(resourceType),
	}
	return req
}

func (s *stream) BuildNACKRequest(resourceType core_model.ResourceType, err error) *envoy_sd.DeltaDiscoveryRequest {
	latestReceived, found := s.latestReceived[resourceType]
	if !found {
		return nil
	}
	req := &envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce:          latestReceived.nonce,
		ResourceNamesSubscribe: []string{"*"},
		TypeUrl:                string(resourceType),
		Node: &envoy_core.Node{
			Id: s.clientID,
		},
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	}
	return req
}

func (s *stream) MarkInitialRequestDone(resourceType core_model.ResourceType) {
	s.Lock()
	s.initialRequestDone[resourceType] = true
	s.Unlock()
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
