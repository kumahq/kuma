package get

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/printers"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/table"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type getDataplanesContext struct {
	*getContext

	tagsArgs struct {
		tags []string
	}
}

func (g *getDataplanesContext) tags() (map[string]string, error) {
	tags := map[string]string{}
	for _, tag := range g.tagsArgs.tags {
		split := strings.Split(tag, "=")
		if len(split) != 2 {
			return nil, errors.Errorf("invalid format of tag %s, it should be key=value", tag)
		}
		tags[split[0]] = split[1]
	}
	return tags, nil
}

func newGetDataplanesCmd(pctx *getContext) *cobra.Command {
	ctx := getDataplanesContext{
		getContext: pctx,
	}
	cmd := &cobra.Command{
		Use:   "dataplanes",
		Short: "Show running Dataplanes",
		Long:  `Show running Dataplanes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.NewDataplaneInspectionClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane client")
			}
			tags, err := ctx.tags()
			if err != nil {
				return err
			}
			inspections, err := client.List(context.Background(), pctx.CurrentMesh(), tags)
			if err != nil {
				return err
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printDataplaneInspections(pctx.Now(), inspections, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(inspections), cmd.OutOrStdout())
			}
		},
	}
	cmd.PersistentFlags().StringSliceVarP(&ctx.tagsArgs.tags, "tag", "", []string{}, "filter by tag in format of key=value. You can provide many tags")
	return cmd
}

func printDataplaneInspections(now time.Time, dataplaneInsights *mesh_core.DataplaneInspectionResourceList, out io.Writer) error {
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
					return s.Status.Total.ResponsesSent
				})
				totalResponsesRejected := dataplaneInsight.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.Status.Total.ResponsesRejected
				})
				onlineStatus := "Offline"
				if dataplaneInsight.IsOnline() {
					onlineStatus = "Online"
				}
				lastUpdated := util_proto.MustTimestampFromProto(lastSubscription.GetStatus().LastUpdateTime)

				tags := map[string]string{}
				for _, inbound := range dataplane.Networking.Inbound {
					for tag, value := range inbound.Tags {
						tags[tag] = value
					}
				}
				var tagsString []string
				for tag, value := range tags {
					tagsString = append(tagsString, fmt.Sprintf("%s=%s", tag, value))
				}
				sort.Strings(tagsString)
				return []string{
					meta.GetMesh(),                       // MESH
					meta.GetName(),                       // NAME,
					strings.Join(tagsString, ", "),       // TAGS
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
