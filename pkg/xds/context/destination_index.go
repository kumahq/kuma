package context

import (
	"fmt"
	"maps"
	"strconv"
	"strings"
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

// GetReachableBackends return map of reachable port by its KRI, and bool to indicate if any backend were match or all destinations were returned
func (di *DestinationIndex) GetReachableBackends(dataplane *core_mesh.DataplaneResource) (map[kri.Identifier]core.Port, bool) {
	outbounds := map[kri.Identifier]core.Port{}

	networking := dataplane.Spec.GetNetworking()

	processRef := func(kind string, name string, namespace string, port *uint32, labels map[string]string) {
		selectorLabels, sectionName := normalizeReachableBackendRef(
			kind,
			name,
			namespace,
			port,
			labels,
			dataplane.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag],
		)
		if len(selectorLabels) == 0 {
			return
		}

		backendRef := common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind:   common_api.TargetRefKind(kind),
				Labels: &selectorLabels,
			},
			Port: port,
		}
		if sectionName != "" {
			backendRef.SectionName = pointer.To(sectionName)
		}

		selectorLabels, sectionName, ok := backendRef.RealResourceSelector("")
		if !ok {
			return
		}

		ids := di.resolveResourceIdentifiersForLabels(core_model.ResourceType(kind), selectorLabels)
		for _, id := range ids {
			if sectionName != "" {
				id = kri.WithSectionName(id, sectionName)
			}

			var dest core.Destination
			if dest = di.getDestinationByKRI(id); dest == nil {
				return
			}

			if p, ok := dest.FindPortByName(id.SectionName); ok {
				outbounds[kri.WithSectionName(id, p.GetName())] = p
				return
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

func normalizeReachableBackendRef(kind, name, namespace string, port *uint32, labels map[string]string, defaultNamespace string) (map[string]string, string) {
	sectionName := ""
	if port != nil && *port > 0 {
		sectionName = fmt.Sprintf("%d", *port)
	}
	if len(labels) > 0 {
		return labels, sectionName
	}
	if name == "" {
		return nil, sectionName
	}

	if common_api.TargetRefKind(kind) == common_api.MeshService && namespace == "" {
		if service, parsedNamespace, parsedPort, ok := parseLegacyMeshServiceTag(name); ok {
			name = service
			namespace = parsedNamespace
			if sectionName == "" && parsedPort > 0 {
				sectionName = fmt.Sprintf("%d", parsedPort)
			}
		}
	}

	normalized := map[string]string{
		mesh_proto.DisplayName: name,
	}
	switch common_api.TargetRefKind(kind) {
	case common_api.MeshService, common_api.MeshExternalService, common_api.MeshMultiZoneService:
		if namespace == "" {
			namespace = defaultNamespace
		}
		if namespace != "" {
			normalized[mesh_proto.KubeNamespaceTag] = namespace
		}
	}
	return normalized, sectionName
}

func parseLegacyMeshServiceTag(name string) (string, string, int32, bool) {
	segments := strings.Split(name, "_")
	if len(segments) != 4 || segments[2] != "svc" {
		return "", "", 0, false
	}

	port, err := strconv.ParseInt(segments[3], 10, 32)
	if err != nil {
		return "", "", 0, false
	}
	return segments[0], segments[1], int32(port), true
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
