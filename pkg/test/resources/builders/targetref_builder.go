package builders

import (
	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
)

func TargetRefMesh() common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.Mesh,
	}
}

func TargetRefMeshSubset(kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshSubset,
		Tags: new(TagsKVToMap(kv)),
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
		Tags: new(TagsKVToMap(kv)),
	}
}

func TargetRefMeshService(name, namespace, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Name:        &name,
		Namespace:   new(namespace),
		SectionName: new(sectionName),
	}
}

func TargetRefMeshHTTPRoute(name, namespace string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:      common_api.MeshHTTPRoute,
		Name:      &name,
		Namespace: new(namespace),
	}
}

func TargetRefMeshServiceLabels(labels map[string]string, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Labels:      new(labels),
		SectionName: new(sectionName),
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
