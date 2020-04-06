package get

import (
	"fmt"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetEntitiesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource TYPE NAME",
		Short: "Show Single Resource",
		Long:  `Show Single Resource.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			resourceTypeArg := args[0]

			var resourceType model.ResourceType
			switch resourceTypeArg {
			case "mesh":
				resourceType = mesh.MeshType
			case "dataplane":
				resourceType = mesh.DataplaneType
			case "healthcheck":
				resourceType = mesh.HealthCheckType
			case "proxytemplate":
				resourceType = mesh.ProxyTemplateType
			case "traffic-log":
				resourceType = mesh.TrafficLogType
			case "traffic-permission":
				resourceType = mesh.TrafficPermissionType
			case "traffic-route":
				resourceType = mesh.TrafficRouteType
			case "traffic-trace":
				resourceType = mesh.TrafficTraceType
			case "fault-injection":
				resourceType = mesh.FaultInjectionType

			default:
				return errors.Errorf("unknown TYPE: %s. Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, traffic-route, traffic-trace, fault-injection", resourceTypeArg)
			}
			fmt.Println("resource", resourceType)
			return nil
		},
	}
	return cmd
}
