package gateway

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

// ListenerGenerator generates Kuma gateway listeners. Not implemented.
type ListenerGenerator struct {
	Resources manager.ReadOnlyResourceManager
}

var _ generator.ResourceGenerator = ListenerGenerator{}

func (ListenerGenerator) Generate(ctx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	return nil, errors.New("not implemented")
}
