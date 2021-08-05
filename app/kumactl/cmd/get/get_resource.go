package get

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewGetResourceCmd(pctx *kumactl_cmd.RootContext, use string, resourceType core_model.ResourceType) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s NAME", use),
		Short: fmt.Sprintf("Show a single %s resource", resourceType),
		Long:  fmt.Sprintf("Show a single %s resource.", resourceType),
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]

			resource, err := registry.Global().NewObject(resourceType)
			if err != nil {
				return err
			}

			if resource.Scope() == core_model.ScopeGlobal {
				if err := rs.Get(context.Background(), resource, store.GetByKey(name, "")); err != nil {
					if store.IsResourceNotFound(err) {
						return errors.New("No resources found")
					}
					return errors.Wrapf(err, "failed to get %s", name)
				}
			} else {
				currentMesh := pctx.CurrentMesh()
				if err := rs.Get(context.Background(), resource, store.GetByKey(name, currentMesh)); err != nil {
					if store.IsResourceNotFound(err) {
						return errors.Errorf("No resources found in %s mesh", currentMesh)
					}
					return errors.Wrapf(err, "failed to get %s in mesh %s", name, currentMesh)
				}
			}

			resources, err := registry.Global().NewList(resourceType)
			if err != nil {
				return err
			}
			if err := resources.AddItem(resource); err != nil {
				return err
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return ResolvePrinter(resourceType, resource.Scope()).Print(pctx.Now(), resources, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(resource), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
