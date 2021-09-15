package merge

import "github.com/kumahq/kuma/pkg/core/resources/model"

// UniqueResources takes a slice of resources that may include duplicates
// and returns a slice with no duplicates. Resources are considered
// duplicates if they have the same type, mesh and name.
func UniqueResources(all []model.Resource) []model.Resource {
	type key struct {
		key model.ResourceKey
		typ model.ResourceType
	}

	set := map[key]model.Resource{}

	for _, r := range all {
		set[key{
			key: model.MetaToResourceKey(r.GetMeta()),
			typ: r.Descriptor().Name,
		}] = r
	}

	var u []model.Resource
	for _, m := range set {
		u = append(u, m)
	}

	return u
}
