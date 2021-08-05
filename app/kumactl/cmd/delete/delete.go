package delete

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/entities"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
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

			def, ok := entities.ByName[resourceTypeArg]
			if !ok {
				return errors.Errorf("unknown TYPE: %s. Allowed values: %s", resourceTypeArg, strings.Join(entities.Names, ", "))
			}
			if def.ReadOnly {
				return errors.Errorf("TYPE: %s is readOnly, can't use it for write action", resourceTypeArg)
			}

			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			var resource model.Resource
			if resource, err = registry.Global().NewObject(def.ResourceType); err != nil {
				return err
			}

			mesh := model.NoMesh
			if resource.Scope() == model.ScopeMesh {
				mesh = pctx.CurrentMesh()
			}

			if err := deleteResource(name, mesh, resource, def.ResourceType, rs); err != nil {
				return err
			}

			cmd.Printf("deleted %s %q\n", def.ResourceType, name)
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
