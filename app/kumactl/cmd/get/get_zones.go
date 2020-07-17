package get

import (
	"context"
	"io"
	"time"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetZonesCmd(pctx *listContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zones",
		Short: "Show Zones",
		Long:  `Show Zone entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			zones := system.ZoneResourceList{}
			if err := rs.List(context.Background(), &zones, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list Zone")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printZones(pctx.Now(), &zones, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&zones), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printZones(rootTime time.Time, zones *system.ZoneResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"NAME", "INGRESS", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(zones.Items) <= i {
					return nil
				}
				zone := zones.Items[i]

				return []string{
					zone.GetMeta().GetName(), // NAME
					zone.Spec.GetIngress().GetAddress(),
					table.TimeSince(zone.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(zones),
	}
	return printers.NewTablePrinter().Print(data, out)
}
