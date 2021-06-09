package inspect

import (
	"context"
	"io"
	"time"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func newInspectZonesCmd(ctx *cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zones",
		Short: "Inspect Zones",
		Long:  `Inspect Zones.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := ctx.CurrentZoneOverviewClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a zone client")
			}
			overviews, err := client.List(context.Background())
			if err != nil {
				return err
			}

			switch format := output.Format(ctx.InspectContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printZoneOverviews(ctx.Now(), overviews, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(overviews), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printZoneOverviews(now time.Time, zoneOverviews *system.ZoneOverviewResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"NAME", "STATUS", "LAST CONNECTED AGO", "LAST UPDATED AGO", "TOTAL UPDATES", "TOTAL ERRORS", "ZONE-CP VERSION"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(zoneOverviews.Items) <= i {
					return nil
				}
				meta := zoneOverviews.Items[i].Meta
				zone := zoneOverviews.Items[i].Spec.Zone
				zoneInsight := zoneOverviews.Items[i].Spec.ZoneInsight

				lastSubscription, lastConnected := zoneInsight.GetLatestSubscription()
				totalResponsesSent := zoneInsight.Sum(func(s *system_proto.KDSSubscription) uint64 {
					return s.GetStatus().GetTotal().GetResponsesSent()
				})
				totalResponsesRejected := zoneInsight.Sum(func(s *system_proto.KDSSubscription) uint64 {
					return s.GetStatus().GetTotal().GetResponsesRejected()
				})
				onlineStatus := "Offline"
				if zoneInsight.IsOnline() && zone.IsEnabled() {
					onlineStatus = "Online"
				}
				lastUpdated := util_proto.MustTimestampFromProto(lastSubscription.GetStatus().GetLastUpdateTime())

				var zoneCPVersion string
				if lastSubscription.GetVersion() != nil {
					if lastSubscription.Version.KumaCp != nil {
						zoneCPVersion = lastSubscription.Version.KumaCp.Version
					}
				}

				return []string{
					meta.GetName(),                       // NAME,
					onlineStatus,                         // STATUS
					table.Ago(lastConnected, now),        // LAST CONNECTED AGO
					table.Ago(lastUpdated, now),          // LAST UPDATED AGO
					table.Number(totalResponsesSent),     // TOTAL UPDATES
					table.Number(totalResponsesRejected), // TOTAL ERRORS
					zoneCPVersion,                        // ZONE-CP VERSION
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
