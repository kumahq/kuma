package generator

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	unified_naming "github.com/kumahq/kuma/v2/pkg/core/naming/unified-naming"
	"github.com/kumahq/kuma/v2/pkg/core/plugins"
	core_system_names "github.com/kumahq/kuma/v2/pkg/core/system_names"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/ordered"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/dynconf"
	generator_core "github.com/kumahq/kuma/v2/pkg/xds/generator/core"
)

func NewGenerator() generator_core.ResourceGenerator {
	return generator{}
}

type generator struct{}

func (generator) Generate(ctx context.Context, rs *xds.ResourceSet, xdsCtx xds_context.Context, proxy *xds.Proxy) (*xds.ResourceSet, error) {
	proxy.OtelPipeBackends = &xds.OtelPipeBackends{}

	for _, policy := range plugins.Plugins().PolicyPlugins(ordered.Policies) {
		if err := policy.Plugin.Apply(rs, xdsCtx, proxy); err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", policy.Name)
		}
	}

	if err := writeUnifiedOtelRoute(rs, xdsCtx, proxy); err != nil {
		return nil, err
	}

	return rs, nil
}

func writeUnifiedOtelRoute(rs *xds.ResourceSet, xdsCtx xds_context.Context, proxy *xds.Proxy) error {
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

	unifiedNamingEnabled := unified_naming.Enabled(proxy.Metadata, xdsCtx.Mesh.Resource)
	getNameOrDefault := core_system_names.GetNameOrDefault(unifiedNamingEnabled)

	return dynconf.AddConfigRoute(
		proxy,
		rs,
		unifiedNamingEnabled,
		getNameOrDefault("otel", xds.OtelDynconfPath),
		xds.OtelDynconfPath,
		data,
	)
}
