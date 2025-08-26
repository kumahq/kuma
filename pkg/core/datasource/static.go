package datasource

import (
	"context"

	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type staticLoader struct {
	secrets map[model.ResourceKey]*system.SecretResource
}

var _ Loader = &staticLoader{}

// NewStaticLoader returns a loader that supports predefined list of secrets
// This implementation is more performant if than dynamic if we already have the list of all secrets
// because we can avoid I/O operations.
func NewStaticLoader(secrets []*system.SecretResource) Loader {
	loader := staticLoader{
		secrets: map[model.ResourceKey]*system.SecretResource{},
	}

	for _, secret := range secrets {
		loader.secrets[model.MetaToResourceKey(secret.GetMeta())] = secret
	}

	return &loader
}

func (s *staticLoader) Load(_ context.Context, mesh string, source *system_proto.DataSource) ([]byte, error) {
	var data []byte
	var err error
	switch source.GetType().(type) {
	case *system_proto.DataSource_Secret:
		data, err = s.loadSecret(mesh, source.GetSecret())
	case *system_proto.DataSource_Inline:
		data, err = source.GetInline().GetValue(), nil
	case *system_proto.DataSource_InlineString:
		data, err = []byte(source.GetInlineString()), nil
	default:
		return nil, errors.New("unsupported type of the DataSource")
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not load data")
	}
	return data, nil
}

func (s *staticLoader) loadSecret(mesh, name string) ([]byte, error) {
	key := model.ResourceKey{
		Mesh: mesh,
		Name: name,
	}

	secret := s.secrets[key]
	if secret == nil {
		return nil, core_store.ErrorResourceNotFound(system.SecretType, name, mesh)
	}
	return secret.Spec.GetData().GetValue(), nil
}
