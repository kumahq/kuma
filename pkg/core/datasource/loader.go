package datasource

import (
	"context"

	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
)

type Loader interface {
	Load(ctx context.Context, mesh string, source *system_proto.DataSource) ([]byte, error)
}
