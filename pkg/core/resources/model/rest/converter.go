package rest

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

var From = &from{}

type from struct{}

func (c *from) Resource(r model.Resource) *Resource {
	var meshName string
	if r.Descriptor().Scope == model.ScopeMesh {
		meshName = r.GetMeta().GetMesh()
	}
	return &Resource{
		Meta: ResourceMeta{
			Mesh:             meshName,
			Type:             string(r.Descriptor().Name),
			Name:             r.GetMeta().GetName(),
			CreationTime:     r.GetMeta().GetCreationTime(),
			ModificationTime: r.GetMeta().GetModificationTime(),
		},
		Spec: r.GetSpec(),
	}
}

func (c *from) ResourceList(rs model.ResourceList) *ResourceList {
	items := make([]*Resource, len(rs.GetItems()))
	for i, r := range rs.GetItems() {
		items[i] = c.Resource(r)
	}
	return &ResourceList{
		Total: rs.GetPagination().Total,
		Items: items,
	}
}

func (r *Resource) ToCore() (model.Resource, error) {
	resource, err := registry.Global().NewObject(model.ResourceType(r.Meta.Type))
	if err != nil {
		return nil, err
	}
	resource.SetMeta(&r.Meta)
	if err := resource.SetSpec(r.Spec); err != nil {
		return nil, err
	}
	return resource, nil
}
