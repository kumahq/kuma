package util

import (
	"fmt"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/util/proto"
)

func ToCoreResourceList(response *envoy.DiscoveryResponse) (model.ResourceList, error) {
	krs := []*mesh_proto.KumaResource{}
	for _, r := range response.Resources {
		kr := &mesh_proto.KumaResource{}
		if err := ptypes.UnmarshalAny(r, kr); err != nil {
			return nil, err
		}
		krs = append(krs, kr)
	}
	return toResources(model.ResourceType(response.TypeUrl), krs)
}

func ToEnvoyResources(rlist model.ResourceList) ([]envoy_types.Resource, error) {
	rv := make([]envoy_types.Resource, 0, len(rlist.GetItems()))
	for _, r := range rlist.GetItems() {
		pbany, err := proto.MarshalAnyDeterministic(r.GetSpec())
		if err != nil {
			return nil, err
		}
		rv = append(rv, &mesh_proto.KumaResource{
			Meta: &mesh_proto.KumaResource_Meta{
				Name:             r.GetMeta().GetName(),
				Mesh:             r.GetMeta().GetMesh(),
				CreationTime:     proto.MustTimestampProto(r.GetMeta().GetCreationTime()),
				ModificationTime: proto.MustTimestampProto(r.GetMeta().GetModificationTime()),
				Version:          r.GetMeta().GetVersion(),
			},
			Spec: pbany,
		})
	}
	return rv, nil
}

func AddPrefixToNames(rs []model.Resource, prefix string) {
	for _, r := range rs {
		newName := fmt.Sprintf("%s.%s", prefix, r.GetMeta().GetName())
		m := NewResourceMeta(newName, r.GetMeta().GetMesh(), r.GetMeta().GetVersion(),
			r.GetMeta().GetCreationTime(), r.GetMeta().GetModificationTime())
		r.SetMeta(m)
	}
}

func AddSuffixToNames(rs []model.Resource, suffix string) {
	for _, r := range rs {
		newName := fmt.Sprintf("%s.%s", r.GetMeta().GetName(), suffix)
		m := NewResourceMeta(newName, r.GetMeta().GetMesh(), r.GetMeta().GetVersion(),
			r.GetMeta().GetCreationTime(), r.GetMeta().GetModificationTime())
		r.SetMeta(m)
	}
}

func ZoneTag(r model.Resource) string {
	dp := r.GetSpec().(*mesh_proto.Dataplane)
	if dp.IsGateway() {
		return dp.GetNetworking().GetGateway().GetTags()[mesh_proto.ZoneTag]
	}
	return dp.GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ZoneTag]
}

func toResources(resourceType model.ResourceType, krs []*mesh_proto.KumaResource) (model.ResourceList, error) {
	list, err := registry.Global().NewList(resourceType)
	if err != nil {
		return nil, err
	}
	for _, kr := range krs {
		obj, err := registry.Global().NewObject(resourceType)
		if err != nil {
			return nil, err
		}
		err = ptypes.UnmarshalAny(kr.Spec, obj.GetSpec())
		if err != nil {
			return nil, err
		}
		obj.SetMeta(kumaResourceMetaToResourceMeta(kr.Meta))
		if err := list.AddItem(obj); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func StatsOf(status *system_proto.KDSSubscriptionStatus, resourceType model.ResourceType) *system_proto.KDSServiceStats {
	if status == nil {
		return &system_proto.KDSServiceStats{}
	}
	stat, ok := status.Stat[string(resourceType)]
	if !ok {
		stat = &system_proto.KDSServiceStats{}
		status.Stat[string(resourceType)] = stat
	}
	return stat
}
