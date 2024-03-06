// Generated by tools/policy-gen
// Run "make generate" to update this file.
package policies

import (
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
)

var NameToModule = map[string]*plugins.PluginInitializer{
	"meshservices": {InitFn: meshservice.InitPlugin, Initialized: false},
}
