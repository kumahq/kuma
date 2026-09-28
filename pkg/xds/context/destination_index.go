package context

import (
	"maps"
	"time"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// DestinationIndex indexes destinations by KRI and labels. It provides optimized access to Kuma destinations. It should
// be used when working with referenceable destination resources like MeshServices, MeshExternalServices or MeshMultiZoneServices
type DestinationIndex struct {
	destinationByIdentifier    map[kri.Identifier]core.Destination
	destinationsByLabelByValue labelsToValuesToResourceIdentifier
	allowAllOutbound           bool
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
			buildLabelValueToServiceNames(ri, destinationsByLabelByValue, destinationIndexLabels(item.GetMeta()))
		}
	}

	return &DestinationIndex{
		destinationByIdentifier:    destinationByIdentifier,
		destinationsByLabelByValue: destinationsByLabelByValue,
	}
}

// WithAllowAllOutbound makes a data plane proxy without reachableBackends reach every destination.
func (di *DestinationIndex) WithAllowAllOutbound(allow bool) *DestinationIndex {
	di.allowAllOutbound = allow
	return di
}

// AllowAllOutbound reports whether a data plane proxy without MeshPassthrough keeps the default outbound passthrough.
func (di *DestinationIndex) AllowAllOutbound() bool {
	return di != nil && di.allowAllOutbound
}

// GetReachableBackends returns reachable ports by KRI, and true when only the returned backends are reachable.
// Without reachableBackends it returns an empty map and true (deny), or every destination and false when allowAllOutbound is set.
func (di *DestinationIndex) GetReachableBackends(dataplane *core_mesh.DataplaneResource) (map[kri.Identifier]core.Port, bool) {
	outbounds := map[kri.Identifier]core.Port{}

	networking := dataplane.Spec.GetNetworking()

	addOutbounds := func(ids []kri.Identifier, sectionName string) {
		for _, id := range ids {
			if sectionName != "" {
				id = kri.WithSectionName(id, sectionName)
			}

			var dest core.Destination
			if dest = di.GetDestinationByKRI(id); dest == nil {
				continue
			}

			// an unnamed port matches the empty section name, so only narrow when one is set
			if id.SectionName != "" {
				if p, ok := dest.FindPortByName(id.SectionName); ok {
					outbounds[kri.WithSectionName(id, p.GetName())] = p
				}
				continue
			}

			for _, p := range dest.GetPorts() {
				outbounds[kri.WithSectionName(id, p.GetName())] = p
			}
		}
	}

	processRef := func(kind string, name string, port *uint32, labels map[string]string) {
		selectorLabels, sectionName := NormalizeBackendRefTarget(
			kind,
			name,
			port,
			labels,
			dataplane.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag],
		)
		if len(selectorLabels) == 0 {
			return
		}

		backendRef := common_api.BackendRef{
			Kind:   common_api.BackendRefKind(kind),
			Labels: &selectorLabels,
			Port:   port,
		}
		if sectionName != "" {
			backendRef.SectionName = pointer.To(sectionName)
		}

		selectorLabels, sectionName, ok := backendRef.RealResourceSelector("")
		if !ok {
			return
		}

		addOutbounds(di.resolveResourceIdentifiersForLabels(core_model.ResourceType(kind), selectorLabels), sectionName)
	}

	// Handle user defined outbound without a transparent proxy
	for _, o := range networking.GetOutbounds(mesh_proto.BackendRefFilter) {
		processRef(o.BackendRef.Kind, o.BackendRef.Name, &o.BackendRef.Port, o.BackendRef.Labels)
	}

	if len(outbounds) > 0 {
		return outbounds, true
	}

	if networking.GetTransparentProxying().GetReachableBackends() == nil {
		if !di.allowAllOutbound {
			return outbounds, true
		}
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

		// Like a Kubernetes label selector, empty labels select every backend of the kind
		if len(ref.Labels) == 0 {
			addOutbounds(di.resourceIdentifiersOfType(core_model.ResourceType(ref.Kind)), "")
			continue
		}

		processRef(ref.Kind, "", port, ref.Labels)
	}

	return outbounds, true
}

func (di *DestinationIndex) GetDestinationByKRI(id kri.Identifier) core.Destination {
	if id.IsEmpty() {
		return nil
	}
	return di.destinationByIdentifier[kri.NoSectionName(id)]
}

// ResolveResourceIdentifier resolves one resource identifier based on the labels.
// If multiple resources match the labels, the oldest one is returned.
// The reason is that picking the oldest one is the less likely to break existing traffic after introducing new resources.
func (di *DestinationIndex) ResolveResourceIdentifier(resType core_model.ResourceType, labels map[string]string) kri.Identifier {
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

func (di *DestinationIndex) resourceIdentifiersOfType(resType core_model.ResourceType) []kri.Identifier {
	var result []kri.Identifier
	for id := range di.destinationByIdentifier {
		if id.ResourceType == resType {
			result = append(result, id)
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

func destinationIndexLabels(meta core_model.ResourceMeta) map[string]string {
	if meta == nil {
		return nil
	}

	labels := map[string]string{}
	maps.Copy(labels, meta.GetLabels())
	if labels[mesh_proto.DisplayName] == "" {
		labels[mesh_proto.DisplayName] = core_model.GetDisplayName(meta)
	}

	return labels
}
