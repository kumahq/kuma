package get

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewGetResourcesCmd(pctx *kumactl_cmd.RootContext, desc model.ResourceTypeDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   desc.KumactlListArg,
		Short: fmt.Sprintf("Show %s", desc.Name),
		Long:  fmt.Sprintf("Show %s entities.", desc.Name),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			resources := desc.NewList()
			currentMesh := pctx.CurrentMesh()
			if resources.Descriptor().Scope == model.ScopeGlobal {
				currentMesh = ""
			}
			if err := rs.List(context.Background(), resources, core_store.ListByMesh(currentMesh), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrapf(err, "failed to list "+string(desc.Name))
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return ResolvePrinter(desc.Name, resources.Descriptor().Scope).Print(pctx.Now(), resources, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(resources), cmd.OutOrStdout())
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	return cmd
}
