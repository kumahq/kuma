package rest

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var From = &from{}

type from struct{}

func (c *from) Resource(r model.Resource) *Resource {
	var meshName string
	switch r.GetType() {
	case mesh.MeshType:
	case system.ZoneType:
	default:
		meshName = r.GetMeta().GetMesh()
	}
	return &Resource{
		Meta: ResourceMeta{
			Mesh:             meshName,
			Type:             string(r.GetType()),
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
