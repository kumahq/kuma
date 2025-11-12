package meshidentity

import (
	"github.com/kumahq/kuma/v2/pkg/core/plugins"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers/bundled"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers/spire"
)

// Map of all providers supported by the MeshIdentity resource.
// This map is not auto-generated, so whenever a new provider is implemented,
// it must be added here manually. The purpose of keeping it manual is to allow
// child projects to extend it.
var NameToModule = map[string]*plugins.PluginInitializer{
	"bundled": {InitFn: bundled.InitProvider, Initialized: false},
	"spire":   {InitFn: spire.InitProvider, Initialized: false},
}
