package context

import (
	"time"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

// DestinationIndex indexes destinations by KRI and labels. It provides optimized access to Kuma destinations. It should
// be used when working with referenceable destination resources like MeshServices, MeshExternalServices or MeshMultiZoneServices
type DestinationIndex struct {
	destinationByIdentifier    map[kri.Identifier]core.Destination
	destinationsByLabelByValue labelsToValuesToResourceIdentifier
	// zero value keeps the legacy behavior: no reachableBackends means every destination
	restrictOutbound bool
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
			ri := kri.From(item)
			destinationByIdentifier[ri] = item.(core.Destination)
			buildLabelValueToServiceNames(ri, destinationsByLabelByValue, item.GetMeta().GetLabels())
		}
	}

	return &DestinationIndex{
		destinationByIdentifier:    destinationByIdentifier,
		destinationsByLabelByValue: destinationsByLabelByValue,
	}
}

// WithAllowAllOutbound makes a data plane proxy without reachableBackends reach every destination.
func (di *DestinationIndex) WithAllowAllOutbound(allow bool) *DestinationIndex {
	di.restrictOutbound = !allow
	return di
}

// AllowAllOutbound reports whether a proxy without reachableBackends reaches every destination.
// An unset index means the legacy permissive behavior.
func (di *DestinationIndex) AllowAllOutbound() bool {
	return di == nil || !di.restrictOutbound
}

// GetReachableBackends return map of reachable port by its KRI, and bool to indicate if any backend were match or all destinations were returned
func (di *DestinationIndex) GetReachableBackends(dataplane *core_mesh.DataplaneResource) (map[kri.Identifier]core.Port, bool) {
	outbounds := map[kri.Identifier]core.Port{}

	networking := dataplane.Spec.GetNetworking()

	processRef := func(kind string, name string, namespace string, port *uint32, labels map[string]string) {
		ids := di.resolveResourceIdentifiersForLabels(core_model.ResourceType(kind), labels)
		if len(ids) == 0 && namespace == "" && isSystemNamespaced(kind) {
			ids = di.resolveResourceIdentifiersForName(core_model.ResourceType(kind), dataplane.GetMeta().GetMesh(), name)
		}
		if len(ids) == 0 {
			ids = []kri.Identifier{
				resolve.TargetRefToKRI(
					kri.From(dataplane),
					common_api.TargetRef{
						Kind:      common_api.TargetRefKind(kind),
						Name:      &name,
						Namespace: &namespace,
						Labels:    &labels,
					},
				),
			}
		}

		for _, id := range ids {
			if port != nil {
				id = kri.WithSectionName(id, *port)
			}

			var dest core.Destination
			if dest = di.getDestinationByKRI(id); dest == nil {
				continue
			}

			if p, ok := dest.FindPortByName(id.SectionName); ok {
				outbounds[kri.WithSectionName(id, p.GetName())] = p
				continue
			}

			for _, p := range dest.GetPorts() {
				outbounds[kri.WithSectionName(id, p.GetName())] = p
			}
		}
	}

	// Handle user defined outbound without a transparent proxy
	for _, o := range networking.GetOutbounds(mesh_proto.BackendRefFilter) {
		processRef(o.BackendRef.Kind, o.BackendRef.Name, "", &o.BackendRef.Port, o.BackendRef.Labels)
	}

	if len(outbounds) > 0 {
		return outbounds, true
	}

	if networking.GetTransparentProxying().GetReachableBackends() == nil {
		if di.restrictOutbound {
			return outbounds, true
		}
		// return all destinations if reachable backends not configured
		for id, dest := range di.destinationByIdentifier {
			for _, port := range dest.GetPorts() {
				outbounds[kri.WithSectionName(id, port.GetName())] = port
			}
		}

		return outbounds, false
	}

	for _, ref := range networking.GetTransparentProxying().GetReachableBackends().GetRefs() {
		var port *uint32
		if ref.Port != nil {
			port = pointer.To(ref.Port.GetValue())
		}

		processRef(ref.Kind, ref.Name, ref.Namespace, port, ref.Labels)
	}

	return outbounds, true
}

func (di *DestinationIndex) getDestinationByKRI(id kri.Identifier) core.Destination {
	if id.IsEmpty() {
		return nil
	}
	return di.destinationByIdentifier[kri.NoSectionName(id)]
}

// resolveResourceIdentifier resolves one resource identifier based on the labels.
// If multiple resources match the labels, the oldest one is returned.
// The reason is that picking the oldest one is the less likely to break existing traffic after introducing new resources.
func (di *DestinationIndex) resolveResourceIdentifier(resType core_model.ResourceType, labels map[string]string) kri.Identifier {
	if len(labels) == 0 {
		return kri.Identifier{}
	}
	var oldestCreationTime *time.Time
	var oldestKri kri.Identifier
	for _, resourceKri := range di.resolveResourceIdentifiersForLabels(resType, labels) {
		resource := di.destinationByIdentifier[kri.NoSectionName(resourceKri)].(core_model.Resource)
		if resource != nil {
			resCreationTime := resource.GetMeta().GetCreationTime()
			if oldestCreationTime == nil || resCreationTime.Before(*oldestCreationTime) {
				oldestCreationTime = &resCreationTime
				oldestKri = resourceKri
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

// MeshExternalService and MeshMultiZoneService live in the system namespace, so a
// ref without a namespace must not default to the data plane proxy namespace
func isSystemNamespaced(kind string) bool {
	switch common_api.TargetRefKind(kind) {
	case common_api.MeshExternalService, common_api.MeshMultiZoneService:
		return true
	default:
		return false
	}
}

func (di *DestinationIndex) resolveResourceIdentifiersForName(resType core_model.ResourceType, mesh string, name string) []kri.Identifier {
	var result []kri.Identifier
	for ri := range di.destinationByIdentifier {
		if ri.ResourceType == resType && ri.Mesh == mesh && ri.Name == name {
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
