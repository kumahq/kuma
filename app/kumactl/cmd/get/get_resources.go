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
			resource := resources.NewItem()
			if resource.Descriptor().Scope == model.ScopeGlobal {
				currentMesh = ""
			}
			if err := rs.List(context.Background(), resources, core_store.ListByMesh(currentMesh), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrap(err, "failed to list "+string(desc.Name))
			}

			format := output.Format(pctx.GetContext.Args.OutputFormat)
			return printers.GenericPrint(format, resources, ResolvePrinter(desc.Name, resource.Descriptor().Scope, pctx.Now()), cmd.OutOrStdout())
		},
	}
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	return cmd
}
