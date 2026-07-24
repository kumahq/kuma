package context

import (
	"hash/fnv"
	"slices"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshextsvc "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	motb_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	meshsvc "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	workload_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/workload/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	meshfaultinjection_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

type ResourceMap map[core_model.ResourceType]core_model.ResourceList

func (rm ResourceMap) listOrEmpty(resourceType core_model.ResourceType) core_model.ResourceList {
	list, ok := rm[resourceType]
	if !ok {
		list, err := registry.Global().NewList(resourceType)
		if err != nil {
			panic(err)
		}
		return list
	}
	return list
}

func (rm ResourceMap) Hash() []byte {
	hasher := fnv.New128a()
	for _, k := range maps.SortedKeys(rm) {
		hasher.Write(resourceListXDSHash(rm[k]))
	}
	return hasher.Sum(nil)
}

type Resources struct {
	MeshLocalResources ResourceMap
	CrossMeshResources map[xds.MeshName]ResourceMap
}

func NewResources() Resources {
	return Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{},
		CrossMeshResources: map[xds.MeshName]ResourceMap{},
	}
}

func (r Resources) Get(id kri.Identifier) core_model.Resource {
	// todo: we can probably optimize it by using indexing on ResourceIdentifier
	list := r.ListOrEmpty(id.ResourceType).GetItems()
	if i := slices.IndexFunc(list, func(r core_model.Resource) bool { return kri.From(r) == kri.NoSectionName(id) }); i >= 0 {
		return list[i]
	}
	return nil
}

func (r Resources) ListOrEmpty(resourceType core_model.ResourceType) core_model.ResourceList {
	return r.MeshLocalResources.listOrEmpty(resourceType)
}

func (r Resources) ServiceInsights() *core_mesh.ServiceInsightResourceList {
	return r.ListOrEmpty(core_mesh.ServiceInsightType).(*core_mesh.ServiceInsightResourceList)
}

func (r Resources) ZoneIngresses() *core_mesh.ZoneIngressResourceList {
	return r.ListOrEmpty(core_mesh.ZoneIngressType).(*core_mesh.ZoneIngressResourceList)
}

func (r Resources) ZoneEgresses() *core_mesh.ZoneEgressResourceList {
	return r.ListOrEmpty(core_mesh.ZoneEgressType).(*core_mesh.ZoneEgressResourceList)
}

func (r Resources) Dataplanes() *core_mesh.DataplaneResourceList {
	return r.ListOrEmpty(core_mesh.DataplaneType).(*core_mesh.DataplaneResourceList)
}

func (r Resources) Secrets() *system.SecretResourceList {
	return r.ListOrEmpty(system.SecretType).(*system.SecretResourceList)
}

func (r Resources) MeshFaultInjections() *meshfaultinjection_api.MeshFaultInjectionResourceList {
	return r.ListOrEmpty(meshfaultinjection_api.MeshFaultInjectionType).(*meshfaultinjection_api.MeshFaultInjectionResourceList)
}

func (r Resources) Meshes() *core_mesh.MeshResourceList {
	return r.ListOrEmpty(core_mesh.MeshType).(*core_mesh.MeshResourceList)
}

func (r Resources) OtherMeshes(localMesh string) *core_mesh.MeshResourceList {
	otherMeshes := core_mesh.MeshResourceList{}
	for _, m := range r.Meshes().Items {
		if m.GetMeta().GetName() != localMesh {
			otherMeshes.Items = append(otherMeshes.Items, m)
		}
	}
	return &otherMeshes
}

func (r Resources) MeshServices() *meshsvc.MeshServiceResourceList {
	list, ok := r.MeshLocalResources[meshsvc.MeshServiceType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(meshsvc.MeshServiceType)
		if err != nil {
			return &meshsvc.MeshServiceResourceList{}
		}
	}
	return list.(*meshsvc.MeshServiceResourceList)
}

func (r Resources) Workloads() *workload_api.WorkloadResourceList {
	list, ok := r.MeshLocalResources[workload_api.WorkloadType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(workload_api.WorkloadType)
		if err != nil {
			return &workload_api.WorkloadResourceList{}
		}
	}
	return list.(*workload_api.WorkloadResourceList)
}

func (r Resources) MeshExternalServices() *meshextsvc.MeshExternalServiceResourceList {
	list, ok := r.MeshLocalResources[meshextsvc.MeshExternalServiceType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(meshextsvc.MeshExternalServiceType)
		if err != nil {
			return &meshextsvc.MeshExternalServiceResourceList{}
		}
	}
	return list.(*meshextsvc.MeshExternalServiceResourceList)
}

func (r Resources) MeshMultiZoneServices() *meshmzservice_api.MeshMultiZoneServiceResourceList {
	list, ok := r.MeshLocalResources[meshmzservice_api.MeshMultiZoneServiceType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(meshmzservice_api.MeshMultiZoneServiceType)
		if err != nil {
			return &meshmzservice_api.MeshMultiZoneServiceResourceList{}
		}
	}
	return list.(*meshmzservice_api.MeshMultiZoneServiceResourceList)
}

func (r Resources) MeshIdentities() *meshidentity_api.MeshIdentityResourceList {
	list, ok := r.MeshLocalResources[meshidentity_api.MeshIdentityType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(meshidentity_api.MeshIdentityType)
		if err != nil {
			return &meshidentity_api.MeshIdentityResourceList{}
		}
	}
	return list.(*meshidentity_api.MeshIdentityResourceList)
}

func (r Resources) MeshOpenTelemetryBackends() *motb_api.MeshOpenTelemetryBackendResourceList {
	list, ok := r.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(motb_api.MeshOpenTelemetryBackendType)
		if err != nil {
			return &motb_api.MeshOpenTelemetryBackendResourceList{}
		}
	}
	return list.(*motb_api.MeshOpenTelemetryBackendResourceList)
}

func (r Resources) MeshTrusts() *meshtrust_api.MeshTrustResourceList {
	list, ok := r.MeshLocalResources[meshtrust_api.MeshTrustType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(meshtrust_api.MeshTrustType)
		if err != nil {
			return &meshtrust_api.MeshTrustResourceList{}
		}
	}
	return list.(*meshtrust_api.MeshTrustResourceList)
}

func (r Resources) MeshZoneAddresses() *meshzoneaddress_api.MeshZoneAddressResourceList {
	list, ok := r.MeshLocalResources[meshzoneaddress_api.MeshZoneAddressType]
	if !ok {
		var err error
		list, err = registry.Global().NewList(meshzoneaddress_api.MeshZoneAddressType)
		if err != nil {
			return &meshzoneaddress_api.MeshZoneAddressResourceList{}
		}
	}
	return list.(*meshzoneaddress_api.MeshZoneAddressResourceList)
}
