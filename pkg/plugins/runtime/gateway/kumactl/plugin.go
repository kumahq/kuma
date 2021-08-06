package kumactl

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/entities"
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
	entities.All = append(entities.All, entities.Definition{
		Singular:     "gateway",
		Plural:       "gateways",
		ResourceType: core_mesh.GatewayType,
	})
}

func (p *Plugin) CustomizeCommand(_ *cobra.Command) {
}
