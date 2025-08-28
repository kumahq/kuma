package config_dump

import (
	envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"

	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type EnvoyConfig struct {
	Boostrap     envoy_admin_v3.BootstrapConfigDump
	Cluster      envoy_admin_v3.ClustersConfigDump
	Endpoints    envoy_admin_v3.EndpointsConfigDump
	Listeners    envoy_admin_v3.ListenersConfigDump
	ScopedRoutes envoy_admin_v3.ScopedRoutesConfigDump
	Routes       envoy_admin_v3.RoutesConfigDump
}

func ParseEnvoyConfig(bs []byte) (*EnvoyConfig, error) {
	var cd envoy_admin_v3.ConfigDump
	if err := proto.FromJSON(bs, &cd); err != nil {
		return nil, err
	}

	var config EnvoyConfig

	for _, raw := range cd.GetConfigs() {
		var err error

		switch {
		case raw.MessageIs(&config.Boostrap):
			err = raw.UnmarshalTo(&config.Boostrap)
		case raw.MessageIs(&config.Cluster):
			err = raw.UnmarshalTo(&config.Cluster)
		case raw.MessageIs(&config.Endpoints):
			err = raw.UnmarshalTo(&config.Endpoints)
		case raw.MessageIs(&config.Listeners):
			err = raw.UnmarshalTo(&config.Listeners)
		case raw.MessageIs(&config.ScopedRoutes):
			err = raw.UnmarshalTo(&config.ScopedRoutes)
		case raw.MessageIs(&config.Routes):
			err = raw.UnmarshalTo(&config.Routes)
		}

		if err != nil {
			return nil, err
		}
	}

	return &config, nil
}

type Metadata struct {
	features xds_types.Features
}

func (m *Metadata) HasFeature(feature string) bool {
	return m.features[feature]
}

func (c *EnvoyConfig) Metadata() *Metadata {
	meta := Metadata{features: xds_types.Features{}}

	if raw, ok := c.Boostrap.GetBootstrap().GetNode().GetMetadata().Fields["features"]; ok {
		for _, value := range raw.GetListValue().GetValues() {
			meta.features[value.GetStringValue()] = true
		}
	}

	return &meta
}
