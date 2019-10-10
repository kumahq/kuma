package system

import (
	"errors"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/gogo/protobuf/types"
)

const (
	SecretType model.ResourceType = "Secret"
)

var _ model.Resource = &SecretResource{}

type SecretResource struct {
	Meta model.ResourceMeta
	Spec types.BytesValue
}

func (t *SecretResource) GetType() model.ResourceType {
	return SecretType
}
func (t *SecretResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *SecretResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *SecretResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *SecretResource) SetSpec(spec model.ResourceSpec) error {
	value, ok := spec.(*types.BytesValue)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *value
		return nil
	}
}

var _ model.ResourceList = &SecretResourceList{}

type SecretResourceList struct {
	Items []*SecretResource
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
	return &SecretResource{}
}
func (l *SecretResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*SecretResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*SecretResource)(nil), r)
	}
}
