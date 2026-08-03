package generator

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/dynconf"
	generator_core "github.com/kumahq/kuma/v3/pkg/xds/generator/core"
)

func NewGenerator() generator_core.ResourceGenerator {
	return generator{}
}

type generator struct{}

func (g generator) Generate(ctx context.Context, rs *xds.ResourceSet, xdsCtx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	proxy.OtelPipeBackends = &xds.OtelPipeBackends{}

	for _, policy := range plugins.Plugins().PolicyPlugins() {
		if err := policy.Plugin.Apply(rs, xdsCtx, proxy); err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", policy.Name)
		}
	}

	if err := writeUnifiedOtelRoute(rs, proxy); err != nil {
		return nil, err
	}

	return rs, nil
}

func writeUnifiedOtelRoute(rs *xds.ResourceSet, proxy *xds.Proxy) error {
	if proxy.OtelPipeBackends.Empty() {
		return nil
	}

	if !proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp) {
		return nil
	}

	dpConfig := xds.OtelDpConfig{
		Backends: proxy.OtelPipeBackends.All(),
	}

	data, err := json.Marshal(dpConfig)
	if err != nil {
		return errors.Wrap(err, "marshaling otel dp config")
	}

	return dynconf.AddConfigRoute(
		proxy,
		rs,
		"otel",
		xds.OtelDynconfPath,
		data,
	)
}
