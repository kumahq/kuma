package ingress

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	envoy "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/topology"
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
	meshGateways []*core_mesh.MeshGatewayResource,
	externalServices []*core_mesh.ExternalServiceResource,
	tagFilters []string,
) []*mesh_proto.ZoneIngress_AvailableService {
	availableServices := GetIngressAvailableServices(skipAvailableServices, allDataplanes, tagFilters)
	availableExternalServices := GetExternalAvailableServices(externalServices)
	availableServices = append(availableServices, availableExternalServices...)

	meshGatewayDataplanes := getMeshGateways(allDataplanes, meshGateways)

	for _, meshGateways := range meshGatewayDataplanes {
		availableMeshGatewayListeners := getIngressAvailableMeshGateways(
			meshGateways.Mesh,
			meshGateways.Gateways,
			meshGateways.Dataplanes,
		)
		availableServices = append(availableServices, availableMeshGatewayListeners...)
	}

	return availableServices
}

// MeshGatewayDataplanes is a helper type to hold the MeshGateways and Dataplanes for a mesh.
type MeshGatewayDataplanes struct {
	Mesh       string
	Gateways   []*core_mesh.MeshGatewayResource
	Dataplanes []*core_mesh.DataplaneResource
}

func getMeshGateways(
	dataplanes []*core_mesh.DataplaneResource,
	meshGateways []*core_mesh.MeshGatewayResource,
) []MeshGatewayDataplanes {
	meshGatewayDataplanes := []MeshGatewayDataplanes{}

	meshGatewaysByMesh := map[xds.MeshName][]*core_mesh.MeshGatewayResource{}
	for _, gateway := range meshGateways {
		gateways := meshGatewaysByMesh[gateway.GetMeta().GetMesh()]
		meshGatewaysByMesh[gateway.GetMeta().GetMesh()] = append(gateways, gateway)
	}

	dataplanesByMesh := map[xds.MeshName][]*core_mesh.DataplaneResource{}
	for _, dataplane := range dataplanes {
		if !dataplane.Spec.IsBuiltinGateway() {
			continue
		}
		dataplanes := dataplanesByMesh[dataplane.GetMeta().GetMesh()]
		dataplanesByMesh[dataplane.GetMeta().GetMesh()] = append(dataplanes, dataplane)
	}

	for meshName, meshGateways := range meshGatewaysByMesh {
		dataplanes := dataplanesByMesh[meshName]

		meshGatewayDataplanes = append(meshGatewayDataplanes, MeshGatewayDataplanes{
			Mesh:       meshName,
			Gateways:   meshGateways,
			Dataplanes: dataplanes,
		})
	}

	return meshGatewayDataplanes
}

func getIngressAvailableMeshGateways(meshName string, meshGateways []*core_mesh.MeshGatewayResource, dataplanes []*core_mesh.DataplaneResource) []*mesh_proto.ZoneIngress_AvailableService {
	endpoints := topology.CrossMeshEndpointTags(meshGateways, dataplanes)

	tagSets := tagSets{}
	for _, endpointTags := range endpoints {
		tagSets.addInstanceOfTags(meshName, mesh_proto.MergeAs[envoy.Tags](
			map[string]string{
				mesh_proto.MeshTag: meshName,
			},
			endpointTags,
		))
	}

	return tagSets.toAvailableServices()
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

func GetExternalAvailableServices(others []*core_mesh.ExternalServiceResource) []*mesh_proto.ZoneIngress_AvailableService {
	tagSets := tagSets{}
	for _, es := range others {
		tagSets.addInstanceOfTags(es.GetMeta().GetMesh(), es.Spec.Tags)
	}

	availableServices := tagSets.toAvailableServices()
	for _, as := range availableServices {
		as.ExternalService = true
	}
	return availableServices
}
