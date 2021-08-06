package kumactl

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type Plugin struct {
	rootContext *kumactl_cmd.RootContext
}

func (p *Plugin) CustomizeContext(root *kumactl_cmd.RootContext) {
	p.rootContext = root

	registry.RegisterType(core_mesh.GatewayResourceTypeDescriptor)
}

func (p *Plugin) CustomizeCommand(_ *cobra.Command) {
}
