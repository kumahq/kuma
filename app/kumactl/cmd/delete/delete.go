package delete

import (
	"context"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete TYPE NAME",
		Short: "Delete Kuma resources",
		Long:  `Delete Kuma resources.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			resourceTypeArg := args[0]
			name := args[1]

			var resource model.Resource
			var resourceType model.ResourceType
			switch resourceTypeArg {
			case "mesh":
				resourceType = mesh.MeshType
			case "dataplane":
				resourceType = mesh.DataplaneType
			case "proxytemplate":
				resourceType = mesh.ProxyTemplateType
			case "traffic-log":
				resourceType = mesh.TrafficLogType
			case "traffic-permission":
				resourceType = mesh.TrafficPermissionType

			default:
				return errors.Errorf("unknown TYPE: %s. Allowed values: mesh, dataplane, proxytemplate, traffic-log, traffic-permission", resourceTypeArg)
			}

			currentMesh := pctx.CurrentMesh()
			if resourceType == mesh.MeshType {
				currentMesh = name
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

func deleteResource(name string, currentMesh string, resource model.Resource, resourceType model.ResourceType, rs store.ResourceStore) error {
	getOptions := store.GetByKey(model.DefaultNamespace, name, currentMesh)
	if err := rs.Get(context.Background(), resource, getOptions); err != nil {
		if store.IsResourceNotFound(err) {
			return errors.Errorf("there is no %s with name %q", resourceType, name)
		}
		return errors.Wrapf(err, "failed to get %s with the name %q", resourceType, name)
	}

	deleteOptions := store.DeleteByKey(model.DefaultNamespace, name, currentMesh)
	if err := rs.Delete(context.Background(), resource, deleteOptions); err != nil {
		return errors.Wrapf(err, "failed to delete %s with the name %q", resourceType, name)
	}

	return nil
}
