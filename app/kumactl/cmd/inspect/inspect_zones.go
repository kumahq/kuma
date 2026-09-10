package inspect

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v3/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/table"
	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"
)

func newInspectZonesCmd(pctx *cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "zones",
		Short: "Inspect Zones",
		Long:  `Inspect Zones.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentZoneOverviewClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a zone client")
			}
			overviews, err := client.List(cmd.Context())
			if err != nil {
				return err
			}

			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, overviews, zoneOverviewTable(pctx.Now()), cmd.OutOrStdout())
		},
	}
}

func zoneOverviewTable(now time.Time) printers.Table {
	return printers.Table{
		Headers: []string{"NAME", "STATUS", "LAST CONNECTED AGO", "LAST UPDATED AGO", "TOTAL UPDATES", "TOTAL ERRORS", "ZONE-CP VERSION", "BACKEND"},
		RowForItem: func(i int, container any) ([]string, error) {
			zoneOverviews := container.(*zone_api.ZoneOverviewResourceList)
			if len(zoneOverviews.Items) <= i {
				return nil, nil
			}
			meta := zoneOverviews.Items[i].Meta
			zone := zoneOverviews.Items[i].Spec.Zone
			zoneInsight := zoneOverviews.Items[i].Spec.ZoneInsight

			lastSubscription := zoneInsight.GetLastSubscription().(*zoneinsight_api.KDSSubscription)
			totalResponsesSent := zoneInsight.Sum(func(s *zoneinsight_api.KDSSubscription) uint64 {
				return s.GetStatus().GetTotal().GetResponsesSent()
			})
			totalResponsesRejected := zoneInsight.Sum(func(s *zoneinsight_api.KDSSubscription) uint64 {
				return s.GetStatus().GetTotal().GetResponsesRejected()
			})
			onlineStatus := "Offline"
			if zoneInsight.IsOnline() && zone.IsEnabled() {
				onlineStatus = "Online"
			}
			lastConnected := lastSubscription.GetConnectTime().AsTime()
			lastUpdated := lastSubscription.GetStatus().GetLastUpdateTime().AsTime()

			var zoneCPVersion string
			if lastSubscription.GetVersion() != nil {
				if lastSubscription.Version.KumaCP != nil {
					zoneCPVersion = lastSubscription.Version.KumaCP.Version
				}
			}

			var backend string
			if lastSubscription.GetConfig() != "" {
				// Unmarshall only what we need to avoid dependency on the whole config
				cfg := struct {
					Store struct {
						Type string `json:"type"`
					} `json:"store"`
				}{}
				if err := json.Unmarshal([]byte(lastSubscription.GetConfig()), &cfg); err != nil {
					return nil, errors.Wrap(err, "could not unmarshal CP config")
				} else {
					backend = cfg.Store.Type
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
				backend,                              // BACKEND
			}, nil
		},
	}
}
