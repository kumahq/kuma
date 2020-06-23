package universal

import (
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	plugin_leader "github.com/Kong/kuma/pkg/plugins/leader"
)

var _ core_plugins.BootstrapPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) Bootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	leaderElector, err := plugin_leader.NewLeaderElector(b)
	if err != nil {
		return err
	}
	b.WithComponentManager(component.NewManager(leaderElector))
	return nil
}
