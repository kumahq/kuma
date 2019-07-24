package rest

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"

	"github.com/pkg/errors"
)

type Mapper interface {
	GetMapping(model.ResourceType) (ResourceMapping, error)
}

type ResourceMapping struct {
	CollectionPath string
}

var _ Mapper = &SimpleMapper{}

type SimpleMapper struct {
	Resources map[model.ResourceType]ResourceMapping
}

func (m *SimpleMapper) GetMapping(typ model.ResourceType) (ResourceMapping, error) {
	mapping, ok := m.Resources[typ]
	if !ok {
		return ResourceMapping{}, errors.Errorf("unknown resource type: %q", typ)
	}
	return mapping, nil
}
