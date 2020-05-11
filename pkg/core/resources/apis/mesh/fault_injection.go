package mesh

import (
	"errors"

	"github.com/Kong/kuma/pkg/core/resources/registry"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	FaultInjectionType model.ResourceType = "FaultInjection"
)

var _ model.Resource = &FaultInjectionResource{}

type FaultInjectionResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.FaultInjection
}

func (f *FaultInjectionResource) GetType() model.ResourceType {
	return FaultInjectionType
}

func (f *FaultInjectionResource) GetMeta() model.ResourceMeta {
	return f.Meta
}

func (f *FaultInjectionResource) SetMeta(m model.ResourceMeta) {
	f.Meta = m
}

func (f *FaultInjectionResource) GetSpec() model.ResourceSpec {
	return &f.Spec
}

func (f *FaultInjectionResource) SetSpec(spec model.ResourceSpec) error {
	faultInjection, ok := spec.(*mesh_proto.FaultInjection)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		f.Spec = *faultInjection
		return nil
	}
}

var _ model.ResourceList = &FaultInjectionResourceList{}

type FaultInjectionResourceList struct {
	Items      []*FaultInjectionResource
	Pagination model.Pagination
}

func (l *FaultInjectionResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *FaultInjectionResourceList) GetItemType() model.ResourceType {
	return FaultInjectionType
}

func (l *FaultInjectionResourceList) NewItem() model.Resource {
	return &FaultInjectionResource{}
}

func (l *FaultInjectionResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*FaultInjectionResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*FaultInjectionResource)(nil), r)
	}
}

func (l *FaultInjectionResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&FaultInjectionResource{})
	registry.RegistryListType(&FaultInjectionResourceList{})
}

func (f *FaultInjectionResource) Sources() []*mesh_proto.Selector {
	return f.Spec.GetSources()
}

func (f *FaultInjectionResource) Destinations() []*mesh_proto.Selector {
	return f.Spec.GetDestinations()
}
