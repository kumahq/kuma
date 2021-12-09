package datasource

import (
	"context"
	"os"

	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type Loader interface {
	Load(ctx context.Context, mesh string, source *system_proto.DataSource) ([]byte, error)
}

type loader struct {
	secretManager manager.ReadOnlyResourceManager
}

var _ Loader = &loader{}

func NewDataSourceLoader(secretManager manager.ReadOnlyResourceManager) Loader {
	return &loader{
		secretManager: secretManager,
	}
}

func (l *loader) Load(ctx context.Context, mesh string, source *system_proto.DataSource) ([]byte, error) {
	var data []byte
	var err error
	switch source.GetType().(type) {
	case *system_proto.DataSource_Secret:
		data, err = l.loadSecret(ctx, mesh, source.GetSecret())
	case *system_proto.DataSource_Inline:
		data, err = source.GetInline().GetValue(), nil
	case *system_proto.DataSource_InlineString:
		data, err = []byte(source.GetInlineString()), nil
	case *system_proto.DataSource_File:
		data, err = os.ReadFile(source.GetFile())
	default:
		return nil, errors.New("unsupported type of the DataSource")
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not load data")
	}
	return data, nil
}

func (l *loader) loadSecret(ctx context.Context, mesh string, secret string) ([]byte, error) {
	if l.secretManager == nil {
		return nil, errors.New("no resource manager")
	}
	resource := system.NewSecretResource()
	if err := l.secretManager.Get(ctx, resource, core_store.GetByKey(secret, mesh)); err != nil {
		return nil, err
	}
	return resource.Spec.GetData().GetValue(), nil
}
