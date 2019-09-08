package universal

import (
	universal_config "github.com/Kong/kuma/pkg/config/plugins/discovery/universal"
	core_discovery "github.com/Kong/kuma/pkg/core/discovery"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	"github.com/pkg/errors"
)

var _ core_plugins.DiscoveryPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) NewDiscoverySource(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_discovery.DiscoverySource, error) {
	cfg, ok := config.(*universal_config.UniversalDiscoveryConfig)
	if !ok {
		return nil, errors.Errorf("wrong type of configuration. Expected: *universal_config.UniversalDiscoveryConfig, got: %T", config)
	}
	source := newStorePollingSource(pc.ResourceStore(), cfg.PollingInterval)
	if err := pc.ComponentManager().Add(source); err != nil {
		return nil, errors.Errorf("could not add store polling source component to the component manager")
	}
	return source, nil
}
