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

type Destination interface {
	GetPorts() []core.Port
	FindPortByName(name string) (core.Port, bool)
}

type DestinationIndex struct {
	MeshServiceByIdentifier             map[kri.Identifier]*meshservice_api.MeshServiceResource
	MeshServicesByLabelByValue          LabelsToValuesToResourceIdentifier
	MeshExternalServiceByIdentifier     map[kri.Identifier]*meshexternalservice_api.MeshExternalServiceResource
	MeshExternalServicesByLabelByValue  LabelsToValuesToResourceIdentifier
	MeshMultiZoneServiceByIdentifier    map[kri.Identifier]*meshmzservice_api.MeshMultiZoneServiceResource
	MeshMultiZoneServicesByLabelByValue LabelsToValuesToResourceIdentifier
}

func (dc *DestinationIndex) GetAllDestinations() map[kri.Identifier]Destination {
	allDestinations := make(map[kri.Identifier]Destination)
	for k, v := range dc.MeshServiceByIdentifier {
		allDestinations[k] = v
	}
	for k, v := range dc.MeshExternalServiceByIdentifier {
		allDestinations[k] = v
	}
	for k, v := range dc.MeshMultiZoneServiceByIdentifier {
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

func (dc *DestinationIndex) GetDestinationByKri(id kri.Identifier) Destination {
	switch id.ResourceType {
	case meshservice_api.MeshServiceType:
		return dc.MeshServiceByIdentifier[id]
	case meshexternalservice_api.MeshExternalServiceType:
		return dc.MeshExternalServiceByIdentifier[id]
	case meshmzservice_api.MeshMultiZoneServiceType:
		return dc.MeshMultiZoneServiceByIdentifier[id]
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
		ms := dc.MeshServiceByIdentifier[key]
		if ms == nil {
			return fmt.Sprintf("%d", port)
		}
		if sectionName, portFound := ms.FindSectionNameByPort(port); portFound {
			return sectionName
		}
		return fmt.Sprintf("%d", port)
	case meshmzservice_api.MeshMultiZoneServiceType:
		mmzs := dc.MeshMultiZoneServiceByIdentifier[key]
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
			matchedResourceIdentifiers, found = dc.MeshExternalServicesByLabelByValue[key]
		case string(meshservice_api.MeshServiceType):
			matchedResourceIdentifiers, found = dc.MeshServicesByLabelByValue[key]
		case string(meshmzservice_api.MeshMultiZoneServiceType):
			matchedResourceIdentifiers, found = dc.MeshMultiZoneServicesByLabelByValue[key]
		}
		if found {
			for ri := range matchedResourceIdentifiers {
				reachable[ri]++
			}
		}
	}
	return reachable
}
