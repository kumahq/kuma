package builders

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func TargetRefMesh() common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.Mesh,
	}
}

func TargetRefMeshSubset(kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshSubset,
		Tags: TagsKVToMap(kv),
	}
}

func TargetRefService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshService,
		Name: name,
	}
}

func TargetRefServiceSubset(name string, kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshServiceSubset,
		Name: name,
		Tags: TagsKVToMap(kv),
	}
}

func TargetRefMeshService(name, namespace, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Name:        name,
		Namespace:   namespace,
		SectionName: sectionName,
	}
}

func TargetRefMeshServiceLabels(labels map[string]string, sectionName string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind:        common_api.MeshService,
		Labels:      labels,
		SectionName: sectionName,
	}
}

func TargetRefMeshExternalService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshExternalService,
		Name: name,
	}
}

func TargetRefMeshGateway(name string) *common_api.TargetRef {
	return &common_api.TargetRef{
		Kind: common_api.MeshGateway,
		Name: name,
	}
}
