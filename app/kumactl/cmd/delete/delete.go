package delete

import (
	"context"

	"github.com/Kong/kuma/pkg/core/resources/apis/system"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func NewDeleteCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete TYPE NAME",
		Short: "Delete Kuma resources",
		Long:  `Delete Kuma resources.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceTypeArg := args[0]
			name := args[1]

			var resource model.Resource
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
			case "circuit-breaker":
				resourceType = mesh.CircuitBreakerType
			case "secret":
				resourceType = system.SecretType
			default:
				return errors.Errorf("unknown TYPE: %s. Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, traffic-route, traffic-trace, fault-injection, circuit-breaker, secret", resourceTypeArg)
			}

			currentMesh := pctx.CurrentMesh()
			if resourceType == mesh.MeshType {
				currentMesh = name
			}

			var rs store.ResourceStore
			var err error
			if resourceType == system.SecretType { // Secret is exposed via Admin Server. It will be merged into API Server eventually.
				rs, err = pctx.CurrentAdminResourceStore()
			} else {
				rs, err = pctx.CurrentResourceStore()
			}
			if err != nil {
				return err
			}

			if resource, err = registry.Global().NewObject(resourceType); err != nil {
				return err
			}

			if err := deleteResource(name, currentMesh, resource, resourceType, rs); err != nil {
				return err
			}

			cmd.Printf("deleted %s %q\n", resourceType, name)
			return nil
		},
	}

	return cmd
}

func deleteResource(name string, mesh string, resource model.Resource, resourceType model.ResourceType, rs store.ResourceStore) error {
	deleteOptions := store.DeleteBy(model.ResourceKey{Mesh: mesh, Name: name})
	if err := rs.Delete(context.Background(), resource, deleteOptions); err != nil {
		if store.IsResourceNotFound(err) {
			return errors.Errorf("there is no %s with name %q", resourceType, name)
		}
		return errors.Wrapf(err, "failed to delete %s with the name %q", resourceType, name)
	}

	return nil
}
