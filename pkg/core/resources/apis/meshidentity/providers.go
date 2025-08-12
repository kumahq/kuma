package meshidentity

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers/bundled"
)

// Map of all providers supported by the MeshIdentity resource.
// This map is not auto-generated, so whenever a new provider is implemented,
// it must be added here manually. The purpose of keeping it manual is to allow
// child projects to extend it.
var NameToModule = map[string]*plugins.PluginInitializer{
	"bundled": {InitFn: bundled.InitProvider, Initialized: false},
}
