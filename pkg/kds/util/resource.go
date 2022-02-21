package util

import (
	"fmt"
	"strings"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func ToCoreResourceList(response *envoy_sd.DiscoveryResponse) (model.ResourceList, error) {
	krs := []*mesh_proto.KumaResource{}
	for _, r := range response.Resources {
		kr := &mesh_proto.KumaResource{}
		if err := util_proto.UnmarshalAnyTo(r, kr); err != nil {
			return nil, err
		}
		krs = append(krs, kr)
	}
	return toResources(model.ResourceType(response.TypeUrl), krs)
}

func ToEnvoyResources(rlist model.ResourceList) ([]envoy_types.Resource, error) {
	rv := make([]envoy_types.Resource, 0, len(rlist.GetItems()))
	for _, r := range rlist.GetItems() {
		pbany, err := util_proto.MarshalAnyDeterministic(r.GetSpec())
		if err != nil {
			return nil, err
		}
		rv = append(rv, &mesh_proto.KumaResource{
			Meta: &mesh_proto.KumaResource_Meta{
				Name: r.GetMeta().GetName(),
				Mesh: r.GetMeta().GetMesh(),
				// KDS ResourceMeta only contains name and mesh.
				// The rest is managed by the receiver of resources anyways. See ResourceSyncer#Sync
				//
				// backwards compatibility
				// Right now we send creation and modification time because the old versions of Kuma CP expects them to be present.
				CreationTime:     util_proto.MustTimestampProto(time.Unix(0, 0)),
				ModificationTime: util_proto.MustTimestampProto(time.Unix(0, 0)),
				Version:          "",
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

func ResourceNameHasAtLeastOneOfPrefixes(resName string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(resName, prefix) {
			return true
		}
	}

	return false
}

func ZoneTag(r model.Resource) string {
	switch res := r.GetSpec().(type) {
	case *mesh_proto.Dataplane:
		if res.GetNetworking().GetGateway() != nil {
			return res.GetNetworking().GetGateway().GetTags()[mesh_proto.ZoneTag]
		}
		return res.GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ZoneTag]
	case *mesh_proto.ZoneIngress:
		return res.GetZone()
	case *mesh_proto.ZoneEgress:
		return res.GetZone()
	default:
		return ""
	}
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
		if err = util_proto.UnmarshalAnyToV2(kr.Spec, obj.GetSpec()); err != nil {
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
