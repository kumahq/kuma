package universal

import (
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	"github.com/pkg/errors"
	"time"
)

var _ core_plugins.DiscoveryPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) NewDiscoverySource(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_discovery.DiscoverySource, error) {
	source := newStorePollingSource(pc.ResourceStore(), time.Second) // todo(jakubdyszkiewicz) parametrize interval
	if err := pc.ComponentManager().Add(source); err != nil {
		return nil, errors.Errorf("could not add store polling source component to the component manager")
	}
	return source, nil
}
