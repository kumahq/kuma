package system

import (
	"errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	GlobalSecretType model.ResourceType = "GlobalSecret"
)

var _ model.Resource = &GlobalSecretResource{}

type GlobalSecretResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Secret
}

func NewGlobalSecretResource() *GlobalSecretResource {
	return &GlobalSecretResource{
		Spec: &system_proto.Secret{},
	}
}

func (t *GlobalSecretResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *GlobalSecretResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *GlobalSecretResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *GlobalSecretResource) SetSpec(spec model.ResourceSpec) error {
	value, ok := spec.(*system_proto.Secret)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = value
		return nil
	}
}

func (t *GlobalSecretResource) Validate() error {
	return nil
}

func (t *GlobalSecretResource) Descriptor() model.ResourceTypeDescriptor {
	return GlobalSecretResourceTypeDescriptor
}

var _ model.ResourceList = &GlobalSecretResourceList{}

type GlobalSecretResourceList struct {
	Items      []*GlobalSecretResource
	Pagination model.Pagination
}

func (l *GlobalSecretResourceList) Descriptor() model.ResourceTypeDescriptor {
	return GlobalSecretResourceTypeDescriptor
}

func (l *GlobalSecretResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *GlobalSecretResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*GlobalSecretResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*GlobalSecretResource)(nil), r)
	}
}

func (l *GlobalSecretResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var GlobalSecretResourceTypeDescriptor model.ResourceTypeDescriptor

func init() {
	GlobalSecretResourceTypeDescriptor = model.ResourceTypeDescriptor{
		Name:           GlobalSecretType,
		Resource:       NewGlobalSecretResource(),
		ResourceList:   &GlobalSecretResourceList{},
		ReadOnly:       false,
		AdminOnly:      true,
		Scope:          model.ScopeGlobal,
		KDSFlags:       model.FromGlobalToZone,
		WsPath:         "global-secrets",
		KumactlArg:     "global-secret",
		KumactlListArg: "global-secrets",
	}
	registry.RegisterType(GlobalSecretResourceTypeDescriptor)
}
