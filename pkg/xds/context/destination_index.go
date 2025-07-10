package context

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
)

// DestinationIndex indexes destinations by KRI and labels. It provides optimized access to Kuma destinations. It should
// be used when working with referencable destination resources like MeshServices, MeshExternalServices or MeshMultizoneServices
type DestinationIndex struct {
	destinationByIdentifier    map[kri.Identifier]core.Destination
	destinationsByLabelByValue labelsToValuesToResourceIdentifier
}
type labelsToValuesToResourceIdentifier map[labelValue]map[kri.Identifier]bool

type labelValue struct {
	label string
	value string
}

func NewDestinationIndex(resources ...[]core_model.Resource) *DestinationIndex {
	destinationByIdentifier := make(map[kri.Identifier]core.Destination)
	destinationsByLabelByValue := labelsToValuesToResourceIdentifier{}
	for _, destinations := range resources {
		for _, item := range destinations {
			ri := kri.From(item, "")
			destinationByIdentifier[ri] = item.(core.Destination)
			buildLabelValueToServiceNames(ri, destinationsByLabelByValue, item.GetMeta().GetLabels())
		}
	}

	return &DestinationIndex{
		destinationByIdentifier:    destinationByIdentifier,
		destinationsByLabelByValue: destinationsByLabelByValue,
	}
}

func (dc *DestinationIndex) GetAllDestinations() map[kri.Identifier]core.Destination {
	return dc.destinationByIdentifier
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
			for _, tri := range dc.resolveResourceIdentifiersForLabels(reachableBackend.Labels) {
				if port := reachableBackend.Port; port != nil {
					tri.SectionName = dc.getSectionNameForPort(tri, int32(reachableBackend.Port.GetValue()))
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
				key.SectionName = dc.getSectionNameForPort(key, int32(reachableBackend.Port.GetValue()))
			}
			out[key] = true
		}
	}
	return &out
}

func (dc *DestinationIndex) GetDestinationByKri(id kri.Identifier) core.Destination {
	return dc.destinationByIdentifier[kri.NoSectionName(id)]
}

func (dc *DestinationIndex) resolveResourceIdentifiersForLabels(labels map[string]string) []kri.Identifier {
	var result []kri.Identifier
	reachable := dc.getDestinationsForLabels(labels)
	for ri, count := range reachable {
		if count == len(labels) {
			result = append(result, ri)
		}
	}
	return result
}

func (dc *DestinationIndex) getSectionNameForPort(key kri.Identifier, port int32) string {
	dest := dc.destinationByIdentifier[key]
	if dest == nil {
		return fmt.Sprintf("%d", port)
	}

	ports := dest.GetPorts()
	for _, destPort := range ports {
		if destPort.GetValue() == port {
			return destPort.GetName()
		}
	}

	return fmt.Sprintf("%d", port)
}

func (dc *DestinationIndex) getDestinationsForLabels(labels map[string]string) map[kri.Identifier]int {
	reachable := map[kri.Identifier]int{}
	for label, value := range labels {
		key := labelValue{
			label: label,
			value: value,
		}

		matchedDestinations, found := dc.destinationsByLabelByValue[key]
		if found {
			for ri := range matchedDestinations {
				reachable[ri]++
			}
		}
	}
	return reachable
}

func buildLabelValueToServiceNames(ri kri.Identifier, resourceNamesByLabels labelsToValuesToResourceIdentifier, labels map[string]string) {
	for label, value := range labels {
		key := labelValue{
			label: label,
			value: value,
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
