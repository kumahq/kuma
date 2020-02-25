package install

import (
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s/metrics"
	kuma_version "github.com/Kong/kuma/pkg/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallMetrics() *cobra.Command {
	args := struct {
		Namespace               string
		KumaPrometheusSdImage   string
		KumaPrometheusSdVersion string
		KumaCpAddress           string
	}{
		Namespace:               "kuma-metrics",
		KumaPrometheusSdImage:   "kong-docker-kuma-docker.bintray.io/kuma-prometheus-sd",
		KumaPrometheusSdVersion: kuma_version.Build.Version,
		KumaCpAddress:           "http://kuma-control-plane.kuma-system:5681",
	}
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Install Metrics backend in Kubernetes cluster",
		Long:  `Install Metrics backend in Kubernetes cluster.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateFiles, err := data.ReadFiles(metrics.Templates)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

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
