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
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewGetResourcesCmd(pctx *kumactl_cmd.RootContext, use string, resourceType model.ResourceType) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Show %s", resourceType),
		Long:  fmt.Sprintf("Show %s entities.", resourceType),
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			resources, err := registry.Global().NewList(resourceType)
			if err != nil {
				return err
			}
			currentMesh := pctx.CurrentMesh()
			resource := resources.NewItem()
			if resource.Scope() == model.ScopeGlobal {
				currentMesh = ""
			}
			if err := rs.List(context.Background(), resources, core_store.ListByMesh(currentMesh), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrapf(err, "failed to list "+string(resourceType))
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return ResolvePrinter(resourceType, resource.Scope()).Print(pctx.Now(), resources, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(resources), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
