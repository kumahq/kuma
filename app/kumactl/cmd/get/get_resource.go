package get

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewGetResourceCmd(pctx *kumactl_cmd.RootContext, desc core_model.ResourceTypeDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s NAME", desc.KumactlArg),
		Short:   fmt.Sprintf("Show a single %s resource", desc.Name),
		Long:    fmt.Sprintf("Show a single %s resource.", desc.Name),
		Aliases: []string{desc.KumactlArgAlias},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]

			resource := desc.NewObject()
			switch desc.Scope {
			case core_model.ScopeGlobal:
				if err := rs.Get(cmd.Context(), resource, store.GetByKey(name, "")); err != nil {
					if store.IsResourceNotFound(err) {
						return errors.New("No resources found")
					}
					return errors.Wrapf(err, "failed to get %s", name)
				}
			case core_model.ScopeMesh:
				currentMesh := pctx.CurrentMesh()
				if err := rs.Get(cmd.Context(), resource, store.GetByKey(name, currentMesh)); err != nil {
					if store.IsResourceNotFound(err) {
						return errors.Errorf("No resources found in %s mesh", currentMesh)
					}
					return errors.Wrapf(err, "failed to get %s in mesh %s", name, currentMesh)
				}
			default:
				return errors.Errorf("Scope %s is unsupported", desc.Scope)
			}

			format := output.Format(pctx.GetContext.Args.OutputFormat)
			return printers.GenericPrint(format, resource, ResolvePrinter(desc.Name, resource.Descriptor().Scope, pctx.Now()), cmd.OutOrStdout())
		},
	}
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	return cmd
}
