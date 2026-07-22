package ingress

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	envoy "github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
)

// tagSets represent map from tags (encoded as string) to number of instances
type tagSets map[serviceKey]uint32

type serviceKey struct {
	mesh string
	tags string
}

type serviceKeySlice []serviceKey

func (s serviceKeySlice) Len() int { return len(s) }
func (s serviceKeySlice) Less(i, j int) bool {
	return s[i].mesh < s[j].mesh || (s[i].mesh == s[j].mesh && s[i].tags < s[j].tags)
}
func (s serviceKeySlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (sk *serviceKey) String() string {
	return fmt.Sprintf("%s.%s", sk.tags, sk.mesh)
}

func (s tagSets) addInstanceOfTags(mesh string, tags envoy.Tags) {
	strTags := tags.String()
	s[serviceKey{tags: strTags, mesh: mesh}]++
}

func (s tagSets) toAvailableServices() []*mesh_proto.ZoneIngress_AvailableService {
	var result []*mesh_proto.ZoneIngress_AvailableService

	var keys []serviceKey
	for key := range s {
		keys = append(keys, key)
	}
	sort.Sort(serviceKeySlice(keys))

	for _, key := range keys {
		tags, _ := envoy.TagsFromString(key.tags) // ignore error since we control how string looks like
		result = append(result, &mesh_proto.ZoneIngress_AvailableService{
			Tags:      tags,
			Instances: s[key],
			Mesh:      key.mesh,
		})
	}
	return result
}

func GetAvailableServices(
	skipAvailableServices map[xds.MeshName]struct{},
	allDataplanes []*core_mesh.DataplaneResource,
	tagFilters []string,
) []*mesh_proto.ZoneIngress_AvailableService {
	return GetIngressAvailableServices(skipAvailableServices, allDataplanes, tagFilters)
}

func GetIngressAvailableServices(
	skipAvailableServices map[xds.MeshName]struct{},
	dataplanes []*core_mesh.DataplaneResource,
	tagFilters []string,
) []*mesh_proto.ZoneIngress_AvailableService {
	tagSets := tagSets{}
	for _, dp := range dataplanes {
		if _, ok := skipAvailableServices[dp.GetMeta().GetMesh()]; ok {
			continue
		}
		for _, dpInbound := range dp.Spec.GetNetworking().GetHealthyInbounds() {
			tags := map[string]string{}
			for key, value := range dpInbound.Tags {
				hasPrefix := func(tagFilter string) bool {
					return strings.HasPrefix(key, tagFilter)
				}
				if len(tagFilters) == 0 || slices.ContainsFunc(tagFilters, hasPrefix) {
					tags[key] = value
				}
			}
			tagSets.addInstanceOfTags(dp.GetMeta().GetMesh(), tags)
		}
	}
	return tagSets.toAvailableServices()
}
