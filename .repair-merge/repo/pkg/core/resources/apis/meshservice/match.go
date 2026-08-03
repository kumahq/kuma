package meshservice

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func MatchDataplanesWithMeshServices(
	dpps []*core_mesh.DataplaneResource,
	meshServices []*meshservice_api.MeshServiceResource,
	matchOnlyHealthy bool,
) map[*meshservice_api.MeshServiceResource][]*core_mesh.DataplaneResource {
	result := map[*meshservice_api.MeshServiceResource][]*core_mesh.DataplaneResource{}

	dppsByName := indexDpsForMatching(dpps, matchOnlyHealthy)
	dppsByMesh := indexDpsByMesh(dpps, matchOnlyHealthy)

	for _, ms := range meshServices {
		switch {
		case ms.Spec.Selector.DataplaneRef != nil:
			result[ms] = matchByRef(ms, dppsByName)
		case ms.Spec.Selector.DataplaneLabels != nil:
			result[ms] = matchByLabels(ms, dppsByMesh)
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
	case meshService.Selector.DataplaneLabels != nil:
		return meshService.Selector.DataplaneLabels.Matches(dpp.Meta.GetLabels())
	default:
		return false
	}
}

func indexDpsForMatching(
	dpps []*core_mesh.DataplaneResource,
	matchOnlyHealthy bool,
) map[model.ResourceKey]*core_mesh.DataplaneResource {
	dppsByName := map[model.ResourceKey]*core_mesh.DataplaneResource{}

	for _, dpp := range dpps {
		inbounds := dpp.Spec.GetNetworking().GetInbound()
		if matchOnlyHealthy {
			inbounds = dpp.Spec.GetNetworking().GetHealthyInbounds()
		}
		if len(inbounds) > 0 {
			dppsByName[model.MetaToResourceKey(dpp.Meta)] = dpp
		}
	}
	return dppsByName
}

func indexDpsByMesh(
	dpps []*core_mesh.DataplaneResource,
	matchOnlyHealthy bool,
) map[string][]*core_mesh.DataplaneResource {
	result := map[string][]*core_mesh.DataplaneResource{}

	for _, dpp := range dpps {
		inbounds := dpp.Spec.GetNetworking().GetInbound()
		if matchOnlyHealthy {
			inbounds = dpp.Spec.GetNetworking().GetHealthyInbounds()
		}
		if len(inbounds) == 0 {
			continue
		}

		mesh := dpp.Meta.GetMesh()
		result[mesh] = append(result[mesh], dpp)
	}

	return result
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

func matchByLabels(
	ms *meshservice_api.MeshServiceResource,
	dppsByMesh map[string][]*core_mesh.DataplaneResource,
) []*core_mesh.DataplaneResource {
	meshDpps := dppsByMesh[ms.GetMeta().GetMesh()]

	var matched []*core_mesh.DataplaneResource
	for _, dpp := range meshDpps {
		if ms.Spec.Selector.DataplaneLabels.Matches(dpp.Meta.GetLabels()) {
			matched = append(matched, dpp)
		}
	}

	return matched
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
