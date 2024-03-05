package core

import (
	"context"

	"github.com/pkg/errors"

	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type ResourceGenerator interface {
	Generate(context.Context, *model.ResourceSet, xds_context.Context, *model.Proxy) (*model.ResourceSet, error)
}

type CompositeResourceGenerator []ResourceGenerator

func (c CompositeResourceGenerator) Generate(
	ctx context.Context, resources *model.ResourceSet, xdsCtx xds_context.Context, proxy *model.Proxy,
) (*model.ResourceSet, error) {
	for _, gen := range c {
		rs, err := gen.Generate(ctx, resources, xdsCtx, proxy)
		if err != nil {
			return nil, errors.Wrapf(err, "%T failed", gen)
		}
		resources.AddSet(rs)
	}
	return resources, nil
}
