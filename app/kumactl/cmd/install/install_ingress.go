package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s/ingress"
	kuma_version "github.com/Kong/kuma/pkg/version"
)

func newInstallIngressCmd() *cobra.Command {
	args := struct {
		Namespace       string
		Image           string
		Version         string
		ImagePullPolicy string
		Mesh            string
		DrainTime       string
		KumaCpAddress   string
		IngressPortType string
	}{
		Namespace:       "kuma-system",
		Image:           "kong-docker-kuma-docker.bintray.io/kuma-dp",
		Version:         kuma_version.Build.Version,
		ImagePullPolicy: "IfNotPresent",
		Mesh:            "default",
		DrainTime:       "30s",
		KumaCpAddress:   "http://kuma-control-plane.kuma-system:5681",
		IngressPortType: "LoadBalancer",
	}
	useNodePort := false
	cmd := &cobra.Command{
		Use:   "ingress",
		Short: "Install Ingress on Kubernetes",
		Long:  `Install Ingress on Kubernetes in a 'kuma-system' namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if useNodePort {
				args.IngressPortType = "NodePort"
			}
			templateFiles, err := data.ReadFiles(ingress.Templates)
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
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install Ingress to")
	cmd.Flags().StringVar(&args.ImagePullPolicy, "image-pull-policy", args.ImagePullPolicy, "image pull policy for Ingress")
	cmd.Flags().StringVar(&args.Version, "version", args.Version, "version of Ingress component")
	cmd.Flags().StringVar(&args.Image, "image", args.Image, "image of the Ingress component")
	cmd.Flags().StringVar(&args.Mesh, "mesh", args.Mesh, "mesh for Ingress")
	cmd.Flags().StringVar(&args.DrainTime, "drain-time", args.DrainTime, "drain time for Envoy proxy")
	cmd.Flags().StringVar(&args.KumaCpAddress, "kuma-cp-address", args.KumaCpAddress, "the address of Kuma CP")
	cmd.Flags().BoolVar(&useNodePort, "use-node-port", false, "use NodePort instead of LoadBalancer")
	return cmd
}
