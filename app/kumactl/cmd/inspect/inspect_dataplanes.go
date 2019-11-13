package inspect

import (
	"context"
	"io"
	"time"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/app/kumactl/pkg/output/table"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type inspectDataplanesContext struct {
	*inspectContext

	tagsArgs struct {
		tags map[string]string
	}
}

func newInspectDataplanesCmd(pctx *inspectContext) *cobra.Command {
	ctx := inspectDataplanesContext{
		inspectContext: pctx,
	}
	cmd := &cobra.Command{
		Use:   "dataplanes",
		Short: "Inspect Dataplanes",
		Long:  `Inspect Dataplanes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentDataplaneOverviewClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane client")
			}
			overviews, err := client.List(context.Background(), pctx.CurrentMesh(), ctx.tagsArgs.tags)
			if err != nil {
				return err
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printDataplaneOverviews(pctx.Now(), overviews, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(overviews), cmd.OutOrStdout())
			}
		},
	}
	cmd.PersistentFlags().StringToStringVarP(&ctx.tagsArgs.tags, "tag", "", map[string]string{}, "filter by tag in format of key=value. You can provide many tags")
	return cmd
}

func printDataplaneOverviews(now time.Time, dataplaneInsights *mesh_core.DataplaneOverviewResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "TAGS", "STATUS", "LAST CONNECTED AGO", "LAST UPDATED AGO", "TOTAL UPDATES", "TOTAL ERRORS"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(dataplaneInsights.Items) <= i {
					return nil
				}
				meta := dataplaneInsights.Items[i].Meta
				dataplane := dataplaneInsights.Items[i].Spec.Dataplane
				dataplaneInsight := dataplaneInsights.Items[i].Spec.DataplaneInsight

				lastSubscription, lastConnected := dataplaneInsight.GetLatestSubscription()
				totalResponsesSent := dataplaneInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.GetStatus().GetTotal().GetResponsesSent()
				})
				totalResponsesRejected := dataplaneInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.GetStatus().GetTotal().GetResponsesRejected()
				})
				onlineStatus := "Offline"
				if dataplaneInsight.IsOnline() {
					onlineStatus = "Online"
				}
				lastUpdated := util_proto.MustTimestampFromProto(lastSubscription.GetStatus().GetLastUpdateTime())

				return []string{
					meta.GetMesh(),                       // MESH
					meta.GetName(),                       // NAME,
					dataplane.Tags().String(),            // TAGS
					onlineStatus,                         // STATUS
					table.Ago(lastConnected, now),        // LAST CONNECTED AGO
					table.Ago(lastUpdated, now),          // LAST UPDATED AGO
					table.Number(totalResponsesSent),     // TOTAL UPDATES
					table.Number(totalResponsesRejected), // TOTAL ERRORS
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
