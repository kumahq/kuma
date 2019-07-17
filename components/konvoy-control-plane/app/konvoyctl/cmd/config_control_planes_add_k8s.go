package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
)

const (
	DefaultKonvoyNamespace = "konvoy-system"
)

func newConfigControlPlanesAddKubernetesCmd(pctx *configControlPlanesAddContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Add a Control Plane installed on Kubernetes",
		Long:  `Add a Control Plane installed on Kubernetes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := konvoyctl_k8s.DetectKubeConfig()
			if err != nil {
				return errors.Wrapf(err, "Failed to detect current `kubectl` context")
			}
			if config.CurrentContext() == "" {
				return errors.Errorf("Failed to detect current `kubectl` context. It seems that you haven't used `kubectl` on this machine yet.")
			}
			client, err := config.NewClient()
			if err != nil {
				return errors.Wrapf(err, "Failed to connect to a target Kubernetes cluster (`kubectl` context %q)", config.CurrentContext())
			}
			installed, err := client.HasNamespace(DefaultKonvoyNamespace)
			if err != nil {
				return errors.Wrapf(err, "Failed to determine whether a target Kubernetes cluster (`kubectl` context %q) has %q namespace", config.CurrentContext(), DefaultKonvoyNamespace)
			}
			if !installed {
				return errors.Errorf("There is no Control Plane installed on a target Kubernetes cluster (`kubectl` context %q)", config.CurrentContext())
			}

			name := pctx.args.name
			if name == "" {
				name = config.CurrentContext()
			}

			return pctx.AddControlPlane(&config_proto.ControlPlane{
				Name: name,
				Coordinates: &config_proto.ControlPlaneCoordinates{
					Type: &config_proto.ControlPlaneCoordinates_Kubernetes_{
						Kubernetes: &config_proto.ControlPlaneCoordinates_Kubernetes{
							Kubeconfig: config.Filename(),
							Context:    config.CurrentContext(),
							Namespace:  DefaultKonvoyNamespace,
						},
					},
				},
			})
		},
	}
	return cmd
}
