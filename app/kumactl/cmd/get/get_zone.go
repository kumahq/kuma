package get

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetZoneCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zone NAME",
		Short: "Show a single Zone resource",
		Long:  `Show a single Zone resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			zone := system.NewZoneResource()
			if err := rs.Get(context.Background(), zone, store.GetByKey(name, model.NoMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No zone resources found")
				}
				return errors.Wrapf(err, "failed to get zone %s", name)
			}
			zones := &system.ZoneResourceList{
				Items: []*system.ZoneResource{zone},
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printZones(pctx.Now(), zones, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(zone), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
