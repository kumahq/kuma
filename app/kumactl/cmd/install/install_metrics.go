// Deprecated: `kumactl install metrics` is deprecated, use `kumactl install observability` instead

package install

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	kumactl_data "github.com/kumahq/kuma/app/kumactl/data"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

func newInstallMetrics(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := pctx.InstallMetricsContext.TemplateArgs

	cmd := &cobra.Command{
		Use:        "metrics",
		Short:      "Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)",
		Long:       `Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in its own namespace.`,
		Deprecated: "We're migrating to `observability`, please use `install observability`",
		RunE: func(cmd *cobra.Command, _ []string) error {
			installMetricsFS := kumactl_data.InstallDeprecatedMetricsFS()
			templateFiles, err := data.ReadFiles(installMetricsFS)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}
			yamlTemplateFiles := templateFiles.Filter(func(file data.File) bool {
				return strings.HasSuffix(file.Name, ".yaml")
			})

			dashboard, err := data.ReadFile(installMetricsFS, "grafana/kuma-dataplane.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-dataplane.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-mesh.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-mesh.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-service-to-service.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-service-to-service.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-cp.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-cp.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-service.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-service.json",
				Content:  dashboard.String(),
			})

			filter := getMetricsExcludePrefixesFilter(args.WithoutPrometheus, args.WithoutGrafana)

			renderedFiles, err := renderFilesWithFilter(yamlTemplateFiles, args, simpleTemplateRenderer, filter)
			if err != nil {
				return errors.Wrap(err, "Failed to render template files")
			}

			sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
			if err != nil {
				return errors.Wrap(err, "Failed to sort resources by kind")
			}

			singleFile := data.JoinYAML(sortedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install metrics to")
	cmd.PersistentFlags().StringVarP(&args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.Flags().StringVar(&args.KumaCpAddress, "kuma-cp-address", args.KumaCpAddress, "the address of Kuma CP")
	cmd.Flags().StringVar(&args.JaegerAddress, "jaeger-address", args.JaegerAddress, "the address of jaeger to query")
	cmd.Flags().StringVar(&args.LokiAddress, "loki-address", args.LokiAddress, "the address of the loki to query")
	cmd.Flags().BoolVar(&args.WithoutPrometheus, "without-prometheus", args.WithoutPrometheus, "disable Prometheus resources generation")
	cmd.Flags().BoolVar(&args.WithoutGrafana, "without-grafana", args.WithoutGrafana, "disable Grafana resources generation")
	return cmd
}

func getMetricsExcludePrefixesFilter(withoutPrometheus, withoutGrafana bool) ExcludePrefixesFilter {
	prefixes := []string{}

	if withoutPrometheus {
		prefixes = append(prefixes, "prometheus")
	}

	if withoutGrafana {
		prefixes = append(prefixes, "grafana")
	}

	return ExcludePrefixesFilter{
		Prefixes: prefixes,
	}
}
