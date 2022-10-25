package builders

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
)

func TargetRefMesh() *common_proto.TargetRef {
	return &common_proto.TargetRef{
		Kind: "Mesh",
	}
}

func TargetRefService(name string) *common_proto.TargetRef {
	return &common_proto.TargetRef{
		Kind: "MeshService",
		Name: name,
	}
}

func TargetRefServiceSubset(name string, kv ...string) *common_proto.TargetRef {
	return &common_proto.TargetRef{
		Kind: "MeshServiceSubset",
		Name: name,
		Tags: tagsKVToMap(kv),
	}
}
