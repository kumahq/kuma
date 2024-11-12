package util

import (
	"fmt"
	"maps"
	"strings"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
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

func ToDeltaCoreResourceList(response *envoy_sd.DeltaDiscoveryResponse) (model.ResourceList, cache_v2.NameToVersion, error) {
	krs := []*mesh_proto.KumaResource{}
	resourceVersions := cache_v2.NameToVersion{}
	for _, r := range response.Resources {
		kr := &mesh_proto.KumaResource{}
		if err := util_proto.UnmarshalAnyTo(r.GetResource(), kr); err != nil {
			return nil, nil, err
		}
		krs = append(krs, kr)
		resourceVersions[kr.GetMeta().GetName()] = r.Version
	}
	list, err := toResources(model.ResourceType(response.TypeUrl), krs)
	if err != nil {
		return list, resourceVersions, err
	}
	return list, resourceVersions, err
}

func ToEnvoyResources(rlist model.ResourceList) ([]envoy_types.Resource, error) {
	rv := make([]envoy_types.Resource, 0, len(rlist.GetItems()))
	for _, r := range rlist.GetItems() {
		pbany, err := model.ToAny(r.GetSpec())
		if err != nil {
			return nil, err
		}
		var pbanyStatus *anypb.Any
		if r.Descriptor().HasStatus {
			pbanyStatus, err = model.ToAny(r.GetStatus())
			if err != nil {
				return nil, err
			}
		}
		rv = append(rv, &mesh_proto.KumaResource{
			Meta: &mesh_proto.KumaResource_Meta{
				Name:    r.GetMeta().GetName(),
				Mesh:    r.GetMeta().GetMesh(),
				Labels:  maps.Clone(r.GetMeta().GetLabels()),
				Version: "",
			},
			Spec:   pbany,
			Status: pbanyStatus,
		})
	}
	return rv, nil
}

func AddPrefixToNames(rs []model.Resource, prefix string) {
	for _, r := range rs {
		r.SetMeta(CloneResourceMeta(
			r.GetMeta(),
			WithName(fmt.Sprintf("%s.%s", prefix, r.GetMeta().GetName())),
		))
	}
}

func AddPrefixToResourceKeyNames(rk []model.ResourceKey, prefix string) []model.ResourceKey {
	for idx, r := range rk {
		rk[idx].Name = fmt.Sprintf("%s.%s", prefix, r.Name)
	}
	return rk
}

func AddSuffixToNames(rs []model.Resource, suffix string) {
	for _, r := range rs {
		r.SetMeta(CloneResourceMeta(
			r.GetMeta(),
			WithName(fmt.Sprintf("%s.%s", r.GetMeta().GetName(), suffix)),
		))
	}
}

func AddSuffixToResourceKeyNames(rk []model.ResourceKey, suffix string) []model.ResourceKey {
	for idx, r := range rk {
		rk[idx].Name = fmt.Sprintf("%s.%s", r.Name, suffix)
	}
	return rk
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
		// todo(jakubdyszkiewicz): consider replacing this whole function with just model.ZoneOfResource(r)
		return model.ZoneOfResource(r)
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
		if err = model.FromAny(kr.Spec, obj.GetSpec()); err != nil {
			return nil, err
		}
		if obj.Descriptor().HasStatus && kr.Status != nil {
			if err = model.FromAny(kr.Status, obj.GetStatus()); err != nil {
				return nil, err
			}
		}
		obj.SetMeta(&resourceMeta{
			name:   kr.GetMeta().GetName(),
			mesh:   kr.GetMeta().GetMesh(),
			labels: maps.Clone(kr.GetMeta().GetLabels()),
		})
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

type cloneResource struct {
	name          string
	withoutStatus bool
}

func WithResourceName(name string) CloneResourceOpt {
	return func(m *cloneResource) {
		m.name = name
	}
}

func WithoutStatus() CloneResourceOpt {
	return func(m *cloneResource) {
		m.withoutStatus = true
	}
}

type CloneResourceOpt func(*cloneResource)

func CloneResource(res core_model.Resource, fs ...CloneResourceOpt) core_model.Resource {
	opts := &cloneResource{}
	for _, f := range fs {
		f(opts)
	}

	newObj := res.Descriptor().NewObject()
	newMeta := CloneResourceMeta(res.GetMeta(), WithName(opts.name))
	newObj.SetMeta(newMeta)
	_ = newObj.SetSpec(res.GetSpec())
	if newObj.Descriptor().HasStatus && !opts.withoutStatus {
		_ = newObj.SetStatus(res.GetStatus())
	}
	return newObj
}
