package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesListCmd(pctx *rootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known Control Planes",
		Long:  `List known Control Planes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			print := &configControlPlanesListTablePrinter{
				out: tabwriter.NewWriter(cmd.OutOrStdout(), 1, 0, 3, ' ', 0),
			}
			defer print.Flush()
			if err := print.Headers(); err != nil {
				return err
			}
			for _, cp := range pctx.Config().ControlPlanes {
				env := "non-k8s"
				if _, ok := cp.Coordinates.Type.(*config_proto.ControlPlaneCoordinates_Kubernetes_); ok {
					env = "k8s"
				}
				if err := print.Row(cp.Name, env); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}

type configControlPlanesListTablePrinter struct {
	out *tabwriter.Writer
}

func (p *configControlPlanesListTablePrinter) Flush() error {
	return p.out.Flush()
}

func (p *configControlPlanesListTablePrinter) Headers() error {
	return p.Row("NAME", "ENVIRONMENT")
}

func (p *configControlPlanesListTablePrinter) Row(columns ...string) error {
	_, err := fmt.Fprintf(p.out, "%s\n", strings.Join(columns, "\t"))
	return err
}
