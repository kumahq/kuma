package builtin

import (
	"github.com/kumahq/kuma/pkg/core/ca"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
)

var _ core_plugins.CaPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.CaBuiltin, &plugin{})
}

func (p plugin) NewCaManager(context core_plugins.PluginContext, config core_plugins.PluginConfig) (ca.Manager, error) {
	return NewBuiltinCaManager(context.ResourceManager()), nil
}
