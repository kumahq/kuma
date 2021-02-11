package install

import (
	"strings"

	"github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/metrics"
)

func newInstallMetrics(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := pctx.InstallMetricsContext.TemplateArgs
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)",
		Long:  `Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in a kuma-metrics namespace`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			args.Mesh = pctx.Args.Mesh

			templateFiles, err := data.ReadFiles(metrics.Templates)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}
			yamlTemplateFiles := templateFiles.Filter(func(file data.File) bool {
				return strings.HasSuffix(file.Name, ".yaml")
			})

			dashboard, err := data.ReadFile(metrics.Templates, "/grafana/kuma-dataplane.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-dataplane.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(metrics.Templates, "/grafana/kuma-mesh.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-mesh.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(metrics.Templates, "/grafana/kuma-service-to-service.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-service-to-service.json",
				Content:  dashboard.String(),
			})

			dashboard, err = data.ReadFile(metrics.Templates, "/grafana/kuma-cp.json")
			if err != nil {
				return err
			}
			args.Dashboards = append(args.Dashboards, context.Dashboard{
				FileName: "kuma-cp.json",
				Content:  dashboard.String(),
			})

			filter := getExcludePrefixesFilter(args.WithoutPrometheus, args.WithoutGrafana)

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
	cmd.Flags().StringVar(&args.KumaPrometheusSdImage, "kuma-prometheus-sd-image", args.KumaPrometheusSdImage, "image name of Kuma Prometheus SD")
	cmd.Flags().StringVar(&args.KumaPrometheusSdVersion, "kuma-prometheus-sd-version", args.KumaPrometheusSdVersion, "version of Kuma Prometheus SD")
	cmd.Flags().StringVar(&args.KumaCpAddress, "kuma-cp-address", args.KumaCpAddress, "the address of Kuma CP")
	cmd.Flags().BoolVar(&args.WithoutPrometheus, "without-prometheus", args.WithoutPrometheus, "disable Prometheus resources generation")
	cmd.Flags().BoolVar(&args.WithoutGrafana, "without-grafana", args.WithoutGrafana, "disable Grafana resources generation")
	return cmd
}

func getExcludePrefixesFilter(withoutPrometheus, withoutGrafana bool) ExcludePrefixesFilter {
	prefixes := []string{}

	if withoutPrometheus {
		prefixes = append(prefixes, "/prometheus")
	}

	if withoutGrafana {
		prefixes = append(prefixes, "/grafana")
	}

	return ExcludePrefixesFilter{
		Prefixes: prefixes,
	}
}
