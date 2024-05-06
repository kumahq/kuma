package inspect

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type inspectDataplanesContext struct {
	args struct {
		tags    map[string]string
		gateway bool
		ingress bool
	}
}

func newInspectDataplanesCmd(pctx *cmd.RootContext) *cobra.Command {
	ctx := inspectDataplanesContext{}
	cmd := &cobra.Command{
		Use:   "dataplanes",
		Short: "Inspect Dataplanes",
		Long:  `Inspect Dataplanes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentDataplaneOverviewClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane client")
			}
			overviews, err := client.List(context.Background(), pctx.CurrentMesh(), ctx.args.tags, ctx.args.gateway, ctx.args.ingress)
			if err != nil {
				return err
			}

			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, overviews, dataplaneOverviewsTable(pctx.Now()), cmd.OutOrStdout())
		},
	}
	cmd.PersistentFlags().StringToStringVarP(&ctx.args.tags, "tag", "", map[string]string{}, "filter by tag in format of key=value. You can provide many tags")
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().BoolVarP(&ctx.args.gateway, "gateway", "", false, "filter gateway dataplanes")
	cmd.PersistentFlags().BoolVarP(&ctx.args.ingress, "ingress", "", false, "filter ingress dataplanes")
	return cmd
}

func dataplaneOverviewsTable(now time.Time) printers.Table {
	return printers.Table{
		Headers: []string{
			"MESH",
			"NAME",
			"TAGS",
			"STATUS",
			"LAST CONNECTED AGO",
			"LAST UPDATED AGO",
			"TOTAL UPDATES",
			"TOTAL ERRORS",
			"CERT REGENERATED AGO",
			"CERT EXPIRATION",
			"CERT REGENERATIONS",
			"CERT BACKEND",
			"SUPPORTED CERT BACKENDS",
			"KUMA-DP VERSION",
			"ENVOY VERSION",
			"DEPENDENCIES VERSIONS",
			"NOTES",
		},
		RowForItem: func(i int, container interface{}) ([]string, error) {
			dataplaneOverviews := container.(*core_mesh.DataplaneOverviewResourceList)
			if len(dataplaneOverviews.Items) <= i {
				return nil, nil
			}
			meta := dataplaneOverviews.Items[i].Meta
			dataplane := dataplaneOverviews.Items[i].Spec.Dataplane
			dataplaneInsight := dataplaneOverviews.Items[i].Spec.DataplaneInsight
			dataplaneOverview := dataplaneOverviews.Items[i]

			lastSubscription := dataplaneInsight.GetLastSubscription().(*mesh_proto.DiscoverySubscription)
			totalResponsesSent := dataplaneInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
				return s.GetStatus().GetTotal().GetResponsesSent()
			})
			totalResponsesRejected := dataplaneInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
				return s.GetStatus().GetTotal().GetResponsesRejected()
			})
			status, errs := dataplaneOverview.Status()
			lastConnected := util_proto.MustTimestampFromProto(lastSubscription.GetConnectTime())
			lastUpdated := util_proto.MustTimestampFromProto(lastSubscription.GetStatus().GetLastUpdateTime())

			var certExpiration *time.Time
			if dataplaneInsight.GetMTLS().GetCertificateExpirationTime() != nil {
				certExpiration = util_proto.MustTimestampFromProto(dataplaneInsight.GetMTLS().GetCertificateExpirationTime())
				// don't use time.Local so we don't have to override it in tests. Instead, use location of current clock (that can be overridden in tests)
				inLocation := certExpiration.In(now.Location())
				certExpiration = &inLocation
			}
			var lastCertGeneration *time.Time
			if dataplaneInsight.GetMTLS().GetLastCertificateRegeneration() != nil {
				lastCertGeneration = util_proto.MustTimestampFromProto(dataplaneInsight.GetMTLS().GetLastCertificateRegeneration())
			}
			dataplaneInsight.GetMTLS().GetCertificateExpirationTime()
			certRegenerations := strconv.Itoa(int(dataplaneInsight.GetMTLS().GetCertificateRegenerations()))
			certBackend := dataplaneInsight.GetMTLS().GetIssuedBackend()
			if dataplaneInsight.GetMTLS() == nil {
				certBackend = "-"
			}
			supportedBackend := strings.Join(dataplaneInsight.GetMTLS().GetSupportedBackends(), ",")

			var kumaDpVersion string
			var envoyVersion string
			var dependenciesVersions []string
			if lastSubscription.GetVersion() != nil {
				if lastSubscription.Version.KumaDp != nil {
					kumaDpVersion = lastSubscription.Version.KumaDp.Version
				}
				if lastSubscription.Version.Envoy != nil {
					envoyVersion = lastSubscription.Version.Envoy.Version
				}
				for name, version := range lastSubscription.GetVersion().GetDependencies() {
					dependenciesVersions = append(
						dependenciesVersions,
						fmt.Sprintf("%s: %s", name, version),
					)
				}
			}

			dependenciesVersionsCell := strings.Join(dependenciesVersions, ", ")
			if dependenciesVersionsCell == "" {
				dependenciesVersionsCell = "-"
			}

			return []string{
				meta.GetMesh(),                       // MESH
				meta.GetName(),                       // NAME,
				dataplane.TagSet().String(),          // TAGS
				status.String(),                      // STATUS
				table.Ago(lastConnected, now),        // LAST CONNECTED AGO
				table.Ago(lastUpdated, now),          // LAST UPDATED AGO
				table.Number(totalResponsesSent),     // TOTAL UPDATES
				table.Number(totalResponsesRejected), // TOTAL ERRORS
				table.Ago(lastCertGeneration, now),   // CERT REGENERATED AGO
				table.Date(certExpiration),           // CERT EXPIRATION
				certRegenerations,                    // CERT REGENERATIONS
				certBackend,                          // CERT BACKEND
				supportedBackend,                     // SUPPORTED CERT BACKENDS
				kumaDpVersion,                        // KUMA-DP VERSION
				envoyVersion,                         // ENVOY VERSION
				dependenciesVersionsCell,             // DEPENDENCIES VERSIONS
				strings.Join(errs, ";"),              // NOTES
			}, nil
		},
	}
}
