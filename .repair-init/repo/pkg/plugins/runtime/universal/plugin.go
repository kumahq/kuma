package universal

import (
	"github.com/kumahq/kuma/v3/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
)

var _ core_plugins.RuntimePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	if rt.Config().Environment != core.UniversalEnvironment {
		return nil
	}
	if rt.Config().Mode == core.Global {
		return nil
	}

	return nil
}
