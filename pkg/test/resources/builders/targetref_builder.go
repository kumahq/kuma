package builders

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func TargetRefMesh() common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.Mesh,
	}
}

func TargetRefDataplaneLabels(kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:   common_api.Dataplane,
		Labels: pointer.To(TagsKVToMap(kv)),
	}
}

func TargetRefDataplaneName(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.Dataplane,
		Labels: pointer.To(map[string]string{
			mesh_proto.DisplayName: name,
		}),
	}
}

func TargetRefService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshService,
		Labels: pointer.To(map[string]string{
			mesh_proto.DisplayName: name,
		}),
	}
}

func TargetRefMeshService(name, namespace, sectionName string) common_api.TargetRef {
	labels := map[string]string{
		mesh_proto.DisplayName: name,
	}
	if namespace != "" {
		labels[mesh_proto.KubeNamespaceTag] = namespace
	}
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Labels:      pointer.To(labels),
		SectionName: pointer.To(sectionName),
	}
}

func TargetRefMeshServiceLabels(labels map[string]string, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Labels:      pointer.To(labels),
		SectionName: pointer.To(sectionName),
	}
}

func TargetRefMeshHTTPRouteLabels(labels map[string]string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:   common_api.MeshHTTPRoute,
		Labels: pointer.To(labels),
	}
}

func TargetRefMeshExternalService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshExternalService,
		Labels: pointer.To(map[string]string{
			mesh_proto.DisplayName: name,
		}),
	}
}

func ToTopLevelTargetRef(ref common_api.TargetRef) common_api.TopLevelTargetRef {
	var kind common_api.TopLevelTargetRefKind
	switch ref.Kind {
	case common_api.Mesh:
		kind = common_api.TopLevelTargetRefKindMesh
	case common_api.Dataplane:
		kind = common_api.TopLevelTargetRefKindDataplane
	default:
		panic("unsupported top-level targetRef kind: " + string(ref.Kind))
	}
	return common_api.TopLevelTargetRef{
		Kind:        kind,
		Labels:      ref.Labels,
		SectionName: ref.SectionName,
	}
}

func ToOutboundTargetRef(ref common_api.TargetRef) common_api.OutboundTargetRef {
	var kind common_api.OutboundTargetRefKind
	switch ref.Kind {
	case common_api.Mesh:
		kind = common_api.OutboundTargetRefKindMesh
	case common_api.MeshService:
		kind = common_api.OutboundTargetRefKindMeshService
	case common_api.MeshExternalService:
		kind = common_api.OutboundTargetRefKindMeshExternalService
	case common_api.MeshMultiZoneService:
		kind = common_api.OutboundTargetRefKindMeshMultiZoneService
	case common_api.MeshHTTPRoute:
		kind = common_api.OutboundTargetRefKindMeshHTTPRoute
	default:
		panic("unsupported outbound targetRef kind: " + string(ref.Kind))
	}
	return common_api.OutboundTargetRef{
		Kind:        kind,
		Labels:      ref.Labels,
		SectionName: ref.SectionName,
	}
}

func BackendRefMeshService(name, namespace, sectionName string, port uint32, weight uint) common_api.BackendRef {
	ref := common_api.BackendRefFrom(TargetRefMeshService(name, namespace, sectionName))
	ref.Port = pointer.To(port)
	ref.Weight = pointer.To(weight)
	return ref
}

func BackendRefMeshServiceLabels(labels map[string]string, sectionName string, weight uint) common_api.BackendRef {
	ref := common_api.BackendRefFrom(TargetRefMeshServiceLabels(labels, sectionName))
	ref.Weight = pointer.To(weight)
	return ref
}

func BackendRefService(name string, weight uint) common_api.BackendRef {
	ref := common_api.BackendRefFrom(TargetRefService(name))
	ref.Weight = pointer.To(weight)
	return ref
}

// BackendRefFrom builds a weighted BackendRef from a TargetRef builder result.
func BackendRefFrom(t common_api.TargetRef, weight uint) common_api.BackendRef {
	ref := common_api.BackendRefFrom(t)
	ref.Weight = pointer.To(weight)
	return ref
}

// BackendRefFromWithPort is BackendRefFrom for a reference that also targets a port.
func BackendRefFromWithPort(t common_api.TargetRef, weight uint, port uint32) common_api.BackendRef {
	ref := BackendRefFrom(t, weight)
	ref.Port = pointer.To(port)
	return ref
}
