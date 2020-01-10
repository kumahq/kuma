package generator

import (
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
)

type ResourceGenerator interface {
	Generate(xds_context.Context, *model.Proxy) ([]*model.Resource, error)
}

type CompositeResourceGenerator []ResourceGenerator

func (c CompositeResourceGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	resources := make([]*model.Resource, 0)
	for _, gen := range c {
		rs, err := gen.Generate(ctx, proxy)
		if err != nil {
			return nil, err
		}
		resources = append(resources, rs...)
	}
	return resources, nil
}
