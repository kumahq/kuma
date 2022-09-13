package universal

import (
	config_core "github.com/kumahq/kuma/pkg/config/core"
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

func (p *plugin) BeforeBootstrap(r core_runtime.Runtime, _ core_plugins.PluginConfig) error {
	if r.Config().Environment != config_core.UniversalEnvironment {
		return nil
	}
	leaderElector, err := plugin_leader.NewLeaderElector(r.Config(), r.GetInstanceId())
	if err != nil {
		return err
	}
	return core_runtime.ApplyOpts(r, core_runtime.WithComponentManager(component.NewManager(leaderElector)))
}

func (p *plugin) AfterBootstrap(_ core_runtime.Runtime, _ core_plugins.PluginConfig) error {
	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return core_plugins.Universal
}

func (p *plugin) Order() int {
	return core_plugins.EnvironmentPreparingOrder
}
