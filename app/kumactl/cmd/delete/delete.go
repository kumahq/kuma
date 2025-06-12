package delete

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewDeleteCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	byName := map[string]model.ResourceTypeDescriptor{}
	allNames := []string{}
	for _, desc := range pctx.Runtime.Registry.ObjectDescriptors(model.HasKumactlEnabled()) {
		byName[desc.KumactlArg] = desc
		if desc.KumactlArgAlias != "" {
			byName[desc.KumactlArgAlias] = desc
			allNames = append(allNames, desc.KumactlArgAlias)
		}
		allNames = append(allNames, desc.KumactlArg)
	}
	sort.Strings(allNames)
	cmd := &cobra.Command{
		Use:   "delete TYPE NAME",
		Short: "Delete Kuma resources",
		Long:  `Delete Kuma resources.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = kumactl_cmd.CheckCompatibility(pctx.FetchServerVersion, cmd.ErrOrStderr())

			resourceTypeArg := args[0]
			name := args[1]

			desc, ok := byName[resourceTypeArg]
			if !ok {
				return errors.Errorf("unknown TYPE: %s. Allowed values: %s", resourceTypeArg, strings.Join(allNames, ", "))
			}
			if desc.ReadOnly {
				return errors.Errorf("TYPE: %s is readOnly, can't use it for write action", resourceTypeArg)
			}

			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			mesh := model.NoMesh
			if desc.Scope == model.ScopeMesh {
				mesh = pctx.CurrentMesh()
			}

			if err := deleteResource(name, mesh, desc, rs); err != nil {
				return err
			}

			cmd.Printf("deleted %s %q\n", desc.Name, name)
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	return cmd
}

func deleteResource(name string, mesh string, desc model.ResourceTypeDescriptor, rs store.ResourceStore) error {
	resource := desc.NewObject()
	deleteOptions := store.DeleteBy(model.ResourceKey{Mesh: mesh, Name: name})
	if err := rs.Delete(context.Background(), resource, deleteOptions); err != nil {
		if store.IsNotFound(err) {
			return errors.Errorf("there is no %s with name %q", desc.Name, name)
		}
		return errors.Wrapf(err, "failed to delete %s with the name %q", desc.Name, name)
	}

	return nil
}
