package ingress

import (
	"context"
	"reflect"
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

// tagSets represent map from tags (encoded as string) to number of instances
type tagSets map[string]uint32

func (s tagSets) addInstanceOfTags(tags envoy.Tags) {
	strTags := tags.String()
	s[strTags]++
}

func (s tagSets) toAvailableServices() []*mesh_proto.Dataplane_Networking_Ingress_AvailableService {
	var result []*mesh_proto.Dataplane_Networking_Ingress_AvailableService

	var keys []string
	for key := range s {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		tags, _ := envoy.TagsFromString(key) // ignore error since we control how string looks like
		result = append(result, &mesh_proto.Dataplane_Networking_Ingress_AvailableService{
			Tags:      tags,
			Instances: s[key],
		})
	}
	return result
}

func UpdateAvailableServices(ctx context.Context, rm manager.ResourceManager, ingress *core_mesh.DataplaneResource, others []*core_mesh.DataplaneResource) error {
	availableServices := GetIngressAvailableServices(others)
	if reflect.DeepEqual(availableServices, ingress.Spec.GetNetworking().GetIngress().GetAvailableServices()) {
		return nil
	}
	ingress.Spec.Networking.Ingress.AvailableServices = availableServices
	if err := rm.Update(ctx, ingress); err != nil {
		return err
	}
	return nil
}

func GetIngressAvailableServices(others []*core_mesh.DataplaneResource) []*mesh_proto.Dataplane_Networking_Ingress_AvailableService {
	tagSets := tagSets{}
	for _, dp := range others {
		if dp.Spec.IsIngress() {
			continue
		}
		for _, dpInbound := range dp.Spec.GetNetworking().GetInbound() {
			tagSets.addInstanceOfTags(dpInbound.Tags)
		}
	}
	return tagSets.toAvailableServices()
}
