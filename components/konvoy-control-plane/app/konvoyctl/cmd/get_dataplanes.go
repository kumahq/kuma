package cmd

import (
	"context"
	"io"
	"time"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/printers"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/table"
	konvoyctl_resources "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/resources"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetDataplanesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataplanes",
		Short: "Show running Dataplanes",
		Long:  `Show running Dataplanes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			controlPlane, err := pctx.CurrentControlPlane()
			if err != nil {
				return err
			}
			rs, err := konvoyctl_resources.NewResourceStore(controlPlane)
			if err != nil {
				return errors.Wrapf(err, "Failed to create a client for a given Control Plane: %s", controlPlane)
			}

			dataplaneStatuses := &mesh_core.DataplaneStatusResourceList{}
			if err := rs.List(context.Background(), dataplaneStatuses); err != nil {
				return errors.Wrapf(err, "Failed to list Dataplanes")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return (&dataplaneStatusesTablePrinter{}).Print(dataplaneStatuses, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(dataplaneStatuses), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

type dataplaneStatusesTablePrinter struct {
}

func (p *dataplaneStatusesTablePrinter) Print(dataplaneStatuses *mesh_core.DataplaneStatusResourceList, out io.Writer) error {
	now := time.Now()
	data := printers.Table{
		Headers: []string{"MESH", "NAMESPACE", "NAME", "SUBSCRIPTIONS", "LAST CONNECTED AGO", "TOTAL UPDATES", "TOTAL ERRORS"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(dataplaneStatuses.Items) <= i {
					return nil
				}
				dataplaneStatus := dataplaneStatuses.Items[i]

				totalSubscriptions := len(dataplaneStatus.Spec.Subscriptions)
				_, lastConnected := dataplaneStatus.Spec.GetLatestSubscription()
				totalResponsesSent := dataplaneStatus.Spec.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.Status.Total.ResponsesSent
				})
				totalResponsesRejected := dataplaneStatus.Spec.Sum(func(s *mesh_proto.DiscoverySubscription) uint64 {
					return s.Status.Total.ResponsesRejected
				})
				return []string{
					dataplaneStatus.Meta.GetMesh(),       // MESH
					dataplaneStatus.Meta.GetNamespace(),  // NAMESPACE
					dataplaneStatus.Meta.GetName(),       // NAME
					table.Number(totalSubscriptions),     // SUBSCRIPTIONS
					table.Ago(lastConnected, now),        // LAST CONNECTED AGO
					table.Number(totalResponsesSent),     // TOTAL UPDATES
					table.Number(totalResponsesRejected), // TOTAL ERRORS
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
