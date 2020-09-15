package get

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetExternalServiceCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-service NAME",
		Short: "Show a single External-Service resource",
		Long:  `Show a single External-Service resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			externalService := &mesh.ExternalServiceResource{}
			if err := rs.Get(context.Background(), externalService, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get mesh %s", currentMesh)
			}
			externalServices := &mesh.ExternalServiceResourceList{
				Items: []*mesh.ExternalServiceResource{externalService},
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printExternalServices(pctx.Now(), externalServices, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(externalService), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
