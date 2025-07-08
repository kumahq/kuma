package context

import (
	"strconv"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
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

// GetReachableBackends return map of reachable port by its KRI, and bool to indicate if any backend were match or all destinations were returned
func (dc *DestinationIndex) GetReachableBackends(mesh *core_mesh.MeshResource, dataplane *core_mesh.DataplaneResource) (map[kri.Identifier]core.Port, bool) {
	outbounds := map[kri.Identifier]core.Port{}
	if dataplane.Spec.GetNetworking().GetOutbound() != nil {
		// TODO handle user defined outbounds on universal without transparent proxy: https://github.com/kumahq/kuma/issues/13868
		return outbounds, true
	}

	reachableBackends := dataplane.Spec.GetNetworking().GetTransparentProxying().GetReachableBackends()
	if reachableBackends == nil {
		// return all destinations if reachable backends not configured
		for destinationKri, destination := range dc.destinationByIdentifier {
			for _, port := range destination.GetPorts() {
				outbounds[kri.WithSectionName(destinationKri, port.GetName())] = port
			}
		}
		return outbounds, false
	}

	for _, reachableBackend := range reachableBackends.GetRefs() {
		if len(reachableBackend.Labels) > 0 {
			for _, destinationKri := range dc.resolveResourceIdentifiersForLabels(reachableBackend.Labels) {
				dest := dc.destinationByIdentifier[destinationKri]
				if dest == nil {
					continue
				}
				if port := reachableBackend.Port; port != nil {
					destPort, ok := dest.FindPortByName(strconv.Itoa(int(reachableBackend.Port.GetValue())))
					if !ok {
						continue
					}
					destinationKri.SectionName = destPort.GetName()
					outbounds[destinationKri] = destPort
				} else {
					for _, port := range dest.GetPorts() {
						outbounds[kri.WithSectionName(destinationKri, port.GetName())] = port
					}
				}
			}
		} else {
			destinationKri := resolve.TargetRefToKRI(dataplane.GetMeta(), common_api.TargetRef{
				Kind:      common_api.TargetRefKind(reachableBackend.Kind),
				Name:      &reachableBackend.Name,
				Namespace: &reachableBackend.Namespace,
			})
			dest := dc.destinationByIdentifier[destinationKri]
			if dest == nil {
				continue
			}
			if port := reachableBackend.Port; port != nil {
				destPort, ok := dest.FindPortByName(strconv.Itoa(int(reachableBackend.Port.GetValue())))
				if !ok {
					continue
				}
				destinationKri.SectionName = destPort.GetName()
				outbounds[destinationKri] = destPort
			} else {
				for _, port := range dest.GetPorts() {
					outbounds[kri.WithSectionName(destinationKri, port.GetName())] = port
				}
			}
		}
	}
	return outbounds, true
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
