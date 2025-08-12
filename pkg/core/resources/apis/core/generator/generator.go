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
	return generator{}
}

type generator struct{}

func (g generator) Generate(ctx context.Context, rs *xds.ResourceSet, xdsCtx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	for pluginName, plugin := range plugins.Plugins().CoreResourcePlugins() {
		if err := plugin.Generate(rs, xdsCtx, proxy); err != nil {
			return nil, errors.Wrapf(err, "could not apply core resource plugin %s", pluginName)
		}
	}
	return rs, nil
}
