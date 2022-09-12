package rest

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

var From = &from{}

type from struct{}

// Resource method generates a resource representation to use with API server or CLI.
// Representation is different depending on the resource type:
//   - new resources which are provided by plugins are converted to 'v1alpha1.Resource'
//   - old resources are converted to 'unversioned.Resource'
//
// The difference between 'v1alpha1.Resource' and 'unversioned.Resource' is 'spec' field
// in 'v1alpha1.Resource':
//
//	type: MeshTrafficPermission
//	name: mtp1
//	mesh: default
//	spec:
//	  targetRef: {...}
//	  from: [...]
//
// while 'unversioned.Resource' is:
//
//	type: CircuitBreaker
//	name: cb1
//	mesh: default
//	sources: [...]
//	destinations: [...]
//	conf: {...}
func (f *from) Resource(r core_model.Resource) Resource {
	if r == nil {
		return nil
	}

	meta := v1alpha1.ResourceMeta{}
	if r.GetMeta() != nil {
		var meshName string
		if r.Descriptor().Scope == core_model.ScopeMesh {
			meshName = r.GetMeta().GetMesh()
		}
		meta = v1alpha1.ResourceMeta{
			Mesh:             meshName,
			Type:             string(r.Descriptor().Name),
			Name:             r.GetMeta().GetName(),
			CreationTime:     r.GetMeta().GetCreationTime(),
			ModificationTime: r.GetMeta().GetModificationTime(),
		}
	}

	if r.Descriptor().IsPluginOriginated {
		return &v1alpha1.Resource{
			ResourceMeta: meta,
			Spec:         r.GetSpec(),
		}
	} else {
		return &unversioned.Resource{
			Meta: meta,
			Spec: r.GetSpec(),
		}
	}
}

func (f *from) ResourceList(rs core_model.ResourceList) *ResourceList {
	items := make([]Resource, len(rs.GetItems()))
	for i, r := range rs.GetItems() {
		items[i] = f.Resource(r)
	}
	return &ResourceList{
		Total: rs.GetPagination().Total,
		Items: items,
	}
}

var To = &to{}

type to struct{}

func (t *to) Core(r Resource) (core_model.Resource, error) {
	resource, err := registry.Global().NewObject(core_model.ResourceType(r.GetMeta().Type))
	if err != nil {
		return nil, err
	}
	resource.SetMeta(r.GetMeta())
	if err := resource.SetSpec(r.GetSpec()); err != nil {
		return nil, err
	}
	return resource, nil
}
