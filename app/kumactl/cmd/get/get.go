package get

import (
	"github.com/spf13/cobra"

	get_context "github.com/kumahq/kuma/app/kumactl/cmd/get/context"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
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
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "meshes", core_mesh.MeshType, printMeshes), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "dataplanes", core_mesh.DataplaneType, printDataplanes), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "external-services", core_mesh.ExternalServiceType, printExternalServices), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "healthchecks", core_mesh.HealthCheckType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "proxytemplates", core_mesh.ProxyTemplateType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "rate-limits", core_mesh.RateLimitType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "traffic-permissions", core_mesh.TrafficPermissionType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "traffic-routes", core_mesh.TrafficRouteType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "traffic-logs", core_mesh.TrafficLogType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "traffic-traces", core_mesh.TrafficTraceType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "fault-injections", core_mesh.FaultInjectionType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "circuit-breakers", core_mesh.CircuitBreakerType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "retries", core_mesh.RetryType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "timeouts", core_mesh.TimeoutType, BasicResourceTablePrinter), &pctx.ListContext))
	cmd.AddCommand(NewGetResourcesCmd(pctx, "secrets", core_system.SecretType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourcesCmd(pctx, "global-secrets", core_system.GlobalSecretType, BasicGlobalResourceTablePrinter))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "zones", core_system.ZoneType, printZones), &pctx.ListContext))
	cmd.AddCommand(WithPaginationArgs(NewGetResourcesCmd(pctx, "zone-ingresses", core_mesh.ZoneIngressType, BasicGlobalResourceTablePrinter), &pctx.ListContext))

	cmd.AddCommand(NewGetResourceCmd(pctx, "mesh", core_mesh.MeshType, printMeshes))
	cmd.AddCommand(NewGetResourceCmd(pctx, "dataplane", core_mesh.DataplaneType, printDataplanes))
	cmd.AddCommand(NewGetResourceCmd(pctx, "external-service", core_mesh.ExternalServiceType, printExternalServices))
	cmd.AddCommand(NewGetResourceCmd(pctx, "healthcheck", core_mesh.HealthCheckType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "proxytemplate", core_mesh.ProxyTemplateType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "rate-limit", core_mesh.RateLimitType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "traffic-permission", core_mesh.TrafficPermissionType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "traffic-route", core_mesh.TrafficRouteType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "traffic-log", core_mesh.TrafficLogType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "traffic-trace", core_mesh.TrafficTraceType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "fault-injection", core_mesh.FaultInjectionType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "circuit-breaker", core_mesh.CircuitBreakerType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "retry", core_mesh.RetryType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "timeout", core_mesh.TimeoutType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "secret", core_system.SecretType, BasicResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "global-secret", core_system.GlobalSecretType, BasicGlobalResourceTablePrinter))
	cmd.AddCommand(NewGetResourceCmd(pctx, "zone", core_system.ZoneType, printZones))
	cmd.AddCommand(NewGetResourceCmd(pctx, "zone-ingress", core_mesh.ZoneIngressType, BasicGlobalResourceTablePrinter))
	return cmd
}

func WithPaginationArgs(cmd *cobra.Command, ctx *get_context.ListContext) *cobra.Command {
	cmd.PersistentFlags().IntVarP(&ctx.Args.Size, "size", "", 0, "maximum number of elements to return")
	cmd.PersistentFlags().StringVarP(&ctx.Args.Offset, "offset", "", "", "the offset that indicates starting element of the resources list to retrieve")
	return cmd
}
