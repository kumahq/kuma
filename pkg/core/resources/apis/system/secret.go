package system

import (
	"errors"
	"fmt"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
)

// Config and Secret are registered unconditionally rather than through the plugin
// initializer map, which is keyed by CRD plural and gated by the enabled resource list.
// Neither has a CRD, and a control plane that cannot read secrets cannot serve identity.
const (
	SecretType model.ResourceType = "Secret"
)

var _ model.Resource = &SecretResource{}

type SecretResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Secret
}

func NewSecretResource() *SecretResource {
	return &SecretResource{
		Spec: &system_proto.Secret{},
	}
}

func (t *SecretResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *SecretResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *SecretResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *SecretResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.Secret)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.Secret{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *SecretResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *SecretResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *SecretResource) Descriptor() model.ResourceTypeDescriptor {
	return SecretResourceTypeDescriptor
}

var _ model.ResourceList = &SecretResourceList{}

type SecretResourceList struct {
	Items      []*SecretResource
	Pagination model.Pagination
}

func (l *SecretResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *SecretResourceList) GetItemType() model.ResourceType {
	return SecretType
}

func (l *SecretResourceList) NewItem() model.Resource {
	return NewSecretResource()
}

func (l *SecretResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*SecretResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*SecretResource)(nil), r)
	}
}

func (l *SecretResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *SecretResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var SecretResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  SecretType,
	Resource:              NewSecretResource(),
	ResourceList:          &SecretResourceList{},
	ReadOnly:              false,
	AdminOnly:             true,
	Scope:                 model.ScopeMesh,
	KDSFlags:              model.GlobalToZonesFlag | model.ProvidedByZoneFlag,
	SkipKDSHash:           true,
	WsPath:                "secrets",
	KumactlArg:            "secret",
	KumactlListArg:        "secrets",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Secret",
	PluralDisplayName:     "Secrets",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: true,
}

func init() {
	registry.RegisterType(SecretResourceTypeDescriptor)
}
