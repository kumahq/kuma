package unversioned

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
)

var From = &from{}

type from struct{}

// Deprecated: use 'rest.From.Resource()' instead
func (c *from) Resource(r core_model.Resource) *Resource {
	var meshName string
	if r.Descriptor().Scope == core_model.ScopeMesh {
		meshName = r.GetMeta().GetMesh()
	}
	return &Resource{
		Meta: v1alpha1.ResourceMeta{
			Mesh:             meshName,
			Type:             string(r.Descriptor().Name),
			Name:             r.GetMeta().GetName(),
			CreationTime:     r.GetMeta().GetCreationTime(),
			ModificationTime: r.GetMeta().GetModificationTime(),
		},
		Spec: r.GetSpec(),
	}
}
