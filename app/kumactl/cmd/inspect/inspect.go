package inspect

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
)

func NewInspectCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect Kuma resources",
		Long:  `Inspect Kuma resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&pctx.InspectContext.Args.OutputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newInspectDataplanesCmd(pctx))
	cmd.AddCommand(newInspectZoneIngressesCmd(pctx))
	cmd.AddCommand(newInspectZonesCmd(pctx))
	cmd.AddCommand(newInspectMeshesCmd(pctx))
	cmd.AddCommand(newInspectServicesCmd(pctx))
	return cmd
}
