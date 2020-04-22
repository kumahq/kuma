package provided

import (
	"github.com/Kong/kuma/pkg/core/ca"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
)

var _ core_plugins.CaPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.CaProvided, &plugin{})
}

func (p plugin) NewCaManager(context core_plugins.PluginContext, config core_plugins.PluginConfig) (ca.CaManager, error) {
	return NewProvidedCaManager(context.DataSourceLoader()), nil
}
