package mesh

import (
	"errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	ProxyTemplateType model.ResourceType = "ProxyTemplate"
)

var _ model.Resource = &ProxyTemplateResource{}

type ProxyTemplateResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.ProxyTemplate
}

func (t *ProxyTemplateResource) GetType() model.ResourceType {
	return ProxyTemplateType
}
func (t *ProxyTemplateResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *ProxyTemplateResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *ProxyTemplateResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *ProxyTemplateResource) SetSpec(spec model.ResourceSpec) error {
	template, ok := spec.(*mesh_proto.ProxyTemplate)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *template
		return nil
	}
}

var _ model.ResourceList = &ProxyTemplateResourceList{}

type ProxyTemplateResourceList struct {
	Items      []*ProxyTemplateResource
	Pagination model.Pagination
}

func (l *ProxyTemplateResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *ProxyTemplateResourceList) GetItemType() model.ResourceType {
	return ProxyTemplateType
}
func (l *ProxyTemplateResourceList) NewItem() model.Resource {
	return &ProxyTemplateResource{}
}
func (l *ProxyTemplateResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ProxyTemplateResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ProxyTemplateResource)(nil), r)
	}
}
func (l *ProxyTemplateResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&ProxyTemplateResource{})
	registry.RegistryListType(&ProxyTemplateResourceList{})
}

func (t *ProxyTemplateResource) Selectors() []*mesh_proto.Selector {
	return t.Spec.Selectors
}
