package context

import (
	"strconv"
	"time"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

// DestinationIndex indexes destinations by KRI and labels. It provides optimized access to Kuma destinations. It should
// be used when working with referenceable destination resources like MeshServices, MeshExternalServices or MeshMultiZoneServices
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
func (di *DestinationIndex) GetReachableBackends(dataplane *core_mesh.DataplaneResource) (map[kri.Identifier]core.Port, bool) {
	outbounds := map[kri.Identifier]core.Port{}
	// Handle user defined outbound without a transparent proxy
	if dataplane.Spec.GetNetworking().GetOutbound() != nil {
		for _, outbound := range dataplane.Spec.GetNetworking().GetOutbound() {
			if outbound.BackendRef == nil {
				continue
			}
			backendRef := common_api.BackendRef{
				TargetRef: common_api.TargetRef{
					Kind:   common_api.TargetRefKind(outbound.BackendRef.Kind),
					Name:   pointer.To(outbound.BackendRef.Name),
					Labels: pointer.To(outbound.BackendRef.Labels),
				},
				Port: pointer.To(outbound.BackendRef.Port),
			}
			ref, ok := resolve.BackendRef(pointer.To(kri.From(dataplane, "")), backendRef, di.ResolveResourceIdentifier)
			if !ok || !ref.ReferencesRealResource() {
				continue
			}
			outboundKri := pointer.Deref(ref.ResourceOrNil())
			dest, ok := di.destinationByIdentifier[kri.NoSectionName(outboundKri)]
			if !ok {
				continue
			}
			port, ok := dest.FindPortByName(outboundKri.SectionName)
			if !ok {
				continue
			}
			outbounds[outboundKri] = port
		}
		return outbounds, true
	}

	reachableBackends := dataplane.Spec.GetNetworking().GetTransparentProxying().GetReachableBackends()
	if reachableBackends == nil {
		// return all destinations if reachable backends not configured
		for destinationKri, destination := range di.destinationByIdentifier {
			for _, port := range destination.GetPorts() {
				outbounds[kri.WithSectionName(destinationKri, port.GetName())] = port
			}
		}
		return outbounds, false
	}

	for _, reachableBackend := range reachableBackends.GetRefs() {
		if len(reachableBackend.Labels) > 0 {
			for _, destinationKri := range di.resolveResourceIdentifiersForLabels(core_model.ResourceType(reachableBackend.Kind), reachableBackend.Labels) {
				dest := di.destinationByIdentifier[destinationKri]
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
			dest := di.destinationByIdentifier[destinationKri]
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

func (di *DestinationIndex) GetDestinationByKri(id kri.Identifier) core.Destination {
	return di.destinationByIdentifier[kri.NoSectionName(id)]
}

// ResolveResourceIdentifier resolves one resource identifier based on the labels.
// If multiple resources match the labels, the oldest one is returned.
// The reason is that picking the oldest one is the less likely to break existing traffic after introducing new resources.
func (di *DestinationIndex) ResolveResourceIdentifier(resType core_model.ResourceType, labels map[string]string) *kri.Identifier {
	if len(labels) == 0 {
		return nil
	}
	var oldestCreationTime *time.Time
	var oldestKri *kri.Identifier
	for _, resourceKri := range di.resolveResourceIdentifiersForLabels(resType, labels) {
		resource := di.destinationByIdentifier[kri.NoSectionName(resourceKri)].(core_model.Resource)
		if resource != nil {
			resCreationTime := resource.GetMeta().GetCreationTime()
			if oldestCreationTime == nil || resCreationTime.Before(*oldestCreationTime) {
				oldestCreationTime = &resCreationTime
				oldestKri = &resourceKri
			}
		}
	}
	return oldestKri
}

func (di *DestinationIndex) resolveResourceIdentifiersForLabels(resType core_model.ResourceType, labels map[string]string) []kri.Identifier {
	var result []kri.Identifier
	reachable := di.getDestinationsForLabels(resType, labels)
	for ri, count := range reachable {
		if count == len(labels) {
			result = append(result, ri)
		}
	}
	return result
}

func (di *DestinationIndex) getDestinationsForLabels(resType core_model.ResourceType, labels map[string]string) map[kri.Identifier]int {
	reachable := map[kri.Identifier]int{}
	for label, value := range labels {
		key := labelValue{
			label: label,
			value: value,
		}

		matchedDestinations, found := di.destinationsByLabelByValue[key]
		if found {
			for ri := range matchedDestinations {
				if ri.ResourceType == resType {
					reachable[ri]++
				}
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
