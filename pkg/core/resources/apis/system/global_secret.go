package system

import (
	"errors"

	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
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

func (r *GlobalSecretResource) GetMeta() model.ResourceMeta {
	return r.Meta
}

func (r *GlobalSecretResource) SetMeta(m model.ResourceMeta) {
	r.Meta = m
}

func (r *GlobalSecretResource) GetSpec() model.ResourceSpec {
	return r.Spec
}

func (r *GlobalSecretResource) SetSpec(spec model.ResourceSpec) error {
	value, ok := spec.(*system_proto.Secret)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		r.Spec = value
		return nil
	}
}

func (*GlobalSecretResource) GetStatus() model.ResourceStatus {
	return nil
}

func (*GlobalSecretResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (*GlobalSecretResource) Validate() error {
	return nil
}

func (*GlobalSecretResource) Descriptor() model.ResourceTypeDescriptor {
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

func (*GlobalSecretResourceList) GetItemType() model.ResourceType {
	return GlobalSecretType
}

func (*GlobalSecretResourceList) NewItem() model.Resource {
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
	GlobalSecretResourceTypeDescriptor.AlternativeWsPath = "global-secrets"
	GlobalSecretResourceTypeDescriptor.KumactlArgAlias = "global-secret"
	GlobalSecretResourceTypeDescriptor.KumactlListArgAlias = "global-secrets"
	GlobalSecretResourceTypeDescriptor.WsPath = "globalsecrets"
	GlobalSecretResourceTypeDescriptor.KumactlArg = "globalsecret"
	GlobalSecretResourceTypeDescriptor.KumactlListArg = "globalsecrets"
	GlobalSecretResourceTypeDescriptor.SingularDisplayName = "Global Secret"
	GlobalSecretResourceTypeDescriptor.PluralDisplayName = "Global Secrets"
	registry.RegisterType(GlobalSecretResourceTypeDescriptor)
}
