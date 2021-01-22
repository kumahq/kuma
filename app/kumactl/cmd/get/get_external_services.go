package get

import (
	"context"
	"io"
	"time"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetExternalServicesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-services",
		Short: "Show ExternalServices",
		Long:  `Show ExternalServices entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			externalServices := mesh.ExternalServiceResourceList{}
			if err := rs.List(context.Background(), &externalServices, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrapf(err, "failed to list ExternalServices")
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printExternalServices(pctx.Now(), &externalServices, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&externalServices), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printExternalServices(rootTime time.Time, externalServices *mesh.ExternalServiceResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "TAGS", "ADDRESS", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(externalServices.Items) <= i {
					return nil
				}
				externalService := externalServices.Items[i]

				return []string{
					externalService.Meta.GetMesh(),                                        // MESH
					externalService.Meta.GetName(),                                        // NAME,
					externalService.Spec.TagSet().String(),                                // TAGS
					externalService.Spec.Networking.Address,                               // ADDRESS
					table.TimeSince(externalService.Meta.GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(externalServices),
	}
	return printers.NewTablePrinter().Print(data, out)
}
