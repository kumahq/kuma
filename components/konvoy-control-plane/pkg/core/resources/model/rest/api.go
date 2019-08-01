package rest

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"

	"github.com/pkg/errors"
)

type Api interface {
	GetResourceApi(model.ResourceType) (ResourceApi, error)
}

type ResourceApi struct {
	CollectionPath string
}

var _ Api = &ApiDescriptor{}

type ApiDescriptor struct {
	Resources map[model.ResourceType]ResourceApi
}

func (m *ApiDescriptor) GetResourceApi(typ model.ResourceType) (ResourceApi, error) {
	mapping, ok := m.Resources[typ]
	if !ok {
		return ResourceApi{}, errors.Errorf("unknown resource type: %q", typ)
	}
	return mapping, nil
}
