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

type TargetRefBuilder struct {
	targetRef *common_api.TargetRef
}

func NewTargetRefBuilder() *TargetRefBuilder {
	targetRef := common_api.TargetRef{}
	b := &TargetRefBuilder{targetRef: &targetRef}
	return b
}

func (b *TargetRefBuilder) WithKind(kind common_api.TargetRefKind) *TargetRefBuilder {
	b.targetRef.Kind = kind
	return b
}

func (b *TargetRefBuilder) WithName(name string) *TargetRefBuilder {
	b.targetRef.Name = name
	return b
}

func (b *TargetRefBuilder) WithTags(tags map[string]string) *TargetRefBuilder {
	b.targetRef.Tags = tags
	return b
}

func (b *TargetRefBuilder) WithMesh(mesh string) *TargetRefBuilder {
	b.targetRef.Mesh = mesh
	return b
}

func (b *TargetRefBuilder) WithProxyTypes(proxyTypes []common_api.TargetRefProxyType) *TargetRefBuilder {
	b.targetRef.ProxyTypes = proxyTypes
	return b
}

func (b *TargetRefBuilder) WithNamespace(namespace string) *TargetRefBuilder {
	b.targetRef.Namespace = namespace
	return b
}

func (b *TargetRefBuilder) WithLabels(labels map[string]string) *TargetRefBuilder {
	b.targetRef.Labels = labels
	return b
}

func (b *TargetRefBuilder) WithSectionName(sectionName string) *TargetRefBuilder {
	b.targetRef.SectionName = sectionName
	return b
}

func (b *TargetRefBuilder) Build() *common_api.TargetRef {
	return b.targetRef
}

