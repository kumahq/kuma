package meshidentity

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers/bundled"
)

var NameToModule = map[string]*plugins.PluginInitializer{
	"bundled": {InitFn: bundled.InitProvider, Initialized: false},
}
