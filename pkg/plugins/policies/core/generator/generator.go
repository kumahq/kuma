package generator

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
)

func NewGenerator() generator_core.ResourceGenerator {
	return generator{
		plugins: plugins.Plugins().PolicyPlugins(),
	}
}

type generator struct {
	plugins map[plugins.PluginName]plugins.PolicyPlugin
}

func (g generator) Generate(ctx context.Context, rs *xds.ResourceSet, xdsCtx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	for _, policy := range g.plugins {
		if err := policy.Apply(rs, xdsCtx, proxy); err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", policy)
		}
	}
	return rs, nil
}
