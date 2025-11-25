package client

import (
	"context"
	"fmt"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/v2/pkg/core/runtime"
	"github.com/kumahq/kuma/v2/pkg/kds"
	"github.com/kumahq/kuma/v2/pkg/kds/util"
	cache_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/cache"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/v2/pkg/version"
)

var _ DeltaKDSStream = &stream{}

type latestReceived struct {
	nonce         string
	nameToVersion cache_v2.NameToVersion
}

type stream struct {
	streamClient       KDSSyncServiceStream
	initialRequestDone map[core_model.ResourceType]bool
	latestReceived     map[core_model.ResourceType]*latestReceived
	clientId           string
	cpConfig           string
	runtimeInfo        core_runtime.RuntimeInfo

	sendCh chan *envoy_sd.DeltaDiscoveryRequest
	recvCh chan *envoy_sd.DeltaDiscoveryResponse
	ctx    context.Context
	cancel context.CancelCauseFunc
}

type KDSSyncServiceStream interface {
	Send(*envoy_sd.DeltaDiscoveryRequest) error
	Recv() (*envoy_sd.DeltaDiscoveryResponse, error)
	Context() context.Context
}

func NewDeltaKDSStream(
	s KDSSyncServiceStream,
	clientId string,
	runtimeInfo core_runtime.RuntimeInfo,
	cpConfig string,
	numberOfDistinctTypes int,
) DeltaKDSStream {
	ctx, cancel := context.WithCancelCause(s.Context())

	// In theory capacity == numberOfDistinctTypes would be enough:
	//
	//   - sendCh: we enqueue one initial DiscoveryRequest per type and only enqueue
	//     an ACK or NACK after the previous message for that type has been sent
	//   - recvCh: the server sends at most one DiscoveryResponse per type and waits
	//     for an ACK or NACK before sending the next one
	//
	// To be safer and to tolerate unexpected client or server behavior and future
	// changes, we add an extra safety margin.
	capacity := 2*numberOfDistinctTypes + 10

	stream := &stream{
		streamClient:       s,
		runtimeInfo:        runtimeInfo,
		initialRequestDone: make(map[core_model.ResourceType]bool),
		latestReceived:     make(map[core_model.ResourceType]*latestReceived),
		clientId:           clientId,
		cpConfig:           cpConfig,
		sendCh:             make(chan *envoy_sd.DeltaDiscoveryRequest, capacity),
		recvCh:             make(chan *envoy_sd.DeltaDiscoveryResponse, capacity),
		ctx:                ctx,
		cancel:             cancel,
	}

	go stream.sendLoop()
	go stream.recvLoop()

	return stream
}

func (s *stream) sendLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case req := <-s.sendCh:
			if err := s.streamClient.Send(req); err != nil {
				s.cancel(err)
				return
			}
		}
	}
}

func (s *stream) recvLoop() {
	for {
		resp, err := s.streamClient.Recv()
		if err != nil {
			s.cancel(err)
			return
		}
		select {
		case <-s.ctx.Done():
			return
		case s.recvCh <- resp:
		}
	}
}

func (s *stream) send(req *envoy_sd.DeltaDiscoveryRequest) error {
	select {
	case <-s.ctx.Done():
		return context.Cause(s.ctx)
	case s.sendCh <- req:
		return nil
	}
}

func (s *stream) recv() (*envoy_sd.DeltaDiscoveryResponse, error) {
	select {
	case <-s.ctx.Done():
		return nil, context.Cause(s.ctx)
	case resp := <-s.recvCh:
		return resp, nil
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
	return s.send(req)
}

func (s *stream) Receive() (UpstreamResponse, error) {
	resp, err := s.recv()
	if err != nil {
		return UpstreamResponse{}, err
	}
	rs, nameToVersion, err := util.ToDeltaCoreResourceList(resp)
	if err != nil {
		return UpstreamResponse{}, err
	}
	// when there isn't nonce it means it's the first request
	isInitialRequest := !s.initialRequestDone[rs.GetItemType()]
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
		InvalidResourcesKey: []core_model.ResourceKey{},
	}, err
}

func (s *stream) ACK(resourceType core_model.ResourceType) error {
	latestReceived := s.latestReceived[resourceType]
	if latestReceived == nil {
		return nil
	}
	err := s.send(&envoy_sd.DeltaDiscoveryRequest{
		ResponseNonce: latestReceived.nonce,
		Node: &envoy_core.Node{
			Id: s.clientId,
		},
		TypeUrl: string(resourceType),
	})
	if err == nil {
		s.initialRequestDone[resourceType] = true
	}
	return err
}

func (s *stream) NACK(resourceType core_model.ResourceType, err error) error {
	latestReceived, found := s.latestReceived[resourceType]
	if !found {
		return nil
	}
	s.initialRequestDone[resourceType] = true
	return s.send(&envoy_sd.DeltaDiscoveryRequest{
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
