package sync

import (
	"context"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type GlobalLoadBalancerProxyBuilder struct {
	*DataplaneProxyBuilder
}

func (p *GlobalLoadBalancerProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey, meshContext xds_context.MeshContext) (*core_xds.Proxy, error) {
	// TODO(nicoche) calculate more stuff to put in the Proxy
	return p.DataplaneProxyBuilder.Build(ctx, key, meshContext)
}
