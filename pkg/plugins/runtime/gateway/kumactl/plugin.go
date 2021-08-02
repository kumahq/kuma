package kumactl

import (
	"github.com/spf13/cobra"

	kumactl_get "github.com/kumahq/kuma/app/kumactl/cmd/get"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/api-server/definitions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type Plugin struct {
	rootContext *kumactl_cmd.RootContext
}

func (p *Plugin) CustomizeContext(root *kumactl_cmd.RootContext) {
	p.rootContext = root

	registry.RegisterType(core_mesh.NewGatewayResource())
	registry.RegistryListType(&core_mesh.GatewayResourceList{})

	// Register the API resource so kumactl can find the URI.
	definitions.All = append(definitions.All,
		definitions.ResourceWsDefinition{
			Type: core_mesh.GatewayType,
			Path: "gateways",
		},
	)

	root.TypeArgs["gateway"] = core_mesh.GatewayType
}

func (p *Plugin) CustomizeCommand(root *cobra.Command) {
	get := kumactl_cmd.FindSubCommand(root, "get")

	get.AddCommand(
		kumactl_get.WithPaginationArgs(
			kumactl_get.NewGetResourcesCmd(p.rootContext, "gateways", core_mesh.GatewayType, kumactl_get.BasicResourceTablePrinter),
			&p.rootContext.ListContext,
		),
	)

	get.AddCommand(
		kumactl_get.NewGetResourceCmd(p.rootContext, "gateway", core_mesh.GatewayType, kumactl_get.BasicResourceTablePrinter),
	)
}
