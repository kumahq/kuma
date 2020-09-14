package install

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/metrics"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type metricsTemplateArgs struct {
	Namespace                 string
	KumaPrometheusSdImage     string
	KumaPrometheusSdVersion   string
	KumaCpAddress             string
	DashboardDataplane        string
	DashboardMesh             string
	DashboardServiceToService string
	DashboardCP               string
}

var DefaultMetricsTemplateArgs = metricsTemplateArgs{
	Namespace:               "kuma-metrics",
	KumaPrometheusSdImage:   "kong-docker-kuma-docker.bintray.io/kuma-prometheus-sd",
	KumaPrometheusSdVersion: kuma_version.Build.Version,
	KumaCpAddress:           "http://kuma-control-plane.kuma-system:5681",
}

func newInstallMetrics() *cobra.Command {
	args := DefaultMetricsTemplateArgs
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)",
		Long:  `Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in a kuma-metrics namespace`,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			args.DashboardDataplane = dashboard.String()

			dashboard, err = data.ReadFile(metrics.Templates, "/grafana/kuma-mesh.json")
			if err != nil {
				return err
			}
			args.DashboardMesh = dashboard.String()

			dashboard, err = data.ReadFile(metrics.Templates, ("/grafana/kuma-service-to-service.json"))
			if err != nil {
				return err
			}
			args.DashboardServiceToService = dashboard.String()

			dashboard, err = data.ReadFile(metrics.Templates, ("/grafana/kuma-cp.json"))
			if err != nil {
				return err
			}
			args.DashboardCP = dashboard.String()

			renderedFiles, err := renderFiles(yamlTemplateFiles, args, simpleTemplateRenderer)
			if err != nil {
				return errors.Wrap(err, "Failed to render template files")
			}

			sortedResources := k8s.SortResourcesByKind(renderedFiles)

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
	return cmd
}
