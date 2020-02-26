package install

import (
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s/metrics"
	kuma_version "github.com/Kong/kuma/pkg/version"
)

func newInstallMetrics() *cobra.Command {
	args := struct {
		Namespace                 string
		KumaPrometheusSdImage     string
		KumaPrometheusSdVersion   string
		KumaCpAddress             string
		DashboardDataplane        string
		DashboardMesh             string
		DashboardServiceToService string
	}{
		Namespace:               "kuma-metrics",
		KumaPrometheusSdImage:   "kong-docker-kuma-docker.bintray.io/kuma-prometheus-sd",
		KumaPrometheusSdVersion: kuma_version.Build.Version,
		KumaCpAddress:           "http://kuma-control-plane.kuma-system:5681",
	}
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Install Metrics backend in Kubernetes cluster",
		Long:  `Install Metrics backend (Prometheus and Grafana) in Kubernetes cluster.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateFiles, err := data.ReadFiles(metrics.Templates)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			dashboard, err := renderDashboard("/grafana/kuma-dataplane.json")
			if err != nil {
				return err
			}
			args.DashboardDataplane = dashboard

			dashboard, err = renderDashboard("/grafana/kuma-mesh.json")
			if err != nil {
				return err
			}
			args.DashboardMesh = dashboard

			dashboard, err = renderDashboard("/grafana/kuma-service-to-service.json")
			if err != nil {
				return err
			}
			args.DashboardServiceToService = dashboard

			renderedFiles, err := renderFiles(templateFiles, args, simpleTemplateRenderer)
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

func renderDashboard(fileName string) (string, error) {
	file, err := data.ReadFile(metrics.Templates, fileName)
	if err != nil {
		return "", err
	}
	// Stored dashboards are prepared for upload to online Grafana repo.
	// We need to replace placeholders with provisioned datasource to use it as provisioned dashboard
	dashboard := strings.ReplaceAll(file.String(), "${DS_PROMETHEUS}", "Prometheus")
	return base64.StdEncoding.EncodeToString([]byte(dashboard)), nil
}
