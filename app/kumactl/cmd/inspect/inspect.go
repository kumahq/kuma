package inspect

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func NewInspectCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect Kuma resources",
		Long:  `Inspect Kuma resources.`,
	}
	inspectCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := kumactl_cmd.RunParentPreRunE(inspectCmd, args); err != nil {
			return err
		}
		if err := pctx.CheckServerVersionCompatibility(); err != nil {
			cmd.PrintErrln(err)
		}
		return nil
	}
	// flags
	inspectCmd.PersistentFlags().StringVarP(&pctx.InspectContext.Args.OutputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	inspectCmd.AddCommand(newInspectDataplanesCmd(pctx))
	inspectCmd.AddCommand(newInspectDataplaneCmd(pctx))
	inspectCmd.AddCommand(newInspectZoneIngressesCmd(pctx))
	inspectCmd.AddCommand(newInspectZoneIngressCmd(pctx))
	inspectCmd.AddCommand(newInspectZoneEgressesCmd(pctx))
	inspectCmd.AddCommand(newInspectZoneEgressCmd(pctx))
	inspectCmd.AddCommand(newInspectZonesCmd(pctx))
	inspectCmd.AddCommand(newInspectMeshesCmd(pctx))
	inspectCmd.AddCommand(newInspectServicesCmd(pctx))

	for _, desc := range registry.Global().ObjectDescriptors(core_model.AllowedToInspect()) {
		inspectCmd.AddCommand(newInspectPolicyCmd(desc, pctx))
	}
	return inspectCmd
}
