package get

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
)

func NewGetCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show Kuma resources",
		Long:  `Show Kuma resources.`,
	}
	// flags
	cmd.PersistentFlags().StringVarP(&pctx.GetContext.Args.OutputFormat, "output", "o", string(output.TableFormat), kuma_cmd.UsageOptions("output format", output.TableFormat, output.YAMLFormat, output.JSONFormat))
	// sub-commands
	cmd.AddCommand(WithPaginationArgs(newGetMeshesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetDataplanesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetExternalServicesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetHealthChecksCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetProxyTemplatesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetTrafficPermissionsCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetTrafficRoutesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetTrafficLogsCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetTrafficTracesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetFaultInjectionsCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetCircuitBreakersCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(newGetRetriesCmd(pctx), &pctx.ListContext))
	cmd.AddCommand(newGetSecretsCmd(pctx))
	cmd.AddCommand(WithPaginationArgs(newGetZonesCmd(pctx), &pctx.ListContext))

	cmd.AddCommand(newGetMeshCmd(pctx))
	cmd.AddCommand(newGetDataplaneCmd(pctx))
	cmd.AddCommand(newGetExternalServiceCmd(pctx))
	cmd.AddCommand(newGetHealthCheckCmd(pctx))
	cmd.AddCommand(newGetProxyTemplateCmd(pctx))
	cmd.AddCommand(newGetTrafficLogCmd(pctx))
	cmd.AddCommand(newGetTrafficPermissionCmd(pctx))
	cmd.AddCommand(newGetTrafficRouteCmd(pctx))
	cmd.AddCommand(newGetTrafficTraceCmd(pctx))
	cmd.AddCommand(newGetFaultInjectionCmd(pctx))
	cmd.AddCommand(newGetCircuitBreakerCmd(pctx))
	cmd.AddCommand(newGetRetryCmd(pctx))
	cmd.AddCommand(newGetSecretCmd(pctx))
	cmd.AddCommand(newGetZoneCmd(pctx))
	return cmd
}

func WithPaginationArgs(cmd *cobra.Command, ctx *kumactl_cmd.ListContext) *cobra.Command {
	cmd.PersistentFlags().IntVarP(&ctx.Args.Size, "size", "", 0, "maximum number of elements to return")
	cmd.PersistentFlags().StringVarP(&ctx.Args.Offset, "offset", "", "", "the offset that indicates starting element of the resources list to retrieve")
	return cmd
}
