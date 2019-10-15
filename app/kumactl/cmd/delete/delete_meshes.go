package delete

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newDeleteMeshCmd(pctx *deleteContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mesh",
		Short: "Delete Mesh",
		Long:  `Delete Mesh.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			meshName := args[0]

			getOptions := store.GetByKey(model.DefaultNamespace, meshName, meshName)
			meshDetails := mesh.MeshResource{}
			if err := rs.Get(context.Background(), &meshDetails, getOptions); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("there is no Mesh with name %q", meshName)
				}
				return errors.Wrapf(err, "failed to get Mesh with the name %q", meshName)
			}

			deleteOptions := store.DeleteByKey(model.DefaultNamespace, meshName, meshName)
			if err := rs.Delete(context.Background(), &mesh.MeshResource{}, deleteOptions); err != nil {
				return errors.Wrapf(err, "failed to delete Mesh with the name %q", meshName)
			}

			cmd.Printf("deleted Mesh %q\n", meshName)
			return nil
		},
	}
	return cmd
}
