package context

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
)

type DestinationIndex struct {
	meshServiceByIdentifier             map[kri.Identifier]*meshservice_api.MeshServiceResource
	meshServicesByLabelByValue          LabelsToValuesToResourceIdentifier
	meshExternalServiceByIdentifier     map[kri.Identifier]*meshexternalservice_api.MeshExternalServiceResource
	meshExternalServicesByLabelByValue  LabelsToValuesToResourceIdentifier
	meshMultiZoneServiceByIdentifier    map[kri.Identifier]*meshmzservice_api.MeshMultiZoneServiceResource
	meshMultiZoneServicesByLabelByValue LabelsToValuesToResourceIdentifier
}

func NewDestinationIndex(resourceMap ResourceMap) *DestinationIndex {
	var meshServices []*meshservice_api.MeshServiceResource
	if resourceMap[meshservice_api.MeshServiceType] != nil {
		meshServices = resourceMap[meshservice_api.MeshServiceType].(*meshservice_api.MeshServiceResourceList).Items
	}

	var meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource
	if resourceMap[meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor.Name] != nil {
		meshExternalServices = resourceMap[meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor.Name].(*meshexternalservice_api.MeshExternalServiceResourceList).Items
	}

	var meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource
	if resourceMap[meshmzservice_api.MeshMultiZoneServiceResourceTypeDescriptor.Name] != nil {
		meshMultiZoneServices = resourceMap[meshmzservice_api.MeshMultiZoneServiceResourceTypeDescriptor.Name].(*meshmzservice_api.MeshMultiZoneServiceResourceList).Items
	}

	return &DestinationIndex{
		meshServiceByIdentifier:             indexByKri(meshServices),
		meshServicesByLabelByValue:          indexByLabelKeyValue(meshServices),
		meshExternalServiceByIdentifier:     indexByKri(meshExternalServices),
		meshExternalServicesByLabelByValue:  indexByLabelKeyValue(meshExternalServices),
		meshMultiZoneServiceByIdentifier:    indexByKri(meshMultiZoneServices),
		meshMultiZoneServicesByLabelByValue: indexByLabelKeyValue(meshMultiZoneServices),
	}
}

func (dc *DestinationIndex) GetAllDestinations() map[kri.Identifier]core.Destination {
	allDestinations := make(map[kri.Identifier]core.Destination)
	for k, v := range dc.meshServiceByIdentifier {
		allDestinations[k] = v
	}
	for k, v := range dc.meshExternalServiceByIdentifier {
		allDestinations[k] = v
	}
	for k, v := range dc.meshMultiZoneServiceByIdentifier {
		allDestinations[k] = v
	}
	return allDestinations
}

func (dc *DestinationIndex) GetReachableBackends(mesh *core_mesh.MeshResource, dataplane *core_mesh.DataplaneResource) *ReachableBackends {
	reachableBackends := dataplane.Spec.GetNetworking().GetTransparentProxying().GetReachableBackends()
	if reachableBackends == nil {
		if mesh.Spec.MeshServicesMode() == mesh_proto.Mesh_MeshServices_ReachableBackends {
			return &ReachableBackends{}
		}
		return nil
	}
	out := ReachableBackends{}
	for _, reachableBackend := range reachableBackends.GetRefs() {
		if len(reachableBackend.Labels) > 0 {
			for _, tri := range dc.resolveResourceIdentifiersForLabels(reachableBackend.Kind, reachableBackend.Labels) {
				if port := reachableBackend.Port; port != nil {
					tri.SectionName = dc.getSectionName(tri.ResourceType, tri, reachableBackend.Port.GetValue())
				}
				out[tri] = true
			}
		} else {
			key := resolve.TargetRefToKRI(dataplane.GetMeta(), common_api.TargetRef{
				Kind:      common_api.TargetRefKind(reachableBackend.Kind),
				Name:      &reachableBackend.Name,
				Namespace: &reachableBackend.Namespace,
			})
			if port := reachableBackend.Port; port != nil {
				key.SectionName = dc.getSectionName(key.ResourceType, key, reachableBackend.Port.GetValue())
			}
			out[key] = true
		}
	}
	return &out
}

func (dc *DestinationIndex) GetDestinationByKri(id kri.Identifier) core.Destination {
	switch id.ResourceType {
	case meshservice_api.MeshServiceType:
		return dc.meshServiceByIdentifier[id]
	case meshexternalservice_api.MeshExternalServiceType:
		return dc.meshExternalServiceByIdentifier[id]
	case meshmzservice_api.MeshMultiZoneServiceType:
		return dc.meshMultiZoneServiceByIdentifier[id]
	}
	return nil
}

func (dc *DestinationIndex) resolveResourceIdentifiersForLabels(kind string, labels map[string]string) []kri.Identifier {
	var result []kri.Identifier
	reachable := dc.getResourceNamesForLabels(kind, labels)
	for ri, count := range reachable {
		if count == len(labels) {
			result = append(result, ri)
		}
	}
	return result
}

func (dc *DestinationIndex) getSectionName(kind core_model.ResourceType, key kri.Identifier, port uint32) string {
	switch kind {
	case meshservice_api.MeshServiceType:
		ms := dc.meshServiceByIdentifier[key]
		if ms == nil {
			return fmt.Sprintf("%d", port)
		}
		if sectionName, portFound := ms.FindSectionNameByPort(port); portFound {
			return sectionName
		}
		return fmt.Sprintf("%d", port)
	case meshmzservice_api.MeshMultiZoneServiceType:
		mmzs := dc.meshMultiZoneServiceByIdentifier[key]
		if mmzs == nil {
			return fmt.Sprintf("%d", port)
		}
		if sectionName, portFound := mmzs.FindSectionNameByPort(port); portFound {
			return sectionName
		}
		return fmt.Sprintf("%d", port)
	default:
		return fmt.Sprintf("%d", port)
	}
}

func (dc *DestinationIndex) getResourceNamesForLabels(kind string, labels map[string]string) map[kri.Identifier]int {
	reachable := map[kri.Identifier]int{}
	for label, value := range labels {
		key := LabelValue{
			Label: label,
			Value: value,
		}
		var matchedResourceIdentifiers map[kri.Identifier]bool
		var found bool
		switch kind {
		case string(meshexternalservice_api.MeshExternalServiceType):
			matchedResourceIdentifiers, found = dc.meshExternalServicesByLabelByValue[key]
		case string(meshservice_api.MeshServiceType):
			matchedResourceIdentifiers, found = dc.meshServicesByLabelByValue[key]
		case string(meshmzservice_api.MeshMultiZoneServiceType):
			matchedResourceIdentifiers, found = dc.meshMultiZoneServicesByLabelByValue[key]
		}
		if found {
			for ri := range matchedResourceIdentifiers {
				reachable[ri]++
			}
		}
	}
	return reachable
}

func indexByKri[T core_model.Resource](list []T) map[kri.Identifier]T {
	byKri := make(map[kri.Identifier]T, len(list))
	for _, item := range list {
		ri := kri.From(item, "")
		byKri[ri] = item
	}
	return byKri
}

func indexByLabelKeyValue[T core_model.Resource](list []T) LabelsToValuesToResourceIdentifier {
	resourceNamesByLabels := LabelsToValuesToResourceIdentifier{}
	for _, item := range list {
		ri := kri.From(item, "")
		buildLabelValueToServiceNames(ri, resourceNamesByLabels, item.GetMeta().GetLabels())
	}
	return resourceNamesByLabels
}

func buildLabelValueToServiceNames(ri kri.Identifier, resourceNamesByLabels LabelsToValuesToResourceIdentifier, labels map[string]string) {
	for label, value := range labels {
		key := LabelValue{
			Label: label,
			Value: value,
		}
		if _, ok := resourceNamesByLabels[key]; ok {
			resourceNamesByLabels[key][ri] = true
		} else {
			resourceNamesByLabels[key] = map[kri.Identifier]bool{
				ri: true,
			}
		}
	}
}
