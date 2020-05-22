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

type listContext struct {
	*getContext
	args struct {
		size   int
		offset string
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
	listCtx := &listContext{getContext: ctx}
	cmd.AddCommand(withPaginationArgs(newGetMeshesCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetDataplanesCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetHealthChecksCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetProxyTemplatesCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetTrafficPermissionsCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetTrafficRoutesCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetTrafficLogsCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetTrafficTracesCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetFaultInjectionsCmd(listCtx), listCtx))
	cmd.AddCommand(withPaginationArgs(newGetCircuitBreakersCmd(listCtx), listCtx))
	cmd.AddCommand(newGetSecretsCmd(ctx))

	cmd.AddCommand(newGetMeshCmd(ctx))
	cmd.AddCommand(newGetDataplaneCmd(ctx))
	cmd.AddCommand(newGetHealthCheckCmd(ctx))
	cmd.AddCommand(newGetProxyTemplateCmd(ctx))
	cmd.AddCommand(newGetTrafficLogCmd(ctx))
	cmd.AddCommand(newGetTrafficPermissionCmd(ctx))
	cmd.AddCommand(newGetTrafficRouteCmd(ctx))
	cmd.AddCommand(newGetTrafficTraceCmd(ctx))
	cmd.AddCommand(newGetFaultInjectionCmd(ctx))
	cmd.AddCommand(newGetCircuitBreakerCmd(ctx))
	cmd.AddCommand(newGetSecretCmd(ctx))
	return cmd
}

func withPaginationArgs(cmd *cobra.Command, ctx *listContext) *cobra.Command {
	cmd.PersistentFlags().IntVarP(&ctx.args.size, "size", "", 0, "maximum number of elements to return")
	cmd.PersistentFlags().StringVarP(&ctx.args.offset, "offset", "", "", "the offset that indicates starting element of the resources list to retrieve")
	return cmd
}
