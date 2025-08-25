package rest

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type Api interface {
	GetResourceApi(model.ResourceType) (ResourceApi, error)
}

type ResourceApi interface {
	List(mesh string) string
	Item(mesh string, name string) string
}

func NewResourceApi(scope model.ResourceScope, path string) ResourceApi {
	switch scope {
	case model.ScopeGlobal:
		return &nonMeshedApi{CollectionPath: path}
	case model.ScopeMesh:
		return &meshedApi{CollectionPath: path}
	default:
		panic("Unsupported scope type")
	}
}

type meshedApi struct {
	CollectionPath string
}

func (r *meshedApi) List(mesh string) string {
	return fmt.Sprintf("/meshes/%s/%s", mesh, r.CollectionPath)
}

func (r meshedApi) Item(mesh string, name string) string {
	return fmt.Sprintf("/meshes/%s/%s/%s", mesh, r.CollectionPath, name)
}

type nonMeshedApi struct {
	CollectionPath string
}

func (r *nonMeshedApi) List(string) string {
	return fmt.Sprintf("/%s", r.CollectionPath)
}

func (r *nonMeshedApi) Item(_, name string) string {
	return fmt.Sprintf("/%s/%s", r.CollectionPath, name)
}

var _ Api = &ApiDescriptor{}

type ApiDescriptor struct {
	Resources map[model.ResourceType]ResourceApi
}

func (m *ApiDescriptor) GetResourceApi(typ model.ResourceType) (ResourceApi, error) {
	mapping, ok := m.Resources[typ]
	if !ok {
		return nil, errors.Errorf("unknown resource type: %q", typ)
	}
	return mapping, nil
}
