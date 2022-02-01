package inspect

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func newInspectZoneEgressesCmd(ctx *cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zoneegresses",
		Short: "Inspect Zone Egresses",
		Long:  `Inspect Zone Egresses.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := ctx.CurrentZoneEgressOverviewClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a zone egress client")
			}
			overviews, err := client.List(context.Background())
			if err != nil {
				return err
			}

			switch format := output.Format(ctx.InspectContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printZoneEgressOverviews(ctx.Now(), overviews, cmd.OutOrStdout())
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

func printZoneEgressOverviews(now time.Time, zoneEgressOverviews *mesh.ZoneEgressOverviewResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"NAME", "STATUS", "LAST CONNECTED AGO", "LAST UPDATED AGO", "TOTAL UPDATES", "TOTAL ERRORS", "KUMA-DP VERSION", "ENVOY VERSION"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(zoneEgressOverviews.Items) <= i {
					return nil
				}
				meta := zoneEgressOverviews.Items[i].Meta
				zoneEgressInsight := zoneEgressOverviews.Items[i].Spec.ZoneEgressInsight

				lastSubscription := zoneEgressInsight.GetLastSubscription().(*mesh_proto.DiscoverySubscription)
				totalResponsesSent := zoneEgressInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.GetStatus().GetTotal().GetResponsesSent()
				})
				totalResponsesRejected := zoneEgressInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.GetStatus().GetTotal().GetResponsesRejected()
				})
				onlineStatus := "Offline"
				if zoneEgressInsight.IsOnline() {
					onlineStatus = "Online"
				}
				lastConnected := util_proto.MustTimestampFromProto(lastSubscription.GetConnectTime())
				lastUpdated := util_proto.MustTimestampFromProto(lastSubscription.GetStatus().GetLastUpdateTime())

				var kumaDpVersion string
				var envoyVersion string
				if lastSubscription.GetVersion() != nil {
					if lastSubscription.Version.KumaDp != nil {
						kumaDpVersion = lastSubscription.Version.KumaDp.Version
					}
					if lastSubscription.Version.Envoy != nil {
						envoyVersion = lastSubscription.Version.Envoy.Version
					}
				}

				return []string{
					meta.GetName(),                       // NAME,
					onlineStatus,                         // STATUS
					table.Ago(lastConnected, now),        // LAST CONNECTED AGO
					table.Ago(lastUpdated, now),          // LAST UPDATED AGO
					table.Number(totalResponsesSent),     // TOTAL UPDATES
					table.Number(totalResponsesRejected), // TOTAL ERRORS
					kumaDpVersion,                        // KUMA-DP VERSION
					envoyVersion,                         // ENVOY VERSION
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
