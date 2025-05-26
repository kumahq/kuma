package builders

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func TargetRefMesh() common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.Mesh,
	}
}

func TargetRefMeshSubset(kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshSubset,
		Tags: pointer.To(TagsKVToMap(kv)),
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
		Kind: common_api.MeshServiceSubset,
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

func TargetRefMeshGateway(name string) *common_api.TargetRef {
	return &common_api.TargetRef{
		Kind: common_api.MeshGateway,
		Name: &name,
	}
}
