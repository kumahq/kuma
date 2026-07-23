package builders

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func TargetRefMesh() common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.Mesh,
	}
}

func TargetRefMeshSubset(kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.LegacyMeshSubsetKind(),
		Tags: pointer.To(TagsKVToMap(kv)),
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
		Name: &name,
	}
}

func TargetRefService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshService,
		Name: &name,
	}
}

func TargetRefServiceSubset(name string, kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.LegacyMeshServiceSubsetKind(),
		Name: &name,
		Tags: pointer.To(TagsKVToMap(kv)),
	}
}

func TargetRefMeshService(name, namespace, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Name:        &name,
		Namespace:   pointer.To(namespace),
		SectionName: pointer.To(sectionName),
	}
}

func TargetRefMeshHTTPRoute(name, namespace string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:      common_api.MeshHTTPRoute,
		Name:      &name,
		Namespace: pointer.To(namespace),
	}
}

func TargetRefMeshServiceLabels(labels map[string]string, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Labels:      pointer.To(labels),
		SectionName: pointer.To(sectionName),
	}
}

func TargetRefMeshExternalService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshExternalService,
		Name: &name,
	}
}

func ToTopLevelTargetRef(ref common_api.TargetRef) common_api.TopLevelTargetRef {
	return common_api.TopLevelTargetRef(ref)
}

func ToOutboundTargetRef(ref common_api.TargetRef) common_api.OutboundTargetRef {
	return common_api.OutboundTargetRef{
		Kind:        ref.Kind,
		Name:        ref.Name,
		Tags:        ref.Tags,
		Mesh:        ref.Mesh,
		ProxyTypes:  ref.ProxyTypes,
		Namespace:   ref.Namespace,
		Labels:      ref.Labels,
		SectionName: ref.SectionName,
	}
}
