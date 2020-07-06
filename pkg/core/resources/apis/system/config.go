package system

import (
	"errors"

	config_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	ConfigType model.ResourceType = "Config"
)

var _ model.Resource = &ConfigResource{}

type ConfigResource struct {
	Meta model.ResourceMeta
	Spec config_proto.Config
}

func (t *ConfigResource) GetType() model.ResourceType {
	return ConfigType
}
func (t *ConfigResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *ConfigResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *ConfigResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *ConfigResource) SetSpec(spec model.ResourceSpec) error {
	value, ok := spec.(*config_proto.Config)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *value
		return nil
	}
}
func (t *ConfigResource) Validate() error {
	return nil
}

var _ model.ResourceList = &ConfigResourceList{}

type ConfigResourceList struct {
	Items      []*ConfigResource
	Pagination model.Pagination
}

func (l *ConfigResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *ConfigResourceList) GetItemType() model.ResourceType {
	return ConfigType
}
func (l *ConfigResourceList) NewItem() model.Resource {
	return &ConfigResource{}
}
func (l *ConfigResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ConfigResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ConfigResource)(nil), r)
	}
}
func (l *ConfigResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&ConfigResource{})
	registry.RegistryListType(&ConfigResourceList{})
}
