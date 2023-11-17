package builders

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func TargetRefMesh() common_api.TargetRef {
	return common_api.TargetRef{
		Kind: "Mesh",
	}
}

func TargetRefMeshSubset(kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: "MeshSubset",
		Tags: TagsKVToMap(kv),
	}
}

func TargetRefService(name string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: "MeshService",
		Name: name,
	}
}

func TargetRefServiceSubset(name string, kv ...string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: "MeshServiceSubset",
		Name: name,
		Tags: TagsKVToMap(kv),
	}
}
