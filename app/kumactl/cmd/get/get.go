package get

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/Kong/kuma/pkg/cmd"
)

type getContext struct {
	*kumactl_cmd.RootContext

	args struct {
		outputFormat string
	}
}

func NewGetCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &getContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show Kuma resources",
		Long:  `Show Kuma resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&ctx.args.outputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(newGetMeshesCmd(ctx))
	cmd.AddCommand(newGetDataplanesCmd(ctx))
	cmd.AddCommand(newGetHealthChecksCmd(ctx))
	cmd.AddCommand(newGetProxyTemplatesCmd(ctx))
	cmd.AddCommand(newGetTrafficPermissionsCmd(ctx))
	cmd.AddCommand(newGetTrafficRoutesCmd(ctx))
	cmd.AddCommand(newGetTrafficLogsCmd(ctx))
	cmd.AddCommand(newGetTrafficTracesCmd(ctx))
	cmd.AddCommand(newGetFaultInjectionsCmd(ctx))
	cmd.AddCommand(newGetFaultInjectionCmd(ctx))
	cmd.AddCommand(newGetMeshCmd(ctx))
	cmd.AddCommand(newGetDataplaneCmd(ctx))
	cmd.AddCommand(newGetHealthCheckCmd(ctx))
	cmd.AddCommand(newGetProxyTemplateCmd(ctx))
	return cmd
}
