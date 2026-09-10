package system

import (
	"errors"
	"fmt"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
)

const (
	ConfigType model.ResourceType = "Config"
)

var _ model.Resource = &ConfigResource{}

type ConfigResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Config
}

func NewConfigResource() *ConfigResource {
	return &ConfigResource{
		Spec: &system_proto.Config{},
	}
}

func (t *ConfigResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ConfigResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ConfigResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ConfigResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.Config)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.Config{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *ConfigResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *ConfigResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *ConfigResource) Descriptor() model.ResourceTypeDescriptor {
	return ConfigResourceTypeDescriptor
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
	return NewConfigResource()
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

func (l *ConfigResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var ConfigResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  ConfigType,
	Resource:              NewConfigResource(),
	ResourceList:          &ConfigResourceList{},
	ReadOnly:              false,
	AdminOnly:             false,
	Scope:                 model.ScopeGlobal,
	KDSFlags:              model.GlobalToZonesFlag | model.ProvidedByZoneFlag,
	SkipKDSHash:           true,
	WsPath:                "",
	KumactlArg:            "",
	KumactlListArg:        "",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Config",
	PluralDisplayName:     "Configs",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: true,
}

func init() {
	registry.RegisterType(ConfigResourceTypeDescriptor)
}
