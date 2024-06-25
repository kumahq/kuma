package meshservice

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/maps"
)

func MatchDataplanesWithMeshServices(
	dpps []*core_mesh.DataplaneResource,
	meshServices []*meshservice_api.MeshServiceResource,
	matchOnlyHealthy bool,
) map[*meshservice_api.MeshServiceResource][]*core_mesh.DataplaneResource {
	result := map[*meshservice_api.MeshServiceResource][]*core_mesh.DataplaneResource{}

	dppsByNameByTag, dppsByName := indexDpsForMatching(dpps, matchOnlyHealthy)

	for _, ms := range meshServices {
		switch {
		case ms.Spec.Selector.DataplaneRef != nil:
			result[ms] = matchByRef(ms, dppsByName)
		case ms.Spec.Selector.DataplaneTags != nil:
			result[ms] = matchByTags(ms, dppsByNameByTag)
		default:
			result[ms] = nil
		}
	}

	return result
}

func indexDpsForMatching(
	dpps []*core_mesh.DataplaneResource,
	matchOnlyHealthy bool,
) (
	map[string]map[string]*core_mesh.DataplaneResource,
	map[string]*core_mesh.DataplaneResource,
) {
	// Map of tag pair in a context of a mesh (ex. app:redis from default mesh) to dpp name to dpp.
	dppsByNameByTag := map[string]map[string]*core_mesh.DataplaneResource{}
	dppsByName := map[string]*core_mesh.DataplaneResource{}

	for _, dpp := range dpps {
		inbounds := dpp.Spec.GetNetworking().GetInbound()
		if matchOnlyHealthy {
			inbounds = dpp.Spec.GetNetworking().GetHealthyInbounds()
		}

		for _, inbound := range inbounds {
			for tagName, tagValue := range inbound.GetTags() {
				tag := fmt.Sprintf("%s;%s;%s", dpp.Meta.GetMesh(), tagName, tagValue)
				dataplanes, ok := dppsByNameByTag[tag]
				if !ok {
					dataplanes = map[string]*core_mesh.DataplaneResource{}
					dppsByNameByTag[tag] = dataplanes
				}
				dataplanes[dpp.Meta.GetName()] = dpp
			}
		}
		if len(inbounds) > 0 {
			dppsByName[fmt.Sprintf("%s;%s", dpp.Meta.GetMesh(), dpp.Meta.GetName())] = dpp
		}
	}
	return dppsByNameByTag, dppsByName
}

func matchByRef(
	ms *meshservice_api.MeshServiceResource,
	dppsByName map[string]*core_mesh.DataplaneResource,
) []*core_mesh.DataplaneResource {
	key := fmt.Sprintf("%s;%s", ms.Meta.GetMesh(), ms.Spec.Selector.DataplaneRef.Name)
	if dpp, ok := dppsByName[key]; ok {
		return []*core_mesh.DataplaneResource{dpp}
	}
	return nil
}

func matchByTags(
	ms *meshservice_api.MeshServiceResource,
	dppsByNameByTag map[string]map[string]*core_mesh.DataplaneResource,
) []*core_mesh.DataplaneResource {
	// Find the shortest map of dpps by name that matches the tags
	// For example, if we have MeshService with selector of `app: redis` and `k8s.kuma.io/namespace: kuma-demo`
	// and DPPs:
	// * `app: redis`, `k8s.kuma.io/namespace: kuma-demo`
	// * `app: demo-app`, `k8s.kuma.io/namespace: kuma-demo`
	// It's better to grab the shortest list of DPPs (app:redis) to reduce number of operations (.Matches) for the next step.
	var shortestDppMap map[string]*core_mesh.DataplaneResource
	for tagName, tagValue := range ms.Spec.Selector.DataplaneTags {
		tagsKey := fmt.Sprintf("%s;%s;%s", ms.GetMeta().GetMesh(), tagName, tagValue)
		if dppsByName, ok := dppsByNameByTag[tagsKey]; ok {
			if shortestDppMap == nil || len(dppsByName) < len(shortestDppMap) {
				shortestDppMap = dppsByName
			}
		} else {
			// No proxies will match this pair of tags, no point in going further.
			shortestDppMap = nil
			break
		}
	}

	// Go over the shortest list of data plane proxies and pick only proxies that matches all tags.
	var dpps []*core_mesh.DataplaneResource
	for _, dppName := range maps.SortedKeys(shortestDppMap) {
		dpp := shortestDppMap[dppName]
		if dpp.Spec.Matches(mesh_proto.TagSelector(ms.Spec.Selector.DataplaneTags)) {
			dpps = append(dpps, dpp)
		}
	}
	return dpps
}
