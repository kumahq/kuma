package get

import (
	"context"
	"io"
	"strings"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetTrafficLogsCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-logs",
		Short: "Show TrafficLogs",
		Long:  `Show TrafficLog entities.`,
		Args: RegexArgs(map[int][]string{
			0: []string{".*\\..*", "<name>.<namespace>"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			var trafficLogs mesh.TrafficLogResourceList

			if len(args) == 1 {
				s := strings.Split(args[0], ".")
				name := s[0]
				namespace := s[1]

				trafficLog := mesh.TrafficLogResource{}
				if err := rs.Get(context.Background(), &trafficLog, core_store.GetByKey(namespace, name, pctx.CurrentMesh())); err != nil {
					return errors.Wrapf(err, "failed to get TrafficLog")
				}

				trafficLogs = mesh.TrafficLogResourceList{
					Items: []*mesh.TrafficLogResource{&trafficLog},
				}

			} else {
				trafficLogs = mesh.TrafficLogResourceList{}
				if err := rs.List(context.Background(), &trafficLogs, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
					return errors.Wrapf(err, "failed to list TrafficLog")
				}
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficLogs(&trafficLogs, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&trafficLogs), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printTrafficLogs(trafficLogs *mesh.TrafficLogResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficLogs.Items) <= i {
					return nil
				}
				trafficLogging := trafficLogs.Items[i]

				return []string{
					trafficLogging.GetMeta().GetMesh(), // MESH
					trafficLogging.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
