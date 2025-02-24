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

func (t *GlobalSecretResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *GlobalSecretResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
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

func (l *GlobalSecretResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *GlobalSecretResourceList) GetItemType() model.ResourceType {
	return GlobalSecretType
}

func (l *GlobalSecretResourceList) NewItem() model.Resource {
	return NewGlobalSecretResource()
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

func (l *GlobalSecretResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var GlobalSecretResourceTypeDescriptor model.ResourceTypeDescriptor

func init() {
	// GlobalSecret is a frankenstein type as it's extracted from the secret one with a different scope
	// so for the type descriptor we copy the one from secret and then change types
	GlobalSecretResourceTypeDescriptor = SecretResourceTypeDescriptor
	GlobalSecretResourceTypeDescriptor.Name = GlobalSecretType
	GlobalSecretResourceTypeDescriptor.Resource = NewGlobalSecretResource()
	GlobalSecretResourceTypeDescriptor.ResourceList = &GlobalSecretResourceList{}
	GlobalSecretResourceTypeDescriptor.Scope = model.ScopeGlobal
	GlobalSecretResourceTypeDescriptor.WsPath = "global-secrets"
	GlobalSecretResourceTypeDescriptor.KumactlArg = "global-secret"
	GlobalSecretResourceTypeDescriptor.KumactlListArg = "global-secrets"
	GlobalSecretResourceTypeDescriptor.SingularDisplayName = "Global Secret"
	GlobalSecretResourceTypeDescriptor.PluralDisplayName = "Global Secrets"
	registry.RegisterType(GlobalSecretResourceTypeDescriptor)
}
