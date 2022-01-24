package universal

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	plugin_leader "github.com/kumahq/kuma/pkg/plugins/leader"
)

var _ core_plugins.BootstrapPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) BeforeBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	leaderElector, err := plugin_leader.NewLeaderElector(b)
	if err != nil {
		return err
	}
	b.WithComponentManager(component.NewManager(leaderElector))
	return nil
}

func (p *plugin) AfterBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return core_plugins.Universal
}

func (p *plugin) Order() int {
	return core_plugins.EnvironmentPreparingOrder
}
