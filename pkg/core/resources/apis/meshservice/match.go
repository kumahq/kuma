package meshservice

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/maps"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

type dppsByNameByTagKey struct {
	mesh     string
	tagName  string
	tagValue string
}

// DppsByNameByTag is a map of tag pair in a context of a mesh (ex. app:redis from default mesh) to dpp name to dpp.
// While this could be a map[dppsByNameByTagKey][]*core_mesh.DataplaneResource, we also want to get rid of duplicates.
// For example, if a DPP has 2 inbounds with app:xyz and MeshService matches app:xyz, we only want to
// have one occurrence of a dpp as a result.
type DppsByNameByTag = map[dppsByNameByTagKey]map[string]*core_mesh.DataplaneResource

func MatchDataplanesWithMeshServices(
	dpps []*core_mesh.DataplaneResource,
	meshServices []*meshservice_api.MeshServiceResource,
	matchOnlyHealthy bool,
) map[*meshservice_api.MeshServiceResource][]*core_mesh.DataplaneResource {
	result := map[*meshservice_api.MeshServiceResource][]*core_mesh.DataplaneResource{}

	dppsByNameByTag, dppsByName, dppsByNameByLabel := indexDpsForMatching(dpps, matchOnlyHealthy)

	for _, ms := range meshServices {
		switch {
		case ms.Spec.Selector.DataplaneRef != nil:
			result[ms] = matchByRef(ms, dppsByName)
		case ms.Spec.Selector.DataplaneTags != nil:
			result[ms] = matchByTags(ms, dppsByNameByTag)
		case ms.Spec.Selector.Dataplane != nil:
			result[ms] = matchByLabels(ms, dppsByNameByLabel)
		default:
			result[ms] = nil
		}
	}

	return result
}

func MatchesDataplane(meshService *meshservice_api.MeshService, dpp *core_mesh.DataplaneResource) bool {
	switch {
	case meshService.Selector.DataplaneRef != nil:
		return meshService.Selector.DataplaneRef.Name == dpp.GetMeta().GetName()
	case meshService.Selector.DataplaneTags != nil:
		return dpp.Spec.Matches(pointer.Deref(meshService.Selector.DataplaneTags))
	case meshService.Selector.Dataplane != nil:
		return meshService.Selector.Dataplane.Matches(dpp.GetMeta().GetLabels())
	default:
		return false
	}
}

// DppsByNameByLabel indexes dataplanes by their resource labels (not inbound tags)
type DppsByNameByLabel = map[dppsByNameByTagKey]map[string]*core_mesh.DataplaneResource

func indexDpsForMatching(
	dpps []*core_mesh.DataplaneResource,
	matchOnlyHealthy bool,
) (
	DppsByNameByTag,
	map[model.ResourceKey]*core_mesh.DataplaneResource,
	DppsByNameByLabel,
) {
	dppsByNameByTag := DppsByNameByTag{}
	dppsByName := map[model.ResourceKey]*core_mesh.DataplaneResource{}
	dppsByNameByLabel := DppsByNameByLabel{}

	for _, dpp := range dpps {
		inbounds := dpp.Spec.GetNetworking().GetInbound()
		if matchOnlyHealthy {
			inbounds = dpp.Spec.GetNetworking().GetHealthyInbounds()
		}

		for _, inbound := range inbounds {
			for tagName, tagValue := range inbound.GetTags() {
				key := dppsByNameByTagKey{
					mesh:     dpp.Meta.GetMesh(),
					tagName:  tagName,
					tagValue: tagValue,
				}
				dataplanes, ok := dppsByNameByTag[key]
				if !ok {
					dataplanes = map[string]*core_mesh.DataplaneResource{}
					dppsByNameByTag[key] = dataplanes
				}
				dataplanes[dpp.Meta.GetName()] = dpp
			}
		}
		if len(inbounds) > 0 {
			dppsByName[model.MetaToResourceKey(dpp.Meta)] = dpp
		}

		// Index by resource labels for dataplane.matchLabels selector
		for labelName, labelValue := range dpp.Meta.GetLabels() {
			key := dppsByNameByTagKey{
				mesh:     dpp.Meta.GetMesh(),
				tagName:  labelName,
				tagValue: labelValue,
			}
			dataplanes, ok := dppsByNameByLabel[key]
			if !ok {
				dataplanes = map[string]*core_mesh.DataplaneResource{}
				dppsByNameByLabel[key] = dataplanes
			}
			dataplanes[dpp.Meta.GetName()] = dpp
		}
	}
	return dppsByNameByTag, dppsByName, dppsByNameByLabel
}

func matchByRef(
	ms *meshservice_api.MeshServiceResource,
	dppsByName map[model.ResourceKey]*core_mesh.DataplaneResource,
) []*core_mesh.DataplaneResource {
	key := model.ResourceKey{
		Mesh: ms.Meta.GetMesh(),
		Name: ms.Spec.Selector.DataplaneRef.Name,
	}
	if dpp, ok := dppsByName[key]; ok {
		return []*core_mesh.DataplaneResource{dpp}
	}
	return nil
}

func matchByTags(
	ms *meshservice_api.MeshServiceResource,
	dppsByNameByTag DppsByNameByTag,
) []*core_mesh.DataplaneResource {
	// For every tag key/value pair of MeshService's selector, find the set of DPPs matched by that pair.
	// Then take the smallest set of all sets.
	var shortestDppMap map[string]*core_mesh.DataplaneResource
	for tagName, tagValue := range pointer.Deref(ms.Spec.Selector.DataplaneTags) {
		tagsKey := dppsByNameByTagKey{
			mesh:     ms.GetMeta().GetMesh(),
			tagName:  tagName,
			tagValue: tagValue,
		}
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
		if dpp.Spec.Matches(pointer.Deref(ms.Spec.Selector.DataplaneTags)) {
			dpps = append(dpps, dpp)
		}
	}
	return dpps
}

func matchByLabels(
	ms *meshservice_api.MeshServiceResource,
	dppsByNameByLabel DppsByNameByLabel,
) []*core_mesh.DataplaneResource {
	// For every label key/value pair of MeshService's selector, find the set of DPPs matched by that pair.
	// Then take the smallest set of all sets.
	var shortestDppMap map[string]*core_mesh.DataplaneResource
	for labelName, labelValue := range pointer.Deref(ms.Spec.Selector.Dataplane.MatchLabels) {
		labelKey := dppsByNameByTagKey{
			mesh:     ms.GetMeta().GetMesh(),
			tagName:  labelName,
			tagValue: labelValue,
		}
		if dppsByName, ok := dppsByNameByLabel[labelKey]; ok {
			if shortestDppMap == nil || len(dppsByName) < len(shortestDppMap) {
				shortestDppMap = dppsByName
			}
		} else {
			// No proxies will match this pair of labels, no point in going further.
			shortestDppMap = nil
			break
		}
	}

	// Go over the shortest list of data plane proxies and pick only proxies that match all labels.
	var dpps []*core_mesh.DataplaneResource
	for _, dppName := range maps.SortedKeys(shortestDppMap) {
		dpp := shortestDppMap[dppName]
		if ms.Spec.Selector.Dataplane.Matches(dpp.Meta.GetLabels()) {
			dpps = append(dpps, dpp)
		}
	}
	return dpps
}

func MatchInboundWithMeshServicePort(inbound *mesh_proto.Dataplane_Networking_Inbound, meshServicePort meshservice_api.Port) bool {
	switch pointer.Deref(meshServicePort.TargetPort).Type {
	case intstr.Int:
		return uint32(pointer.Deref(meshServicePort.TargetPort).IntVal) == inbound.Port
	case intstr.String:
		return pointer.Deref(meshServicePort.TargetPort).StrVal == inbound.Name
	}
	return false
}
