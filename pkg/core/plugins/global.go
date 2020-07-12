package plugins

import (
	"os"

	"github.com/kumahq/kuma/pkg/core"
)

var global = NewRegistry()

func Plugins() Registry {
	return global
}

func Register(name PluginName, plugin Plugin) {
	if err := global.Register(name, plugin); err != nil {
		core.Log.Error(err, "failed to register a plugin", "name", name, "plugin", plugin)
		os.Exit(1)
	}
}
