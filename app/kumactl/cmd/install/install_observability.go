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

var components = []string{"prometheus", "grafana", "loki", "jaeger"}

func newInstallObservability(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := pctx.InstallObservabilityContext.TemplateArgs
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "Install Observability (Metrics, Logging, Tracing) backend in Kubernetes cluster (Prometheus + Grafana + Loki + Jaeger + Zipkin)",
		Long:  `Install Observability (Metrics, Logging, Tracing) backend in Kubernetes cluster (Prometheus + Grafana + Loki + Jaeger + Zipkin) in its own namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			componentsMap := map[string]bool{}
			for _, component := range args.Components {
				componentsMap[component] = true
			}
			args.ComponentsMap = componentsMap
			metrics, err := getMetrics(&args)
			if err != nil {
				return err
			}
			logging, err := getLogging(&args)
			if err != nil {
				return err
			}
			tracing, err := getTracing(&args)
			if err != nil {
				return err
			}
			combinedResources := make([]data.File, len(metrics)+len(logging)+len(tracing))
			combinedResources = append(combinedResources, metrics...)
			combinedResources = append(combinedResources, logging...)
			combinedResources = append(combinedResources, tracing...)

			singleFile := data.JoinYAML(combinedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install observability to")
	cmd.PersistentFlags().StringVarP(&args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.Flags().StringVar(&args.KumaCpAddress, "kuma-cp-address", args.KumaCpAddress, "the address of Kuma CP")
	cmd.Flags().StringVar(&args.JaegerAddress, "jaeger-address", args.JaegerAddress, "the address of jaeger to query")
	cmd.Flags().StringVar(&args.LokiAddress, "loki-address", args.LokiAddress, "the address of the loki to query")
	cmd.Flags().StringVar(&args.PrometheusAddress, "prometheus-address", args.PrometheusAddress, "the address of the prometheus server")
	cmd.Flags().StringSliceVar(&args.Components, "components", args.Components, "list of components")
	return cmd
}

func getMetrics(args *context.ObservabilityTemplateArgs) ([]data.File, error) {
	installMetricsFS := kumactl_data.InstallMetricsFS()
	templateFiles, err := data.ReadFiles(installMetricsFS)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read template files")
	}
	yamlTemplateFiles := templateFiles.Filter(func(file data.File) bool {
		return strings.HasSuffix(file.Name, ".yaml")
	})

	dashboard, err := data.ReadFile(installMetricsFS, "grafana/kuma-dataplane.json")
	if err != nil {
		return nil, err
	}
	args.Dashboards = append(args.Dashboards, context.Dashboard{
		FileName: "kuma-dataplane.json",
		Content:  dashboard.String(),
	})

	dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-mesh.json")
	if err != nil {
		return nil, err
	}
	args.Dashboards = append(args.Dashboards, context.Dashboard{
		FileName: "kuma-mesh.json",
		Content:  dashboard.String(),
	})

	dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-service-to-service.json")
	if err != nil {
		return nil, err
	}
	args.Dashboards = append(args.Dashboards, context.Dashboard{
		FileName: "kuma-service-to-service.json",
		Content:  dashboard.String(),
	})

	dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-cp.json")
	if err != nil {
		return nil, err
	}
	args.Dashboards = append(args.Dashboards, context.Dashboard{
		FileName: "kuma-cp.json",
		Content:  dashboard.String(),
	})
	dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-gateway.json")
	if err != nil {
		return nil, err
	}
	args.Dashboards = append(args.Dashboards, context.Dashboard{
		FileName: "kuma-gateway.json",
		Content:  dashboard.String(),
	})

	dashboard, err = data.ReadFile(installMetricsFS, "grafana/kuma-service.json")
	if err != nil {
		return nil, err
	}
	args.Dashboards = append(args.Dashboards, context.Dashboard{
		FileName: "kuma-service.json",
		Content:  dashboard.String(),
	})

	filter := getExcludePrefixesFilter(args)
	renderedFiles, err := renderFilesWithFilter(yamlTemplateFiles, args, simpleTemplateRenderer, filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to render template files")
	}

	sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sort resources by kind")
	}
	return sortedResources, nil
}

func getLogging(args *context.ObservabilityTemplateArgs) ([]data.File, error) {
	templateFiles, err := data.ReadFiles(kumactl_data.InstallLoggingFS())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read template files")
	}

	filter := getExcludePrefixesFilter(args)
	renderedFiles, err := renderFilesWithFilter(templateFiles, args, simpleTemplateRenderer, filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to render template files")
	}

	sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sort resources by kind")
	}
	return sortedResources, nil
}

func getTracing(args *context.ObservabilityTemplateArgs) ([]data.File, error) {
	templateFiles, err := data.ReadFiles(kumactl_data.InstallTracingFS())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read template files")
	}

	filter := getExcludePrefixesFilter(args)
	renderedFiles, err := renderFilesWithFilter(templateFiles, args, simpleTemplateRenderer, filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to render template files")
	}

	sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sort resources by kind")
	}
	return sortedResources, nil
}

func getExcludePrefixesFilter(args *context.ObservabilityTemplateArgs) ExcludePrefixesFilter {
	prefixes := []string{}
	for _, key := range components {
		if !args.ComponentsMap[key] {
			prefixes = append(prefixes, key)
		}
	}
	return ExcludePrefixesFilter{
		Prefixes: prefixes,
	}
}
