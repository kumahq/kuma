package inspect

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func newInspectZoneIngressesCmd(pctx *cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zoneingresses",
		Short:   "Inspect Zone Ingresses",
		Long:    `Inspect Zone Ingresses.`,
		Aliases: []string{"zone-ingresses"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentZoneIngressOverviewClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a zone ingress client")
			}
			overviews, err := client.List(cmd.Context())
			if err != nil {
				return err
			}

			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, overviews, zoneIngressOverviewTable(pctx.Now()), cmd.OutOrStdout())
		},
	}
	return cmd
}

func zoneIngressOverviewTable(now time.Time) printers.Table {
	return printers.Table{
		Headers: []string{"NAME", "STATUS", "LAST CONNECTED AGO", "LAST UPDATED AGO", "TOTAL UPDATES", "TOTAL ERRORS", "KUMA-DP VERSION", "ENVOY VERSION"},
		RowForItem: func(i int, container interface{}) ([]string, error) {
			zoneIngressOverviews := container.(*mesh.ZoneIngressOverviewResourceList)
			if len(zoneIngressOverviews.Items) <= i {
				return nil, nil
			}
			meta := zoneIngressOverviews.Items[i].Meta
			zoneIngressInsight := zoneIngressOverviews.Items[i].Spec.ZoneIngressInsight

			lastSubscription := zoneIngressInsight.GetLastSubscription().(*mesh_proto.DiscoverySubscription)
			totalResponsesSent := zoneIngressInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
				return s.GetStatus().GetTotal().GetResponsesSent()
			})
			totalResponsesRejected := zoneIngressInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
				return s.GetStatus().GetTotal().GetResponsesRejected()
			})
			onlineStatus := "Offline"
			if zoneIngressInsight.IsOnline() {
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
			}, nil
		},
	}
}
