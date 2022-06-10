package core

import (
	"github.com/pkg/errors"

	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type ResourceGenerator interface {
	Generate(xds_context.Context, *model.Proxy) (*model.ResourceSet, error)
}

type CompositeResourceGenerator []ResourceGenerator

func (c CompositeResourceGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for _, gen := range c {
		rs, err := gen.Generate(ctx, proxy)
		if err != nil {
			return nil, errors.Wrapf(err, "%T failed", gen)
		}
		resources.AddSet(rs)
	}
	return resources, nil
}
